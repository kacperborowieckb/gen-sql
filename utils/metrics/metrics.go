package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status_code"},
	)

	HTTPRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status_code"},
	)

	HTTPRequestInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)
)

// Middleware returns a chi middleware that records Prometheus metrics
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment in-flight requests
		HTTPRequestInFlight.Inc()
		defer HTTPRequestInFlight.Dec()

		// Wrap response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Process request
		next.ServeHTTP(ww, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(ww.Status())
		route := getRoutePattern(r)

		HTTPRequestDuration.WithLabelValues(r.Method, route, statusCode).Observe(duration)
		HTTPRequestTotal.WithLabelValues(r.Method, route, statusCode).Inc()
	})
}

func getRoutePattern(r *http.Request) string {
	if routePattern := chi.RouteContext(r.Context()).RoutePattern(); routePattern != "" {
		return routePattern
	}
	return r.URL.Path
}
