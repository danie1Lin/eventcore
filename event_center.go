package eventcore

import (
	"fmt"

	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
)

// EventCenter 單一process的EventHub
type EventCenter struct {
	uuid          string
	eventHandlers map[EventType]map[string]EventHandler
}

func (h *EventCenter) GetInfo() DispatcherInfo {
	if h.uuid == "" {
		h.uuid = uuid.New().String()
	}
	return DispatcherInfo{
		Name: "EventCenter",
		ID:   h.uuid,
		Type: "EventCenter",
	}
}

func (h *EventCenter) Subscribe(eventType EventType, handler EventHandler, handlerName string) {
	if h.eventHandlers == nil {
		h.eventHandlers = make(map[EventType]map[string]EventHandler)
	}
	if orgHandlers, ok := h.eventHandlers[eventType]; !ok {
		h.eventHandlers[eventType] = map[string]EventHandler{handlerName: handler}
	} else {
		if _, ok := orgHandlers[handlerName]; ok {
			level.Warn(Logger).Log("message", fmt.Sprintf("handler %s", handlerName))
		} else {
			h.eventHandlers[eventType][handlerName] = handler
		}
	}
}

func (h *EventCenter) Emit(event Event) error {
	event.AddNode(h.GetInfo())
	if handlers, ok := h.eventHandlers[event.GetType()]; ok {
		for _, handler := range handlers {
			if err := handler(event); err != nil {
				level.Error(Logger).Log("error", err)
				return err
			}
		}
	}
	return nil
}
