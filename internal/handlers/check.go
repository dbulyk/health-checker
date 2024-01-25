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

	if cpuUsage.LoadZone == services.DangerZone ||
		memUsage.LoadZone == services.DangerZone ||
		netUsage.LoadZone == services.DangerZone ||
		diskUsage.LoadZone == services.DangerZone {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	html := "<html><head><title>Health Checker</title></head><body><h1>Health Checker</h1><table>"
	html = writeUtilization(html, "CPU", cpuUsage)
	html = writeUtilization(html, "RAM", memUsage)
	html = writeUtilization(html, "Network", netUsage)
	html = writeUtilization(html, "Disk", diskUsage)

	html += "</table>"

	_, err := fmt.Fprint(w, html)
	if err != nil {
		slog.Error("error while writing response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeUtilization(html string, name string, usage *models.Utilization) string {
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

	html += fmt.Sprintf("<tr><td>%s</td><td style='color: %s'>%s</td><td>%s</td></tr>", name, color, usage.Value, message)
	return html
}
