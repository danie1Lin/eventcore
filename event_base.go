package eventcore

import (
	"encoding/json"
	"errors"
	"reflect"
	//log "github.com/sirupsen/logrus"
)

type EventBase struct {
	Message string
	Type    EventType
	Emitter string
	Event   Event `json:"-"`
}

func (b *EventBase) Base() *EventBase {
	return b
}

func (b *EventBase) SetBase(base *EventBase) {
}

func (b *EventBase) GetType() EventType {
	return EventType(reflect.Indirect(reflect.ValueOf(b.Event)).Type().String())
}

func (b *EventBase) Unserialize(data []byte) (Event, error) {
	if err := json.Unmarshal(data, b); err != nil {
		return nil, err
	}
	if f := eventUnserailizers.getUnserializer(EventType(b.Type)); f != nil {
		if e, err := f(data); err != nil {
			return nil, err
		} else {
			e.BindSelf(b)
			return e, nil
		}
	}

	return nil, errors.New("unknow event type")
}

func (b *EventBase) Serialize() (data []byte, err error) {
	b.Type = b.GetType()
	data, err = json.Marshal(b.Event)
	return data, err
}

func (b *EventBase) BindSelf(e Event) {
	b.Event = e
}
