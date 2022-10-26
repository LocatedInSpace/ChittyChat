package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/LocatedInSpace/ChittyChat/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedChittyChatServer        // You need this line if you have a server
	port                               string // Not required but useful if your server needs to know what port it's listening to

	lamport      Lamport                // logical time
	mutex        sync.Mutex             // used to lock the server to avoid race conditions.
	participants map[string]Participant // token -> Participant
	unusedId     int32
}

type Participant struct {
	id   int32
	name string
	// user joined/left *going* to Participant
	statuses chan gRPC.StatusChange
}

var msgChannels []chan<- gRPC.MessageRecv

var msgLock sync.Mutex

type Lamport struct {
	timestamp int64 // logical time
}

func (l *Lamport) Get(time int64) int64 {
	if l.timestamp > time {
		l.timestamp++
		return l.timestamp - 1
	} else {
		l.timestamp = time + 1
		return l.timestamp
	}
}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var port = flag.String("port", "5400", "Server port") // set with "-port <port>" in terminal

func main() {

	// f := setLog() //uncomment this line to log to a log.txt file instead of the console
	// defer f.Close()

	// This parses the flags and sets the correct/given corresponding values.
	flag.Parse()
	fmt.Println(".:server is starting:.")

	// starts a goroutine executing the launchServer method.
	launchServer()
}

func launchServer() {
	log.Printf("Server: Attempts to create listener on port %s\n", *port)

	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server: Failed to listen on port %s: %v\n", *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	l := new(Lamport)
	l.timestamp = 1
	server := &Server{
		port:         *port,
		lamport:      *l,
		participants: make(map[string]Participant),
	}

	gRPC.RegisterChittyChatServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server: Listening at %v\n", list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

func (s *Server) GetUnusedId() int32 {
	/*in practice, this should reuse id's of stale connections*/
	s.unusedId++
	return s.unusedId - 1
}

func (s *Server) Join(in *gRPC.Information, stream gRPC.ChittyChat_JoinServer) error {
	// new participant
	nP := new(Participant)
	nP.name = in.ClientName
	for _, p := range s.participants {
		if p.name == nP.name {
			// name isnt unique, so say bad person, and then terminate stream
			rsp := &gRPC.StatusChange{Joined: false}
			stream.Send(rsp)
			log.Printf("Duplicate client name (%s) tried to join, denied.\n", nP.name)
			// this error (possibly due to datarace) does not get transmitted
			// so it could be blank with no change in functionality
			return errors.New("Duplicate name, please change your client name")
		}
	}

	status := gRPC.StatusChange{Joined: true, ClientName: nP.name, Id: s.GetUnusedId(), Lamport: s.lamport.Get(0)}
	// send StatusChange to all participants
	for _, p := range s.participants {
		// non blocking send
		select {
		case p.statuses <- status:
			log.Printf("Sent StatusChange(joined) of %s to %s\n", nP.name, p.name)
		default:
			log.Printf("Could not send StatusChange(joined) of %s to %s\n", nP.name, p.name)
		}
		//p.statuses <- status

	}
	// send StatusChange with added token to new participant
	token := uuid.New().String()
	log.Printf("Generated token for %s: %s\n", nP.name, token)
	status.Token = token
	stream.Send(&status)
	log.Printf("Sent StatusChange(joined) of %s to %s | Lamport: %v\n", nP.name, nP.name, status.Lamport)

	nP.id = status.Id
	// this gets set by PollMessages
	//nP.messages = make(chan gRPC.MessageRecv)
	nP.statuses = make(chan gRPC.StatusChange)

	s.participants[token] = *nP
	var err error
awaitStream:
	for {
		select {

		case <-stream.Context().Done():
			err = stream.Context().Err()
			// not ideal, since stream.Context().Done() would ideally only fire when err != nil
			// but for some reason it fires constantly, so we have to check
			if err != nil {
				break awaitStream
			}
		case msg := <-nP.statuses:
			err = stream.Send(&msg)
			if err != nil {
				break awaitStream
			}
		}
	}
	log.Println(err)
	status.Lamport = s.lamport.Get(0)
	log.Printf("Participant %s left Chitty-Chat at Lamport time %v\n", nP.name, status.Lamport)
	s.mutex.Lock()
	delete(s.participants, token)
	s.mutex.Unlock()

	for _, p := range s.participants {
		// non blocking send
		status.Token = ""
		status.Joined = false
		select {
		case p.statuses <- status:
			log.Printf("Sent StatusChange(left) of %s to %s\n", nP.name, p.name)
		default:
			log.Printf("Could not send StatusChange(left) of %s to %s\n", nP.name, p.name)
		}
		//p.statuses <- status

	}
	return err
}

func (s *Server) QueryUsername(ctx context.Context, id *gRPC.Id) (*gRPC.NameOfId, error) {
	for _, p := range s.participants {
		if p.id == id.Id {
			return &gRPC.NameOfId{Exists: true, Name: p.name}, nil
		}
	}
	return &gRPC.NameOfId{Exists: false}, nil
}

func (s *Server) PollMessages(stream gRPC.ChittyChat_PublishServer, msgs chan<- gRPC.MessageRecv) {

	msgLock.Lock()
	msgChannels = append(msgChannels, msgs)
	msgLock.Unlock()
	for {
		// get the next message from the stream
		msg, err := stream.Recv()

		// the stream is closed so we can exit the loop
		if err == io.EOF {
			break
		}
		// some other error
		if err != nil {
			break
		}

		if _, ok := s.participants[msg.Token]; ok {
			log.Printf("Received message from %v: %s | Lamport: %v, Corrected-Lamport: %v\n", s.participants[msg.Token].name, msg.Message, msg.Lamport, s.lamport.Get(msg.Lamport)-1)
			s.lamport.timestamp--
			// add this to every participants channel
			msg := gRPC.MessageRecv{Id: s.participants[msg.Token].id, Message: msg.Message, Lamport: s.lamport.Get(msg.Lamport)}
			for _, pMsgs := range msgChannels {
				pMsgs <- msg
			}
		} else {
			log.Printf("Invalid token (%v) given.\n", msg.Token)
			break
			//return errors.New("Invalid token")
		}
	}
	msgLock.Lock()
	index := 0
	for i, m := range msgChannels {
		if m == msgs {
			index = i
		}
	}
	msgChannels = append(msgChannels[:index], msgChannels[index+1:]...)
	msgLock.Unlock()

	log.Printf("Ended PollMessages")
}

func (s *Server) Publish(stream gRPC.ChittyChat_PublishServer) error {
	msgs := make(chan gRPC.MessageRecv)
	go s.PollMessages(stream, msgs)

	var err error
awaitStream:
	for {
		select {

		case <-stream.Context().Done():
			err = stream.Context().Err()
			// not ideal, since stream.Context().Done() would ideally only fire when err != nil
			// but for some reason it fires constantly, so we have to check
			if err != nil {
				break awaitStream
			}
		case msg := <-msgs:
			err = stream.Send(&msg)
			if err != nil {
				break awaitStream
			}
		}
	}

	return err
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v\n", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
