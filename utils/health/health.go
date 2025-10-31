package health

import (
	"log"
	"net/http"

	"github.com/kacperborowieckb/gen-sql/utils/json"
)

type HealthStatus struct {
	Status string `json:"status"`
}

// Handler is a generic, simple health check handler.
// It reports "ok" if the server is running.
func Handler(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{Status: "ok"}

	err := json.WriteJSON(w, http.StatusOK, status)
	if err != nil {
		log.Printf("Error writing health check response: %v", err)
	}
}
