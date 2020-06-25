package eventcore

import (
	"context"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	log "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
)

type EventCenterCluster struct {
	uuid uuid.UUID
	EventCenter
	Backend
	receivers *EventReceivers
	server    *EventCenterWebsocketServer
}

type Receiver <-chan *message.Message
type EventReceivers struct {
	sync.Map
}

func (rs *EventReceivers) Loop(f func(e Event) error) error {
	var err error
	rs.Range(func(key, value interface{}) bool {
		if r, ok := value.(Receiver); ok {
			select {
			case msg, ok := <-r:
				if !ok {
					rs.Delete(key)
					level.Info(Logger).Log("message", "topic reciever close")
					return true
				}
				level.Debug(Logger).Log("received message", msg.UUID, "payload", string(msg.Payload))
				var e Event = &EventBase{}
				e, err = e.Unserialize(msg.Payload)
				if err != nil {
					level.Error(Logger).Log("error", err)
					return true
				}
				level.Debug(Logger).Log("received message", msg.UUID, "event", e)
				if err := f(e); err != nil {
					level.Error(Logger).Log("error", err)
					return true
				}
				msg.Ack()
			default:
			}
		}
		return true
	})
	return err
}

func (h *EventCenterCluster) UUID() uuid.UUID {
	return h.uuid
}

func NewEventCenterCluster(backend Backend, workers int, uuid uuid.UUID) *EventCenterCluster {
	h := &EventCenterCluster{
		uuid:      uuid,
		receivers: &EventReceivers{},
		Backend:   backend,
	}
	h.init(backend)
	for i := 0; i < workers; i++ {
		go h.Run()
	}
	return h
}

func (h *EventCenterCluster) init(backend Backend) {
	h.Backend.SetConsumerGroupName(h.uuid.String())
}

func (h *EventCenterCluster) Run() {
	logger := log.With(Logger, "event_center_id", h.uuid)
	for {
		h.receivers.Loop(func(e Event) error {
			h.Dispatch(e)
			level.Info(logger).Log("event", e)
			return nil
		})
	}
}

func (h *EventCenterCluster) Subscribe(eventType EventType, handler EventHandler, handlerName string) error {
	h.EventCenter.Subscribe(eventType, handler, handlerName)
	messages, err := h.Backend.Subscriber().Subscribe(context.Background(), string(eventType))
	if err != nil {
		return err
	}
	h.receivers.Store(eventType, Receiver(messages))
	return nil
}

func (h *EventCenterCluster) Dispatch(event Event) {
	if h.server != nil {
		h.server.Emit(event)
	}
	h.EventCenter.Emit(event)
}

func (h *EventCenterCluster) Emit(event Event) {
	event.AddNode(h.GetInfo())
	event.BindSelf(event)
	payload, err := event.Serialize()
	if err != nil {
		level.Error(Logger).Log(err)
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	h.Publisher().Publish(string(event.GetType()), msg)
}

func (h *EventCenterCluster) Stop() {
	err := h.Backend.Close()
	if err != nil {
		level.Error(Logger).Log(err)
	}
}

func (h *EventCenterCluster) GetInfo() DispatcherInfo {
	return DispatcherInfo{
		Name: "EventCenterCluster",
		ID:   h.uuid.String(),
		Type: "EventCenterCluster",
	}
}
