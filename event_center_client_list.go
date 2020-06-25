package eventcore

import (
	"net"
	"sync"

	"github.com/daniel840829/eventcore/proto"
)

type EventCenterClient struct {
	ID     string
	Tunnel proto.EventCenter_EventTunnelServer
	conn   net.Conn
}

type EventCenterClientList struct {
	sync.RWMutex
	list []string
}

func NewEventCenterClientList() *EventCenterClientList {
	return &EventCenterClientList{
		list: make([]string, 0),
	}
}

func (l *EventCenterClientList) Len() int {
	result := 0
	l.RLock()
	defer l.Unlock()
	result = len(l.list)
	return result
}

func (l *EventCenterClientList) Get(index int) string {
	result := ""
	l.RLock()
	defer l.RUnlock()
	if index > len(l.list) {
		return ""
	}
	result = l.list[index]
	return result
}

func (l *EventCenterClientList) Append(ID string) {
	l.Lock()
	defer l.Unlock()
	l.list = append(l.list, ID)
}

func (l *EventCenterClientList) AppendIfNonExist(ID string) bool {
	l.Lock()
	defer l.Unlock()
	for _, v := range l.list {
		if v == ID {
			return false
		}
	}
	l.list = append(l.list, ID)
	return true
}

func (l *EventCenterClientList) Range(f func(clientID string) bool) {
	l.RLock()
	defer l.RUnlock()
	for i := range l.list {
		if !f(l.list[i]) {
			return
		}
	}
}
