package main

import (
	"context"
	"errors"
	"health-checker/internal/configs"
	"health-checker/internal/handlers"
	"health-checker/internal/services"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := configs.GetCheckerCfg()

	if cfg.DebugMode {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}

		handler := slog.NewTextHandler(os.Stdout, opts)
		slog.SetDefault(slog.New(handler))
		slog.Debug("debug mode enabled")
	}

	if cfg.Interval < 0 {
		slog.Error("invalid interval, it should be greater than 0", "interval", cfg.Interval)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	monitor := services.NewMonitor()
	monitor.Start(ctx, cfg)

	address := cfg.Address + ":" + cfg.Port

	mux := handlers.NewRouter(monitor, cfg)

	srv := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           mux,
	}
	slog.Info("server initialized at", "address", address)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	slog.Info("server is shutting down. Please wait...")
	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := srv.Shutdown(shutdownContext); err != nil {
		slog.Error("server shutdown error,", "error", err)
	}

	<-shutdownContext.Done()
	slog.Info("server stopped")
}
