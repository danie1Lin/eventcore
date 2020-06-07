package eventcore

import (
	"encoding/json"
	"reflect"

	"errors"
	//log "github.com/sirupsen/logrus"
)

type EventBase struct {
	Message      string
	Type         EventType
	Emitter      string
	EventContent []byte
	Event        Event `json:"-"`
}

func (b *EventBase) Base() *EventBase {
	return b
}

func (b *EventBase) SetBase(base *EventBase) {
}

func (b *EventBase) GetType() EventType {
	if b.Event == nil {
		return "EventBase"
	}
	return EventType(reflect.Indirect(reflect.ValueOf(b.Event)).Type().String())
}

func (b *EventBase) Unserialize(data []byte) (Event, error) {
	if err := json.Unmarshal(data, b); err != nil {
		return nil, err
	}
	if b.EventContent != nil {
		if f := eventUnserailizers.getUnserializer(EventType(b.Type)); f != nil {
			if e, err := f(b.EventContent); err != nil {
				return nil, err
			} else {
				return e, nil
			}

		}
	}
	return b, nil
}

func (b *EventBase) Serialize() (data []byte, err error) {
	if b.Event != nil {
		if b.EventContent, err = b.ParentSerialize(); err != nil {
			return nil, err
		}
	}
	b.Type = b.GetType()
	data, err = json.Marshal(b)
	return data, err
}

// Overrride it to customize your serializer
func (b *EventBase) ParentSerialize() (data []byte, err error) {
	return json.Marshal(b.Event)
}

//
func (b *EventBase) ParentUnserializer() EventUnserializer {
	return func(data []byte) (Event, error) {
		return nil, errors.New("No Implement")
	}
}
