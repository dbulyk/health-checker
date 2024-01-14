package handlers

import (
	"fmt"
	"health-checker/internal/configs"
	"health-checker/internal/services"
	"log/slog"
	"net/http"
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

	if cpuUsage > cfg.Threshold || memoryUsage > cfg.Threshold {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_, err := fmt.Fprintf(w, "Утилизация процессора: %.2f%%\nУтилизация памяти: %.2f%%", cpuUsage, memoryUsage)
	if err != nil {
		slog.Error("ошибка записи ответа", "ошибка", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
