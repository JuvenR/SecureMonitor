package storage

//  logs buffer (ring buffer).
var (
	recentLogs  []string
	maxLogsSize = 200
)

//  appends a line to the recent logs buffer.
func AddLog(line string) {
	storeMu.Lock()
	defer storeMu.Unlock()

	recentLogs = append(recentLogs, line)
	if len(recentLogs) > maxLogsSize {
		recentLogs = recentLogs[len(recentLogs)-maxLogsSize:]
	}
}

//  returns a copy of the recent log buffer.
func GetLogs() []string {
	storeMu.Lock()
	defer storeMu.Unlock()

	out := make([]string, len(recentLogs))
	copy(out, recentLogs)
	return out
}
