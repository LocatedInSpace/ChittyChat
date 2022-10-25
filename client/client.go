package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	gRPC "github.com/LocatedInSpace/ChittyChat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.ChittyChatClient //the server
var ServerConn *grpc.ClientConn  //the server connection

var token string

// id to name
var participants map[int32]string

type Lamport struct {
	timestamp int64 // logical time
}

var l Lamport = Lamport{timestamp: 0}

func (l *Lamport) Get(time int64) int64 {
	if l.timestamp > time {
		l.timestamp++
		return l.timestamp - 1
	} else {
		l.timestamp = time + 1
		return l.timestamp
	}
}

// looks up name in participants
// if id does not have an associated name, then
// query server for name
func GetName(i int32) string {
	if _, ok := participants[i]; ok {
		return participants[i]
	} else {
		// meant to query server, for now, just like.. return default
		return "NO NAME"
	}
}

func LogUserUpdate(s *gRPC.StatusChange) {
	if s.Joined {
		participants[s.Id] = s.ClientName
		log.Printf("Participant %s joined Chitty-Chat at Lamport time %v\n", s.ClientName, s.Lamport)
	} else {
		if _, ok := participants[s.Id]; ok {
			delete(participants, s.Id)
		}
		log.Printf("Participant %s left Chitty-Chat at Lamport time %v\n", s.ClientName, s.Lamport)
	}
}

func UserUpdates(s gRPC.ChittyChat_JoinClient) {
	for {
		status, err := s.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		LogUserUpdate(status)
	}
}

func main() {
	//parse flag/arguments
	flag.Parse()

	participants = make(map[int32]string)

	fmt.Println("--- CLIENT APP ---")

	//log to file instead of console
	//f := setLog()
	//defer f.Close()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer()
	defer ServerConn.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Message to send to server")
	fmt.Println("--------------------")

	// status stream
	sStream, err := server.Join(context.Background(), &gRPC.Information{ClientName: *clientsName})
	if err != nil {
		log.Println(err)
		fmt.Println(err)
		return
	}
	status, err := sStream.Recv()
	if err != nil || !status.Joined {
		log.Println("Duplicate name, please change your client name")
		return
	}
	token = status.Token
	LogUserUpdate(status)
	go UserUpdates(sStream)

	// message stream
	mStream, err := server.Publish(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	var lastTimestamp int64 = 0
	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim input

		if !conReady(server) {
			log.Printf("Client %s: something was wrong with the connection to the server :(", *clientsName)
			continue
		}

		mStream.Send(&gRPC.MessageSent{Token: token, Message: input, Lamport: l.Get(lastTimestamp)})
		msg, err := mStream.Recv()
		log.Printf("Received message from %s: %s | Lamport: %v", GetName(msg.Id), msg.Message, msg.Lamport)
		lastTimestamp = msg.Lamport

		//Convert string to int64, return error if the int is larger than 32bit or not a number
		/*val, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			if input == "hi" {
				sayHi()
			}
			continue
		}
		incrementVal(val)*/
	}
}

// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	server = gRPC.NewChittyChatClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(s gRPC.ChittyChatClient) bool {
	return ServerConn.GetState().String() == "READY"
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
