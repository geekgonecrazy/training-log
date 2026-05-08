// Package metrics declares the Prometheus collectors used across the app.
//
// Pattern follows flockledger: package-level promauto vars on the default
// registry, called directly from instrumentation sites with `metrics.X.Inc()`.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"net/http"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tracker_http_requests_total",
			Help: "Total HTTP requests handled by the public router.",
		},
		[]string{"method", "route", "code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tracker_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	HTTPRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tracker_http_requests_in_flight",
		Help: "Number of HTTP requests currently being served.",
	})

	DBOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tracker_db_operations_total",
			Help: "Total database operations by type and table.",
		},
		[]string{"operation", "table"},
	)

	DBOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tracker_db_operation_duration_seconds",
			Help:    "Database operation duration in seconds, labeled by the calling method.",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"operation", "method"},
	)

	AuthTokenValidationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tracker_auth_token_validations_total",
			Help: "Auth token validation attempts by outcome.",
		},
		[]string{"status"},
	)

	BuildInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tracker_build_info",
			Help: "Build information.",
		},
		[]string{"version", "commit"},
	)
)

func init() {
	// Default registry skips runtime/process collectors; register them so
	// /metrics exposes goroutines, GC, RSS, etc.
	prometheus.DefaultRegisterer.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
}

// Init records build info, called once at startup.
func Init(version, commit string) {
	BuildInfo.WithLabelValues(version, commit).Set(1)
}

// Handler returns the /metrics HTTP handler bound to the default registry.
func Handler() http.Handler { return promhttp.Handler() }
