package monitor

import "sync"

var (
	simMu        sync.Mutex
	simSSHFails  = make(map[string]int)
	simFTPFails  = make(map[string]int)
	simApacheErr = make(map[string]int)
)

//  schedules n simulated SSH failures for a given IP.
func AddSimulatedSSH(ip string, n int) {
	if n <= 0 || ip == "" {
		return
	}
	simMu.Lock()
	defer simMu.Unlock()
	simSSHFails[ip] += n
}

//  schedules n simulated FTP failures for a given IP.
func AddSimulatedFTP(ip string, n int) {
	if n <= 0 || ip == "" {
		return
	}
	simMu.Lock()
	defer simMu.Unlock()
	simFTPFails[ip] += n
}

//  schedules n simulated Apache errors for a given IP.
func AddSimulatedApache(ip string, n int) {
	if n <= 0 || ip == "" {
		return
	}
	simMu.Lock()
	defer simMu.Unlock()
	simApacheErr[ip] += n
}

//  returns and clears the current SSH simulated events.
func drainSimulatedSSH() map[string]int {
	simMu.Lock()
	defer simMu.Unlock()

	out := simSSHFails
	simSSHFails = make(map[string]int)
	return out
}

//  returns and clears the current FTP simulated events.
func drainSimulatedFTP() map[string]int {
	simMu.Lock()
	defer simMu.Unlock()

	out := simFTPFails
	simFTPFails = make(map[string]int)
	return out
}

//  returns and clears the current Apache simulated events.
func drainSimulatedApache() map[string]int {
	simMu.Lock()
	defer simMu.Unlock()

	out := simApacheErr
	simApacheErr = make(map[string]int)
	return out
}
