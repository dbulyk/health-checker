package handlers

import (
	"health-checker/internal/services"
	"log/slog"
	"net/http"
)

var (
	monitor *services.Monitor
)

func NewRouter(m *services.Monitor) *http.ServeMux {
	monitor = m

	mux := http.NewServeMux()
	mux.HandleFunc("/check", checkUtilization)
	return mux
}

func checkUtilization(w http.ResponseWriter, _ *http.Request) {
	cpuUsage := monitor.GetCPUUtilizationValue()
	switch cpuUsage.LoadZone {
	case services.WarningZone:
		_, err := w.Write([]byte("CPU utilization exceeds 75%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
		}
	case services.DangerZone:
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := w.Write([]byte("CPU utilization exceeds 90%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
			return
		}
	}

	memUsage := monitor.GetRAMUtilizationValue()
	switch memUsage.LoadZone {
	case services.WarningZone:
		_, err := w.Write([]byte("RAM utilization exceeds 75%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
		}
	case services.DangerZone:
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := w.Write([]byte("RAM utilization exceeds 90%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
			return
		}
	}

	netUsage := monitor.GetNetUtilizationValue()
	switch netUsage.LoadZone {
	case services.WarningZone:
		_, err := w.Write([]byte("Network utilization exceeds 75%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
		}
	case services.DangerZone:
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := w.Write([]byte("Network utilization exceeds 90%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
			return
		}
	}

	slog.Debug("System utilization:", "cpu", cpuUsage.Value, "RAM", memUsage.Value, "network", netUsage.Value)

	w.WriteHeader(http.StatusOK)
}
