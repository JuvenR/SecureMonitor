// internal/api/routes.go
package api

import "net/http"

func registerRoutes(mux *http.ServeMux) {
	// API routes.
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/logs", handleLogs)
	mux.HandleFunc("/api/blocked", handleBlocked)
	mux.HandleFunc("/api/unblock", handleUnblock)
	mux.HandleFunc("/api/stats", handleStats)
	mux.HandleFunc("/api/alerts", handleAlerts)
	mux.HandleFunc("/api/dashboard", handleDashboard)

	// simulation endpoint for demo/testing.
	mux.HandleFunc("/api/simulate", handleSimulate)

	// static dashboard assets.
	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/", fs)
}
