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
		INSERT INTO generation_projects (project_id, ddl_schema, generation_instructions, max_rows)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (project_id) DO NOTHING;
	`

	_, err := s.dbPool.ExecContext(ctx, insertSQL,
		in.ProjectId,
		in.DdlSchema,
		in.GenerationInstructions,
		in.MaxRows,
	)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save project data")
	}

	event := messaging.ProjectCreatedEvent{
		ProjectID:              in.ProjectId,
		DdlSchema:              in.DdlSchema,
		GenerationInstructions: in.GenerationInstructions,
		MaxRows:                in.MaxRows,
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
		// Replaced errors.InternalServerError
		return nil, status.Error(codes.Internal, "failed to begin transaction")
	}
	defer tx.Rollback()

	const tablesQuery = `SELECT tablename FROM pg_tables WHERE schemaname = $1`

	tableRows, err := tx.QueryContext(ctx, tablesQuery, projectID)
	if err != nil {
		log.Printf("Failed to query project schema: %v", err)
		// Replaced errors.InternalServerError
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
		// Replaced errors.InternalServerError
		return nil, status.Error(codes.Internal, "failed to set search path")
	}

	finalResponse := make(map[string]interface{})

	for _, tableName := range tableNames {
		query := fmt.Sprintf("SELECT * FROM %s", pq.QuoteIdentifier(tableName))

		rows, err := tx.QueryContext(ctx, query)

		if err != nil {
			log.Printf("Failed to query table %s: %v", tableName, err)
			// Replaced errors.InternalServerError
			return nil, status.Error(codes.Internal, "failed to query table")
		}

		tableData, err := scanDynamicRows(rows) // Helper function is unchanged

		if err != nil {
			log.Printf("Failed to scan data from table %s: %v", tableName, err)
			// Replaced errors.InternalServerError
			return nil, status.Error(codes.Internal, "failed to scan data from table")
		}

		rows.Close()

		finalResponse[tableName] = tableData
	}

	// Marshal the final map to a JSON string
	jsonBytes, err := json.Marshal(finalResponse)
	if err != nil {
		log.Printf("Failed to marshal final response: %v", err)
		// Replaced errors.InternalServerError
		return nil, status.Error(codes.Internal, "failed to marshal JSON response")
	}

	// Replaced jsonUtils.WriteJSON(w, http.StatusOK, finalResponse)
	return &pb.GetProjectDataResponse{
		JsonData: string(jsonBytes),
	}, nil
}

// scanDynamicRows (Helper function)
// This function is unchanged as its logic is independent of HTTP or gRPC.
func scanDynamicRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold the values for scanning
		values := make([]interface{}, len(columns))
		// Create a slice of *interface{} to pass to rows.Scan
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		// Scan the row into the slice of pointers
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map to hold the row data [column_name] -> [value]
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Convert []byte (raw bytes) to string for cleaner JSON
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		results = append(results, rowMap)
	}

	return results, nil
}
