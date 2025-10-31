package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kacperborowieckb/gen-sql/utils/env"
	"github.com/kacperborowieckb/gen-sql/utils/health"
	"github.com/kacperborowieckb/gen-sql/utils/shutdown"

	"github.com/go-chi/chi/v5"
)

func main() {
	port := env.GetString("PORT", "8082")
	// TODO: implement AMQP
	// amqpURL := env.GetString("AMQP_URL", "")
	// queueName := env.GetString("GENERATOR_QUEUE", "gensql.jobs")

	r := chi.NewRouter()

	r.Get("/health", health.Handler)

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Printf("generator service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	shutdown.WaitForShutdown(srv, 5*time.Second)
}
