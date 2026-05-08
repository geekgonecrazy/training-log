// Package metrics owns the Prometheus registry and the HTTP instrumentation
// used by the public router. The handler exposed here is mounted by the
// internal router on /metrics.
package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry        *prometheus.Registry
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	inFlight        prometheus.Gauge
}

// New builds a Metrics with its own registry (no global state) plus the
// standard Go runtime + process collectors.
func New() *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		registry: reg,
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total HTTP requests handled by the public router.",
			},
			[]string{"method", "route", "code"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "route"},
		),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being served.",
		}),
	}

	reg.MustRegister(
		m.requestsTotal,
		m.requestDuration,
		m.inFlight,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return m
}

// Handler returns the /metrics handler bound to this registry.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		Registry: m.registry,
	})
}

// Middleware records request count + latency. The route label is bucketed to
// the first two path segments so templated IDs (/v1/habits/{id}) don't blow
// up cardinality.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := routeLabel(r.URL.Path)
		start := time.Now()
		mrw := &metricsResponseWriter{ResponseWriter: w, status: http.StatusOK}

		m.inFlight.Inc()
		defer m.inFlight.Dec()

		next.ServeHTTP(mrw, r)

		elapsed := time.Since(start).Seconds()
		m.requestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(mrw.status)).Inc()
		m.requestDuration.WithLabelValues(r.Method, route).Observe(elapsed)
	})
}

// routeLabel reduces a path to at most its first two segments so high-cardinality
// path params (IDs, tokens) don't explode the metric series.
func routeLabel(p string) string {
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return "/"
	}
	parts := strings.SplitN(p, "/", 3)
	if len(parts) < 2 {
		return "/" + parts[0]
	}
	return "/" + parts[0] + "/" + parts[1]
}

type metricsResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (m *metricsResponseWriter) WriteHeader(code int) {
	if m.wroteHeader {
		return
	}
	m.status = code
	m.wroteHeader = true
	m.ResponseWriter.WriteHeader(code)
}
