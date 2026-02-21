#!/usr/bin/env bash
set -euo pipefail

DATA_DIR="/var/lib/postgresql/18/docker"
PRIMARY_HOST="${PRIMARY_HOST:-ecommerce-postgres}"
REPLICATION_USER="${REPLICATION_USER:-replicator}"
REPLICATION_PASSWORD="${REPLICATION_PASSWORD:-replicator_pass}"

if [ ! -s "${DATA_DIR}/PG_VERSION" ]; then
  echo "Replica data not found. Taking base backup from primary..."
  mkdir -p "${DATA_DIR}"
  find "${DATA_DIR}" -mindepth 1 -maxdepth 1 -exec rm -rf {} +

  export PGPASSWORD="${REPLICATION_PASSWORD}"
  until pg_isready -h "${PRIMARY_HOST}" -U "${REPLICATION_USER}" >/dev/null 2>&1; do
    sleep 1
  done

  pg_basebackup \
    -h "${PRIMARY_HOST}" \
    -U "${REPLICATION_USER}" \
    -D "${DATA_DIR}" \
    -Fp \
    -Xs \
    -P \
    -R
fi

exec docker-entrypoint.sh postgres
