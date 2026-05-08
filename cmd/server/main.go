package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/geekgonecrazy/training-log/config"
	"github.com/geekgonecrazy/training-log/controllers"
	"github.com/geekgonecrazy/training-log/core/metrics"
	"github.com/geekgonecrazy/training-log/router"
	"github.com/geekgonecrazy/training-log/store/sqlite"
	"github.com/geekgonecrazy/training-log/webfs"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config.yaml")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.Log)

	st, err := sqlite.Open(cfg.Database.Path)
	if err != nil {
		logger.Error("open db", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	cookieOpts := controllers.CookieOptions{
		Domain: cfg.Auth.CookieDomain,
		Secure: cfg.Auth.CookieSecure,
	}
	authCtrl := controllers.NewAuthController(st, cfg.Auth, cookieOpts)
	habitCtrl := controllers.NewHabitController(st)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	handler, err := router.New(ctx, router.Deps{
		Cfg:        cfg,
		Logger:     logger,
		Auth:       authCtrl,
		Habit:      habitCtrl,
		StaticRoot: webfs.FS(),
	})
	if err != nil {
		logger.Error("build router", "err", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	logger.Info("listening", "address", cfg.Server.Address)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			cancel()
		}
	}()

	// Internal listener: /health, /health/ready, /metrics. Bind to a separate
	// address so it can be firewalled off from external traffic. An empty
	// internal_address disables the listener entirely.
	var internalSrv *http.Server
	if cfg.Server.InternalAddress != "" {
		internalSrv = &http.Server{
			Addr: cfg.Server.InternalAddress,
			Handler: router.NewInternal(router.InternalDeps{
				DB:             st,
				MetricsHandler: metrics.Handler(),
			}),
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		}
		logger.Info("internal listening", "address", cfg.Server.InternalAddress)
		go func() {
			if err := internalSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("internal server error", "err", err)
				cancel()
			}
		}()
	}

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
	if internalSrv != nil {
		_ = internalSrv.Shutdown(shutdownCtx)
	}
}

func newLogger(c config.LogConfig) *slog.Logger {
	level := slog.LevelInfo
	switch c.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: level}
	if c.Format == "json" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}
