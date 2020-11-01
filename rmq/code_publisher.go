package rmq

import (
	"fmt"

	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"

	"github.com/dimuls/mycode"
)

type CodePublisher struct {
	connection *rabbitmq.Connection
	channel    *rabbitmq.Channel
}

func NewCodePublisher(rmqURI string) (*CodePublisher, error) {

	connection, err := rabbitmq.Dial(rmqURI)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	_, err = channel.QueueDeclare(runQueue, true, false,
		false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	err = channel.Confirm(false)
	if err != nil {
		return nil, fmt.Errorf("set channel confirm mode: %w", err)
	}

	return &CodePublisher{
		connection: connection,
		channel:    channel,
	}, err
}

func (cp *CodePublisher) Close() error {

	err := cp.channel.Close()
	if err != nil {
		return fmt.Errorf("close channel: %w", err)
	}

	err = cp.connection.Close()
	if err != nil {
		return fmt.Errorf("close connection: %w", err)
	}

	return nil
}

func (cp *CodePublisher) PublishCode(c *mycode.Code) error {
	codeProto, err := proto.Marshal(c)
	if err != nil {
		return fmt.Errorf("proto marshal code: %w", err)
	}
	return cp.channel.Publish("", codeQueue, false,
		false, amqp.Publishing{
			ContentType:  "application/protobuf",
			DeliveryMode: amqp.Persistent,
			Body:         codeProto,
		})
}
