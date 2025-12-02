// internal/api/handlers.go
package api

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"securemonitor/internal/firewall"
	"securemonitor/internal/monitor"
	"securemonitor/internal/storage"
)

//  is the aggregated payload returned by /api/dashboard.
type DashboardSnapshot struct {
	Status  map[string]string `json:"status"`
	Stats   map[string]int    `json:"stats"`
	Logs    []string          `json:"logs"`
	Alerts  []storage.Alert   `json:"alerts"`
	Blocked []string          `json:"blocked"`
}

// returns a simple status block for health checks and the dashboard.
func buildStatusSnapshot() map[string]string {
	return map[string]string{
		"status": "OK",
		"msg":    "SecureMonitor is running",
	}
}

// tries to resolve the real client IP.
func clientIPFromRequest(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ----------- JSON handlers -----------

func handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, buildStatusSnapshot())
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	logs := storage.GetLogs()
	writeJSON(w, http.StatusOK, logs)
}

func handleBlocked(w http.ResponseWriter, r *http.Request) {
	ips := storage.ListBlocked()
	writeJSON(w, http.StatusOK, ips)
}

func handleUnblock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, "missing ip parameter", http.StatusBadRequest)
		return
	}

	firewall.UnblockIP(ip)
	storage.RemoveBlocked(ip)
	storage.AddLog("[FIREWALL] unblocked via dashboard: " + ip)

	w.WriteHeader(http.StatusNoContent)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]int{
		"ssh":    monitor.GetSSHCount(),
		"ftp":    monitor.GetFTPCount(),
		"apache": monitor.GetApacheCount(),
	}
	writeJSON(w, http.StatusOK, stats)
}

func handleAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := storage.GetAlerts()
	writeJSON(w, http.StatusOK, alerts)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	snap := DashboardSnapshot{
		Status: buildStatusSnapshot(),
		Stats: map[string]int{
			"ssh":    monitor.GetSSHCount(),
			"ftp":    monitor.GetFTPCount(),
			"apache": monitor.GetApacheCount(),
		},
		Logs:    storage.GetLogs(),
		Alerts:  storage.GetAlerts(),
		Blocked: storage.ListBlocked(),
	}

	writeJSON(w, http.StatusOK, snap)
}

// ------------Simulation handler -----------

// handleSimulate allows injecting fake SSH/FTP/Apache events for testing.
// It returns JSON by default (POST or Accept: application/json) or a small HTML view for GET in a browser.
func handleSimulate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	kind := r.URL.Query().Get("kind") // ssh | ftp | apache
	if kind == "" {
		kind = "ssh"
	}

	n := 10 // default: 10 events
	if ns := r.URL.Query().Get("n"); ns != "" {
		if parsed, err := strconv.Atoi(ns); err == nil && parsed > 0 && parsed <= 1000 {
			n = parsed
		}
	}

	ip := clientIPFromRequest(r)
	now := time.Now()

	switch kind {
	case "ssh":
		monitor.AddSimulatedSSH(ip, n)
	case "ftp":
		monitor.AddSimulatedFTP(ip, n)
	case "apache":
		monitor.AddSimulatedApache(ip, n)
	default:
		http.Error(w, "invalid kind (use ssh|ftp|apache)", http.StatusBadRequest)
		return
	}

	storage.AddLog("[SIM] scheduled " + strconv.Itoa(n) + " " + kind + " events from " + ip)

	resp := map[string]interface{}{
		"ok":        true,
		"kind":      kind,
		"count":     n,
		"ip":        ip,
		"timestamp": now.Format(time.RFC3339),
	}

	accept := r.Header.Get("Accept")
	if r.Method == http.MethodPost || strings.Contains(accept, "application/json") {
		// JSON response (curl / frontend usage).
		writeJSON(w, http.StatusOK, resp)
		return
	}

	// compact HTML result for GET in a browser.
	renderSimulateHTML(w, kind, n, ip, now)
}
