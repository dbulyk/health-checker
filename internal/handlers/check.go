package handlers

import (
	"fmt"
	"health-checker/internal/configs"
	"health-checker/internal/services"
	"log/slog"
	"net/http"
	"strconv"
)

var (
	monitor *services.Monitor
	cfg     configs.Checker
)

func NewRouter(m *services.Monitor, c configs.Checker) *http.ServeMux {
	monitor = m
	cfg = c

	mux := http.NewServeMux()
	mux.HandleFunc("/check", checkUtilization)
	return mux
}

func checkUtilization(w http.ResponseWriter, _ *http.Request) {
	cpuUsage := monitor.GetCPUUtilizationValue()
	memoryUsage := monitor.GetRAMUtilizationValue()

	slog.Debug("utilization", "cpu", strconv.FormatFloat(cpuUsage, 'f', 2, 64),
		"memory", strconv.FormatFloat(memoryUsage, 'f', 2, 64))

	if cpuUsage > cfg.Threshold || memoryUsage > cfg.Threshold {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := fmt.Fprintf(w, "CPU load: %.2f%%\nRAM load: %.2f%%", cpuUsage, memoryUsage)
		if err != nil {
			slog.Error("error writing response", "error", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "CPU load: %.2f%%\nRAM load: %.2f%%", cpuUsage, memoryUsage)
	if err != nil {
		slog.Error("error writing response", "error", err)
	}
}
