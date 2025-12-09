package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"github.com/kacperborowieckb/gen-sql/shared/messaging"
	"github.com/kacperborowieckb/gen-sql/utils/env"
	"github.com/kacperborowieckb/gen-sql/utils/health"
	"github.com/kacperborowieckb/gen-sql/utils/shutdown"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type apiServer struct {
	dataClient  pb.DataServiceClient
	queryClient pb.QueryServiceClient
	mqClient    *messaging.RabbitMQ
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

	dataConn, err := grpc.NewClient(dataServiceAddress, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to data service: %v", err)
	}
	defer dataConn.Close()

	dataClient := pb.NewDataServiceClient(dataConn)

	queryServiceAddress := env.GetString("QUERY_SERVICE_ADDR", "localhost:8083")
	log.Printf("Connecting to query service at %s", queryServiceAddress)

	queryConn, err := grpc.NewClient(queryServiceAddress, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to query service: %v", err)
	}
	defer queryConn.Close()

	queryClient := pb.NewQueryServiceClient(queryConn)

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
		dataClient:  dataClient,
		queryClient: queryClient,
		mqClient:    mqClient,
	}
	// --- End gRPC Client Setup ---

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(httprate.Limit(
		1,
		3*time.Second,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	))

	r.Get("/health", health.Handler)
	r.Route("/projects", func(r chi.Router) {
		r.Get("/", s.handleGetProjects)
		r.Post("/", s.handleStartDataGeneration)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.handleGetProjectData)
			r.Delete("/", s.handleDeleteProject)
			r.Post("/query", s.handleQuery)
		})
	})

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Printf("api service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	shutdown.WaitForShutdown(srv, 5*time.Second)
}
