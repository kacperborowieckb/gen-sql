package main

import (
	"context"
	"log"

	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *dataServer) StartDataGeneration(ctx context.Context, in *pb.StartDataGenerationRequest) (*pb.StartDataGenerationResponse, error) {
	log.Printf("Received StartDataGeneration request for project: %s", in.ProjectId)

	if in.ProjectId == "" || in.DdlSchema == "" {
		return nil, status.Error(codes.InvalidArgument, "projectId and ddlSchema are required")
	}

	return &pb.StartDataGenerationResponse{
		GenerationJobId: "mockJobId",
		Message:         "Data generation job successfully queued.",
		Success:         true,
	}, nil
}
