package monitor

import "strings"

// parseFTPFailuresFromLines aggregates failed FTP login attempts per IP
// from raw log lines (e.g. vsftpd logs).
func parseFTPFailuresFromLines(lines []string) map[string]int {
	failures := make(map[string]int)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// Only consider FTP entries from vsftpd.
		if !strings.Contains(line, "vsftpd") {
			continue
		}

		// Count authentication failures for each source IP.
		if strings.Contains(line, "authentication failure") {
			ip := extractIP(line)
			if ip != "" {
				failures[ip]++
			}
		}
	}

	return failures
}

// parseFTPFailures reads new FTP log lines from disk and returns per-IP failure counts.
func parseFTPFailures(path string) map[string]int {
	lines, err := ReadNewLines(path)
	if err != nil || len(lines) == 0 {
		return map[string]int{}
	}
	return parseFTPFailuresFromLines(lines)
}
