package monitor

import (
	"fmt"
	"strings"
	"time"

	"securemonitor/internal/config"
	"securemonitor/internal/firewall"
	"securemonitor/internal/storage"
)

// encapsulates service-specific logic: stats, alerts, blocking.
type ServiceStrategy interface {
	// Name returns the identifier of the service (ssh, ftp, apache, ...).
	Name() string

	// handles the events detected in the current scan cycle.
	// events: ip -> count of new events in this cycle.
	ProcessEvents(events map[string]int, cfg config.Config, now time.Time, whitelist map[string]struct{})
}

//  builds a log prefix like [SSH], [FTP], [APACHE] from a service name.
func logPrefix(service string) string {
	return "[" + strings.ToUpper(service) + "]"
}


type loginServiceConfig struct {
	name          string                      // "ssh" or "ftp"
	defaultThresh int                         // fallback if no cfg values
	getSpecific   func(cfg config.Config) int // cfg.SSHMaxFailures / cfg.FTPMaxFailures
	incCounter    func(n int)                 // IncSSHBy / IncFTPBy
}

type LoginServiceStrategy struct {
	cfg    loginServiceConfig
	totals map[string]int // accumulated failures per IP across scans
}

// builds a strategy for SSH failed-logins.
func NewSSHStrategy() *LoginServiceStrategy {
	return &LoginServiceStrategy{
		cfg: loginServiceConfig{
			name:          "ssh",
			defaultThresh: 3,
			getSpecific: func(cfg config.Config) int {
				return cfg.SSHMaxFailures
			},
			incCounter: IncSSHBy,
		},
		totals: make(map[string]int),
	}
}

//  builds a strategy for FTP failed-logins.
func NewFTPStrategy() *LoginServiceStrategy {
	return &LoginServiceStrategy{
		cfg: loginServiceConfig{
			name:          "ftp",
			defaultThresh: 3,
			getSpecific: func(cfg config.Config) int {
				return cfg.FTPMaxFailures
			},
			incCounter: IncFTPBy,
		},
		totals: make(map[string]int),
	}
}

func (s *LoginServiceStrategy) Name() string {
	return s.cfg.name
}

//  processes SSH/FTP login failures per cycle and enforces
// stats, alerts and firewall blocks.
func (s *LoginServiceStrategy) ProcessEvents(events map[string]int, cfg config.Config, now time.Time, whitelist map[string]struct{}) {
	if len(events) == 0 {
		return
	}

	service := s.cfg.name
	prefix := logPrefix(service)

	threshold := effectiveThreshold(
		s.cfg.getSpecific(cfg),
		cfg.MaxFailures,
		s.cfg.defaultThresh,
	)

	for ip, newFails := range events {
		if newFails <= 0 {
			continue
		}

		// 1) Update stats.
		s.cfg.incCounter(newFails)

		// 2) Update accumulated total per IP.
		oldTotal := s.totals[ip]
		total := oldTotal + newFails
		s.totals[ip] = total

		// 3) Log and alert.
		storage.AddLog(fmt.Sprintf(
			"%s %d new failed logins from %s (total=%d)",
			prefix, newFails, ip, total,
		))

		severity := classifySeverity(service, newFails, total, threshold)
		country := lookupCountry(ip)

		storage.AddAlert(storage.Alert{
			Timestamp: now.Format(time.RFC3339),
			Service:   service,
			IP:        ip,
			Country:   country,
			Severity:  severity,
			Message: fmt.Sprintf(
				"%d new %s failed logins from %s (total=%d)",
				newFails, strings.ToUpper(service), ip, total,
			),
		})

		// 4) Block decision.
		if isLoopback(ip) || isWhitelisted(ip, whitelist) {
			continue
		}

		if total >= threshold {
			storage.AddLog(fmt.Sprintf(
				"%s Blocking %s (total fails=%d, threshold=%d)",
				prefix, ip, total, threshold,
			))
			firewall.BlockIP(ip)
			storage.AddBlocked(ip)
			// Optionally: s.totals[ip] = 0
		}
	}
}

// ----------------APACHE---------------

type ApacheStrategy struct{}

// builds a strategy for Apache error monitoring.
func NewApacheStrategy() *ApacheStrategy {
	return &ApacheStrategy{}
}

func (s *ApacheStrategy) Name() string {
	return "apache"
}

// processes Apache 4xx/5xx errors for this cycle, updates stats,
// generates alerts and may block IPs according to config.
func (s *ApacheStrategy) ProcessEvents(apacheErrors map[string]int, cfg config.Config, now time.Time, whitelist map[string]struct{}) {
	if len(apacheErrors) == 0 {
		return
	}

	threshold := effectiveThreshold(
		cfg.ApacheErrorThreshold,
		cfg.MaxFailures,
		10, // previous default for Apache
	)

	// Total errors for this cycle (global stats).
	totalApacheErrors := 0
	for _, c := range apacheErrors {
		totalApacheErrors += c
	}
	if totalApacheErrors == 0 {
		return
	}

	// Update global Apache counter.
	IncApacheBy(totalApacheErrors)

	storage.AddLog(fmt.Sprintf(
		"[APACHE] Errors detected this cycle: %d (ips=%d)",
		totalApacheErrors, len(apacheErrors),
	))

	// One alert per IP.
	for ip, count := range apacheErrors {
		severity := classifySeverity("apache", count, count, threshold)
		country := lookupCountry(ip)

		storage.AddAlert(storage.Alert{
			Timestamp: now.Format(time.RFC3339),
			Service:   "apache",
			IP:        ip,
			Country:   country,
			Severity:  severity,
			Message: fmt.Sprintf(
				"%d Apache 4xx/5xx errors from %s this cycle",
				count, ip,
			),
		})

		// Optional blocking policy for Apache.
		if cfg.ApacheBlockOnThreshold &&
			!isLoopback(ip) &&
			!isWhitelisted(ip, whitelist) &&
			count >= threshold {

			storage.AddLog(fmt.Sprintf(
				"[APACHE] Blocking %s (errors this cycle=%d, threshold=%d)",
				ip, count, threshold,
			))
			firewall.BlockIP(ip)
			storage.AddBlocked(ip)
		}
	}
}
