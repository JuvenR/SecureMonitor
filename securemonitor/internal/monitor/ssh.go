package monitor

import "strings"

//  aggregates failed SSH login attempts
// for each IP from raw log lines.
func parseSSHFailuresFromLines(lines []string) map[string]int {
	failures := make(map[string]int)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// Only consider entries from sshd.
		if !strings.Contains(line, "sshd") {
			continue
		}

		//  Failed password" events per source IP.
		if strings.Contains(line, "Failed password") {
			ip := extractIP(line)
			if ip != "" {
				failures[ip]++
			}
		}
	}

	return failures
}

//  reads new SSH log lines from disk and returns per-IP failure counts.
func parseSSHFailures(path string) map[string]int {
	lines, err := ReadNewLines(path)
	if err != nil || len(lines) == 0 {
		return map[string]int{}
	}
	return parseSSHFailuresFromLines(lines)
}
