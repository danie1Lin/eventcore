package eventcore

import (
	"github.com/ThreeDotsLabs/watermill/message"
)

type Backend interface {
	Publisher() message.Publisher
	Subscriber() message.Subscriber
	SetConsumerGroupName(name string)
	Close() error
}
