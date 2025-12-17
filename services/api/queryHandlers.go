package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"github.com/kacperborowieckb/gen-sql/utils/errors"
	"github.com/kacperborowieckb/gen-sql/utils/json"
	"google.golang.org/grpc/status"
)

func (s *apiServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID := chi.URLParam(r, "id")

	if projectID == "" {
		errors.BadRequestResponse(w, r, fmt.Errorf("project ID is required"))
		return
	}

	type QueryPayload struct {
		Query string `json:"query"`
	}

	var payload QueryPayload

	err := json.ReadJSON(w, r, &payload)

	if err != nil {
		errors.BadRequestResponse(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	query := payload.Query

	if query == "" {
		errors.BadRequestResponse(w, r, fmt.Errorf("query field cannot be empty in the request body"))
		return
	}

	queryGrpcRequest := &pb.QueryRequest{
		ProjectId: projectID,
		Query:     query,
	}

	queryGrpcResponse, err := s.queryClient.Query(ctx, queryGrpcRequest)

	if err != nil {
		st, _ := status.FromError(err)
		httpCode := GrpcStatusCodeToHTTP(st.Code())
		json.WriteJSONError(w, httpCode, st.Message())
		return
	}

	json.WriteJSON(w, http.StatusOK, queryGrpcResponse)
}

func (s *apiServer) handleExportQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cacheID := chi.URLParam(r, "cacheId")

	if cacheID == "" {
		errors.BadRequestResponse(w, r, fmt.Errorf("cacheId is required"))
		return
	}

	exportResp, err := s.queryClient.ExportCsv(ctx, &pb.ExportCsvRequest{
		CacheId: cacheID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		httpCode := GrpcStatusCodeToHTTP(st.Code())
		json.WriteJSONError(w, httpCode, st.Message())
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="export-%s.csv"`, cacheID))
	w.WriteHeader(http.StatusOK)
	if _, writeErr := w.Write([]byte(exportResp.CsvData)); writeErr != nil {
		log.Printf("failed to write csv response: %v", writeErr)
	}
}
