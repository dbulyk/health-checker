package main

import (
	"context"
	"errors"
	"fmt"
	"health-checker/config"
	"log"
	"log/slog"
	"net/http"
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
	var (
		percentages []float64
		percentage  float64
		err         error
	)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			percentages, err = cpu.Percent(interval, true)
			if err != nil {
				return err
			}

			cpuLoadLock.Lock()
			lastCPULoad = 0.0
			for _, percentage = range percentages {
				lastCPULoad += percentage
			}
			cpuLoadLock.Unlock()

			slog.Info("cpu", "load", strconv.FormatFloat(lastCPULoad, 'f', 2, 64))
		case <-ctx.Done():
			slog.Info("cpu load update stopped")
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

	slog.Info("ram", "load", memoryUsage)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			memoryInfo, err = mem.VirtualMemory()
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

			//slog.Info("ram", "load", memoryUsage)
		case <-ctx.Done():
			slog.Info("memory utilization update stopped")
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

	slog.Info("utilization", "cpu", cpuLoad, "memory", memoryUsage)

	if cpuLoad > cfg.Threshold || memoryUsage > cfg.Threshold {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "CPU load: %.2f%%\nRAM load: %.2f%%", cpuLoad, memoryUsage)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "CPU load: %.2f%%\nRAM load: %.2f%%", cpuLoad, memoryUsage)
}
