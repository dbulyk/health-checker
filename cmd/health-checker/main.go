package main

import (
	"context"
	"errors"
	"health-checker/config"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

var (
	lastCPULoad float64
	loadLock    sync.Mutex
	cfg         config.Checker
)

func main() {
	cfg = config.GetCheckerCfg()

	if cfg.Interval < 0 {
		slog.Error("invalid interval, it should be greater than 0", "interval", cfg.Interval)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		err := updateCPULoad(ctx, cfg.Interval)
		if err != nil {
			slog.Error("cpu load error", "error", err)
		}
	}()

	address := cfg.Address + ":" + cfg.Port

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           mux,
	}
	mux.HandleFunc("/", checkCPUAndRAMLoad)
	slog.Info("server initialized at", "address", address)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownContext); err != nil {
		slog.Error("server shutdown error,", "error", err)
	}

	<-shutdownContext.Done()
	slog.Info("server stopped")
}

func updateCPULoad(ctx context.Context, interval time.Duration) error {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		slog.Error("cpu load error", "error", err)
		return err
	}

	loadLock.Lock()
	lastCPULoad = percentages[0]
	loadLock.Unlock()
	slog.Info("cpu load updated", "load", lastCPULoad)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			percentages, err = cpu.Percent(interval, false)
			if err != nil {
				slog.Error("cpu load error", "error", err)
				time.Sleep(interval)
				return err
			}

			loadLock.Lock()
			lastCPULoad = percentages[0]
			loadLock.Unlock()
			slog.Info("cpu load updated", "load", lastCPULoad)
		case <-ctx.Done():
			slog.Info("cpu load update stopped")
			return nil
		}
	}
}

func checkCPUAndRAMLoad(w http.ResponseWriter, _ *http.Request) {
	loadLock.Lock()
	cpuLoad := lastCPULoad
	loadLock.Unlock()

	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		slog.Error("memory info error", "error", err)
	}
	memoryUsage := memoryInfo.UsedPercent

	slog.Info("utilization", "cpu", cpuLoad, "memory", memoryUsage)

	if cpuLoad > cfg.Threshold || memoryUsage > cfg.Threshold {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
