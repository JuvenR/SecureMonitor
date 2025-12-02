package storage

import (
	"os"
	"strings"
	"sync"
	"time"
)

//  holds an IP, when it was blocked and how many times
// it has been blocked in this daemon lifetime.
type BlockedEntry struct {
	IP        string    `json:"ip"`
	BlockedAt time.Time `json:"blocked_at"`
	Strikes   int       `json:"strikes"`
}

var (
	storeMu      sync.Mutex
	blockedIPs   = make(map[string]BlockedEntry)
	strikeCounts = make(map[string]int)
)

// restores blocked IPs from a file (one IP per line).
// Since the file does not store timestamps nor strikes, we use time.Now()
func LoadBlockedFromFile(path string) {
	storeMu.Lock()
	defer storeMu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	now := time.Now()
	lines := strings.Split(string(data), "\n")
	for _, raw := range lines {
		ip := strings.TrimSpace(raw)
		if ip == "" {
			continue
		}

		current := strikeCounts[ip]
		if current < 1 {
			current = 1
		}
		strikeCounts[ip] = current

		blockedIPs[ip] = BlockedEntry{
			IP:        ip,
			BlockedAt: now,
			Strikes:   current,
		}
	}
}

// persists only the IPs (one per line) to disk.
func SaveBlockedToFile(path string) {
	storeMu.Lock()
	defer storeMu.Unlock()

	var b strings.Builder
	for ip := range blockedIPs {
		b.WriteString(ip)
		b.WriteByte('\n')
	}

	_ = os.WriteFile(path, []byte(b.String()), 0o644)
}

// AddBlocked marks an IP as blocked.
func AddBlocked(ip string) {
	storeMu.Lock()
	defer storeMu.Unlock()

	ip = strings.TrimSpace(ip)
	if ip == "" {
		return
	}

	if entry, ok := blockedIPs[ip]; ok {
		if entry.Strikes <= 0 {
			curr := strikeCounts[ip]
			if curr < 1 {
				curr = 1
			}
			entry.Strikes = curr
			blockedIPs[ip] = entry
		}
		return
	}

	prevStrikes := strikeCounts[ip]
	if prevStrikes < 0 {
		prevStrikes = 0
	}
	newStrikes := prevStrikes + 1
	if newStrikes <= 0 {
		newStrikes = 1
	}

	strikeCounts[ip] = newStrikes

	blockedIPs[ip] = BlockedEntry{
		IP:        ip,
		BlockedAt: time.Now(),
		Strikes:   newStrikes,
	}
}

//  removes an IP from the in-memory blocked map.
func RemoveBlocked(ip string) {
	storeMu.Lock()
	defer storeMu.Unlock()

	delete(blockedIPs, strings.TrimSpace(ip))
}

//  returns only the IPs (for the /api/blocked handler).
func ListBlocked() []string {
	storeMu.Lock()
	defer storeMu.Unlock()

	ips := make([]string, 0, len(blockedIPs))
	for ip := range blockedIPs {
		ips = append(ips, ip)
	}
	return ips
}

//  returns IP + BlockedAt + Strikes for auto-unblock logic.
func ListBlockedEntries() []BlockedEntry {
	storeMu.Lock()
	defer storeMu.Unlock()

	out := make([]BlockedEntry, 0, len(blockedIPs))
	for _, entry := range blockedIPs {
		out = append(out, entry)
	}
	return out
}
