package eventcore

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kit/kit/log/level"
)

type AmqpBackend struct {
	suffix         string
	Url            string
	configuaration amqp.Config
	publisher      *amqp.Publisher
	subscriber     *amqp.Subscriber
}

func NewAmqpBackend(connectionString string) *AmqpBackend {
	return &AmqpBackend{
		Url: connectionString,
	}
}

func (a *AmqpBackend) Close() error {
	if a.publisher != nil {
		err := a.publisher.Close()
		if err != nil {
			level.Error(Logger).Log(err)
		}
	}
	if a.subscriber != nil {
		err := a.subscriber.Close()
		if err != nil {
			level.Error(Logger).Log(err)
		}
	}

	return nil
}

func (a *AmqpBackend) SetConsumerGroupName(name string) {
	a.suffix = name
}

func (a *AmqpBackend) config() amqp.Config {
	level.Debug(Logger).Log("connection", a.Url, "consummer_group", a.suffix)
	a.configuaration = amqp.NewDurablePubSubConfig(
		a.Url,
		amqp.GenerateQueueNameTopicNameWithSuffix(a.suffix),
	)

	return a.configuaration
}

func (a *AmqpBackend) Subscriber() message.Subscriber {
	if a.subscriber == nil {
		var err error
		a.subscriber, err = amqp.NewSubscriber(a.config(), watermill.NewStdLogger(false, false))
		if err != nil {
			panic(err)
		}
	}
	return a.subscriber
}

func (a *AmqpBackend) Publisher() message.Publisher {
	if a.publisher == nil {
		var err error
		a.publisher, err = amqp.NewPublisher(a.config(), watermill.NewStdLogger(false, false))
		if err != nil {
			panic(err)
		}
	}
	return a.publisher
}
