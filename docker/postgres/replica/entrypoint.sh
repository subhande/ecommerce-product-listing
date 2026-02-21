#!/usr/bin/env bash
set -e

# Custom entrypoint for the replica container.
# 1) Waits for the primary to be fully ready.
# 2) Runs pg_basebackup to seed the replica data directory.
# 3) Hands off to the standard postgres entrypoint.

PRIMARY_HOST="${PRIMARY_HOST:-postgres-primary}"
REPLICATION_USER="${REPLICATION_USER:-replicator}"
REPLICATION_PASSWORD="${REPLICATION_PASSWORD:-replicator_pass}"
PGDATA="${PGDATA:-/var/lib/postgresql/data}"

# ── Wait for primary ──────────────────────────────────────────────
echo "Replica: waiting for primary at '${PRIMARY_HOST}:5432' ..."
until PGPASSWORD="${REPLICATION_PASSWORD}" pg_isready -h "${PRIMARY_HOST}" -U "${REPLICATION_USER}" -d postgres -q 2>/dev/null; do
  sleep 1
done
echo "Replica: primary is ready."

# ── Bootstrap with pg_basebackup (only if PGDATA is empty) ───────
if [ -z "$(ls -A "${PGDATA}" 2>/dev/null)" ]; then
  echo "Replica: PGDATA is empty — running pg_basebackup ..."
  PGPASSWORD="${REPLICATION_PASSWORD}" pg_basebackup \
    -h "${PRIMARY_HOST}" \
    -U "${REPLICATION_USER}" \
    -D "${PGDATA}" \
    -Fp -Xs -P -R

  # Ensure correct ownership
  chown -R postgres:postgres "${PGDATA}"
  chmod 0700 "${PGDATA}"
  echo "Replica: base backup complete."
else
  echo "Replica: PGDATA already populated — skipping pg_basebackup."
fi

# ── Start Postgres via the official entrypoint ────────────────────
exec docker-entrypoint.sh postgres
