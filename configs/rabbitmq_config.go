package configs

import (
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
	URI          string
	Exchange     string
	ExchangeType string
}

type RabbitMQProducer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  *RabbitMQConfig
}

type RabbitMQConsumer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  *RabbitMQConfig
}

func NewRabbitMQConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		URI:          os.Getenv("RABBITMQ_URI"),
		Exchange:     os.Getenv("RABBITMQ_EXCHANGE"),
		ExchangeType: os.Getenv("RABBITMQ_EXCHANGE_TYPE"),
	}
}

func NewProducerRabbitMQ(config *RabbitMQConfig) (*RabbitMQProducer, error) {
	conn, err := amqp091.Dial(config.URI)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	err = ch.ExchangeDeclare(
		config.Exchange,
		config.ExchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQProducer{
		conn:    conn,
		channel: ch,
		config:  config,
	}, nil
}

func (p *RabbitMQProducer) Publish(routingKey string, body []byte) error {
	return p.channel.Publish(
		p.config.Exchange,
		routingKey,
		false, 
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}

func (p *RabbitMQProducer) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

func NewConsumerRabbitMQ(config *RabbitMQConfig, queueName string) (*RabbitMQConsumer, error) {
	conn, err := amqp091.Dial(config.URI)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	err = ch.QueueBind(
		q.Name,
		queueName, // routing key
		config.Exchange,
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQConsumer{
		conn:    conn,
		channel: ch,
		config:  config,
	}, nil
}

func (c *RabbitMQConsumer) Consume(autoAck bool) (<-chan amqp091.Delivery, error) {
	return c.channel.Consume(
		"",       
		"",       
		autoAck, 
		false,    
		false,    
		false,    
		nil,     
	)
}

func (c *RabbitMQConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
