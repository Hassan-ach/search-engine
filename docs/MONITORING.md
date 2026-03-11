# Monitoring Service

> **TL;DR** — The monitoring service is the **host-level brain** of Boogle. It boots
> the Docker stack, waits for infrastructure to be healthy, then ticks every 15 s to
> auto-scale workers (spider, indexer, ranking), clean up junk data, and keep the engine
> alive. It exposes a tiny HTTP API for health checks and in-memory metrics.

---

## Table of Contents

1. [Role in the System](#1-role-in-the-system)
2. [High-Level Architecture](#2-high-level-architecture)
3. [Source Layout](#3-source-layout)
4. [Startup Sequence](#4-startup-sequence)
5. [Tick Lifecycle](#5-tick-lifecycle)
6. [Policy Engine](#6-policy-engine)
7. [Workers](#7-workers)
8. [Metrics HTTP API](#8-metrics-http-api)
9. [Configuration Reference](#9-configuration-reference)
10. [Running the Service](#10-running-the-service)
11. [Known Issues & Limitations](#11-known-issues--limitations)

---

## 1. Role in the System

```
                     ┌──────────────────────────────────┐
                     │         HOST MACHINE             │
                     │                                  │
                     │   ┌──────────────────────────┐   │
                     │   │   monitoring  (Node.js)  │   │  ← runs here, NOT in a container
                     │   │   port 7070              │   │
                     │   └──────┬───────────────────┘   │
                     │          │ docker compose API     │
                     │  ┌───────▼──────────────────────┐ │
                     │  │  Docker engine (dockerd)     │ │
                     │  └──┬──┬──┬──┬──┬──────────────┘ │
                     │     │  │  │  │  │                 │
                     │   psql redis spider indexer engine│
                     └──────────────────────────────────┘
```

Unlike every other Boogle service, **monitoring runs directly on the host** (not in a
container). This is intentional — it needs to shell out to `docker compose` in order to
start, stop, and scale sibling containers. Running it in Docker would require mounting
the Docker socket and matching host filesystem paths, which adds complexity (see
[Running the Service](#10-running-the-service) for the containerised option).

---

## 2. High-Level Architecture

```
src/index.ts
  └─ app.ts  (createApp)
       ├─ ensureServicesUp(psql, redis)          ─── docker compose up -d
       ├─ waitForDb / waitForRedis               ─── retry loops
       ├─ ensureMonitorStateTable()              ─── CREATE TABLE IF NOT EXISTS
       ├─ express app  →  GET /metrics, /health
       └─ Scheduler.start()
            │
            └─ every TICK_INTERVAL_MS (default 15 s)
                 ├─ _ensureBaseline()            ─── re-up psql + redis
                 ├─ _tickIndexer()               ─── spawn indexer jobs
                 ├─ _tickSpider()                ─── ensure spider is up
                 ├─ _tickRanking()               ─── trigger ranking jobs
                 ├─ CleanupWorker.run()          ─── clear indexed HTML (every N ticks)
                 └─ EngineGuardian.check()       ─── restart engine if down
```

### Data stores used by monitoring

| Store | Purpose |
|---|---|
| PostgreSQL | Reads `pages` (backlog counts, crawl rate, total indexed); owns `monitor_state` k/v table |
| Redis | Reads `urls` sorted-set cardinality → crawl queue depth |

### monitor_state table

A lightweight key-value table in PostgreSQL that survives process restarts. Used for:

| Key | Value stored |
|---|---|
| `last_indexer_spawn_ms` | `Date.now()` when last indexer job was launched |
| `last_spider_spawn_ms` | `Date.now()` when spider was last started |
| `last_ranking_indexed_count` | `totalIndexedPages` at the time ranking last ran |

> **Note:** `migration/schema.sql` already creates this table; `db/queries.ts` also
> issues `CREATE TABLE IF NOT EXISTS` on startup. Both are safe — the `IF NOT EXISTS`
> guard makes the second call a no-op.

---

## 3. Source Layout

```
services/monitoring/
├── package.json          ESM Node.js project, scripts: build / dev / start
├── tsconfig.json         strict ES2022 + NodeNext modules
├── .env.example          all supported environment variables with defaults
├── src/
│   ├── index.ts          entry point — calls createApp(), wires SIGINT/SIGTERM
│   ├── app.ts            createApp(): boots infra, waits for DB/Redis, starts HTTP + scheduler
│   ├── config/
│   │   ├── schema.ts     Zod schema — validates & types every env var
│   │   └── env.ts        loads .env, resolves MONITORING_ROOT & PROJECT_ROOT, exports config
│   ├── db/
│   │   ├── client.ts     pg.Pool with max=5, connectionString from DATABASE_URL
│   │   ├── queries.ts    SQL helpers: page counts, HTML cleanup, monitor_state CRUD
│   │   └── inspector.ts  collectDbMetrics() — bundles the three page-count queries
│   ├── redis/
│   │   ├── client.ts     redis createClient(), connect/disconnect helpers
│   │   └── inspector.ts  getCrawlQueueSize() → ZCARD "urls"
│   ├── docker/
│   │   ├── compose.ts    composeMap (infra/engine/spider/indexer/ranking paths) + services map
│   │   ├── controller.ts docker compose wrappers: ensureServiceUp, runOneOffJob, ps helpers
│   │   └── lock.ts       AsyncMutex — serialises concurrent docker compose calls
│   ├── orchestration/
│   │   └── backoff.ts    withRetry<T>(fn, {maxAttempts, baseDelayMs, label})
│   ├── policy/
│   │   ├── indexer-policy.ts  pure-function: decideIndexer(input) → {shouldSpawn, reason}
│   │   ├── spider-policy.ts   pure-function: decideSpider(input)  → {shouldSpawn, reason}
│   │   └── ranking-policy.ts  pure-function: decideRanking(input) → {shouldRun, reason}
│   ├── scheduler/
│   │   └── scheduler.ts  Scheduler class — owns the setInterval, drives all tick logic
│   ├── workers/
│   │   ├── cleanup-worker.ts   CleanupWorker — trims `html` column from indexed pages
│   │   └── engine-guardian.ts  EngineGuardian — container check + HTTP probe + restart
│   ├── metrics/
│   │   ├── types.ts      MetricEntry, MetricsSnapshot types
│   │   ├── registry.ts   in-memory store, counter() / gauge() / getSnapshot()
│   │   └── http.ts       Express router: GET /metrics, GET /health
│   └── logger/
│       └── logger.ts     pino with two transports: JSON file (debug+) and pretty console (info+)
```

---

## 4. Startup Sequence

```
1. Load .env + validate all env vars via Zod (exit 1 on bad config)
2. docker compose up -d psql redis          (ensureServicesUp — up to 4 retries, exponential backoff)
3. [optional] docker compose up -d adminer  (if MANAGE_ADMINER=true)
4. Poll SELECT 1 every 3 s, up to 20 attempts  →  DB ready
5. Poll redis.connect() every 3 s, up to 20 attempts  →  Redis ready
6. CREATE TABLE IF NOT EXISTS monitor_state
7. Start Express on MONITOR_PORT (default 7070)
8. Scheduler.start() — runs first tick immediately, then every TICK_INTERVAL_MS
```

If step 4 or 5 times out, the process throws and exits — relying on the host init
system (or Docker restart policy) to retry.

---

## 5. Tick Lifecycle

Each tick runs the following stages **sequentially** (a slow stage blocks the next):

```
_ensureBaseline()
  └─ docker compose up -d psql redis   (idempotent; ensures infra never drifts down)

_tickIndexer()
  ├─ collectDbMetrics()          → unindexedPages, totalIndexedPages, pagesCrawledLastMinute
  ├─ countRunningJobsByImage()   → how many indexer containers are currently up
  ├─ monitorStateGet(last_indexer_spawn_ms)
  ├─ decideIndexer(...)          → policy decision
  └─ [if shouldSpawn] docker compose run -d --rm indexer
       + monitorStateSet(last_indexer_spawn_ms, now)

_tickSpider()
  ├─ collectDbMetrics()          (⚠ called again — see Known Issues)
  ├─ getCrawlQueueSize()         → ZCARD "urls"
  ├─ countRunningJobsByImage()   → how many spider containers are up
  ├─ monitorStateGet(last_spider_spawn_ms)
  ├─ decideSpider(...)           → policy decision
  └─ [if shouldSpawn] docker compose up -d spider
       + monitorStateSet(last_spider_spawn_ms, now)

_tickRanking()
  ├─ collectDbMetrics()          (shares result from _tickIndexer? — NO, separate call)
  ├─ countRunningJobsByImage()   → how many ranking containers are up
  ├─ monitorStateGet(last_ranking_indexed_count)
  ├─ decideRanking(...)          → policy decision
  └─ [if shouldRun] docker compose run -d --rm ranking
       + monitorStateSet(last_ranking_indexed_count, totalIndexedPages)

CleanupWorker.run()   (every CLEANUP_EVERY_N_TICKS ticks, default every 4th)
  └─ UPDATE pages SET html = '' WHERE indexed = true AND html <> '' LIMIT batchSize

EngineGuardian.check()
  ├─ docker ps  →  isContainerRunning(ENGINE_CONTAINER_NAME)
  └─ [if running] HTTP GET ENGINE_HTTP_CHECK_URL with 3 s timeout
       [if not running] _handleDown() → rate-limited restart
```

---

## 6. Policy Engine

All three policies are **pure functions** (no side effects, easy to unit-test):

### Indexer Policy (`indexer-policy.ts`)

Spawn a new one-off indexer container when **all** of:
- `unindexedPages >= INDEXER_BACKLOG_THRESHOLD`
- `runningIndexerJobs < INDEXER_MAX_PARALLEL`
- Cooldown elapsed (`now - lastSpawnedAt >= INDEXER_COOLDOWN_MS`)

Indexer is run as `docker compose run -d --rm indexer` — a one-shot job that exits when
done. Multiple parallel jobs are allowed up to `INDEXER_MAX_PARALLEL`.

### Spider Policy (`spider-policy.ts`)

Ensure spider is running. Spawn when:
- `runningSpiders < SPIDER_MIN_INSTANCES` — always maintain minimum coverage, OR
- `crawlQueueSize >= SPIDER_QUEUE_HIGH_THRESHOLD` **AND**
  `crawlRatePerMinute <= SPIDER_RATE_LOW_THRESHOLD` **AND**
  `runningSpiders < SPIDER_MAX_INSTANCES` **AND**
  cooldown has elapsed

Spider is a **persistent service** (not a one-off job), so it uses
`docker compose up -d spider`.  
⚠ See [Known Issue #1](#1-spider-scaling-not-implemented).

### Ranking Policy (`ranking-policy.ts`)

Trigger ranking when:
- `runningRankingJobs < RANKING_MAX_PARALLEL`
- `totalIndexedPages - lastRankingIndexedCount >= RANKING_TRIGGER_DELTA`

Ranking is run as `docker compose run -d --rm ranking` — a one-shot job like the indexer.

---

## 7. Workers

### CleanupWorker (`workers/cleanup-worker.ts`)

Runs every `CLEANUP_EVERY_N_TICKS` ticks (default every 4th = every ~60 s).

```sql
UPDATE pages
SET html = '', updated_at = NOW()
WHERE id IN (
  SELECT id FROM pages
  WHERE indexed = true AND html <> ''
  LIMIT <CLEANUP_BATCH_SIZE>
)
```

Purpose: the indexer has finished with the raw HTML; keeping it in PostgreSQL wastes
disk. The spider re-writes it if a URL is re-crawled anyway. Clears up to
`CLEANUP_BATCH_SIZE` (default 500) rows per run.

### EngineGuardian (`workers/engine-guardian.ts`)

Runs on every tick.

1. `docker ps` → checks if `ENGINE_CONTAINER_NAME` (`engine`) is in the running list.
2. If running: fires an HTTP GET to `ENGINE_HTTP_CHECK_URL` with a 3-second timeout.
   - Status < 500 → healthy.
   - Status ≥ 500 or timeout → warns but does **not** restart (container is up).
3. If **not** running: calls `_handleDown()`:
   - Prunes restart timestamps older than 5 minutes from the sliding window.
   - If `timestamps.length >= ENGINE_MAX_RESTARTS_PER_5M` → enters **degraded mode**
     (logs error, stops attempting restarts until engine comes back on its own).
   - Otherwise: `docker compose up -d engine`, records restart timestamp.

---

## 8. Metrics HTTP API

The monitoring service exposes two endpoints:

### `GET /health`
```json
{ "status": "ok", "ts": "2026-03-11T14:00:00.000Z" }
```
Always returns 200 while the process is alive.

### `GET /metrics`
```json
{
  "uptime_seconds": 3600,
  "metrics": {
    "scheduler_tick_count":            { "type": "gauge",   "value": 240,  "description": "current scheduler tick number" },
    "indexer_unindexed_backlog":        { "type": "gauge",   "value": 87,   "description": "unindexed page count" },
    "indexer_running_jobs":             { "type": "gauge",   "value": 1,    "description": "active indexer containers" },
    "indexer_jobs_spawned_total":        { "type": "counter", "value": 12,   "description": "total indexer jobs spawned" },
    "spider_queue_size":                 { "type": "gauge",   "value": 423,  "description": "redis crawl queue size" },
    "spider_crawl_rate_per_min":         { "type": "gauge",   "value": 38,   "description": "pages crawled last minute" },
    "spider_running_count":              { "type": "gauge",   "value": 1,    "description": "running spider containers" },
    "spider_instances_started_total":    { "type": "counter", "value": 3,    "description": "total spider instances started" },
    "ranking_running_jobs":              { "type": "gauge",   "value": 0,    "description": "active ranking containers" },
    "ranking_total_indexed":             { "type": "gauge",   "value": 5000, "description": "total indexed pages" },
    "ranking_jobs_triggered_total":      { "type": "counter", "value": 7,    "description": "total ranking jobs triggered" },
    "engine_running":                    { "type": "gauge",   "value": 1,    "description": "1 if engine container is running" },
    "engine_http_alive":                 { "type": "gauge",   "value": 1,    "description": "1 if engine HTTP probe succeeded" },
    "engine_restarts_total":             { "type": "counter", "value": 0,    "description": "total engine restart attempts" },
    "cleanup_last_batch_rows":           { "type": "gauge",   "value": 500,  "description": "rows cleared in last cleanup batch" },
    "cleanup_rows_total":                { "type": "counter", "value": 4500, "description": "total html rows cleared" },
    "scheduler_tick_errors_total":       { "type": "counter", "value": 0,    "description": "total tick errors" }
  }
}
```

Metrics are stored in-memory; they reset on process restart.

---

## 9. Configuration Reference

Copy `.env.example` to `.env` and adjust values.

| Variable | Default | Description |
|---|---|---|
| `MONITOR_PORT` | `7070` | Port for the HTTP metrics/health server |
| `LOG_PATH` | `logs/monitor.log` | Log file path (relative to `services/monitoring/`) |
| `TICK_INTERVAL_MS` | `15000` | Scheduler tick interval in ms |
| `DATABASE_URL` | *(required)* | PostgreSQL connection string, e.g. `postgresql://admin:pass@localhost:5432/se` |
| `REDIS_URL` | *(required)* | Redis URL, e.g. `redis://localhost:6379/1` |
| `ROOT_COMPOSE_FILE` | `../../docker-compose.yml` | Path to root infra compose file (relative to `services/monitoring/` unless absolute) |
| `ENGINE_COMPOSE_FILE` | `../engine/docker-compose.yml` | Path to engine compose file |
| `SPIDER_COMPOSE_FILE` | `../spider/docker-compose.yml` | Path to spider compose file |
| `INDEXER_COMPOSE_FILE` | `../indexer/docker-compose.yml` | Path to indexer compose file |
| `RANKING_COMPOSE_FILE` | `../ranking/docker-compose.yml` | Path to ranking compose file |
| `ENGINE_CONTAINER_NAME` | `engine` | Name of the engine Docker container |
| `ENGINE_HTTP_CHECK_URL` | `http://localhost:1323/` | URL used for engine liveness probe |
| `ENGINE_RESTART_BACKOFF_MS` | `5000` | (reserved — not actively used in current code) |
| `ENGINE_MAX_RESTARTS_PER_5M` | `5` | Max engine container restarts in any 5-minute window |
| `INDEXER_BACKLOG_THRESHOLD` | `100` | Min unindexed pages before spawning an indexer job |
| `INDEXER_MAX_PARALLEL` | `3` | Max parallel indexer containers |
| `INDEXER_COOLDOWN_MS` | `20000` | Minimum ms between indexer spawns |
| `SPIDER_MIN_INSTANCES` | `1` | Min spider containers to keep alive |
| `SPIDER_MAX_INSTANCES` | `4` | Max spider containers (⚠ see Known Issues) |
| `SPIDER_RATE_LOW_THRESHOLD` | `20` | Pages/min below which crawl rate is considered "low" |
| `SPIDER_QUEUE_HIGH_THRESHOLD` | `500` | Queue depth above which scale-up is considered |
| `SPIDER_COOLDOWN_MS` | `30000` | Minimum ms between spider starts |
| `RANKING_TRIGGER_DELTA` | `50` | Newly indexed pages needed since last ranking run |
| `RANKING_MAX_PARALLEL` | `1` | Max parallel ranking containers |
| `CLEANUP_BATCH_SIZE` | `500` | Max rows to clear per cleanup pass |
| `CLEANUP_EVERY_N_TICKS` | `4` | Run cleanup every Nth tick |
| `MANAGE_ADMINER` | `false` | If `true`, ensures the adminer container is up |

---

## 10. Running the Service

### Recommended: host mode

```bash
# from project root
cd services/monitoring

cp .env.example .env
# --- edit .env: set DATABASE_URL, REDIS_URL, credentials ---

npm install
npm run build
npm start
# or for hot-reload during development:
npm run dev
```

The process exits cleanly on SIGINT / SIGTERM. Existing Docker containers (psql, redis,
spider, indexer, engine, ranking) **continue running** after the monitor exits.

### Alternative: containerised mode

> ⚠  Containerising the monitoring service has important constraints — read the
> notes in `services/monitoring/docker-compose.yml` before proceeding.

```bash
# 1. Export the absolute project root so that Docker can mount it at the same path
#    as it exists on the host (required for docker compose build contexts to resolve)
export BOOGLE_PROJECT_ROOT=$(git rev-parse --show-toplevel)

# 2. Build and start
cd services/monitoring
cp .env.example .env
# edit .env as above

docker compose up -d

# 3. Follow logs
docker compose logs -f monitoring
```

Key constraints for containerised mode:

- **Docker socket mount:** The container communicates with the host Docker daemon via
  `/var/run/docker.sock`. This grants the container full Docker control — treat it
  accordingly (do not expose the monitoring port to the public internet).
- **Same-path project mount:** The project root must be mounted at its exact host path
  inside the container. Docker Compose resolves build contexts relative to compose files
  on the **host** filesystem; if the paths differ, `docker compose build` inside the
  container will fail.
- **Host networking:** `network_mode: host` is used so that
  `localhost:5432` / `localhost:6379` / `localhost:1323` resolve to the host services
  the same way they do in host mode.

---

## 11. Known Issues & Limitations

### 1. Spider scaling not implemented

**File:** [`src/scheduler/scheduler.ts`](../services/monitoring/src/scheduler/scheduler.ts) — `_tickSpider()`

The spider policy supports 1–`SPIDER_MAX_INSTANCES` instances. However,
`_tickSpider()` calls `ensureServiceUp()`, which maps to:

```
docker compose -f spider/docker-compose.yml up -d spider
```

`up -d` without `--scale` always targets **exactly 1 replica**. If 2 spiders are
already running (e.g. started manually), the next tick would scale it **down** to 1.
The `SPIDER_MAX_INSTANCES` guard in the policy is therefore never actually reached.

**Effect:** The monitoring service can only ever maintain a single spider instance.

**Fix (not applied — code change required):**
```typescript
// Instead of ensureServiceUp, use:
await runDockerCompose(services.spider.compose, [
  "up", "-d", "--scale", `spider=${targetCount}`,
]);
```

---

### 2. `collectDbMetrics()` called twice per tick

**File:** [`src/scheduler/scheduler.ts`](../services/monitoring/src/scheduler/scheduler.ts)

`_tickIndexer()` and `_tickSpider()` each call `collectDbMetrics()` independently
within the same tick. This executes **6 PostgreSQL queries per tick** (3 + 3) instead
of the necessary 3. The results from `_tickIndexer()` are not reused by `_tickSpider()`.

**Effect:** Doubled read load on PostgreSQL; minor at current scale.

---

### 3. Redis client not recreated on reconnect

**File:** [`src/redis/client.ts`](../services/monitoring/src/redis/client.ts)

`redisClient` is a module-level singleton created once at import time. The comment
`// Create a fresh client each attempt so it's not in an error state` appears in
`connectRedis()`, but no fresh client is actually created — the same instance is reused.
If the initial `connect()` fails and leaves the client in an error state, subsequent
`connectRedis()` calls retry `.connect()` on the same broken object.

**Effect:** In practice, the `waitForRedis()` retry loop in `app.ts` catches errors and
retries, so startup eventually succeeds. The stale-client risk surfaces only if Redis
drops mid-operation (the `error` event handler logs it, but reconnection is handled by
the Redis client library automatically in most cases).

---

### 4. `_ensureBaseline()` fires on every tick

**File:** [`src/scheduler/scheduler.ts`](../services/monitoring/src/scheduler/scheduler.ts)

Every tick (default every 15 s), the scheduler calls:
```
docker compose up -d psql redis
```
`up -d` is idempotent but still invokes the Docker API, generates compose logs, and
acquires the `dockerMutex`. This is intentional as a safety net (ensures infra is never
inadvertently stopped), but it is noisy and slightly slows down each tick.

---

### 5. No stack teardown on monitor shutdown

When the monitoring process exits (SIGINT / SIGTERM), the shutdown sequence stops the
scheduler, closes the DB pool, and disconnects from Redis. It does **not** stop any
Docker containers. The Spider, Indexer, Ranking, Engine, PostgreSQL, and Redis
containers keep running.

**This is by design** — the search stack should outlive the monitoring process.
However, it is worth noting that restarting the monitor does not disturb any crawling
or indexing in progress.

---

### 6. `monitor_state` table created redundantly

**Files:** [`migration/schema.sql`](../migration/schema.sql),
[`src/db/queries.ts`](../services/monitoring/src/db/queries.ts)

The `monitor_state` table is declared in `schema.sql` (applied by PostgreSQL's
`docker-entrypoint-initdb.d` mechanism on first start) **and** created again by
`ensureMonitorStateTable()` at monitoring startup via `CREATE TABLE IF NOT EXISTS`. The
second call is a no-op, so this is harmless, but it creates ambiguity about who owns
the table schema.

---

### 7. Engine HTTP probe uses plain `http` module

**File:** [`src/workers/engine-guardian.ts`](../services/monitoring/src/workers/engine-guardian.ts)

The `_httpProbe()` method uses Node.js's built-in `http` module directly. If the engine
is ever configured with HTTPS, the probe would need to be updated to use the `https`
module (or a library that handles both).
