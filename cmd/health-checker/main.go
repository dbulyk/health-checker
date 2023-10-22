package main

import (
	"context"
	"health-checker/config"
	"log/slog"
	"net/http"
	"os"
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
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	cfg = config.GetCheckerCfg()

	if cfg.Interval < 0 {
		slog.Error("invalid interval, it should be greater than 0", "interval", cfg.Interval)
		return
	}

	go func() {
		err := updateCPULoad(cfg.Interval, sigs)
		if err != nil {
			slog.Error("cpu load error", "error", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	address := cfg.Address + ":" + cfg.Port

	srv := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("server initialized at", "address", address)
	http.HandleFunc("/", checkCPUAndRAMLoad)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("server error,", "error", err)
		}
	}()

	<-sigs
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error,", "error", err)
	}
	slog.Info("server stopped")
}

func updateCPULoad(interval time.Duration, sigs chan os.Signal) error {
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
		case <-sigs:
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
