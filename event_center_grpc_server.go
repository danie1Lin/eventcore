package eventcore

import (
	"context"
	"fmt"
	"sync"

	"github.com/daniel840829/eventcore/proto"
	"github.com/go-kit/kit/log/level"
	"google.golang.org/grpc/metadata"
)

type EventCenterServer struct {
	clients            sync.Map // ClientID : *EventCenterClient
	clientSubscription sync.Map // EventType : *EventCenterClientList
	EventCenter        *EventCenterCluster
}

func (server *EventCenterServer) Subscript(c context.Context, info *proto.SubscriptInfo) (*proto.SubscriptResult, error) {
	actual, _ := server.clientSubscription.LoadOrStore(info.EventType, NewEventCenterClientList())
	list := actual.(*EventCenterClientList)
	list.Append(info.ClientID)
	Logger.Log("Get subscriber")
	server.EventCenter.Subscribe(EventType(info.EventType), func(e Event) error {
		Logger.Log("to grpc client", e)
		return nil
	}, info.EventType)
	return &proto.SubscriptResult{Success: true}, nil
}

func (server *EventCenterServer) EventTunnel(tunnel proto.EventCenter_EventTunnelServer) error {
	md, ok := metadata.FromIncomingContext(tunnel.Context())
	if !ok {
		return fmt.Errorf("No metadata")
	}
	data := md.Get("Client ID")
	if data == nil || len(data) != 1 {
		return fmt.Errorf("No client id")
	}
	id := data[0]
	_, isStored := server.clients.LoadOrStore(id, &EventCenterClient{ID: id, Tunnel: tunnel})
	if !isStored {
		return fmt.Errorf("not support multi tunnel for same client id: %s", id)
	}

	for {
		payload, err := tunnel.Recv()
		if err != nil {
			level.Error(Logger).Log("error", err, "client id", id, "message", "received error")
			break
		}
		var e Event = &EventBase{}
		e, err = e.Unserialize(payload.Content)
		server.EventCenter.Emit(e)
	}
	return nil
}
