package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"github.com/kacperborowieckb/gen-sql/utils/env"
	"github.com/kacperborowieckb/gen-sql/utils/health"
	"github.com/kacperborowieckb/gen-sql/utils/json"
	"github.com/kacperborowieckb/gen-sql/utils/shutdown"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type apiServer struct {
	dataClient pb.TestServiceClient
}

func main() {
	port := env.GetString("PORT", "8080")

	// --- gRPC Client Setup ---
	dataSvcAddr := env.GetString("DATA_SERVICE_ADDR", "localhost:8081")
	isInsecure := env.GetString("DATA_SERVICE_INSECURE", "true") == "true"

	log.Printf("Connecting to data service at %s (insecure: %v)", dataSvcAddr, isInsecure)

	var opts []grpc.DialOption
	if isInsecure {
		// Use insecure for local development (no TLS)
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// TODO: add secure credentials
		log.Println("Using default secure credentials")
	}

	conn, err := grpc.NewClient(dataSvcAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to data service: %v", err)
	}
	defer conn.Close()

	dataClient := pb.NewTestServiceClient(conn)

	s := &apiServer{
		dataClient: dataClient,
	}
	// --- End gRPC Client Setup ---

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", health.Handler)
	// Add the new route to test the gRPC connection
	r.Get("/ping-data", s.handlePingData)

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Printf("api service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	shutdown.WaitForShutdown(srv, 5*time.Second)
}

func (s *apiServer) handlePingData(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	log.Println("Sending gRPC Ping to data service...")

	res, err := s.dataClient.Ping(ctx, &pb.PingRequest{})

	if err != nil {
		log.Printf("gRPC Ping failed: %v", err)
		http.Error(w, "Failed to ping data service", http.StatusInternalServerError)
		return
	}

	log.Printf("gRPC Ping successful: %s", res.String())

	json.WriteJSON(w, http.StatusOK, map[string]string{
		"status":                "success",
		"data_service_response": res.String(),
	})
}
