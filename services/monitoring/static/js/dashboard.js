const POLL_INTERVAL_MS = 5000;
const SPIDER_QUEUE_HIGH_THRESHOLD = 500;
const INDEXER_BACKLOG_THRESHOLD = 100;

const els = {
  uptime: document.getElementById("uptime"),
  lastUpdate: document.getElementById("last-update"),
  liveIndicator: document.getElementById("live-indicator"),
  alertBanner: document.getElementById("alert-banner"),

  statusEngine: document.getElementById("status-engine"),
  statusEngineHttp: document.getElementById("status-engine-http"),
  statusSpider: document.getElementById("status-spider"),
  statusIndexer: document.getElementById("status-indexer"),
  statusRanking: document.getElementById("status-ranking"),
  statusScheduler: document.getElementById("status-scheduler"),

  spiderQueue: document.getElementById("spider-queue"),
  spiderRate: document.getElementById("spider-rate"),
  spiderRunning: document.getElementById("spider-running"),
  spiderStarted: document.getElementById("spider-started"),
  spiderQueueBar: document.getElementById("spider-queue-bar"),
  spiderQueueLabel: document.getElementById("spider-queue-label"),

  indexerBacklog: document.getElementById("indexer-backlog"),
  indexerRunning: document.getElementById("indexer-running"),
  indexerSpawned: document.getElementById("indexer-spawned"),
  indexerBacklogBar: document.getElementById("indexer-backlog-bar"),
  indexerBacklogLabel: document.getElementById("indexer-backlog-label"),

  rankingTotalIndexed: document.getElementById("ranking-total-indexed"),
  rankingRunning: document.getElementById("ranking-running"),
  rankingTriggered: document.getElementById("ranking-triggered"),

  engineRunning: document.getElementById("engine-running"),
  engineHttp: document.getElementById("engine-http"),
  engineRestarts: document.getElementById("engine-restarts"),
  engineRestartFailures: document.getElementById("engine-restart-failures"),
  engineRateLimited: document.getElementById("engine-rate-limited"),

  cleanupLastBatch: document.getElementById("cleanup-last-batch"),
  cleanupTotal: document.getElementById("cleanup-total"),

  schedulerTick: document.getElementById("scheduler-tick"),
  schedulerErrors: document.getElementById("scheduler-errors"),
};

let lastSchedulerErrors = 0;
let pollFailed = false;

function getMetric(snapshot, key, fallback = 0) {
  const value = snapshot?.metrics?.[key]?.value;
  return Number.isFinite(value) ? value : fallback;
}

function formatUptime(totalSeconds) {
  const s = Math.max(0, Math.floor(totalSeconds));
  const h = String(Math.floor(s / 3600)).padStart(2, "0");
  const m = String(Math.floor((s % 3600) / 60)).padStart(2, "0");
  const sec = String(s % 60).padStart(2, "0");
  return `${h}:${m}:${sec}`;
}

function setText(el, text) {
  if (!el) return;
  el.textContent = String(text);
}

function setState(el, text, className) {
  if (!el) return;
  el.textContent = text;
  el.classList.remove("ok", "warn", "critical");
  if (className) el.classList.add(className);
}

function setMeter(bar, label, value, threshold) {
  if (!bar || !label) return;
  const percent = Math.min(100, Math.round((value / threshold) * 100));
  bar.style.width = `${percent}%`;
  bar.classList.remove("warn", "critical");

  if (percent >= 100) {
    bar.classList.add("critical");
  } else if (percent >= 70) {
    bar.classList.add("warn");
  }

  label.textContent = `${value} / ${threshold}`;
}

function updateAlert(message, isVisible) {
  if (!els.alertBanner) return;
  if (!isVisible) {
    els.alertBanner.classList.add("hidden");
    els.alertBanner.textContent = "";
    return;
  }
  els.alertBanner.classList.remove("hidden");
  els.alertBanner.textContent = message;
}

function render(snapshot) {
  const uptime = Number(snapshot?.uptime_seconds ?? 0);

  const spiderQueue = getMetric(snapshot, "spider_queue_size");
  const spiderRate = getMetric(snapshot, "spider_crawl_rate_per_min");
  const spiderRunning = getMetric(snapshot, "spider_running_count");
  const spiderStarted = getMetric(snapshot, "spider_instances_started_total");

  const indexerBacklog = getMetric(snapshot, "indexer_unindexed_backlog");
  const indexerRunning = getMetric(snapshot, "indexer_running_jobs");
  const indexerSpawned = getMetric(snapshot, "indexer_jobs_spawned_total");

  const rankingTotalIndexed = getMetric(snapshot, "ranking_total_indexed");
  const rankingRunning = getMetric(snapshot, "ranking_running_jobs");
  const rankingTriggered = getMetric(snapshot, "ranking_jobs_triggered_total");

  const engineRunning = getMetric(snapshot, "engine_running");
  const engineHttpAlive = getMetric(snapshot, "engine_http_alive");
  const engineRestarts = getMetric(snapshot, "engine_restarts_total");
  const engineRestartFailures = getMetric(snapshot, "engine_restart_failures_total");
  const engineRateLimited = getMetric(snapshot, "engine_restarts_rate_limited");

  const cleanupLastBatch = getMetric(snapshot, "cleanup_last_batch_rows");
  const cleanupTotal = getMetric(snapshot, "cleanup_rows_total");

  const schedulerTick = getMetric(snapshot, "scheduler_tick_count");
  const schedulerErrors = getMetric(snapshot, "scheduler_tick_errors_total");

  setText(els.uptime, `UPTIME: ${formatUptime(uptime)}`);
  setText(els.lastUpdate, `LAST UPDATE: ${new Date().toLocaleTimeString()}`);

  setState(els.statusEngine, engineRunning > 0 ? "RUNNING" : "DOWN", engineRunning > 0 ? "ok" : "critical");
  setState(els.statusEngineHttp, engineHttpAlive > 0 ? "ALIVE" : "DEAD", engineHttpAlive > 0 ? "ok" : "critical");
  setState(els.statusSpider, `${spiderRunning}`, spiderRunning > 0 ? "ok" : "warn");
  setState(els.statusIndexer, `${indexerRunning}`, indexerRunning > 0 ? "ok" : "warn");
  setState(els.statusRanking, `${rankingRunning}`, rankingRunning > 0 ? "ok" : "warn");
  setState(els.statusScheduler, schedulerErrors > 0 ? "ERRORS" : "OK", schedulerErrors > 0 ? "critical" : "ok");

  setText(els.spiderQueue, spiderQueue);
  setText(els.spiderRate, spiderRate);
  setText(els.spiderRunning, spiderRunning);
  setText(els.spiderStarted, spiderStarted);
  setMeter(els.spiderQueueBar, els.spiderQueueLabel, spiderQueue, SPIDER_QUEUE_HIGH_THRESHOLD);

  setText(els.indexerBacklog, indexerBacklog);
  setText(els.indexerRunning, indexerRunning);
  setText(els.indexerSpawned, indexerSpawned);
  setMeter(els.indexerBacklogBar, els.indexerBacklogLabel, indexerBacklog, INDEXER_BACKLOG_THRESHOLD);

  setText(els.rankingTotalIndexed, rankingTotalIndexed);
  setText(els.rankingRunning, rankingRunning);
  setText(els.rankingTriggered, rankingTriggered);

  setState(els.engineRunning, engineRunning > 0 ? "YES" : "NO", engineRunning > 0 ? "ok" : "critical");
  setState(els.engineHttp, engineHttpAlive > 0 ? "YES" : "NO", engineHttpAlive > 0 ? "ok" : "critical");
  setText(els.engineRestarts, engineRestarts);
  setText(els.engineRestartFailures, engineRestartFailures);
  setText(els.engineRateLimited, engineRateLimited);

  setText(els.cleanupLastBatch, cleanupLastBatch);
  setText(els.cleanupTotal, cleanupTotal);

  setText(els.schedulerTick, schedulerTick);
  setText(els.schedulerErrors, schedulerErrors);

  const alertMessages = [];
  if (engineRateLimited > 0) {
    alertMessages.push("ENGINE RESTART RATE LIMITED");
  }
  if (schedulerErrors > lastSchedulerErrors) {
    alertMessages.push("SCHEDULER ERRORS INCREASED");
  }
  updateAlert(alertMessages.join(" | "), alertMessages.length > 0);
  lastSchedulerErrors = schedulerErrors;
}

async function tick() {
  try {
    const res = await fetch("/metrics", { cache: "no-store" });
    if (!res.ok) throw new Error(`metrics request failed with ${res.status}`);

    const snapshot = await res.json();
    render(snapshot);
    pollFailed = false;
    if (els.liveIndicator) {
      setState(els.liveIndicator, "LIVE 5s", "ok");
    }
  } catch (_err) {
    pollFailed = true;
    if (els.liveIndicator) {
      setState(els.liveIndicator, "STALE", "warn");
    }
    updateAlert("MONITOR_UNREACHABLE: failed to refresh metrics", true);
  }
}

void tick();
setInterval(() => {
  void tick();
}, POLL_INTERVAL_MS);
