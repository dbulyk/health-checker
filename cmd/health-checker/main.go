package main

import (
	"context"
	"healthChecker/config"
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
	cfg         = config.GetCheckerCfg()
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go updateCPULoad()

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

func updateCPULoad() {
	for {
		percentages, err := cpu.Percent(cfg.Interval, false)
		if err != nil {
			slog.Error("cpu load error", "error", err)
			time.Sleep(cfg.Interval)
			continue
		}

		loadLock.Lock()
		lastCPULoad = percentages[0]
		loadLock.Unlock()
		slog.Info("cpu load updated", "load", lastCPULoad)
	}
}

func checkCPUAndRAMLoad(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "CPU or RAM utilization above the threshold", http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
