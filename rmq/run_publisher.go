package rmq

import (
	"fmt"

	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"

	"github.com/dimuls/mycode"
)

type RunPublisher struct {
	connection *rabbitmq.Connection
	channel    *rabbitmq.Channel
}

func NewRunPublisher(rmqURI string) (*RunPublisher, error) {

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

	return &RunPublisher{
		connection: connection,
		channel:    channel,
	}, err
}

func (cp *RunPublisher) Close() error {

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

func (cp *RunPublisher) PublishRun(c *mycode.Run) error {
	runProto, err := proto.Marshal(c)
	if err != nil {
		return fmt.Errorf("proto marshal run: %w", err)
	}
	return cp.channel.Publish("", runQueue, false,
		false, amqp.Publishing{
			ContentType:  "application/protobuf",
			DeliveryMode: amqp.Persistent,
			Body:         runProto,
		})
}
