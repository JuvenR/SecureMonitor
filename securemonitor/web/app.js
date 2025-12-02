// fetch json and surface http errors
async function fetchJSON(url) {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error("http " + res.status);
  }
  return res.json();
}

// dashboard state
const dashboardState = {
  status: null,
  stats: null,
  logs: [],
  alerts: [],
  blocked: [],
};

let consecutiveFailures = 0;

// state for logs and alert pill
let lastLogsLength = 0;
let lastFailureIndex = -1;
let lastAlertAt = null;

// normalize log entry to a single line
function formatLogEntry(entry) {
  if (typeof entry === "string") {
    return entry;
  }
  if (!entry || typeof entry !== "object") {
    return String(entry);
  }

  const timestamp =
    entry.timestamp ||
    entry.time ||
    entry.ts ||
    "";
  const level = (entry.level || entry.severity || entry.type || "").toUpperCase();
  const message =
    entry.message ||
    entry.msg ||
    entry.error ||
    entry.err ||
    entry.text ||
    "";
  const extra =
    entry.detail ||
    entry.details ||
    entry.context ||
    entry.info ||
    "";

  let line = "";
  if (timestamp) line += `[${timestamp}] `;
  if (level) line += `${level} `;
  line += message;
  if (extra) line += ` – ${extra}`;
  return line.trim();
}

// simple check for security failures in logs
function isFailureEntry(entry) {
  const line = formatLogEntry(entry).toLowerCase();
  return (
    line.includes("error") ||
    line.includes("failed") ||
    line.includes("fail ") ||
    line.includes("critical") ||
    line.includes("denied") ||
    line.includes("block")
  );
}

// relative time in spanish
function formatRelativeFromNow(date) {
  if (!date) return "Sin alertas recientes";

  const now = new Date();
  const diffMs = now.getTime() - date.getTime();

  if (diffMs < 0) {
    return "Hace un momento";
  }

  const diffSec = Math.round(diffMs / 1000);

  if (diffSec < 60) {
    return "Ahora";
  }

  let value;
  let unit;

  if (diffSec < 3600) {
    value = -Math.round(diffSec / 60);
    unit = "minute";
  } else if (diffSec < 86400) {
    value = -Math.round(diffSec / 3600);
    unit = "hour";
  } else {
    value = -Math.round(diffSec / 86400);
    unit = "day";
  }

  if (typeof Intl !== "undefined" && Intl.RelativeTimeFormat) {
    const rtf = new Intl.RelativeTimeFormat("es", { numeric: "auto" });
    return rtf.format(value, unit);
  }

  const n = Math.abs(value);
  const sufijos = {
    minute: "min",
    hour: "h",
    day: "d",
  };
  return `hace ${n}${sufijos[unit] || ""}`;
}

// update alert pill text and style
function updateLastAlertPill() {
  const pill = document.getElementById("lastAlertPill");
  if (!pill) return;

  if (!lastAlertAt) {
    pill.textContent = "Sin alertas recientes";
    pill.className = "pill pill-small";
    return;
  }

  const relative = formatRelativeFromNow(lastAlertAt);
  pill.textContent = `Última alerta: ${relative}`;
  pill.className = "pill pill-small pill-warn";
}

// map severity to css class
function severityClass(sev) {
  const s = (sev || "").toUpperCase();
  if (s === "HIGH") return "sev-high";
  if (s === "MEDIUM") return "sev-medium";
  return "sev-low";
}

// map service id to label
function serviceLabel(service) {
  switch ((service || "").toLowerCase()) {
    case "ssh":
      return "SSH";
    case "ftp":
      return "FTP";
    case "apache":
      return "Apache";
    default:
      return service || "—";
  }
}

// apply status info to header
function applyStatus(st) {
  const badge = document.getElementById("statusText");
  const pill = document.getElementById("daemonPill");
  if (!badge || !pill || !st) return;

  const ok = st.status === "OK";
  badge.textContent = ok ? `OK – ${st.msg || ""}` : "ERROR";
  badge.className = ok ? "badge badge-ok" : "badge badge-warn";

  pill.textContent = ok ? "Daemon activo" : "Daemon con problemas";
  pill.className = ok ? "pill pill-ok" : "pill pill-warn";
}

// show degraded connection in header
function applyConnectionDegraded(failures) {
  const badge = document.getElementById("statusText");
  const pill = document.getElementById("daemonPill");
  if (!badge || !pill) return;

  if (failures < 3) {
    badge.textContent = "Retraso de red…";
    badge.className = "badge badge-warn";
  } else {
    badge.textContent = "SIN CONEXIÓN";
    badge.className = "badge badge-warn";
    pill.textContent = "Sin conexión con daemon";
    pill.className = "pill pill-warn";
  }
}

// update stats cards
function applyStats(stats) {
  if (!stats) return;
  const sshEl = document.getElementById("sshCount");
  const ftpEl = document.getElementById("ftpCount");
  const apacheEl = document.getElementById("apacheCount");

  if (sshEl) sshEl.textContent = stats.ssh ?? 0;
  if (ftpEl) ftpEl.textContent = stats.ftp ?? 0;
  if (apacheEl) apacheEl.textContent = stats.apache ?? 0;
}

// incremental logs rendering
function applyLogs(allLogs) {
  const container = document.getElementById("logs");
  if (!container || !Array.isArray(allLogs)) return;

  const wasAtBottom =
    container.scrollHeight === 0
      ? true
      : container.scrollTop + container.clientHeight >=
        container.scrollHeight - 8;

  if (allLogs.length < lastLogsLength) {
    container.innerHTML = "";
    lastLogsLength = 0;
  }

  let latestFailureIndexLocal = -1;

  for (let idx = lastLogsLength; idx < allLogs.length; idx++) {
    const entry = allLogs[idx];
    const div = document.createElement("div");
    div.textContent = formatLogEntry(entry);
    container.appendChild(div);

    if (isFailureEntry(entry)) {
      latestFailureIndexLocal = idx;
    }
  }

  lastLogsLength = allLogs.length;

  if (wasAtBottom) {
    container.scrollTop = container.scrollHeight;
  }

  const hasNewFailure =
    latestFailureIndexLocal !== -1 &&
    latestFailureIndexLocal > lastFailureIndex;

  if (hasNewFailure) {
    container.classList.add("logs-alert");
    setTimeout(() => {
      container.classList.remove("logs-alert");
    }, 2000);
  }

  lastFailureIndex = latestFailureIndexLocal;
}

// render alerts table
function renderAlerts(alerts) {
  const tbody = document.getElementById("alertsBody");
  const emptyMsg = document.getElementById("alertsEmptyMsg");
  const countEl = document.getElementById("alertsCount");

  if (!tbody) return;

  const list = Array.isArray(alerts) ? alerts : [];

  if (countEl) {
    countEl.textContent = `${list.length} alerta${list.length === 1 ? "" : "s"}`;
  }

  tbody.innerHTML = "";

  if (list.length === 0) {
    if (emptyMsg) emptyMsg.style.display = "block";
    lastAlertAt = null;
    updateLastAlertPill();
    return;
  }

  if (emptyMsg) emptyMsg.style.display = "none";

  const sorted = [...list].reverse().slice(0, 50);

  const newest = sorted[0];
  let newestTs = newest.timestamp || newest.time || newest.ts || "";
  if (newestTs) {
    const d = new Date(newestTs);
    if (!Number.isNaN(d.getTime())) {
      lastAlertAt = d;
    } else {
      lastAlertAt = new Date();
    }
    updateLastAlertPill();
  }

  for (const alert of sorted) {
    const tr = document.createElement("tr");

    let when = alert.timestamp || "";
    try {
      if (when) {
        const d = new Date(when);
        when = d.toLocaleString();
      }
    } catch {
      // keep original if parse fails
    }

    const sev = alert.severity || "LOW";
    const sevCls = severityClass(sev);
    const svc = serviceLabel(alert.service);
    const ip = alert.ip || "—";
    const msg = alert.message || "";
    const country = alert.country || "—";

    if (sevCls === "sev-high") {
      tr.classList.add("sev-row-high");
    } else if (sevCls === "sev-medium") {
      tr.classList.add("sev-row-medium");
    } else {
      tr.classList.add("sev-row-low");
    }

    tr.innerHTML = `
      <td data-label="Hora">${when}</td>
      <td data-label="Servicio">
        <span class="service-tag">${svc}</span>
      </td>
      <td data-label="IP">${ip}</td>
      <td data-label="País">${country}</td>
      <td data-label="Severidad">
        <span class="severity-pill ${sevCls}">
          ${sev.toUpperCase()}
        </span>
      </td>
      <td data-label="Mensaje">${msg}</td>
    `;

    tbody.appendChild(tr);
  }
}

// small wrapper so render and apply stay separate
function applyAlerts(alerts) {
  renderAlerts(alerts);
}

// render blocked ip list
function applyBlocked(ips) {
  const list = document.getElementById("blockedList");
  const empty = document.getElementById("blockedEmptyMsg");
  if (!list) return;

  list.innerHTML = "";

  const arr = Array.isArray(ips) ? ips : [];

  if (empty) {
    empty.style.display = arr.length === 0 ? "block" : "none";
  }

  const sortedIps = [...arr].sort((a, b) => a.localeCompare(b, "en"));

  sortedIps.forEach((ip) => {
    const li = document.createElement("li");

    const span = document.createElement("span");
    span.textContent = ip;

    const btn = document.createElement("button");
    btn.textContent = "Unblock";
    btn.onclick = async () => {
      try {
        await fetch("/api/unblock?ip=" + encodeURIComponent(ip), {
          method: "POST",
        });
        refreshDashboard();
      } catch (_) {
        // ignore unblock errors in ui
      }
    };

    li.appendChild(span);
    li.appendChild(btn);
    list.appendChild(li);
  });
}

// main snapshot loop
async function refreshDashboard() {
  try {
    const data = await fetchJSON("/api/dashboard");
    consecutiveFailures = 0;

    dashboardState.status  = data.status  || dashboardState.status;
    dashboardState.stats   = data.stats   || dashboardState.stats;
    dashboardState.logs    = Array.isArray(data.logs)    ? data.logs    : dashboardState.logs;
    dashboardState.alerts  = Array.isArray(data.alerts)  ? data.alerts  : dashboardState.alerts;
    dashboardState.blocked = Array.isArray(data.blocked) ? data.blocked : dashboardState.blocked;

    applyStatus(dashboardState.status);
    applyStats(dashboardState.stats);
    applyLogs(dashboardState.logs);
    applyAlerts(dashboardState.alerts);
    applyBlocked(dashboardState.blocked);
  } catch (e) {
    consecutiveFailures++;
    applyConnectionDegraded(consecutiveFailures);
  }
}

// bootstrap and polling
document.addEventListener("DOMContentLoaded", () => {
  refreshDashboard();
  setInterval(refreshDashboard, 5000);
});
