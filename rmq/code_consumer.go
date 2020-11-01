package rmq

import (
	"context"
	"fmt"
	"sync"

	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"

	"github.com/dimuls/mycode"
)

type CodeHandler interface {
	HandleCode(context.Context, *mycode.Code) error
}

type CodeConsumer struct {
	codeHandler  CodeHandler
	connection   *rabbitmq.Connection
	channel      *rabbitmq.Channel
	log          *logrus.Entry
	stopHandlers func()
	wg           sync.WaitGroup
}

const codeQueue = "code"

func NewCodeConsumer(rmqURI string, qos int, ch CodeHandler) (
	cc *CodeConsumer, err error) {

	cc = &CodeConsumer{
		codeHandler: ch,
		log:         logrus.WithField("subsystem", "nats_code_consumer"),
	}

	cc.connection, err = rabbitmq.Dial(rmqURI)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	cc.channel, err = cc.connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	_, err = cc.channel.QueueDeclare(codeQueue, true, false,
		false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	err = cc.channel.Qos(qos, 0, false)
	if err != nil {
		return nil, fmt.Errorf("set qos: %w", err)
	}

	msgs, err := cc.channel.Consume(codeQueue, "", false,
		false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consume messages: %w", err)
	}

	var ctx context.Context

	ctx, cc.stopHandlers = context.WithCancel(context.Background())

	cc.wg.Add(1)
	go func() {
		defer cc.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				cc.wg.Add(1)
				go cc.handleMsg(ctx, msg)
			}
		}
	}()

	return
}

func (cc *CodeConsumer) handleMsg(ctx context.Context, msg amqp.Delivery) {

	defer cc.wg.Done()

	code := &mycode.Code{}

	err := proto.Unmarshal(msg.Body, code)
	if err != nil {
		logrus.WithError(err).Error(
			"failed to JSON unmarshal code: %w", err)
		err = msg.Nack(false, false)
		if err != nil {
			cc.log.WithError(err).Error("failed to nack")
		}
		return
	}

	log := cc.log.WithField("solution_test_id", code.SolutionTestId)

	err = cc.codeHandler.HandleCode(ctx, code)
	if err != nil {
		err = msg.Nack(false, false)
		if err != nil {
			log.WithError(err).Error("failed to nack")
		}
		return
	}

	err = msg.Ack(false)
	if err != nil {
		log.WithError(err).Error("failed to ack")
	}
}

func (cc *CodeConsumer) Close() error {

	cc.stopHandlers()
	cc.wg.Wait()

	err := cc.channel.Close()
	if err != nil {
		return fmt.Errorf("close channel: %w", err)
	}

	err = cc.connection.Close()
	if err != nil {
		return fmt.Errorf("close connection: %w", err)
	}

	return nil
}
