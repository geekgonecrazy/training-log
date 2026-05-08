// Package middleware contains HTTP middleware applied in front of the gateway mux.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/geekgonecrazy/training-log/controllers"
	"github.com/geekgonecrazy/training-log/core/auth"
	"github.com/geekgonecrazy/training-log/core/metrics"
)

// Auth reads the access cookie, validates the JWT, and stores the user_id in
// the request context. If no cookie or the JWT is invalid, the request still
// proceeds — the controller will reject unauthenticated calls.
//
// publicPathPrefixes are matched against r.URL.Path; when a request matches,
// the middleware skips authentication entirely (used for /v1/auth/*).
func Auth(secret []byte, publicPathPrefixes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			isPublic := false
			for _, p := range publicPathPrefixes {
				if strings.HasPrefix(path, p) {
					isPublic = true
					break
				}
			}

			c, err := r.Cookie(controllers.CookieAccessToken)
			switch {
			case err != nil || c.Value == "":
				if !isPublic {
					metrics.AuthTokenValidationsTotal.WithLabelValues("missing_token").Inc()
				}
			default:
				if claims, perr := auth.ParseAccessToken(secret, c.Value, time.Now()); perr == nil {
					r = r.WithContext(auth.ContextWithUserID(r.Context(), claims.Subject))
					metrics.AuthTokenValidationsTotal.WithLabelValues("success").Inc()
				} else {
					// Bad/expired access tokens silently fall through; the client
					// will get a 401 from /v1/auth/me or any protected RPC and can
					// then call /v1/auth/refresh.
					metrics.AuthTokenValidationsTotal.WithLabelValues("invalid_token").Inc()
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestID attaches a short random ID to the request context, also mirrors it
// onto the request and response headers so downstream middleware (logging) and
// the client both see it.
type ctxKeyRequestID struct{}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 6)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
			r.Header.Set("X-Request-ID", id)
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Recover catches panics, logs them with stack, and returns 500.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic", "err", rec, "stack", string(debug.Stack()))
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logging emits one log line per request with method, path, status, duration.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(lrw, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", lrw.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", r.Header.Get("X-Request-ID"),
			)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (l *loggingResponseWriter) WriteHeader(code int) {
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}
