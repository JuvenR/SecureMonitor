package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"securemonitor/internal/config"
	"securemonitor/internal/firewall"
	"securemonitor/internal/storage"
)


// attempts to parse a client IP address from a log line
// uses known patterns (Apache client, rhost=, " from ").
func extractIP(line string) string {
	// Apache client pattern.
	if strings.Contains(line, "Client \"") {
		parts := strings.Split(line, "Client \"")
		if len(parts) > 1 {
			rest := parts[1]
			ip := strings.SplitN(rest, "\"", 2)[0]
			return strings.TrimSpace(ip)
		}
	}

	// SSH and FTP pattern.
	if strings.Contains(line, "rhost=") {
		parts := strings.Split(line, "rhost=")
		if len(parts) > 1 {
			ip := strings.Fields(parts[1])[0]
			return strings.TrimSpace(ip)
		}
	}

	// Generic " from " pattern.
	if strings.Contains(line, " from ") {
		parts := strings.Split(line, " from ")
		if len(parts) > 1 {
			ip := strings.Fields(parts[1])[0]
			return strings.TrimSpace(ip)
		}
	}

	return ""
}

// assigns a severity level based on new events,
// accumulated totals and a threshold.
func classifySeverity(service string, newEvents, total, threshold int) string {
	if threshold <= 0 {
		threshold = 1
	}

	// Strong spike in one cycle or reached the limit means HIGH.
	if newEvents >= threshold || total >= threshold {
		return "HIGH"
	}

	// Approaching the threshold means MEDIUM.
	if total*2 >= threshold {
		return "MEDIUM"
	}

	// Otherwise means LOW.
	return "LOW"
}


//  GEO IP (country) via ip-api.com
var (
	httpClient = &http.Client{Timeout: 2 * time.Second}
	geoCache   = make(map[string]string)
)

// models the subset of the response.
type ipAPIResponse struct {
	Status      string `json:"status"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
}

// lookupCountry resolves the country label for an IP:
// - local/private IPs -> "Local"
// - failures -> empty string.
func lookupCountry(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}

	parsed := net.ParseIP(ip)
	if parsed != nil && (parsed.IsLoopback() || parsed.IsPrivate()) {
		return "Local"
	}

	if c, ok := geoCache[ip]; ok {
		return c
	}

	url := "http://ip-api.com/json/" + ip + "?fields=status,country,countryCode"
	resp, err := httpClient.Get(url)
	if err != nil {
		log.Printf("geo: lookup failed for %s: %v", ip, err)
		geoCache[ip] = ""
		return ""
	}
	defer resp.Body.Close()

	var data ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("geo: decode failed for %s: %v", ip, err)
		geoCache[ip] = ""
		return ""
	}

	if data.Status != "success" {
		geoCache[ip] = ""
		return ""
	}

	label := data.Country
	if data.CountryCode != "" {
		label += " (" + data.CountryCode + ")"
	}
	geoCache[ip] = label
	return label
}


// Makes autoblock incremental.
//
// Base: cfg.AutoUnblockMinutes (ej. 5 min)
// Strikes:
//   1er strike  -> 1x base   (5 min)
//   2do strike  -> 3x base   (15 min)
//   3er strike  -> 5x base   (25 min)
//   4to strike  -> 7x base   (35 min)
//   ...
func autoUnblockExpired(cfg config.Config, now time.Time) {
	if cfg.AutoUnblockMinutes <= 0 {
		return
	}

	base := time.Duration(cfg.AutoUnblockMinutes) * time.Minute
	entries := storage.ListBlockedEntries()

	for _, e := range entries {
		strikes := e.Strikes
		if strikes <= 0 {
			strikes = 1
		}

		factor := 1 + (strikes-1)*2
		maxAge := base * time.Duration(factor)

		age := now.Sub(e.BlockedAt)
		if age >= maxAge {
			storage.AddLog(fmt.Sprintf(
				"[FW] Auto-unblock %s (age=%s, strikes=%d, maxAge≈%s)",
				e.IP,
				age.Truncate(time.Second),
				strikes,
				maxAge.Truncate(time.Second),
			))
			firewall.UnblockIP(e.IP)
			storage.RemoveBlocked(e.IP)
		}
	}
}

// effectiveThreshold resolves the threshold for a service using:
// 1) a specific value (if > 0),
// 2) a global fallback (if > 0),
// 3) a hardcoded default.
func effectiveThreshold(specific, global, fallback int) int {
	if specific > 0 {
		return specific
	}
	if global > 0 {
		return global
	}
	return fallback
}



// reads new SSH/FTP failures and injects simulated events,
// taking into account the case where both services share the same log file.
func readSSHAndFTP(cfg config.Config) (map[string]int, map[string]int) {
	var sshFails map[string]int
	var ftpFails map[string]int

	if cfg.SSHLogPath == cfg.FTPLogPath {
		lines, err := ReadNewLines(cfg.SSHLogPath)
		if err != nil {
			lines = []string{}
		}
		sshFails = parseSSHFailuresFromLines(lines)
		ftpFails = parseFTPFailuresFromLines(lines)
	} else {
		sshFails = parseSSHFailures(cfg.SSHLogPath)
		ftpFails = parseFTPFailures(cfg.FTPLogPath)
	}

	// Inject simulated events.
	for ip, c := range drainSimulatedSSH() {
		sshFails[ip] += c
	}
	for ip, c := range drainSimulatedFTP() {
		ftpFails[ip] += c
	}

	return sshFails, ftpFails
}

// reads new Apache errors and injects simulated events.
func readApache(cfg config.Config) map[string]int {
	apacheErrors := parseApacheErrors(cfg.ApacheAccessLogPath)
	for ip, c := range drainSimulatedApache() {
		apacheErrors[ip] += c
	}
	return apacheErrors
}

//-----------------IMPORTANT---------------

// RunLoop is the main monitoring loop that periodically scans logs,
// updates stats, generates alerts and enforces firewall blocks.
func RunLoop(cfg config.Config) {
	ticker := time.NewTicker(time.Duration(cfg.CheckIntervalSeconds) * time.Second)
	defer ticker.Stop()

	// Service strategies.
	sshStrategy := NewSSHStrategy()
	ftpStrategy := NewFTPStrategy()
	apacheStrategy := NewApacheStrategy()

	for {
		now := time.Now()

		// Reload whitelist each cycle (small file, cheap enough).
		whitelist := loadWhitelist(cfg.WhitelistFile)

		// Auto-unblock old IPs.
		autoUnblockExpired(cfg, now)

		storage.AddLog("[SCAN START] " + now.Format(time.RFC3339))

		// Read events for SSH/FTP and Apache.
		sshFails, ftpFails := readSSHAndFTP(cfg)
		apacheErrors := readApache(cfg)

		log.Printf(
			"monitor loop: ssh_ips=%d ftp_ips=%d apache_ips=%d",
			len(sshFails),
			len(ftpFails),
			len(apacheErrors),
		)

		// Delegate per-service logic to strategies.
		sshStrategy.ProcessEvents(sshFails, cfg, now, whitelist)
		ftpStrategy.ProcessEvents(ftpFails, cfg, now, whitelist)
		apacheStrategy.ProcessEvents(apacheErrors, cfg, now, whitelist)

		// Persist blocked snapshot to disk.
		storage.SaveBlockedToFile(cfg.BlockedIPsFile)

		<-ticker.C
	}
}
