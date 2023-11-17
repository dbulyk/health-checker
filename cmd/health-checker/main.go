package main

import (
	"context"
	"errors"
	"fmt"
	"health-checker/config"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	lastCPULoad      float64
	totalMemoryUsage float64
	cpuLoadLock      sync.Mutex
	ramLoadLock      sync.Mutex
	cfg              config.Checker
	pollCount        atomic.Int64
)

func main() {
	cfg = config.GetCheckerCfg()

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

	go func() {
		err := updateCPULoad(ctx, cfg.Interval)
		if err != nil {
			slog.Error("cpu load error", "error", err)
			stop()
		}
	}()

	go func() {
		err := updateMemoryLoad(ctx, cfg.Interval)
		if err != nil {
			slog.Error("memory load error", "error", err)
			stop()
		}
	}()
	address := cfg.Address + ":" + cfg.Port

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           mux,
	}
	mux.HandleFunc("/check", checkCPUAndRAMLoad)
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

func updateCPULoad(ctx context.Context, interval time.Duration) error {
	var (
		percentages []float64
		err         error
	)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			percentages, err = cpu.PercentWithContext(ctx, interval, false)
			if err != nil && errors.Is(err, context.Canceled) {
				return nil
			} else if err != nil {
				return err
			}

			cpuLoadLock.Lock()
			lastCPULoad = percentages[0]
			cpuLoadLock.Unlock()

			slog.Debug("cpu load", "load", strconv.FormatFloat(lastCPULoad, 'f', 2, 64))
		case <-ctx.Done():
			slog.Debug("cpu load update stopped")
			return nil
		}
	}
}

func updateMemoryLoad(ctx context.Context, interval time.Duration) error {
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	memoryUsage := memoryInfo.UsedPercent

	pollCount.Store(1)
	ramLoadLock.Lock()
	totalMemoryUsage += memoryUsage
	ramLoadLock.Unlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			memoryInfo, err = mem.VirtualMemoryWithContext(ctx)
			if err != nil {
				return err
			}
			memoryUsage = memoryInfo.UsedPercent

			ramLoadLock.Lock()
			if pollCount.Load() > 5 {
				totalMemoryUsage = memoryUsage
				pollCount.Swap(1)
			} else {
				totalMemoryUsage += memoryUsage
				pollCount.Add(1)
			}
			ramLoadLock.Unlock()
			slog.Debug("memory utilization", "utilization", strconv.FormatFloat(memoryUsage, 'f', 2, 64))
		case <-ctx.Done():
			slog.Debug("memory utilization update stopped")
			return nil
		}
	}
}

func checkCPUAndRAMLoad(w http.ResponseWriter, _ *http.Request) {
	cpuLoadLock.Lock()
	cpuLoad := lastCPULoad
	cpuLoadLock.Unlock()

	ramLoadLock.Lock()
	memoryUsage := totalMemoryUsage / float64(pollCount.Load())
	ramLoadLock.Unlock()

	slog.Debug("utilization", "cpu", strconv.FormatFloat(cpuLoad, 'f', 2, 64),
		"memory", strconv.FormatFloat(totalMemoryUsage, 'f', 2, 64))

	if cpuLoad > cfg.Threshold || memoryUsage > cfg.Threshold {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := fmt.Fprintf(w, "CPU load: %.2f%%\nRAM load: %.2f%%", cpuLoad, memoryUsage)
		if err != nil {
			slog.Error("error writing response", "error", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "CPU load: %.2f%%\nRAM load: %.2f%%", cpuLoad, memoryUsage)
	if err != nil {
		slog.Error("error writing response", "error", err)
	}
}
