package storage

// represents a security alert for the dashboard.
type Alert struct {
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
	IP        string `json:"ip,omitempty"`
	Country   string `json:"country,omitempty"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

// alerts buffer (ring buffer).
var (
	alerts        []Alert
	maxAlertsSize = 100
)

//  appends an alert to the buffer, trimming if needed.
func AddAlert(a Alert) {
	storeMu.Lock()
	defer storeMu.Unlock()

	alerts = append(alerts, a)
	if len(alerts) > maxAlertsSize {
		alerts = alerts[len(alerts)-maxAlertsSize:]
	}
}

// returns a snapshot of the alerts in memory.
func GetAlerts() []Alert {
	storeMu.Lock()
	defer storeMu.Unlock()

	out := make([]Alert, len(alerts))
	copy(out, alerts)
	return out
}
