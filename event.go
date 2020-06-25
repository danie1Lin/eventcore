package eventcore

type EventType string

type EventHandler func(Event) error

type EventUnserializer func([]byte) (Event, error)

type Event interface {
	GetType() EventType
	GetSourceID() string
	Serialize() ([]byte, error)
	BindSelf(Event)
	Unserialize(data []byte) (Event, error)
	AddNode(DispatcherInfo)
}

type Dispatcher interface {
	GetDispatcherInfo() DispatcherInfo
}

type DispatcherInfo struct {
	Metadata map[string]string
	Name     string
	Type     string
	ID       string
}

func (d *DispatcherInfo) GetType() string {
	return d.Type
}

func (d *DispatcherInfo) GetName() string {
	return d.Name
}

func (d *DispatcherInfo) GetID() string {
	return d.ID
}

func (d *DispatcherInfo) GetMetadata() map[string]string {
	return d.Metadata
}
