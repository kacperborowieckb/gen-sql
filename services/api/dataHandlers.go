package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"github.com/kacperborowieckb/gen-sql/utils/errors"
	"github.com/kacperborowieckb/gen-sql/utils/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *apiServer) handleStartDataGeneration(w http.ResponseWriter, r *http.Request) {
	// parse the multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		errors.BadRequestResponse(w, r, fmt.Errorf("error parsing multipart form: %w", err))
		return
	}

	instructions := r.FormValue("generationInstructions")
	maxRowsStr := r.FormValue("maxRows")

	maxRows, err := strconv.ParseInt(maxRowsStr, 10, 32)
	if err != nil {
		errors.BadRequestResponse(w, r, fmt.Errorf("invalid maxRows: must be an integer: %w", err))
		return
	}

	file, fileHeader, err := r.FormFile("ddlFile")
	if err != nil {
		errors.BadRequestResponse(w, r, fmt.Errorf("error retrieving 'ddlFile': %w", err))
		return
	}
	defer file.Close()

	filename := fileHeader.Filename
	if !(strings.HasSuffix(filename, ".sql") || strings.HasSuffix(filename, ".ddl")) {
		errors.BadRequestResponse(w, r, fmt.Errorf("invalid file format: only .sql or .ddl files are allowed, got %s", filename))
		return
	}

	ddlBytes, err := io.ReadAll(file)
	if err != nil {
		errors.InternalServerError(w, r, fmt.Errorf("error reading file content: %w", err))
		return
	}
	ddlSchema := string(ddlBytes)

	if ddlSchema == "" {
		errors.BadRequestResponse(w, r, fmt.Errorf("ddlFile content cannot be empty"))
		return
	}
	if maxRows <= 0 {
		errors.BadRequestResponse(w, r, fmt.Errorf("maxRows must be greater than 0"))
		return
	}

	projectId := uuid.New().String()

	grpcReq := &pb.StartDataGenerationRequest{
		ProjectId:              projectId,
		DdlSchema:              ddlSchema,
		GenerationInstructions: instructions,
		MaxRows:                int32(maxRows),
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	log.Printf("Sending StartDataGeneration gRPC request for new project %s", projectId)
	resp, err := s.dataClient.StartDataGeneration(ctx, grpcReq)
	if err != nil {
		if grpcStatus, ok := status.FromError(err); ok {
			errors.InternalServerError(w, r, fmt.Errorf("gRPC error: [%s] %s", grpcStatus.Code(), grpcStatus.Message()))
		} else {
			errors.InternalServerError(w, r, fmt.Errorf("failed to start data generation: %w", err))
		}
		return
	}

	log.Printf("gRPC call successful. Job ID: %s", resp.GenerationJobId)

	responsePayload := map[string]string{
		"projectId":       projectId,
		"generationJobId": resp.GenerationJobId,
		"message":         resp.Message,
	}

	statusCode := http.StatusCreated

	if !resp.Success {
		statusCode = http.StatusInternalServerError
	}

	json.WriteJSON(w, statusCode, responsePayload)
}

func (s *apiServer) HandleGetProjectData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID := chi.URLParam(r, "id")

	if projectID == "" {
		errors.BadRequestResponse(w, r, fmt.Errorf("project ID is required"))
		return
	}

	grpcRequest := &pb.GetProjectDataRequest{
		ProjectId: projectID,
	}

	log.Printf("Gateway: Forwarding GetProjectData request for %s to DataService", projectID)
	grpcResponse, err := s.dataClient.GetProjectData(ctx, grpcRequest)

	if err != nil {
		st, _ := status.FromError(err)
		httpCode := grpcStatusCodeToHTTP(st.Code())
		json.WriteJSONError(w, httpCode, st.Message())
		return
	}

	json.WriteRawJSON(w, http.StatusOK, []byte(grpcResponse.JsonData))
}

func grpcStatusCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Internal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
