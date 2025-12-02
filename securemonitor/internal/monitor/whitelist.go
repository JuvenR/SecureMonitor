package monitor

import (
	"net"
	"os"
	"strings"
)

// isLoopback reports whether an IP belongs to loopback/private ranges.
func isLoopback(ip string) bool {
	p := net.ParseIP(strings.TrimSpace(ip))
	if p == nil {
		return false
	}
	return p.IsLoopback() || p.IsPrivate()
}

// loadWhitelist loads a simple IP whitelist file where each line may contain:
// - an IP address
// - comments starting with '#'
func loadWhitelist(path string) map[string]struct{} {
	m := make(map[string]struct{})
	if path == "" {
		return m
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return m
	}

	lines := strings.Split(string(data), "\n")
	for _, raw := range lines {
		// Support comments with '#'.
		s := strings.Split(raw, "#")[0]
		ip := strings.TrimSpace(s)
		if ip == "" {
			continue
		}
		m[ip] = struct{}{}
	}
	return m
}

// isWhitelisted returns true if the IP is present in the whitelist map.
func isWhitelisted(ip string, wl map[string]struct{}) bool {
	_, ok := wl[strings.TrimSpace(ip)]
	return ok
}
