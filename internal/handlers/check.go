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
	case "normal":
		w.WriteHeader(http.StatusOK)
	case "warning":
		_, err := w.Write([]byte("CPU utilisation exceeds 75%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
			return
		}
	case "danger":
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := w.Write([]byte("CPU utilisation exceeds 90%."))
		if err != nil {
			slog.Error("response recording error", "error", err)
			return
		}
	}

	//memoryUsage := monitor.GetRAMutilisationValue()

	//if cpuUsage > cfg.Threshold || memoryUsage > cfg.Threshold {
	//	w.WriteHeader(http.StatusServiceUnavailable)
	//} else {
	//	w.WriteHeader(http.StatusOK)
	//}

	//_, err := fmt.Fprintf(w, "cpu utilisation: %.2f%%\n", cpuUsage)
	//if err != nil {
	//	slog.Error("response recording error", "error", err)
	//	http.Error(w, "response recording error", http.StatusInternalServerError)
	//	return
	//}
}
