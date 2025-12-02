package api

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

//  holds the data passed to the HTML template.
type simulateViewData struct {
	Kind         string
	KindUpper    string
	KindReadable string
	Count        int
	IP           string
	Timestamp    string
	BaseURL      string
}

//  returns a label for a given simulation kind.
func kindLabel(kind string) string {
	switch strings.ToLower(kind) {
	case "ssh":
		return "SSH · login attempts"
	case "ftp":
		return "FTP · failed authentications"
	case "apache":
		return "Apache · HTTP 4xx/5xx errors"
	default:
		return kind
	}
}

// HTML template for  simulation
const simulateResultHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>SecureMonitor · Simulation completed</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <link rel="stylesheet" href="/styles.css?v=2" />
  <style>
    .simulate-main {
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 1.75rem 1.25rem;
    }

    .simulate-card {
      max-width: 720px;
      width: 100%;
      display: flex;
      flex-direction: column;
      gap: 1.25rem;
    }

    .simulate-header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 1rem;
      margin-bottom: 0.2rem;
    }

    .simulate-eyebrow {
      font-size: 0.75rem;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--text-dim);
      margin-bottom: 0.25rem;
    }

    .simulate-title {
      font-size: 1.2rem;
      margin: 0;
    }

    .simulate-subtitle {
      margin: 0;
      font-size: 0.86rem;
      color: var(--text-dim);
    }

    .simulate-pill {
      white-space: nowrap;
    }

    .simulate-grid {
      margin-top: 0.75rem;
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 0.9rem;
    }

    .simulate-metric {
      padding: 0.55rem 0.7rem;
      border-radius: 0.75rem;
      border: 1px solid rgba(51, 65, 85, 0.9);
      background: rgba(15, 23, 42, 0.88);
    }

    .simulate-metric-label {
      font-size: 0.72rem;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      color: var(--text-dim);
      margin-bottom: 0.2rem;
      display: block;
    }

    .simulate-metric-value {
      font-size: 0.9rem;
      font-weight: 500;
    }

    .simulate-section {
      margin-top: 0.6rem;
    }

    .simulate-section-title {
      font-size: 0.9rem;
      margin: 0 0 0.4rem 0;
    }

    /* Curl block with thin horizontal scrollbar */
    .simulate-code {
      margin: 0;
      padding: 0.7rem 0.85rem;
      border-radius: 0.8rem;
      background: rgba(2, 6, 23, 0.9);
      border: 1px solid rgba(51, 65, 85, 0.9);
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
      font-size: 0.8rem;
      overflow-x: auto;
      white-space: pre;
      scrollbar-width: thin; /* Firefox */
      scrollbar-color: rgba(148, 163, 184, 0.85) transparent;
    }

    .simulate-code::-webkit-scrollbar {
      height: 6px;
    }

    .simulate-code::-webkit-scrollbar-track {
      background: transparent;
    }

    .simulate-code::-webkit-scrollbar-thumb {
      background: rgba(148, 163, 184, 0.75);
      border-radius: 999px;
    }

    .simulate-code::-webkit-scrollbar-thumb:hover {
      background: rgba(148, 163, 184, 1);
    }

    .simulate-footer {
      margin-top: 0.6rem;
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
    }

    .simulate-link {
      display: inline-flex;
      align-items: center;
      gap: 0.35rem;
      font-size: 0.82rem;
      text-decoration: none;
      color: #bfdbfe;
    }

    .simulate-link:hover {
      text-decoration: underline;
    }

    /* Responsive: tablets / phones */
    @media (max-width: 768px) {
      .simulate-main {
        align-items: flex-start;
        padding: 1.25rem 0.9rem;
      }

      .simulate-card {
        margin-top: 0.75rem;
        margin-bottom: 1.25rem;
        gap: 1rem;
      }

      .simulate-header {
        flex-direction: column;
        align-items: flex-start;
      }

      .simulate-pill {
        align-self: flex-start;
        margin-top: 0.25rem;
      }

      .simulate-title {
        font-size: 1.05rem;
      }

      .simulate-subtitle {
        font-size: 0.82rem;
      }
    }

    @media (max-width: 640px) {
      .simulate-grid {
        grid-template-columns: 1fr;
      }
    }
  </style>
</head>
<body>
  <div class="page simulate-main">
    <section class="card simulate-card">
      <header class="simulate-header">
        <div>
          <div class="simulate-eyebrow">Internal tool · /api/simulate</div>
          <h1 class="simulate-title">Simulation for {{.KindUpper}}</h1>
          <p class="simulate-subtitle">
            SecureMonitor injected simulated events into the daemon so you can
            exercise the dashboard without relying on real traffic.
          </p>
        </div>

        <span class="pill pill-ok simulate-pill">
          Simulation completed
        </span>
      </header>

      <div class="simulate-grid">
        <div class="simulate-metric">
          <span class="simulate-metric-label">Type</span>
          <span class="simulate-metric-value">{{.KindReadable}}</span>
        </div>
        <div class="simulate-metric">
          <span class="simulate-metric-label">Generated events</span>
          <span class="simulate-metric-value">{{.Count}}</span>
        </div>
        <div class="simulate-metric">
          <span class="simulate-metric-label">IP used</span>
          <span class="simulate-metric-value">{{.IP}}</span>
        </div>
      </div>

      <div class="simulate-section">
        <p class="simulate-subtitle">
          Simulation timestamp: <code>{{.Timestamp}}</code>
        </p>
      </div>

      <section class="simulate-section">
        <h2 class="simulate-section-title">Test from a terminal (curl)</h2>
        <pre class="simulate-code"><code>curl -X POST "{{.BaseURL}}/api/simulate?kind={{.Kind}}&n={{.Count}}"</code></pre>
      </section>

      <section class="simulate-section simulate-footer">
        <p class="muted">
          Now open the main dashboard to see how the stats, recent alerts and blocked
          IPs have changed.
        </p>
        <a href="../dashboard.html" class="simulate-link">
          Go to dashboard
        </a>
      </section>
    </section>
  </div>
</body>
</html>
`

var simulateResultTmpl = template.Must(
	template.New("simulateResult").Parse(simulateResultHTML),
)

// renders the simulation result page.
func renderSimulateHTML(w http.ResponseWriter, kind string, count int, ip string, ts time.Time) {
	data := simulateViewData{
		Kind:         kind,
		KindUpper:    strings.ToUpper(kind),
		KindReadable: kindLabel(kind),
		Count:        count,
		IP:           ip,
		Timestamp:    ts.Format(time.RFC3339),
		// Public base URL for the curl snippet.
		BaseURL: "https://securemonitor.juvenr.com",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := simulateResultTmpl.Execute(w, data); err != nil {
		log.Printf("template error in /api/simulate: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
