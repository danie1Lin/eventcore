package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	. "github.com/daniel840829/eventcore"
	"github.com/daniel840829/eventcore/proto"
	"github.com/google/uuid"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
)

func init() {
}

var amqpURI = "amqp://guest:guest@localhost:5672/"

// You can customize your event
type EventTest struct {
	EventBase           // embedding EventBase to have basic function
	CostumeField string // add whatever fields you need to add
}

func NewEventTest() *EventTest {
	e := &EventTest{}
	// binding event in EventBase to let EventBase serialize your custome event
	e.BindSelf(e)
	return e
}

func main() {
	DebugMode()
	// You can customize your backend which implements Backend interface.
	uid1, _ := uuid.Parse("2917bd76-31ec-4546-8465-2e14cf825d65")
	hub1 := NewEventCenterCluster(NewAmqpBackend(amqpURI), 4, uid1)
	uid2, _ := uuid.Parse("a68a9790-7a94-4253-b145-c6e7c2853e5f")
	hub2 := NewEventCenterCluster(NewAmqpBackend(amqpURI), 1, uid2)

	// Register Event type
	eventType := RegisterEvent(NewEventTest())

	// Subscribe event in event center with hanlder
	hub1.Subscribe(eventType, func(e Event) error {
		Logger.Log("hub1", e)
		return nil
	}, "hub1")

	// hub1 and hub2 can both receive same event
	hub2.Subscribe(eventType, func(e Event) error {
		Logger.Log("hub2", e)
		return nil
	}, "hub2")

	grpcServer := grpc.NewServer()
	proto.RegisterEventCenterServer(grpcServer, &EventCenterServer{
		EventCenter: &hub1.EventCenter,
	})
	wrappedGrpc := grpcweb.WrapServer(grpcServer)
	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(req) {
			wrappedGrpc.ServeHTTP(resp, req)
		}
		// Fall back to other servers.
		//		http.DefaultServeMux.ServeHTTP(resp, req)
	})
	http.HandleFunc("/main.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "webclient/main.html")
	})
	http.HandleFunc("/dist/main.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "webclient/dist/main.js")
	})
	go http.ListenAndServe(":8080", nil)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	for i := 0; ; i++ {
		select {
		case <-quit:
			Logger.Log("message", "quit")
			if err := hub1.Close(); err != nil {
				Logger.Log("error", err)
			}
			if err := hub2.Close(); err != nil {
				Logger.Log("error", err)
			}
			os.Exit(0)
		default:
			// e := NewEventTest()
			// e.CostumeField = fmt.Sprintf("no:%d", i)
			// // Emit the event
			// hub2.Emit(e)
			// //time.Sleep(1 * time.Second)
		}
	}
}
