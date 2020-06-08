package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	. "github.com/daniel840829/eventcore"
)

var amqpURI = "amqp://guest:guest@localhost:5672/"

// You can customize your event
type EventTest struct {
	EventBase           // embedding EventBase to have basic function
	CostumeField string // add whatever fields you need to add
}

func NewEventTest() *EventTest {
	e := &EventTest{}
	// binding event in EventBase to let EventBase serialize your custome event
	e.EventBase.Event = e
	return e
}

// you must implement your ParentUnserializer to let EventCenter unserialize your custome event
func (e *EventTest) ParentUnserializer() EventUnserializer {
	return func(data []byte) (Event, error) {
		e := &EventTest{}

		// You can use your way to umarshal. It's not necessary to be json.
		// But if you want to use other way, you need to override Serialize() method to package it.
		if err := json.Unmarshal(data, e); err != nil {
			return nil, err
		}
		// rebind event
		e.EventBase.Event = e
		return e, nil
	}
}

func main() {
	Debug = true

	// You can customize your backend which implements Backend interface.
	hub1 := NewEventCenterCluster(NewAmqpBackend(amqpURI))
	hub2 := NewEventCenterCluster(NewAmqpBackend(amqpURI))

	// Register Event type
	RegisterEvent(NewEventTest())

	// Subscribe event in event center with hanlder
	hub1.Subscribe(NewEventTest().GetType(), func(e Event) error {
		Logger.Log("hub1", e)
		return nil
	}, "hub1")

	// hub1 and hub2 can both receive same event
	hub2.Subscribe(NewEventTest().GetType(), func(e Event) error {
		Logger.Log("hub2", e)
		return nil
	}, "hub2")

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
			e := NewEventTest()
			e.Message = "Hi"
			e.CostumeField = fmt.Sprintf("no:%d", i)
			// Emit the event
			hub2.Emit(e)
			time.Sleep(1 * time.Second)
		}
	}
}
