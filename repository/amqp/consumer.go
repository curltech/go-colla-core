package amqp

import (
	"github.com/curltech/go-colla-core/logger"
	"github.com/streadway/amqp"
	"time"
)

type Consumer struct {
	Uri          string //amqp://guest:guest@localhost:5672/
	ExchangeName string // "test-exchange"
	ExchangeType string // "direct", "Exchange type - direct|fanout|topic|x-custom"
	QueueName    string //("test-key", "AMQP routing key"
	BindingKey   string
	ReceiveFunc  func(data []byte)
	lifetime     time.Duration
	conn         *amqp.Connection
	channel      *amqp.Channel
	tag          string
	done         chan error
}

func NewConsumer(uri, exchange, exchangeType, queueName, key, ctag string) (*Consumer, error) {
	c := &Consumer{
		Uri:          uri,
		ExchangeName: exchange,
		ExchangeType: exchangeType,
		QueueName:    queueName,
		BindingKey:   key,
		conn:         nil,
		channel:      nil,
		tag:          ctag,
		done:         make(chan error),
	}

	var err error

	logger.Sugar.Infof("dialing %q", uri)
	c.conn, err = amqp.Dial(uri)
	if err != nil {
		logger.Sugar.Errorf("Dial: %s", err)
		return nil, err
	}

	go func() {
		logger.Sugar.Infof("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	logger.Sugar.Infof("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		logger.Sugar.Errorf("Channel: %s", err)
		return nil, err
	}

	logger.Sugar.Infof("got Channel, declaring Exchange (%q)", exchange)
	if err = c.channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		logger.Sugar.Errorf("Exchange Declare: %s", err)
		return nil, err
	}

	logger.Sugar.Infof("declared Exchange, declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		logger.Sugar.Errorf("Queue Declare: %s", err)
		return nil, err
	}

	logger.Sugar.Infof("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		logger.Sugar.Errorf("Queue Bind: %s", err)
		return nil, err
	}

	logger.Sugar.Infof("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		logger.Sugar.Errorf("Queue Consume: %s", err)
		return nil, err
	}

	go c.handle(deliveries, c.done)

	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		logger.Sugar.Errorf("Consumer cancel failed: %s", err)
		return err
	}

	if err := c.conn.Close(); err != nil {
		logger.Sugar.Errorf("AMQP connection close error: %s", err)
		return err
	}

	defer logger.Sugar.Infof("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func (c *Consumer) handle(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		logger.Sugar.Infof(
			"got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
		d.Ack(false)
		if c.ReceiveFunc != nil {
			c.ReceiveFunc(d.Body)
		}
	}
	logger.Sugar.Infof("handle: deliveries channel closed")
	done <- nil
}
