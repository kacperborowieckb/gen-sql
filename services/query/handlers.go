package main

import (
	"context"
	"log"

	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *queryServer) Query(ctx context.Context, in *pb.QueryRequest) (*pb.QueryResponse, error) {
	log.Printf("Received Query request for project: %s", in.ProjectId)

	if in.ProjectId == "" || in.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "projectId and query are required")
	}

	return &pb.QueryResponse{
		JsonData: "Query result",
	}, nil
}
