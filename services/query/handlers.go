package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"github.com/kacperborowieckb/gen-sql/utils/parsing"
	"github.com/lib/pq"
	"google.golang.org/genai"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *queryServer) Query(ctx context.Context, in *pb.QueryRequest) (*pb.QueryResponse, error) {
	log.Printf("Received Query request for project: %s", in.ProjectId)

	if in.ProjectId == "" || in.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "projectId and query are required")
	}

	var ddlSchema string
	const selectSQL = `SELECT ddl_schema FROM generation_projects WHERE project_id = $1`

	// Execute the query to fetch the ddl_schema
	err := s.dbPool.QueryRowContext(ctx, selectSQL, in.ProjectId).Scan(&ddlSchema)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Project %s not found in generation_projects", in.ProjectId)
			return nil, status.Error(codes.NotFound, fmt.Sprintf("project ID %s not found", in.ProjectId))
		}

		log.Printf("Failed to retrieve ddl_schema for project %s: %v", in.ProjectId, err)
		return nil, status.Error(codes.Internal, "failed to query project schema from database")
	}

	log.Printf("Successfully retrieved DDL Schema for project %s.", in.ProjectId)

	result, err := s.genaiClient.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(GetQueryPrompt(ddlSchema, in.Query)),
		&genai.GenerateContentConfig{},
	)

	if err != nil {
		return nil, fmt.Errorf("call to gemini failed: %w", err)
	}

	generatedQuery := result.Text()

	if generatedQuery == "" {
		return nil, status.Error(codes.Internal, "Gemini returned an empty query string")
	}

	tx, err := s.dbPool.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to begin transaction")
	}
	defer tx.Rollback()

	setSearchPathSQL := fmt.Sprintf("SET LOCAL search_path = %s", pq.QuoteIdentifier(in.ProjectId))
	if _, err := tx.ExecContext(ctx, setSearchPathSQL); err != nil {
		log.Printf("Failed to set search path for project %s: %v", in.ProjectId, err)
		return nil, status.Error(codes.Internal, "failed to set search path")
	}

	log.Printf("Executing generated SQL: %s", generatedQuery)
	rows, err := tx.QueryContext(ctx, generatedQuery)
	if err != nil {
		log.Printf("Failed to execute generated query for project %s: %v", in.ProjectId, err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("SQL query failed: %v", err))
	}
	defer rows.Close()

	queryData, err := parsing.ScanDynamicRows(rows)
	if err != nil {
		log.Printf("Failed to scan query results for project %s: %v", in.ProjectId, err)
		return nil, status.Error(codes.Internal, "failed to process query results")
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return nil, status.Error(codes.Internal, "failed to commit transaction")
	}

	jsonBytes, err := json.Marshal(queryData)
	if err != nil {
		log.Printf("Failed to marshal query result to JSON: %v", err)
		return nil, status.Error(codes.Internal, "failed to encode JSON response")
	}

	cacheID := uuid.New().String()
	s.cache.Set(cacheID, string(jsonBytes), 5*time.Minute)

	return &pb.QueryResponse{
		GeneratedQuery: generatedQuery,
		JsonData:       string(jsonBytes),
		CacheId:        cacheID,
	}, nil
}

func (s *queryServer) ExportCsv(ctx context.Context, in *pb.ExportCsvRequest) (*pb.ExportCsvResponse, error) {
	if in.GetCacheId() == "" {
		return nil, status.Error(codes.InvalidArgument, "cacheId is required")
	}

	cached, ok := s.cache.Get(in.CacheId)
	if !ok {
		return nil, status.Error(codes.NotFound, "cached result not found or expired")
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(cached), &rows); err != nil {
		log.Printf("Failed to unmarshal cached data for cacheId %s: %v", in.CacheId, err)
		return nil, status.Error(codes.Internal, "failed to read cached data")
	}

	headerSet := make(map[string]struct{})
	for _, row := range rows {
		for k := range row {
			headerSet[k] = struct{}{}
		}
	}

	if len(headerSet) == 0 {
		return nil, status.Error(codes.NotFound, "no data available for export")
	}

	headers := make([]string, 0, len(headerSet))
	for k := range headerSet {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err := writer.Write(headers); err != nil {
		log.Printf("Failed to write CSV headers for cacheId %s: %v", in.CacheId, err)
		return nil, status.Error(codes.Internal, "failed to generate csv")
	}

	for _, row := range rows {
		record := make([]string, len(headers))
		for i, col := range headers {
			if val, ok := row[col]; ok && val != nil {
				record[i] = fmt.Sprint(val)
			} else {
				record[i] = ""
			}
		}

		if err := writer.Write(record); err != nil {
			log.Printf("Failed to write CSV record for cacheId %s: %v", in.CacheId, err)
			return nil, status.Error(codes.Internal, "failed to generate csv")
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Printf("Failed to flush CSV writer for cacheId %s: %v", in.CacheId, err)
		return nil, status.Error(codes.Internal, "failed to generate csv")
	}

	return &pb.ExportCsvResponse{
		CsvData: buf.String(),
	}, nil
}
