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

type EventTest struct {
	EventBase
	CostumeField string
}

func NewEventTest() *EventTest {
	e := &EventTest{}
	// binding event in EventBase
	e.EventBase.Event = e
	return e
}

func (e *EventTest) ParentUnserializer() EventUnserializer {
	return func(data []byte) (Event, error) {
		e := &EventTest{}
		if err := json.Unmarshal(data, e); err != nil {
			return nil, err
		}
		e.EventBase.Event = e
		return e, nil
	}
}

func main() {
	Debug = true
	hub1 := NewEventCenterCluster(NewAmqpBackend(amqpURI))
	hub2 := NewEventCenterCluster(NewAmqpBackend(amqpURI))
	RegisterEvent(NewEventTest())

	hub1.Subscribe(NewEventTest().GetType(), func(e Event) error {
		Logger.Log("hub1", e)
		return nil
	}, "hub1")

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
			hub2.Emit(e)
			time.Sleep(1 * time.Second)
		}
	}
}
