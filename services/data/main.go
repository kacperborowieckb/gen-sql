package main

import (
	"context"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// dataServer implements the gRPC TestServiceServer interface.
// It holds dependencies like the database pool.
type dataServer struct {
	pb.UnimplementedTestServiceServer
	dbPool *sql.DB
}

func (s *dataServer) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	log.Println("Received Ping request")

	if err := s.dbPool.PingContext(ctx); err != nil {
		log.Printf("Failed to ping database: %v", err)
		return nil, status.Errorf(codes.Internal, "database ping failed: %v", err)
	}

	log.Println("Database ping successful")

	return &pb.PingResponse{}, nil
}

func NewDataServer(dbPool *sql.DB) *dataServer {
	return &dataServer{
		dbPool: dbPool,
	}
}

func main() {
	port := env.GetString("PORT", "8081")

	dbConfig, err := db.DBConfig()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	dbPool, err := db.NewConnection(dbConfig)
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

	s := NewDataServer(dbPool)

	pb.RegisterTestServiceServer(grpcServer, s)

	reflection.Register(grpcServer)

	log.Printf("gRPC data service listening on %s", lis.Addr())

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
