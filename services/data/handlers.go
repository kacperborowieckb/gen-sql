package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/kacperborowieckb/gen-sql/shared/contracts"
	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"github.com/kacperborowieckb/gen-sql/shared/messaging"
	"github.com/kacperborowieckb/gen-sql/utils/parsing"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *dataServer) StartDataGeneration(ctx context.Context, in *pb.StartDataGenerationRequest) (*pb.StartDataGenerationResponse, error) {
	log.Printf("Received StartDataGeneration request for project: %s", in.ProjectId)

	if in.ProjectId == "" || in.DdlSchema == "" {
		return nil, status.Error(codes.InvalidArgument, "projectId and ddlSchema are required")
	}

	const insertSQL string = `
		INSERT INTO generation_projects (project_id, ddl_schema, instructions, rows_to_generate, temperature)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (project_id) DO NOTHING;
	`

	_, err := s.dbPool.ExecContext(ctx, insertSQL,
		in.ProjectId,
		in.DdlSchema,
		in.Instructions,
		in.RowsToGenerate,
		in.Temperature,
	)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save project data")
	}

	event := messaging.ProjectCreatedEvent{
		ProjectID:      in.ProjectId,
		DdlSchema:      in.DdlSchema,
		Instructions:   in.Instructions,
		RowsToGenerate: in.RowsToGenerate,
		Temperature:    in.Temperature,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal ProjectCreatedEvent: %v", err)
		return nil, status.Error(codes.Internal, "failed to marshal event data")
	}

	amqpMsg := contracts.AmqpMessage{
		OwnerId: in.ProjectId,
		Data:    eventData,
	}

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

func (s *dataServer) GetProjectData(ctx context.Context, in *pb.GetProjectDataRequest) (*pb.GetProjectDataResponse, error) {
	projectID := in.ProjectId

	log.Printf("Received GetProjectData request for project: %s", projectID)

	if projectID == "" {
		return nil, status.Error(codes.InvalidArgument, "project ID is required")
	}

	tx, err := s.dbPool.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return nil, status.Error(codes.Internal, "failed to begin transaction")
	}
	defer tx.Rollback()

	const tablesQuery = `SELECT tablename FROM pg_tables WHERE schemaname = $1`

	tableRows, err := tx.QueryContext(ctx, tablesQuery, projectID)
	if err != nil {
		log.Printf("Failed to query project schema: %v", err)
		return nil, status.Error(codes.Internal, "failed to query project schema")
	}
	defer tableRows.Close()

	var tableNames []string

	for tableRows.Next() {
		var tableName string

		if err := tableRows.Scan(&tableName); err != nil {
			log.Printf("Failed to scan table name: %v", err)
			return nil, status.Error(codes.Internal, "failed to scan table name")
		}

		tableNames = append(tableNames, tableName)
	}

	if len(tableNames) == 0 {
		log.Printf("No tables found for schema: %s", projectID)
		return &pb.GetProjectDataResponse{JsonData: "{}"}, nil
	}

	setSearchPathSQL := fmt.Sprintf("SET LOCAL search_path = %s", pq.QuoteIdentifier(projectID))
	if _, err := tx.ExecContext(ctx, setSearchPathSQL); err != nil {
		log.Printf("Failed to set search path: %v", err)
		return nil, status.Error(codes.Internal, "failed to set search path")
	}

	finalResponse := make(map[string]interface{})

	for _, tableName := range tableNames {
		query := fmt.Sprintf("SELECT * FROM %s", pq.QuoteIdentifier(tableName))

		rows, err := tx.QueryContext(ctx, query)

		if err != nil {
			log.Printf("Failed to query table %s: %v", tableName, err)
			return nil, status.Error(codes.Internal, "failed to query table")
		}

		tableData, err := parsing.ScanDynamicRows(rows)

		if err != nil {
			log.Printf("Failed to scan data from table %s: %v", tableName, err)
			return nil, status.Error(codes.Internal, "failed to scan data from table")
		}

		rows.Close()

		finalResponse[tableName] = tableData
	}

	jsonBytes, err := json.Marshal(finalResponse)
	if err != nil {
		log.Printf("Failed to marshal final response: %v", err)
		return nil, status.Error(codes.Internal, "failed to marshal JSON response")
	}

	return &pb.GetProjectDataResponse{
		JsonData: string(jsonBytes),
	}, nil
}

func (s *dataServer) GetProjects(ctx context.Context, in *pb.GetProjectsRequest) (*pb.GetProjectsResponse, error) {
	const query = `SELECT project_id FROM generation_projects`

	rows, err := s.dbPool.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Failed to query project IDs: %v", err)
		return nil, status.Error(codes.Internal, "failed to fetch project list")
	}
	defer rows.Close()

	projectIDs := []string{}

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Printf("Failed to scan project ID: %v", err)
			return nil, status.Error(codes.Internal, "failed to scan project data")
		}
		projectIDs = append(projectIDs, id)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, status.Error(codes.Internal, "error reading project list")
	}

	return &pb.GetProjectsResponse{
		ProjectIds: projectIDs,
	}, nil
}

func (s *dataServer) DeleteProject(ctx context.Context, in *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	projectID := in.ProjectId

	log.Printf("Received DeleteProject request for project: %s", projectID)

	if projectID == "" {
		return nil, status.Error(codes.InvalidArgument, "project ID is required")
	}

	// Use a transaction to ensure both operations succeed or fail together
	tx, err := s.dbPool.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return nil, status.Error(codes.Internal, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Drop the schema with CASCADE
	dropSchemaSQL := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pq.QuoteIdentifier(projectID))
	if _, err := tx.ExecContext(ctx, dropSchemaSQL); err != nil {
		log.Printf("Failed to drop schema %s: %v", projectID, err)
		return nil, status.Error(codes.Internal, "failed to drop schema")
	}

	log.Printf("Successfully dropped schema %s", projectID)

	// Delete from generation_projects table
	const deleteSQL = `DELETE FROM generation_projects WHERE project_id = $1`
	result, err := tx.ExecContext(ctx, deleteSQL, projectID)
	if err != nil {
		log.Printf("Failed to delete project record: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete project record")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get rows affected: %v", err)
		return nil, status.Error(codes.Internal, "failed to verify deletion")
	}

	if rowsAffected == 0 {
		log.Printf("Project %s not found in generation_projects table", projectID)
		return &pb.DeleteProjectResponse{
			Success: false,
			Message: "Project not found",
		}, nil
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return nil, status.Error(codes.Internal, "failed to commit transaction")
	}

	log.Printf("Successfully deleted project %s", projectID)

	return &pb.DeleteProjectResponse{
		Success: true,
		Message: "Project deleted successfully",
	}, nil
}
