package handlers

import (
	"health-checker/internal/configs"
	"health-checker/internal/services"
	"log/slog"
	"net/http"
)

var (
	monitor *services.Monitor
)

func NewRouter(m *services.Monitor, c configs.Checker) *http.ServeMux {
	monitor = m

	mux := http.NewServeMux()
	mux.HandleFunc("/check", checkUtilization)
	return mux
}

func checkUtilization(w http.ResponseWriter, _ *http.Request) {
	cpuUsage := monitor.GetCPUUtilizationValue()
	if cpuUsage.HighLoad {
		slog.Debug("cpu utilization is high", "value", cpuUsage.Value)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	//memoryUsage := monitor.GetRAMUtilizationValue()

	//if cpuUsage > cfg.Threshold || memoryUsage > cfg.Threshold {
	//	w.WriteHeader(http.StatusServiceUnavailable)
	//} else {
	//	w.WriteHeader(http.StatusOK)
	//}

	//_, err := fmt.Fprintf(w, "cpu utilization: %.2f%%\n", cpuUsage)
	//if err != nil {
	//	slog.Error("response recording error", "error", err)
	//	http.Error(w, "response recording error", http.StatusInternalServerError)
	//	return
	//}
}
