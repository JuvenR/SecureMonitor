// internal/api/server.go
package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// boots the HTTP API on the provided address.
func StartServer(addr string) {
	mux := http.NewServeMux()

	// registers all routes
	registerRoutes(mux)

	go func() {
		log.Printf("api listening on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("api server error: %v", err)
		}
	}()
}

// sends a json response with a given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("api json encode error: %v", err)
	}
}
