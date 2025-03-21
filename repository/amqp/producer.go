package amqp

import (
	"github.com/curltech/go-colla-core/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	Uri          string //amqp://guest:guest@localhost:5672/
	ExchangeName string // "test-exchange"
	ExchangeType string // "direct", "Exchange type - direct|fanout|topic|x-custom"
	RoutingKey   string //("test-key", "AMQP routing key"
	Reliable     bool
	SuccessFunc  func(confirms *amqp.Confirmation) // "Wait for the publisher confirmation before exiting"
	FailFunc     func(confirms *amqp.Confirmation)
	connection   *amqp.Connection
	channel      *amqp.Channel
}

func NewProducer(uri, exchange, exchangeType, routingKey string, reliable bool) (*Producer, error) {
	producer := &Producer{
		uri,
		exchange,
		exchangeType,
		routingKey,
		false,
		nil,
		nil,
		nil,
		nil,
	}
	// This function dials, connects, declares, publishes, and tears down,
	// all in one go. In a real service, you probably want to maintain a
	// long-lived connection as state, and publish against that.

	logger.Sugar.Infof("dialing %q", uri)
	var err error
	producer.connection, err = amqp.Dial(uri)
	if err != nil {
		logger.Sugar.Errorf("Dial: %s", err.Error())

		return nil, err
	}
	logger.Sugar.Infof("got Connection, getting Channel")
	producer.channel, err = producer.connection.Channel()
	if err != nil {
		logger.Sugar.Errorf("Channel: %s", err)
		return nil, err
	}

	logger.Sugar.Infof("got Channel, declaring %q Exchange (%q)", producer.ExchangeType, producer.ExchangeName)
	if err = producer.channel.ExchangeDeclare(
		producer.ExchangeName, // name
		producer.ExchangeType, // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // noWait
		nil,                   // arguments
	); err != nil {
		logger.Sugar.Errorf("Exchange Declare: %s", err)
		return nil, err
	}

	// Reliable publisher confirms require confirm.select support from the
	// connection.
	if producer.Reliable {
		logger.Sugar.Infof("enabling publishing confirms.")
		if err = producer.channel.Confirm(false); err != nil {
			logger.Sugar.Errorf("Channel could not be put into confirm mode: %s", err)
			return nil, err
		}

		confirms := producer.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

		defer producer.confirm(confirms)
	}

	return producer, nil
}

func (this *Producer) Close() error {
	err := this.channel.Close()
	if err != nil {
		logger.Sugar.Errorf("Producer close failed: %s", err)
		return err
	}
	err = this.connection.Close()
	if err != nil {
		logger.Sugar.Errorf("AMQP connection close error: %s", err)
		return err
	}

	return nil
}

func (this *Producer) Send(data []byte) error {
	logger.Sugar.Infof("declared Exchange, publishing %dB body (%q)", len(data), data)
	if err := this.channel.Publish(
		this.ExchangeName, // publish to an exchange
		this.RoutingKey,   // routing to 0 or more queues
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            data,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		logger.Sugar.Errorf("Exchange Publish: %s", err)
		return err
	}

	return nil
}

// One would typically keep a channel of publishings, a sequence number, and a
// set of unacknowledged sequence numbers and loop until the publishing channel
// is closed.
func (this *Producer) confirm(confirms <-chan amqp.Confirmation) {
	logger.Sugar.Infof("waiting for confirmation of one publishing")

	if confirmed := <-confirms; confirmed.Ack {
		logger.Sugar.Infof("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
		if this.SuccessFunc != nil {
			this.SuccessFunc(&confirmed)
		}
	} else {
		logger.Sugar.Errorf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
		if this.FailFunc != nil {
			this.FailFunc(&confirmed)
		}
	}
}
