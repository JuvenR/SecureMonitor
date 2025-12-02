package main

import (
	"log"

	"securemonitor/internal/api"
	"securemonitor/internal/config"
	"securemonitor/internal/monitor"
	"securemonitor/internal/storage"
)

//entrypoint for the SecureMonitor daemon process.
func main() {
	log.Println("securemonitor starting up")

	//load configuration from disk.
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	//preload any previously blocked IP addresses.
	storage.LoadBlockedFromFile(cfg.BlockedIPsFile)
	log.Printf("loaded blocked ip list from %s", cfg.BlockedIPsFile)

	//start the http API server in the background.
	log.Println("starting api server on :9000")
	api.StartServer(":9000")

	//enter the main monitoring loop (blocking call).
	log.Println("entering monitoring loop")
	monitor.RunLoop(cfg)
}
