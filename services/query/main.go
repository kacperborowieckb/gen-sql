package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/kacperborowieckb/gen-sql/shared/gen/proto"
	"github.com/kacperborowieckb/gen-sql/utils/db"
	"github.com/kacperborowieckb/gen-sql/utils/env"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type queryServer struct {
	pb.UnimplementedQueryServiceServer
	dbPool *sql.DB
}

func NewQueryServer(dbPool *sql.DB) *queryServer {
	return &queryServer{
		dbPool: dbPool,
	}
}

func main() {
	port := env.GetString("PORT", "8083")

	dbPool, err := db.NewConnection(env.GetString("DATABASE_URL", ""))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// --- gRPC Server Setup ---
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	s := NewQueryServer(dbPool)

	pb.RegisterQueryServiceServer(grpcServer, s)

	reflection.Register(grpcServer)

	log.Printf("gRPC query service listening on %s", lis.Addr())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// --- Graceful Shutdown ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down gRPC server...")

	grpcServer.GracefulStop()

	log.Println("gRPC server stopped")
}
