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
	"strings"
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

func (l *Lamport) Get() int64 {
	l.timestamp++
	return l.timestamp
}

func (l *Lamport) Correct(time int64, info string) string {
	if time-l.timestamp == 1 {
		l.timestamp++
		return fmt.Sprintf("Lamport: %v", l.timestamp)
	} else if l.timestamp < time {
		log.Printf("%s | Correcting own Lamport (%v), instead using supplied Lamport-time %v\n", info, l.timestamp, time)
		l.timestamp = time
		return fmt.Sprintf("Corrected Lamport: %v", l.timestamp)
	} else if time < l.timestamp {
		l.timestamp++
		log.Printf("%s | Received Lamport-time %v, instead using own Lamport++ (%v)\n", info, time, l.timestamp)
		return fmt.Sprintf("Received Lamport: %v, Corrected: %v", time, l.timestamp)
	}
	return ""
}

func Validated(s string, max int) string {
	r := strings.ReplaceAll(s, "\n", "")
	r = strings.ReplaceAll(r, "\r", "")
	r = strings.ReplaceAll(r, "\t", "")
	// no longer than 128
	end := len(r)
	if end > max {
		end = max
	}
	r = r[:end]
	return r
}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var port = flag.String("port", "5400", "Server port") // set with "-port <port>" in terminal

func main() {

	f := setLog()
	defer f.Close()

	flag.Parse()

	log.Printf("Creating listener on port %s\n", *port)

	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Failed to open listener on port %s: %v\n", *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	grpcServer := grpc.NewServer()

	l := new(Lamport)
	l.timestamp = 0
	server := &Server{
		port:         *port,
		lamport:      *l,
		participants: make(map[string]Participant),
	}

	gRPC.RegisterChittyChatServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Listening at %v\n", list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("Failed to serve %v", err)
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
	nP.name = Validated(in.ClientName, 20)
	for _, p := range s.participants {
		if p.name == nP.name {
			// name isnt unique, so say bad person, and then terminate stream
			rsp := &gRPC.StatusChange{Joined: false}
			stream.Send(rsp)
			log.Printf("Join() | Duplicate client name (%s) tried to join, denied.\n", nP.name)
			// this error (possibly due to datarace) does not get transmitted
			// so it could be blank with no change in functionality
			return errors.New("Duplicate name, please change your client name")
		}
	}
	log.Printf("Join() | Join request received from %s | Lamport: %v\n", in.ClientName, in.Lamport)
	s.lamport.Correct(in.Lamport, "Join()")
	status := gRPC.StatusChange{Joined: true, ClientName: nP.name, Id: s.GetUnusedId(), Lamport: s.lamport.Get()}
	// send StatusChange to all participants
	for _, p := range s.participants {
		// non blocking send
		select {
		case p.statuses <- status:
			log.Printf("Join() | Sent StatusChange(joined) of %s to %s | Lamport: %v\n", nP.name, p.name, status.Lamport)
		default:
			log.Printf("Join() | Could not send StatusChange(joined) of %s to %s\n", nP.name, p.name)
		}
		//p.statuses <- status

	}
	// send StatusChange with added token to new participant
	token := uuid.New().String()
	log.Printf("Join() | Generated token for %s: %s\n", nP.name, token)
	status.Token = token
	stream.Send(&status)
	log.Printf("Join() | Sent StatusChange(joined) of %s to %s | Lamport: %v\n", nP.name, nP.name, status.Lamport)

	nP.id = status.Id
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
	if !strings.Contains(err.Error(), "context canceled") {
		log.Printf("Join() | %s\n", err)
	}
	status.Lamport = s.lamport.Get()
	log.Printf("Join() | Participant %s left Chitty-Chat at Lamport time %v\n", nP.name, status.Lamport)
	s.mutex.Lock()
	delete(s.participants, token)
	s.mutex.Unlock()

	for _, p := range s.participants {
		// non blocking send
		status.Token = ""
		status.Joined = false
		select {
		case p.statuses <- status:
			log.Printf("Join() | Sent StatusChange(left) of %s to %s | Lamport: %v\n", nP.name, p.name, status.Lamport)
		default:
			log.Printf("Join() | Could not send StatusChange(left) of %s to %s\n", nP.name, p.name)
		}
		//p.statuses <- status

	}
	return err
}

func (s *Server) QueryUsername(ctx context.Context, id *gRPC.Id) (*gRPC.NameOfId, error) {
	for _, p := range s.participants {
		if p.id == id.Id {
			log.Printf("QueryUsername() | Received request about existing id %v: %s\n", id.Id, p.name)
			return &gRPC.NameOfId{Exists: true, Name: p.name}, nil
		}
	}
	log.Printf("QueryUsername() | Received request about unknown id %v\n", id.Id)
	return &gRPC.NameOfId{Exists: false}, nil
}

func (s *Server) PollMessages(stream gRPC.ChittyChat_PublishServer, msgs chan<- gRPC.MessageRecv) {

	msgLock.Lock()
	// append received channel of messages to msgChannels
	// every participant has one such channel, and it is how each thread
	// sends messages back to main thread (and other threads)
	msgChannels = append(msgChannels, msgs)
	msgLock.Unlock()
	// variable for storing the name of the client owning the thread
	// this is needed for logging which thread exited, since participants & msgChannels do NOT know which is which
	lastName := ""
	for {
		// get the next message from the stream
		msg, err := stream.Recv()
		if err != nil {
			break
		}
		// technically we check this in client.go, but lets check on server
		// just in case there's a malicious client
		msg.Message = Validated(msg.Message, 128)

		if _, ok := s.participants[msg.Token]; ok {
			lastName = s.participants[msg.Token].name
			log.Printf("Publish() | Received '%s' from %v | %s\n", msg.Message, lastName, s.lamport.Correct(msg.Lamport, "Publish()"))
			// add this to every participants channel
			msg := gRPC.MessageRecv{Id: s.participants[msg.Token].id, Message: msg.Message, Lamport: s.lamport.Get()}
			for _, pMsgs := range msgChannels {
				pMsgs <- msg
			}
		} else {
			log.Printf("Publish() | Invalid token (%v) given.\n", msg.Token)
			break
		}
	}
	msgLock.Lock()
	index := 0
	for i, m := range msgChannels {
		if m == msgs {
			index = i
		}
	}
	// remove self from list of channels
	msgChannels = append(msgChannels[:index], msgChannels[index+1:]...)
	msgLock.Unlock()

	//log.Printf("Ended PollMessages for %s", lastName)
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
	if err := os.Truncate("server-log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v\n", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("server-log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	// print to both file and console
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	return f
}
