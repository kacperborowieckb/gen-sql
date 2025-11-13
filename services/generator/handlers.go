package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/kacperborowieckb/gen-sql/shared/contracts"
	"github.com/kacperborowieckb/gen-sql/shared/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (s *generatorServer) handleProjectCreated(d amqp.Delivery) error {
	log.Printf("Received a message with routing key: %s", d.RoutingKey)

	var amqpMsg contracts.AmqpMessage

	if err := json.Unmarshal(d.Body, &amqpMsg); err != nil {
		log.Printf("Failed to unmarshal AmqpMessage: %v. Body: %s", err, string(d.Body))
		return fmt.Errorf("failed to unmarshal outer AmqpMessage: %w", err)
	}

	var event messaging.ProjectCreatedEvent

	if err := json.Unmarshal(amqpMsg.Data, &event); err != nil {
		log.Printf("Failed to unmarshal ProjectCreatedEvent: %v. Data: %s", err, string(amqpMsg.Data))
		return fmt.Errorf("failed to unmarshal inner ProjectCreatedEvent: %w", err)
	}

	return nil
}
