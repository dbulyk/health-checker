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
	"runtime"
	"syscall"
	"time"
)

func main() {
	if runtime.GOOS != "windows" {
		slog.Info("поддерживается только Windows")
		os.Exit(1)
	}

	cfg := configs.GetCheckerCfg()

	if cfg.DebugMode {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}

		handler := slog.NewTextHandler(os.Stdout, opts)
		slog.SetDefault(slog.New(handler))
		slog.Debug("сервер запущен в режиме отладки")
	}

	if cfg.Interval < 0 {
		slog.Error("неккоректный интервал, укажите >= 0", "интервал", cfg.Interval)
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
	slog.Info("сервер запущен", "адрес", address)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ошибка: %v", err)
		}
	}()

	<-ctx.Done()
	slog.Info("сервер останавливается. Пожалуйста, подождите...")
	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := srv.Shutdown(shutdownContext); err != nil {
		slog.Error("ошибка остановки сервера,", "ошибка", err)
	}

	<-shutdownContext.Done()
	slog.Info("сервер остановлен")
}
