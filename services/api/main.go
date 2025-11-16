package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"github.com/kacperborowieckb/gen-sql/shared/messaging"
	"github.com/kacperborowieckb/gen-sql/utils/env"
	"github.com/kacperborowieckb/gen-sql/utils/health"
	"github.com/kacperborowieckb/gen-sql/utils/shutdown"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type apiServer struct {
	dataClient pb.DataServiceClient
	mqClient   *messaging.RabbitMQ
}

func main() {
	port := env.GetString("PORT", "8080")

	// --- gRPC Client Setup ---
	dataServiceAddress := env.GetString("DATA_SERVICE_ADDR", "localhost:8081")
	isInsecure := env.GetString("DATA_SERVICE_INSECURE", "true") == "true"

	log.Printf("Connecting to data service at %s (insecure: %v)", dataServiceAddress, isInsecure)

	var opts []grpc.DialOption
	if isInsecure {
		// Use insecure for local development (no TLS)
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// TODO: add secure credentials
		log.Println("Using default secure credentials")
	}

	conn, err := grpc.NewClient(dataServiceAddress, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to data service: %v", err)
	}
	defer conn.Close()

	dataClient := pb.NewDataServiceClient(conn)

	// --- RabbitMQ Client Setup ---
	rabbitMQURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	mqClient, err := messaging.NewRabbitMQ(rabbitMQURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	defer mqClient.Close()

	if err := mqClient.SetupAppTopology(); err != nil {
		log.Fatalf("Failed to setup RabbitMQ topology: %v", err)
	}
	// --- End RabbitMQ Client Setup ---

	s := &apiServer{
		dataClient: dataClient,
		mqClient:   mqClient,
	}
	// --- End gRPC Client Setup ---

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", health.Handler)
	r.Post("/projects", s.handleStartDataGeneration)
	r.Get("/projects/{id}", s.HandleGetProjectData)

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Printf("api service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	shutdown.WaitForShutdown(srv, 5*time.Second)
}
