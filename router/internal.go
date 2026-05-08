package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger is satisfied by anything with a context-aware liveness check; the
// sqlite Store implements it.
type Pinger interface {
	Ping(ctx context.Context) error
}

// InternalDeps wires the internal listener — health probes and the Prometheus
// scrape endpoint. This handler is intended for a private listener, not the
// public address.
type InternalDeps struct {
	DB             Pinger
	MetricsHandler http.Handler
}

// NewInternal returns the http.Handler for the internal listener.
//
//	GET /health         — liveness, always 200 if the process is up
//	GET /health/ready   — readiness, 200 only if the DB ping succeeds
//	GET /metrics        — Prometheus exposition
func NewInternal(d InternalDeps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if d.DB != nil {
			if err := d.DB.Ping(ctx); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{
					"status": "unready",
					"error":  err.Error(),
				})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	if d.MetricsHandler != nil {
		mux.Handle("/metrics", d.MetricsHandler)
	}

	return mux
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
