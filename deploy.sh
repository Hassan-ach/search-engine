#!/usr/bin/env bash
# deploy.sh — deploy the full Boogle stack on a VPS
# Usage: ./deploy.sh [--pull] [--build] [--restart] [--help]
#
# Environment variable overrides:
#   BOOGLE_PROJECT_ROOT   Absolute path to this repo (default: directory containing this script)

set -euo pipefail

# ── Colours ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GRN='\033[0;32m'
YLW='\033[1;33m'
CYN='\033[0;36m'
BLD='\033[1m'
RST='\033[0m'

log()  { echo -e "${CYN}[boogle]${RST} $*"; }
ok()   { echo -e "${GRN}[  ok  ]${RST} $*"; }
warn() { echo -e "${YLW}[ warn ]${RST} $*"; }
die()  { echo -e "${RED}[error ]${RST} $*" >&2; exit 1; }

# ── CLI flags ─────────────────────────────────────────────────────────────────
OPT_PULL=false
OPT_BUILD=false
OPT_RESTART=false

usage() {
  cat <<EOF
Usage: $(basename "$0") [options]

Options:
  --pull      git pull latest changes before deploying
  --build     force rebuild of all Docker images (--no-cache)
  --restart   bring all services down before starting (full restart)
  --help      show this help

By default the script does an incremental deploy:
  - Creates missing .env files from .env.example
  - Creates the Docker network if absent
  - Pulls Docker images for infra (postgres, redis, adminer)
  - Builds application images
  - Starts / updates each service with docker compose up -d --build

Set BOOGLE_PROJECT_ROOT to override the auto-detected project root.
EOF
}

for arg in "$@"; do
  case "$arg" in
    --pull)    OPT_PULL=true ;;
    --build)   OPT_BUILD=true ;;
    --restart) OPT_RESTART=true ;;
    --help)    usage; exit 0 ;;
    *) die "Unknown option: $arg  (run with --help)" ;;
  esac
done

# ── Directories ───────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="${BOOGLE_PROJECT_ROOT:-$SCRIPT_DIR}"
export BOOGLE_PROJECT_ROOT="$ROOT"

SVC="$ROOT/services"

# ── Prerequisites ─────────────────────────────────────────────────────────────
log "Checking prerequisites…"

for cmd in docker git; do
  command -v "$cmd" &>/dev/null || die "'$cmd' is not installed."
done

# Docker Compose V2 (docker compose subcommand)
docker compose version &>/dev/null || die "'docker compose' plugin not found. Install Docker Compose V2."

ok "Prerequisites satisfied"

# ── Optional git pull ─────────────────────────────────────────────────────────
if $OPT_PULL; then
  log "Pulling latest changes…"
  git -C "$ROOT" pull --ff-only || die "git pull failed. Resolve conflicts manually."
  ok "Repository up to date"
fi

# ── .env bootstrap ───────────────────────────────────────────────────────────
# Copy .env.example → .env when the target does not yet exist.
bootstrap_env() {
  local dir="$1"
  local example="$dir/.env.example"
  local target="$dir/.env"

  if [[ ! -f "$example" ]]; then
    return
  fi

  if [[ ! -f "$target" ]]; then
    cp "$example" "$target"
    warn "$target created from .env.example — review and set real secrets before relying on it"
  fi
}

log "Bootstrapping .env files…"
bootstrap_env "$ROOT"
bootstrap_env "$SVC/engine"
bootstrap_env "$SVC/spider"
bootstrap_env "$SVC/indexer"
bootstrap_env "$SVC/ranking"
bootstrap_env "$SVC/monitoring"
ok ".env files ready"

# ── Fix monitoring .env compose paths to absolute paths ───────────────────────
# The .env.example uses relative paths suited for local dev.
# On the VPS those paths must point to BOOGLE_PROJECT_ROOT.
MONITORING_ENV="$SVC/monitoring/.env"
if grep -q '^\(ROOT\|ENGINE\|SPIDER\|INDEXER\|RANKING\)_COMPOSE_FILE=\.\.' "$MONITORING_ENV" 2>/dev/null; then
  log "Rewriting relative compose paths in monitoring .env to absolute paths…"
  sed -i \
    -e "s|ROOT_COMPOSE_FILE=.*|ROOT_COMPOSE_FILE=${ROOT}/docker-compose.yml|" \
    -e "s|ENGINE_COMPOSE_FILE=.*|ENGINE_COMPOSE_FILE=${SVC}/engine/docker-compose.yml|" \
    -e "s|SPIDER_COMPOSE_FILE=.*|SPIDER_COMPOSE_FILE=${SVC}/spider/docker-compose.yml|" \
    -e "s|INDEXER_COMPOSE_FILE=.*|INDEXER_COMPOSE_FILE=${SVC}/indexer/docker-compose.yml|" \
    -e "s|RANKING_COMPOSE_FILE=.*|RANKING_COMPOSE_FILE=${SVC}/ranking/docker-compose.yml|" \
    "$MONITORING_ENV"
  ok "Compose paths updated"
fi

# ── Docker network ────────────────────────────────────────────────────────────
NETWORK="search_net"
if ! docker network inspect "$NETWORK" &>/dev/null; then
  log "Creating Docker network '$NETWORK'…"
  docker network create "$NETWORK"
  ok "Network created"
else
  ok "Network '$NETWORK' already exists"
fi

# ── Build flag ────────────────────────────────────────────────────────────────
BUILD_FLAG="--build"
$OPT_BUILD && BUILD_FLAG="--build --no-cache"

# ── Restart: tear everything down first ───────────────────────────────────────
if $OPT_RESTART; then
  warn "Stopping all services (--restart)…"
  docker compose -f "$ROOT/docker-compose.yml"         down --remove-orphans 2>/dev/null || true
  docker compose -f "$SVC/engine/docker-compose.yml"   down --remove-orphans 2>/dev/null || true
  docker compose -f "$SVC/spider/docker-compose.yml"   down --remove-orphans 2>/dev/null || true
  docker compose -f "$SVC/indexer/docker-compose.yml"  down --remove-orphans 2>/dev/null || true
  docker compose -f "$SVC/ranking/docker-compose.yml"  down --remove-orphans 2>/dev/null || true
  docker compose -f "$SVC/monitoring/docker-compose.yml" \
    --env-file "$MONITORING_ENV" down --remove-orphans 2>/dev/null || true
  ok "All services stopped"
fi

# ── Helper: bring up a compose project ────────────────────────────────────────
compose_up() {
  local label="$1"
  local file="$2"
  shift 2
  log "Starting ${label}…"
  # shellcheck disable=SC2068
  docker compose -f "$file" $@ up -d $BUILD_FLAG
  ok "${label} running"
}

# ── 1. Infrastructure: postgres + redis + adminer ─────────────────────────────
log "Pulling infra images…"
docker compose -f "$ROOT/docker-compose.yml" pull --quiet
compose_up "infrastructure (postgres, redis, adminer)" "$ROOT/docker-compose.yml"

# Give postgres a moment to finish its init (only matters on first boot)
if ! docker exec psql pg_isready -q 2>/dev/null; then
  log "Waiting for postgres to become ready…"
  for i in $(seq 1 30); do
    sleep 2
    docker exec psql pg_isready -q 2>/dev/null && break
    [[ $i -eq 30 ]] && die "Postgres did not become ready in 60 seconds."
  done
fi
ok "Postgres is ready"

# ── 2. Engine ─────────────────────────────────────────────────────────────────
compose_up "engine" "$SVC/engine/docker-compose.yml"

# ── 3. Spider ─────────────────────────────────────────────────────────────────
compose_up "spider" "$SVC/spider/docker-compose.yml"

# ── 4. Indexer ────────────────────────────────────────────────────────────────
compose_up "indexer" "$SVC/indexer/docker-compose.yml"

# ── 5. Ranking ────────────────────────────────────────────────────────────────
compose_up "ranking" "$SVC/ranking/docker-compose.yml"

# ── 6. Monitoring ─────────────────────────────────────────────────────────────
log "Starting monitoring…"
BOOGLE_PROJECT_ROOT="$ROOT" \
  docker compose \
    -f "$SVC/monitoring/docker-compose.yml" \
    --env-file "$MONITORING_ENV" \
  up -d $BUILD_FLAG
ok "Monitoring running"

# ── Status summary ────────────────────────────────────────────────────────────
echo ""
echo -e "${BLD}── Container status ──────────────────────────────────────────${RST}"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" \
  --filter "network=$NETWORK" \
  --filter "name=monitoring"
echo ""
echo -e "${GRN}${BLD}Deploy complete.${RST}"
echo -e "  Dashboard : ${CYN}http://<VPS_IP>:7070${RST}"
echo -e "  Engine    : ${CYN}http://<VPS_IP>:1323${RST}"
echo -e "  Adminer   : ${CYN}http://<VPS_IP>:8080${RST}"
echo ""
echo -e "Logs: ${YLW}docker compose -f services/monitoring/docker-compose.yml logs -f${RST}"
