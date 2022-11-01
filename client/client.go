package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	gRPC "github.com/LocatedInSpace/ChittyChat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"github.com/LocatedInSpace/ChittyChat/chatlist"
	"golang.design/x/clipboard"
)

var serverPort = flag.String("server", "5400", "Tcp server")
var logName = flag.String("logname", "log", "Name of -log.txt file output")

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
	return l.timestamp
}

func (l *Lamport) Correct(time int64) {
	if time-l.timestamp == 1 {
		l.timestamp++
	} else if l.timestamp < time {
		log.Printf("Correcting own Lamport (%v), instead using supplied Lamport-time %v\n", l.timestamp, time)
		l.timestamp = time
	} else if time < l.timestamp {
		log.Printf("Received Lamport-time %v, instead using own Lamport++ (%v)\n", time, l.timestamp+1)
		l.timestamp++
	}
}

func Validated(s string, max int) string {
	r := strings.ReplaceAll(s, "\n", "")
	r = strings.ReplaceAll(r, "\r", "")
	r = strings.ReplaceAll(r, "\t", "")
	runes := []rune(r)
	end := len(runes)
	if end > max {
		end = max
	}
	return string(runes[:end])
}

func main() {
	flag.Parse()

	//log to file instead of console
	f := setLog()
	defer f.Close()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// Init returns an error if the package is not ready for use.
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	participants = make(map[int32]string)
	updated = make(chan bool)

	li := chatlist.NewList()
	li.Rows = make([]string, 0)
	li.Title = "Chat (Disconnected)"
	li.WrapText = true

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

	msgs := make(chan string, 10)

	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<C-c>":
				return
			case "<C-<Backspace>>":
				runes := []rune(input.Text)
				end := len(runes)
				if end > 0 {
					input.Text = string(runes[:end-1])
				}
			case "<Space>":
				input.Text += " "
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(1, 1, payload.Width-1, payload.Height-1)
			case "<Down>", "<MouseWheelDown>":
				if len(li.Rows) == 0 {
					continue
				}
				li.ScrollDown()
			case "<Up>", "<MouseWheelUp>":
				if len(li.Rows) == 0 {
					continue
				}
				li.ScrollUp()
			case "<Enter>":
				if len(input.Text) > 0 {
					// not authenticated
					if token == "" {
						li.Rows = append(li.Rows, "[Attempting to authenticate with server...](bg:yellow,fg:black)")
						li.ScrollBottom()
						ui.Render(grid)
						displayName = strings.TrimSpace(input.Text)
						token = TryAuthenticate(li, msgs)
					} else {
						msgs <- strings.TrimSpace(input.Text)
					}
					input.Text = ""
				}
			default:
				switch e.Type {
				case ui.KeyboardEvent: // handle all key presses
					if e.ID == "<C-v>" {
						input.Text += string(clipboard.Read(clipboard.FmtText))
					} else if len(e.ID) <= 4 {
						input.Text += e.ID
					}
				}
			}
		case <-updated:
		}
		// removes newlines, checks length, etc.
		if token == "" {
			input.Text = Validated(input.Text, 20)
		} else {
			input.Text = Validated(input.Text, 128)
		}

		ui.Clear()
		if token == "" {
			li.Title = "Chat (Disconnected)"
			input.Title = "Enter your displayname"
		} else {
			li.Title = "Chat (Connected)"
			input.Title = "Hello, " + displayName
		}
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
		log.Printf("Calling QueryUsername with id: %v", i)
		msg, _ := server.QueryUsername(context.Background(), &gRPC.Id{Id: i})
		if msg.Exists {
			participants[i] = msg.Name
			return msg.Name
		}
		return "<UNKNOWN>"
	}
}

func LogUserUpdate(s *gRPC.StatusChange, li *chatlist.List) {
	if s.Joined {
		participants[s.Id] = s.ClientName
		log.Printf("Participant %s joined Chitty-Chat at Lamport time %v\n", s.ClientName, s.Lamport)
		li.Rows = append(li.Rows, fmt.Sprintf("[Participant %s joined Chitty-Chat at Lamport time %v](fg:green)", s.ClientName, s.Lamport))
		//li.Rows = append(li.Rows, fmt.Sprintf("[Lamport: %v](fg:black,bg:white) %s: %s", err))
	} else {
		// we dont need to check if s.Id exists, since it will no-op if not valid
		delete(participants, s.Id)
		log.Printf("Participant %s left Chitty-Chat at Lamport time %v\n", s.ClientName, s.Lamport)
		li.Rows = append(li.Rows, fmt.Sprintf("[Participant %s left Chitty-Chat at Lamport time %v](fg:yellow)", s.ClientName, s.Lamport))
	}
	li.ScrollBottom()
	// ui update event
	updated <- true
}

func LogMessage(li *chatlist.List, msg *gRPC.MessageRecv) {
	log.Printf("%s: %s | Lamport: %v", GetName(msg.Id), msg.Message, msg.Lamport)
	li.Rows = append(li.Rows, fmt.Sprintf("[Lamport: %v](fg:black,bg:white) [%s](fg:cyan): %s", msg.Lamport, GetName(msg.Id), msg.Message))
	li.ScrollBottom()
	// ui update event
	updated <- true
}

func LogError(li *chatlist.List, err error) {
	log.Println(err)
	li.Rows = append(li.Rows, fmt.Sprintf("[%s](fg:red)", err))
	li.ScrollBottom()
	// ui update event
	updated <- true
}

func TryAuthenticate(li *chatlist.List, msgs <-chan string) string {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(5*time.Second), grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("%s: Attempts to dial on port %s\n", displayName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			LogError(li, errors.New("Server could not be reached"))
		} else {
			LogError(li, err)
		}
		return ""
	}

	server = gRPC.NewChittyChatClient(conn)
	ServerConn = conn

	lam := l.Get()
	log.Printf("Sending join message to server | Lamport:  %v\n", lam)
	// status stream
	sStream, err := server.Join(context.Background(), &gRPC.Information{ClientName: displayName, Lamport: lam})
	if err != nil {
		LogError(li, err)
		return ""
	}

	status, err := sStream.Recv()
	if err != nil || !status.Joined {
		LogError(li, errors.New("Duplicate name, please change your client name"))
		return ""
	}

	log.Printf("Server accepted our name %s, our token is: %s", status.ClientName, status.Token)

	l.Correct(status.Lamport)
	// the different Logs are meant to be ran in a goroutine
	// since they use a channel for ui-update, if not goroutine it will cause a deadlock waiting
	go LogUserUpdate(status, li)

	// message stream
	mStream, err := server.Publish(context.Background())
	if err != nil {
		LogError(li, err)
		return ""
	}

	go HandleStreams(sStream, mStream, msgs, li)

	return status.Token
}

func HandleStreams(sStream gRPC.ChittyChat_JoinClient, mStream gRPC.ChittyChat_PublishClient, msgs <-chan string, li *chatlist.List) {
	errs := make(chan error)
	done := make(chan bool)
	go UserUpdates(sStream, li, errs)
	go RecvMessages(mStream, li, errs)
	go SendMessages(mStream, msgs, li, errs, done)

	// we deliberately only wait on one error, the other two (if not SendMessages) will fail by themselves
	err := <-errs
	// tell SendMessages to error
	done <- true
	LogError(li, err)
	token = ""
	ServerConn.Close()
}

func UserUpdates(s gRPC.ChittyChat_JoinClient, li *chatlist.List, errs chan<- error) {
	for {
		status, err := s.Recv()
		if err != nil {
			if strings.Contains(err.Error(), "forcibly closed by the remote host") {
				err = errors.New("Server has went offline")
			}
			// errs will only wait *once*
			// so, it should see if it's waiting on us
			// if not, then RecvMessage or SendMessage failed first, and that error is displayed
			select {
			case errs <- err:
			default:
			}
			break
		}
		l.Correct(status.Lamport)
		LogUserUpdate(status, li)
	}
}

func RecvMessages(mStream gRPC.ChittyChat_PublishClient, li *chatlist.List, errs chan<- error) {
	for {
		msg, err := mStream.Recv()
		if err != nil {
			if strings.Contains(err.Error(), "forcibly closed by the remote host") {
				err = errors.New("Server has went offline")
			}
			// see comment for UserUpdates
			select {
			case errs <- err:
			default:
			}
			break
		}
		LogMessage(li, msg)
		l.Correct(msg.Lamport)
	}
}

func SendMessages(mStream gRPC.ChittyChat_PublishClient, msgs <-chan string, li *chatlist.List, errs chan<- error, exit <-chan bool) {
sendloop:
	for {
		// since SendMessages is not like UserUpdates/RecvMessages in that it receives from server
		// it needs an "updater" akin to our exit-channel, so that it does not forever wait on <-msgs update
		select {
		case msg := <-msgs:
			lam := l.Get()
			log.Printf("Sending '%s' to server with Lamport time: %v", msg, lam)
			err := mStream.Send(&gRPC.MessageSent{Token: token, Message: msg, Lamport: lam})
			if err != nil {
				// see comment for UserUpdates
				select {
				case errs <- err:
				default:
				}
				break sendloop
			}
		case <-exit:
			break sendloop
		}
	}
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	f, err := os.OpenFile(*logName+"-log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
