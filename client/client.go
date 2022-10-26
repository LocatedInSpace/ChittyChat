package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	gRPC "github.com/LocatedInSpace/ChittyChat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.ChittyChatClient //the server
var ServerConn *grpc.ClientConn  //the server connection

// id to name
var participants map[int32]string
var token string
var displayName string
var updated chan bool

type Lamport struct {
	timestamp int64 // logical time
}

var l Lamport = Lamport{timestamp: 0}

func (l *Lamport) Get() int64 {
	l.timestamp++
	return l.timestamp - 1
}

func (l *Lamport) Correct(time int64) {
	if l.timestamp < time {
		l.timestamp = time
		l.timestamp++
	}
}

func main() {
	//log to file instead of console
	//f := setLog()
	//defer f.Close()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	flag.Parse()
	participants = make(map[int32]string)
	updated = make(chan bool)

	li := widgets.NewList()
	li.Title = "Chat (Disconnected)"
	/*li.Rows = []string{
		"[1] [dumbdwa dwa jdoawd uwahd uiawhdiuawh duiawhduiawhuidhwaiu](fg:blue)",
		"[2] [library](fg:red)",
		"[3] [color](fg:white,bg:green) output",
		"[4] output.go",
		"[5] random_out.go",
		"[6] dashboard.go",
		"[7] foo",
		"[8] bar",
		"[9] baz",
	}*/
	//li.TextStyle = ui.NewStyle(ui.ColorWhite)
	li.WrapText = true
	li.SelectedRowStyle = ui.NewStyle(ui.ColorClear, ui.ColorBlue)
	//li.SetRect(0, 0, 200, 200)

	input := widgets.NewParagraph()
	input.Title = "Enter your displayname"
	input.Text = ""
	input.WrapText = true

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(1, 1, termWidth-1, termHeight-1)

	grid.Set(
		ui.NewRow(0.8,
			ui.NewCol(1.0, li),
		),
		ui.NewRow(0.2,
			ui.NewCol(1.0, input),
		),
	)

	ui.Render(grid)
	//termWidth, termHeight = ui.TerminalDimensions()
	//grid.SetRect(1, 1, termWidth-1, termHeight-1)
	//ui.Clear()
	//ui.Render(grid)

	msgs := make(chan string)

	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<C-c>":
				return
			case "<C-<Backspace>>":
				input.Text = input.Text[:len(input.Text)-1]
			case "<Space>":
				input.Text += " "
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(1, 1, payload.Width-1, payload.Height-1)
			case "<Down>", "<MouseWheelDown>":
				li.ScrollAmount(1)
			case "<Up>", "<MouseWheelUp>":
				li.ScrollAmount(-1)
			case "<Enter>":
				if len(input.Text) > 0 {
					// not authenticated
					if token == "" {
						displayName = strings.TrimSpace(input.Text)
						token = TryAuthenticate(li, msgs)
						if token != "" {
							li.Title = "Chat (Connected)"
							input.Title = "Hello, " + displayName
						}
					} else {
						msgs <- strings.TrimSpace(input.Text)
					}
					input.Text = ""
				}
				li.ScrollBottom()
			default:
				switch e.Type {
				case ui.KeyboardEvent: // handle all key presses
					if len(e.ID) <= 2 {
						input.Text += e.ID
					}
					//log.Print(e.ID) // keypress string
				}
			}
		case <-updated:
		}

		ui.Clear()
		ui.Render(grid)
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
		msg, _ := server.QueryUsername(context.Background(), &gRPC.Id{Id: i})
		if msg.Exists {
			participants[i] = msg.Name
			return msg.Name
		}
		return "<UNKNOWN>"
	}
}

func LogUserUpdate(s *gRPC.StatusChange, li *widgets.List) {
	l.Correct(s.Lamport)
	if s.Joined {
		participants[s.Id] = s.ClientName
		log.Printf("Participant %s joined Chitty-Chat at Lamport time %v\n", s.ClientName, s.Lamport)
		li.Rows = append(li.Rows, fmt.Sprintf("[Participant %s joined Chitty-Chat at Lamport time %v](fg:green)", s.ClientName, s.Lamport))
		//li.Rows = append(li.Rows, fmt.Sprintf("[Lamport: %v](fg:black,bg:white) %s: %s", err))
	} else {
		if _, ok := participants[s.Id]; ok {
			delete(participants, s.Id)
		}
		log.Printf("Participant %s left Chitty-Chat at Lamport time %v\n", s.ClientName, s.Lamport)
		li.Rows = append(li.Rows, fmt.Sprintf("[Participant %s left Chitty-Chat at Lamport time %v](fg:yellow)", s.ClientName, s.Lamport))
	}
	// tells ui to update, if it can
	select {
	case updated <- true:
	default:
	}
}

func LogMessage(li *widgets.List, msg *gRPC.MessageRecv) {
	log.Printf("%s: %s | Lamport: %v", GetName(msg.Id), msg.Message, msg.Lamport)
	li.Rows = append(li.Rows, fmt.Sprintf("[Lamport: %v](fg:black,bg:white) %s: %s", msg.Lamport, GetName(msg.Id), msg.Message))
	// tells ui to update, if it can
	select {
	case updated <- true:
	default:
	}
}

func LogError(li *widgets.List, err error) {
	log.Println(err)
	li.Rows = append(li.Rows, fmt.Sprintf("[%s](fg:red)", err))
	// tells ui to update, if it can
	select {
	case updated <- true:
	default:
	}
}

func TryAuthenticate(li *widgets.List, msgs <-chan string) string {

	ConnectToServer()
	//defer ServerConn.Close()

	// status stream
	sStream, err := server.Join(context.Background(), &gRPC.Information{ClientName: displayName})
	if err != nil {
		LogError(li, err)
		return ""
	}
	status, err := sStream.Recv()
	if err != nil || !status.Joined {
		LogError(li, errors.New("Duplicate name, please change your client name"))
		return ""
	}
	LogUserUpdate(status, li)
	go UserUpdates(sStream, li)

	// message stream
	mStream, err := server.Publish(context.Background())
	if err != nil {
		LogError(li, err)
		return ""
	}

	//Infinite loop to listen for clients input.
	go RecvMessages(mStream, li)
	go SendMessages(mStream, msgs, li)
	return status.Token
}

func UserUpdates(s gRPC.ChittyChat_JoinClient, li *widgets.List) {
	for {
		status, err := s.Recv()
		if err != nil {
			LogError(li, err)
			break
		}
		LogUserUpdate(status, li)
	}
	token = ""
	LogError(li, errors.New("Ended UserUpdates"))
}

func RecvMessages(mStream gRPC.ChittyChat_PublishClient, li *widgets.List) {
	for {
		msg, err := mStream.Recv()
		// the stream is closed so we can exit the loop
		if err == io.EOF {
			break
		}
		// some other error
		if err != nil {
			break
		}
		LogMessage(li, msg)
		l.Correct(msg.Lamport)
	}
	token = ""
	LogError(li, errors.New("Ended RecvMessages"))
}

func SendMessages(mStream gRPC.ChittyChat_PublishClient, msgs <-chan string, li *widgets.List) {
	for {
		// receive input text from UI loop
		msg := <-msgs

		if !conReady(server) {
			LogError(li, errors.New("Connection to server is faulty :("))
			continue
		}

		err := mStream.Send(&gRPC.MessageSent{Token: token, Message: msg, Lamport: l.Get()})
		if err != nil {
			LogError(li, err)
			break
		}
	}
	token = ""
	LogError(li, errors.New("Ended SendMessages"))
}

// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", "dummy", *serverPort)
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
