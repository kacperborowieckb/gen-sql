package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kacperborowieckb/gen-sql/shared/messaging"
	"github.com/kacperborowieckb/gen-sql/utils/db"
	"github.com/kacperborowieckb/gen-sql/utils/env"
)

type generatorServer struct {
	dbPool   *sql.DB
	mqClient *messaging.RabbitMQ
}

func NewGeneratorServer(dbPool *sql.DB, mqClient *messaging.RabbitMQ) *generatorServer {
	return &generatorServer{
		dbPool:   dbPool,
		mqClient: mqClient,
	}
}

func main() {
	log.Println("Starting generator service...")

	// --- Database Setup ---
	dbConfig, err := db.DBConfig()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	dbPool, err := db.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

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

	// --- Create Server Instance ---
	s := NewGeneratorServer(dbPool, mqClient)

	// --- Start Consuming Messages ---
	log.Println("Starting consumer for queue:", messaging.DataGenerationQueue)
	// refactor to have some consumer struct and handle based on routing key
	if err := mqClient.ConsumeMessages(messaging.DataGenerationQueue, s.handleProjectCreated); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	// --- Graceful Shutdown ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down generator service...")
	log.Println("Generator service stopped.")
}
