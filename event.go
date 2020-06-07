package eventcore

type EventType string

type EventHandler func(Event) error

type EventUnserializer func([]byte) (Event, error)

type Event interface {
	GetType() EventType
	Serialize() ([]byte, error)
	Unserialize([]byte) (Event, error)
	ParentSerialize() ([]byte, error)
	ParentUnserializer() EventUnserializer
}
