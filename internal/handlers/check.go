package handlers

import (
	"fmt"
	"health-checker/internal/models"
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
	w.Header().Set("Content-Type", "text/html")

	cpuUsage := monitor.GetCPUUtilizationValue()
	memUsage := monitor.GetRAMUtilizationValue()
	netUsage := monitor.GetNetUtilizationValue()
	diskUsage := monitor.GetDiskUtilizationValue()

	memUsage.LoadZone = services.WarningZone
	if cpuUsage.LoadZone == services.DangerZone ||
		memUsage.LoadZone == services.DangerZone ||
		netUsage.LoadZone == services.DangerZone ||
		diskUsage.LoadZone == services.DangerZone {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_, err := fmt.Fprintf(w, "<table>")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Error("response recording error", "error", err)
		return
	}

	writeUtilization(w, "CPU", cpuUsage)
	writeUtilization(w, "RAM", memUsage)
	writeUtilization(w, "Network", netUsage)
	writeUtilization(w, "Disk", diskUsage)

	_, err = fmt.Fprintf(w, "/<table>")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Error("response recording error", "error", err)
		return
	}
}

func writeUtilization(w http.ResponseWriter, name string, usage *models.Utilization) {
	var color string
	var message string

	switch usage.LoadZone {
	case services.WarningZone:
		color = "orange"
		message = "Warning: High utilization"
	case services.DangerZone:
		color = "red"
		message = "Danger: Critical utilization"
	default:
		color = "green"
	}

	_, err := fmt.Fprintf(w, "<tr><td>%s</td><td style='color: %s'>%s</td><td>%s</td></tr>", name, color, usage.Value, message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Error("response recording error", "error", err)
		return
	}
}
