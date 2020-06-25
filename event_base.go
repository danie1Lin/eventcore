package eventcore

import (
	"encoding/json"
	"fmt"
	"reflect"
	//log "github.com/sirupsen/logrus"
)

type EventBase struct {
	Message string
	Type    EventType
	Emitter string
	From    string
	Event   Event `json:"-"`
	Traces  []DispatcherInfo
}

func (b *EventBase) AddNode(d DispatcherInfo) {
	if b.Traces == nil {
		b.Traces = []DispatcherInfo{d}
	} else {
		b.Traces = append(b.Traces, d)
	}
}

func (b *EventBase) GetSourceID() string {
	return b.From
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
			e.BindSelf(e)
			return e, nil
		}
	}
	return nil, fmt.Errorf("unknow event type %s", b.Type)
}

func (b *EventBase) Serialize() (data []byte, err error) {
	b.Type = b.GetType()
	data, err = json.Marshal(b.Event)
	return data, err
}

func (b *EventBase) BindSelf(e Event) {
	b.Event = e
}
