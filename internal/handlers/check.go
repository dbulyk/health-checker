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
	mux.HandleFunc("/check", checkUtilisation)
	return mux
}

func checkUtilisation(w http.ResponseWriter, _ *http.Request) {
	cpuUsage := monitor.GetCPUUtilisationValue()

	switch cpuUsage.LoadZone {
	case services.WarningZone:
		_, err := w.Write([]byte("CPU utilisation exceeds 75%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
			return
		}
	case services.DangerZone:
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := w.Write([]byte("CPU utilisation exceeds 90%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
			return
		}
	}

	memUsage := monitor.GetRAMUtilisationValue()
	switch memUsage.LoadZone {
	case services.WarningZone:
		_, err := w.Write([]byte("RAM utilisation exceeds 75%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
			return
		}
	case services.DangerZone:
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := w.Write([]byte("RAM utilisation exceeds 90%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
