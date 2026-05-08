package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/geekgonecrazy/training-log/core/metrics"
)

// Metrics records request count, duration, and in-flight gauge for every
// public request. The route label is bucketed to the first two path segments
// so high-cardinality path params (IDs, tokens) don't blow up the series.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := routeLabel(r.URL.Path)
		start := time.Now()
		mrw := &metricsResponseWriter{ResponseWriter: w, status: http.StatusOK}

		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		next.ServeHTTP(mrw, r)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(mrw.status)).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
	})
}

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
