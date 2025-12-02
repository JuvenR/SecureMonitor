package monitor

import "sync"

// In-memory aggregated metrics exposed via the HTTP API.
var (
	countersMu  sync.Mutex
	sshCount    int
	ftpCount    int
	apacheCount int
)

// increments the SSH failed login counter by 1.
func IncSSH() {
	IncSSHBy(1)
}

//  increments the FTP failed login counter by 1.
func IncFTP() {
	IncFTPBy(1)
}

//  increments the Apache error counter by 1.
func IncApache() {
	IncApacheBy(1)
}

//  increments the SSH failed login counter by n.
func IncSSHBy(n int) {
	if n <= 0 {
		return
	}
	countersMu.Lock()
	sshCount += n
	countersMu.Unlock()
}

//  increments the FTP failed login counter by n.
func IncFTPBy(n int) {
	if n <= 0 {
		return
	}
	countersMu.Lock()
	ftpCount += n
	countersMu.Unlock()
}

//  increments the Apache error counter by n.
func IncApacheBy(n int) {
	if n <= 0 {
		return
	}
	countersMu.Lock()
	apacheCount += n
	countersMu.Unlock()
}

//  returns the current SSH failed login counter.
func GetSSHCount() int {
	countersMu.Lock()
	defer countersMu.Unlock()
	return sshCount
}

//  returns the current FTP failed login counter.
func GetFTPCount() int {
	countersMu.Lock()
	defer countersMu.Unlock()
	return ftpCount
}

//  returns the current Apache error counter.
func GetApacheCount() int {
	countersMu.Lock()
	defer countersMu.Unlock()
	return apacheCount
}
