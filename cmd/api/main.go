package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eyex-api/eyex/internal/config"
	"github.com/eyex-api/eyex/internal/handlers"
	"github.com/eyex-api/eyex/internal/middleware"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	api := handlers.New("frontend/html-js")
	handler := middleware.Chain(api.Routes(), middleware.Recover, middleware.Logging, middleware.CORS(cfg.AllowedOrigin))

	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("EyeX API listening", "address", cfg.Address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
