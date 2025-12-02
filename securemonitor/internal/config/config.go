package config

import (
	"encoding/json"
	"os"
)

// holds runtime configuration loaded from a JSON file.
type Config struct {
	ServicesToWatch      []string `json:"services_to_watch"`
	SSHLogPath           string   `json:"ssh_log_path"`
	ApacheAccessLogPath  string   `json:"apache_access_log_path"`
	ApacheErrorLogPath   string   `json:"apache_error_log_path"`
	FTPLogPath           string   `json:"ftp_log_path"`

	MaxFailures          int `json:"max_failures"` // global fallback

	SSHMaxFailures       int `json:"ssh_max_failures"`
	FTPMaxFailures       int `json:"ftp_max_failures"`
	ApacheErrorThreshold int `json:"apache_error_threshold"`

	CheckIntervalSeconds int    `json:"check_interval_seconds"`
	BlockedIPsFile       string `json:"blocked_ips_file"`
	WhitelistFile        string `json:"whitelist_file"`

	AutoUnblockMinutes     int  `json:"auto_unblock_minutes"`      // 0 = disabled
	ApacheBlockOnThreshold bool `json:"apache_block_on_threshold"` // true = Apache also blocks
}

// reads configuration from the JSON file path.
func Load(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
