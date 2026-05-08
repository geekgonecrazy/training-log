// Package router builds the top-level http.Handler.
//
// Wire diagram:
//
//	  HTTP request
//	       │
//	       ▼
//	┌─────────────────┐
//	│  middleware     │  RequestID, Logging, Recover, Auth
//	│  stack          │
//	└─────────────────┘
//	       │
//	       ▼  (path == /v1/* )      (else)
//	┌─────────────┐         ┌─────────────────────┐
//	│ gateway mux │         │ embedded SPA static │
//	│ (in-process │         │ (web/build/...)     │
//	│  to gRPC)   │         └─────────────────────┘
//	└─────────────┘
package router

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/geekgonecrazy/training-log/config"
	"github.com/geekgonecrazy/training-log/controllers"
	habitv1 "github.com/geekgonecrazy/training-log/models/habit/v1"
	"github.com/geekgonecrazy/training-log/router/middleware"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
)

type Deps struct {
	Cfg        config.Config
	Logger     *slog.Logger
	Auth       *controllers.AuthController
	Habit      *controllers.HabitController
	StaticRoot fs.FS // expected to point at the built SvelteKit app's root (with index.html at root)
}

// New builds the http.Handler.
func New(ctx context.Context, d Deps) (http.Handler, error) {
	gwMux := runtime.NewServeMux(
		runtime.WithMetadata(extractAuthMetadata),
		runtime.WithOutgoingHeaderMatcher(outgoingHeaderMatcher),
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
	)

	if err := habitv1.RegisterAuthServiceHandlerServer(ctx, gwMux, d.Auth); err != nil {
		return nil, err
	}
	if err := habitv1.RegisterHabitServiceHandlerServer(ctx, gwMux, d.Habit); err != nil {
		return nil, err
	}

	root := http.NewServeMux()
	root.Handle("/v1/", gwMux)
	root.Handle("/", spaHandler(d.StaticRoot, d.Logger))

	stack := chain(
		middleware.RequestID,
		middleware.Metrics,
		middleware.Logging(d.Logger),
		middleware.Recover(d.Logger),
		middleware.Auth([]byte(d.Cfg.Auth.JWTSecret), []string{"/v1/auth/"}),
	)

	return stack(root), nil
}

func chain(mws ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		// Apply in reverse so the first listed runs outermost.
		for i := len(mws) - 1; i >= 0; i-- {
			h = mws[i](h)
		}
		return h
	}
}

// extractAuthMetadata copies cookies + user_id (if previously authed by middleware)
// + user-agent into the gRPC incoming metadata for the in-process handler.
func extractAuthMetadata(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}

	// User ID injected by Auth middleware → into context.
	if uid, ok := userIDFromCtx(ctx); ok {
		md.Set("auth-user-id", itoa(uid))
	}
	// Cookies → into metadata (refresh endpoint reads these).
	if c, err := req.Cookie(controllers.CookieAccessToken); err == nil {
		md.Set(controllers.MDAccessTokenKey, c.Value)
	}
	if c, err := req.Cookie(controllers.CookieRefreshToken); err == nil {
		md.Set(controllers.MDRefreshTokenKey, c.Value)
	}
	if ua := req.UserAgent(); ua != "" {
		md.Set(controllers.MDUserAgentKey, ua)
	}
	return md
}

// outgoingHeaderMatcher allows Set-Cookie to pass straight through to the HTTP response.
func outgoingHeaderMatcher(s string) (string, bool) {
	if strings.EqualFold(s, "set-cookie") {
		return "Set-Cookie", true
	}
	return "", false
}

// userIDFromCtx reads the user_id auth context value, if present.
func userIDFromCtx(ctx context.Context) (int64, bool) {
	// Late binding: avoid pulling core/auth ctx key into this file. We import via
	// a small helper to keep the dep direction clean.
	return ctxUserID(ctx)
}

// spaHandler serves the SvelteKit static build from staticFS, falling back
// to index.html for any path not found (so client-side routes work on hard reload).
func spaHandler(staticFS fs.FS, logger *slog.Logger) http.Handler {
	if staticFS == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "frontend not built; run: cd web && npm run build", http.StatusServiceUnavailable)
		})
	}
	fileServer := http.FileServer(http.FS(staticFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(r.URL.Path, "/")
		if clean == "" {
			clean = "index.html"
		}
		if _, err := fs.Stat(staticFS, clean); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				logger.Warn("static stat", "path", clean, "err", err)
			}
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
