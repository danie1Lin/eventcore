package eventcore

import "sync"

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

func RegisterEvent(e Event) {
	eventUnserailizers.setUnserializer(e.GetType(), e.ParentUnserializer())
}
