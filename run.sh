#!/usr/bin/env bash
set -euo pipefail

PORT="${PORT:-8080}"

# Use all available CPU cores unless GOMAXPROCS is already provided.
if [[ -z "${GOMAXPROCS:-}" ]]; then
  if command -v getconf >/dev/null 2>&1; then
    export GOMAXPROCS="$(getconf _NPROCESSORS_ONLN)"
  elif command -v nproc >/dev/null 2>&1; then
    export GOMAXPROCS="$(nproc)"
  else
    export GOMAXPROCS="4"
  fi
fi

# Kill anything running on the configured port (no sudo required).
pids="$(lsof -ti tcp:"${PORT}" || true)"
if [[ -n "${pids}" ]]; then
  kill -9 ${pids}
  echo "Killed existing process(es) on port ${PORT}: ${pids}"
else
  echo "No process is using port ${PORT}"
fi

echo "Starting server with GOMAXPROCS=${GOMAXPROCS} on port ${PORT}"
go run main.go

# GOMAXPROCS=12 PORT=8080 bash run.sh

# GOMAXPROCS=4 PORT=8080 bash run.sh