package main

import (
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

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedChittyChatServer        // You need this line if you have a server
	port                               string // Not required but useful if your server needs to know what port it's listening to

	lamport int64      // logical time
	mutex   sync.Mutex // used to lock the server to avoid race conditions.
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
		log.Printf("Server: Failed to listen on port %s: %v", *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	// makes a new server instance using the name and port from the flags.
	server := &Server{
		port:    *port,
		lamport: 0, // gives default value, but not sure if it is necessary
	}

	gRPC.RegisterChittyChatServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server: Listening at %v\n", list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

func (s *Server) Publish(msgStream gRPC.ChittyChat_PublishServer) error {
	for {
		// get the next message from the stream
		msg, err := msgStream.Recv()

		// the stream is closed so we can exit the loop
		if err == io.EOF {
			break
		}
		// some other error
		if err != nil {
			return err
		}
		// log the message
		log.Printf("Received message from %v: %s", msg.Token, msg.Message)

		s.mutex.Lock()
		rsp := &gRPC.MessageRecv{Id: 1, Message: msg.Message, Lamport: s.lamport}
		msgStream.Send(rsp)
		s.lamport++
		s.mutex.Unlock()
	}

	// be a nice server and say goodbye to the client :)
	//ack := &gRPC.Farewell{Message: "Goodbye"}
	//msgStream.SendAndClose(ack)

	return nil
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
