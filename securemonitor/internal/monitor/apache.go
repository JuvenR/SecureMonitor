package monitor

import (
	"log"
	"strings"
)

// reports whether the Apache access log line
func isApacheErrorLine(line string) bool {
	// currently we track 404 and 500 as signals.
	return strings.Contains(line, " 404 ") || strings.Contains(line, " 500 ")
}

// tries to extract the client IP from a common"
func extractApacheAccessIP(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}

	ip := strings.TrimSpace(fields[0])
	if ip == "-" {
		return ""
	}
	return ip
}

// reads new lines from the access log and returns
func parseApacheErrors(path string) map[string]int {
	lines, err := ReadNewLines(path)
	if err != nil || len(lines) == 0 {
		return map[string]int{}
	}

	errorsByIP := make(map[string]int)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		if !isApacheErrorLine(line) {
			continue
		}

		ip := extractApacheAccessIP(line)
		if ip == "" {
			ip = extractIP(line)
		}
		if ip == "" {
			ip = "unknown"
		}

		errorsByIP[ip]++
		log.Printf("apache: matched error from %s: %s", ip, line)
	}

	return errorsByIP
}
