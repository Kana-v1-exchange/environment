package rmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RMQSettings struct {
	User     string
	Password string
	Host     string
	Port     string
}

type RmqHandler interface {
	Write(msg string) error
	Read() (<-chan amqp.Delivery, error)
}

type rmqClient struct {
	ch *amqp.Channel
}

func (rmqS *RMQSettings) Connect() RmqHandler {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s", rmqS.User, rmqS.Password, rmqS.Host, rmqS.Port))

	if err != nil {
		panic(fmt.Sprintf("cannot connect to the rmq; err: %v", err))
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("rmq connection cannot create a channel; err: %v", err))
	}

	_, err = ch.QueueDeclare(
		"exchanges",
		true,
		false,
		true,
		false,
		nil,
	)

	if err != nil {
		panic(fmt.Errorf("cannot create the 'exchange' queue; err: %v", err))
	}

	return &rmqClient{
		ch: ch,
	}
}

func (rc *rmqClient) Write(msg string) error {
	err := rc.ch.Publish(
		"",
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		},
	)

	if err != nil {
		return fmt.Errorf("cannot publish message '%s'; err: %v", msg, err)
	}

	return nil
}

func (rc *rmqClient) Read() (<-chan amqp.Delivery, error) {
	msgs, err := rc.ch.Consume(
		"exchanges",
		"",
		true,
		false,
		false,
		true,
		amqp.Table{},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot get messages from the queue 'exchanges'; err: %v", err)
	}

	return msgs, nil
}
