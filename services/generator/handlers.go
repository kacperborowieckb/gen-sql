package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/kacperborowieckb/gen-sql/shared/contracts"
	"github.com/kacperborowieckb/gen-sql/shared/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/genai"
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

	ctx := context.Background()

	schemaName := fmt.Sprintf("\"project_%s\"", event.ProjectID)

	tx, err := s.dbPool.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	createSchemaSQL := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	if _, err := tx.ExecContext(ctx, createSchemaSQL); err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schemaName, err)
	}

	setSearchPathSQL := fmt.Sprintf("SET LOCAL search_path = %s", schemaName)
	if _, err := tx.ExecContext(ctx, setSearchPathSQL); err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}
	log.Printf("Set transaction search_path to %s", schemaName)

	if _, err := tx.ExecContext(ctx, event.DdlSchema); err != nil {
		log.Printf("Failed to execute user DDL in schema %s. DDL: %s", schemaName, event.DdlSchema)
		return fmt.Errorf("failed to execute user DDL: %w", err)
	}
	log.Printf("Successfully executed user DDL for project %s", event.ProjectID)

	prompt := BuildGenerationPrompt(event.DdlSchema, event.GenerationInstructions)

	result, err := s.genaiClient.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)

	if err != nil {
		return fmt.Errorf("call to gemini failed: %w", err)
	}

	generatedSQL := result.Text()
	if generatedSQL == "" {
		return fmt.Errorf("gemini returned an empty response")
	}

	log.Printf("Received %d bytes of SQL from AI", len(generatedSQL))

	if _, err := tx.ExecContext(ctx, generatedSQL); err != nil {
		logSnippet := generatedSQL

		if len(logSnippet) > 200 {
			logSnippet = logSnippet[:200] + "..."
		}

		log.Printf("Failed to execute generated SQL: %s", logSnippet)

		return fmt.Errorf("failed to execute generated SQL: %w", err)
	}

	log.Printf("Successfully executed generated SQL for project %s", event.ProjectID)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully completed data generation for project %s", event.ProjectID)

	return nil
}
