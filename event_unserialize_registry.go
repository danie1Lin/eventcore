package eventcore

import (
	"encoding/json"
	"reflect"
	"sync"
)

func makeJsonUnserializer(e Event) EventUnserializer {
	_type := reflect.Indirect(reflect.ValueOf(e)).Type()
	return func(data []byte) (Event, error) {
		e := reflect.New(_type).Interface().(Event)
		if err := json.Unmarshal(data, e); err != nil {
			return nil, err
		}
		return e, nil
	}

}

type eventUnserailizeRegistry struct {
	sync.Map
}

var eventUnserailizers *eventUnserailizeRegistry

func (s *eventUnserailizeRegistry) getUnserializer(eventType EventType) EventUnserializer {
	if v, ok := s.Load(eventType); ok {
		if h, ok := v.(EventUnserializer); ok {
			return h
		}
	}
	return nil
}

func (s *eventUnserailizeRegistry) setUnserializer(eventType EventType, unserializer EventUnserializer) {
	s.Store(eventType, unserializer)
}

func init() {
	eventUnserailizers = &eventUnserailizeRegistry{}
}

// RegisterEvent assign event with a eventType
func RegisterEvent(e Event, unserializer ...EventUnserializer) EventType {
	e.BindSelf(e)
	if len(unserializer) > 0 {
		eventUnserailizers.setUnserializer(e.GetType(), unserializer[0])
		return e.GetType()
	}
	eventUnserailizers.setUnserializer(e.GetType(), makeJsonUnserializer(e))
	return e.GetType()
}
