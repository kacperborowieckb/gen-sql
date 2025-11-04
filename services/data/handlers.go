package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/kacperborowieckb/gen-sql/shared/contracts"
	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"github.com/kacperborowieckb/gen-sql/shared/messaging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *dataServer) StartDataGeneration(ctx context.Context, in *pb.StartDataGenerationRequest) (*pb.StartDataGenerationResponse, error) {
	log.Printf("Received StartDataGeneration request for project: %s", in.ProjectId)

	if in.ProjectId == "" || in.DdlSchema == "" {
		return nil, status.Error(codes.InvalidArgument, "projectId and ddlSchema are required")
	}

	// Create ProjectCreatedEvent
	event := messaging.ProjectCreatedEvent{
		ProjectID:              in.ProjectId,
		DdlSchema:              in.DdlSchema,
		GenerationInstructions: in.GenerationInstructions,
		MaxRows:                in.MaxRows,
	}

	// Marshal event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal ProjectCreatedEvent: %v", err)
		return nil, status.Error(codes.Internal, "failed to marshal event data")
	}

	// Wrap in AmqpMessage
	amqpMsg := contracts.AmqpMessage{
		OwnerId: in.ProjectId,
		Data:    eventData,
	}

	// Publish message to RabbitMQ
	if err := s.mqClient.PublishMessage(ctx, messaging.ProjectsExchange, contracts.ProjectCreatedRoutingKey, amqpMsg); err != nil {
		log.Printf("Failed to publish message to RabbitMQ: %v", err)
		return nil, status.Error(codes.Internal, "failed to publish message to queue")
	}

	log.Printf("Successfully published ProjectCreatedEvent for project: %s", in.ProjectId)

	return &pb.StartDataGenerationResponse{
		GenerationJobId: "mockJobId",
		Message:         "Data generation job successfully queued.",
		Success:         true,
	}, nil
}
