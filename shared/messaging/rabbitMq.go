package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/kacperborowieckb/gen-sql/shared/contracts"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ProjectsExchange = "projects_exchange"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create channel: %v", err)
	}

	return &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}, nil
}

func (r *RabbitMQ) DeclareExchange(name, kind string) error {
	return r.Channel.ExchangeDeclare(
		name,  // name
		kind,  // type (e.g., "topic", "direct", "fanout")
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}

func (r *RabbitMQ) DeclareQueue(name string) (amqp.Queue, error) {
	return r.Channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}

func (r *RabbitMQ) BindQueue(queueName, routingKey, exchangeName string) error {
	return r.Channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
}

// MessageHandler is the function signature for processing a delivered message.
// Return an error to Nack (reject) the message, or nil to Ack (acknowledge) it.
type MessageHandler func(d amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	msgs, err := r.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (we want manual ack)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			err := handler(d)
			if err != nil {
				// Nack the message and drop it (don't requeue)
				log.Printf("Failed to handle message: %v", err)
				d.Nack(false, false)
			} else {
				// Ack the message
				d.Ack(false)
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, exchange, routingKey string, message contracts.AmqpMessage) error {
	log.Printf("Publishing message with routing key: %s", routingKey)

	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         jsonMsg,
	}

	return r.Channel.PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		msg,
	)
}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
	if r.Channel != nil {
		r.Channel.Close()
	}
}

// SetupAppTopology declares all the exchanges, queues, and bindings
// required for the application to run.
func (r *RabbitMQ) SetupAppTopology() error {
	log.Println("Setting up RabbitMQ application topology...")

	if err := r.DeclareExchange(ProjectsExchange, "topic"); err != nil {
		return err
	}

	if err := r.declareAndBind(DataGenerationQueue, ProjectsExchange, contracts.ProjectCreatedRoutingKey); err != nil {
		return err
	}

	log.Println("RabbitMQ application topology setup complete.")

	return nil
}

func (r *RabbitMQ) declareAndBind(queueName, exchangeName, routingKey string) error {
	q, err := r.DeclareQueue(queueName)
	if err != nil {
		return err
	}

	if err := r.BindQueue(q.Name, routingKey, exchangeName); err != nil {
		return err
	}

	return nil
}
