# SecureMonitor 🛡️

**SecureMonitor** is a Linux security daemon written in Go that:

- Monitors logs from **SSH**, **FTP (vsftpd)** and **Apache**.
- Generates **statistics** and **alerts** with severity levels (LOW / MEDIUM / HIGH).
- Integrates with **UFW** to **block malicious IPs** (with optional auto-unblock).
- Exposes a dark **web dashboard** on port `9000` with:
  - Service stats cards (SSH / FTP / Apache),
  - Recent alerts table,
  - Blocked IP list with “Unblock” actions,
  - Live-ish log viewer.

It was originally built as an **educational / lab tool** for a *Linux Networks* course, and then evolved into a small, real-world style monitoring daemon.

---

## 🚀 Quick demo

Dashboard (once the daemon is running):

```text
http://YOUR-SERVER:9000/dashboard.html
```
Simulate attacks (no real traffic needed):

```text
curl -X POST "http://YOUR-SERVER:9000/api/simulate?kind=ssh&n=15"
curl -X POST "http://YOUR-SERVER:9000/api/simulate?kind=ftp&n=25"
curl -X POST "http://YOUR-SERVER:9000/api/simulate?kind=apache&n=12"
```
Optionally you can expose it over the internet using something like Cloudflare Tunnel, e.g.:
```text
cloudflared tunnel run securemonitor-tunnel
```
## ✨ Features

### 🔍 Log monitoring

- **SSH** → `auth.log`
- **FTP (vsftpd)** → `vsftpd.log`
- **Apache** → access & error logs (`access.log`, `error.log`)

---

### 📊 Web dashboard

- Aggregated counters per service (SSH / FTP / Apache).
- Alerts with severity levels: `LOW`, `MEDIUM`, `HIGH`.
- Recent logs panel with visual highlight for failures.
- Blocked IP list with UI controls to unblock IPs.
- Alerts include the IPs and countries from the attackers.

---

### 🔒 Firewall integration (UFW)

- Automatic blocking of IPs above configurable thresholds.
- Optional auto-unblock after **N** minutes.
- Whitelist for IPs / ranges that must never be blocked.

---

### 🧪 Attack simulator

- Internal endpoint:  
  `GET /api/simulate?kind=ssh|ftp|apache&n=N`  
  `POST /api/simulate?kind=ssh|ftp|apache&n=N`
- Injects fake events into the daemon, perfect for demos and testing.

---

### ⚙️ Simple JSON configuration

- Log paths
- Thresholds per service
- Auto-unblock behavior
- Whitelist

---

### 🧱 Modular Go architecture

- `internal/api` – HTTP API and minimal HTML views.
- `internal/monitor` – Monitoring logic and counters.
- `internal/firewall` – UFW integration (block/unblock).
- `internal/storage` – In-memory storage + basic persistence.
- `internal/config` – Configuration loading.
- `web/` – Dashboard and documentation site (HTML/CSS/JS).
---
## ⚙️ Requirements

- **OS:** Linux (tested on Ubuntu Server 22.04)
- **Go:** 1.21+ (recommended)
- **Optional but typical monitored services:**
  - `openssh-server`
  - `vsftpd`
  - `apache2`
- **Firewall:** UFW enabled if you want firewall integration.

---

## 🧩 Installation

### 1. Clone and build

```bash
git clone https://github.com/JuvenR/securemonitor.git
cd securemonitor

go build -o securemonitor ./cmd/securemonitor
sudo mv securemonitor /usr/local/bin/

```
### 2. Create a dedicated service user (recommended)
```bash
sudo adduser --system --no-create-home --group securemon
sudo mkdir -p /var/lib/securemonitor
sudo chown securemon:securemon /var/lib/securemonitor
```

### 3. Configuration
Create a config file, for example: /etc/securemonitor/config.json
```bash
{
  "services_to_watch": ["ssh", "ftp", "apache"],
  "ssh_log_path": "/var/log/auth.log",
  "apache_access_log_path": "/var/log/apache2/securemonitor_access.log",
  "apache_error_log_path": "/var/log/apache2/error.log",
  "ftp_log_path": "/var/log/auth.log",

  "max_failures": 3,             
  "ssh_max_failures": 5,         
  "ftp_max_failures": 3,          
  "apache_error_threshold": 10,  
  "auto_unblock_minutes": 1,

  "check_interval_seconds": 5,
  "blocked_ips_file": "blocked_ips.txt",
  "whitelist_file": "whitelist.txt",
  "apache_block_on_threshold": false


}
```
Adjust paths to match your distro / log configuration.

### 4. systemd service
Create /etc/systemd/securemonitor.service:
```bash
# /etc/systemd/system/securemonitor.service
[Unit]
Description=Secure Monitor Daemon
After=network.target

[Service]
ExecStart=/home/juveen/securemonitor/securemonitor
WorkingDirectory=/home/juveen/securemonitor
StandardOutput=journal
StandardError=journal
Restart=always

[Install]
WantedBy=multi-user.target
```

Reload and enable:
```bash
sudo systemctl daemon-reload
sudo systemctl enable securemonitor
sudo systemctl start securemonitor
```
Check status:
```bash
sudo systemctl status securemonitor
```

## 🌐 Web dashboard

The daemon serves both the API and the static files from `web/` on the configured port (default **9000**).

### Main entry points

- **Dashboard UI**
 ```bash
  GET /dashboard.html
 ```
- **Unified snapshot for the UI**
```bash
  GET /api/dashboard 
 ```
```bash
{
  "status": { ... },
  "stats": { "ssh": 0, "ftp": 0, "apache": 0 },
  "logs": [ ... ],
  "alerts": [ ... ],
  "blocked": [ "1.2.3.4", "5.6.7.8" ]
}
```


