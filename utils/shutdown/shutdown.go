package shutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server interface {
	Shutdown(ctx context.Context) error
}

func WaitForShutdown(srv Server, timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Shutting down server... Received signal: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped.")
}
