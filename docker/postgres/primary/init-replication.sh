#!/usr/bin/env bash
set -e

# This script runs inside /docker-entrypoint-initdb.d/ on the PRIMARY.
# It creates the replication user, appends a pg_hba rule, and configures
# synchronous replication based on the SYNC_STATE env var.
#
# SYNC_STATE values:
#   async     – (default) all standbys are asynchronous
#   sync      – one standby is fully synchronous  (FIRST 1)
#   potential  – same as sync; extra replicas appear as "potential"
#   quorum    – quorum-based synchronous commit    (ANY 1)

SYNC_STATE="${SYNC_STATE:-async}"

# ── 1. Create the replication role ────────────────────────────────
psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" <<-SQL
  DO \$\$
  BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${REPLICATION_USER}') THEN
      CREATE ROLE ${REPLICATION_USER} WITH REPLICATION LOGIN PASSWORD '${REPLICATION_PASSWORD}';
    END IF;
  END
  \$\$;
SQL

# ── 2. Allow replication connections from any host ────────────────
HBA_ENTRY="host replication ${REPLICATION_USER} 0.0.0.0/0 md5"
if ! grep -q "${HBA_ENTRY}" "${PGDATA}/pg_hba.conf"; then
  echo "${HBA_ENTRY}" >> "${PGDATA}/pg_hba.conf"
fi

# ── 3. Configure synchronous replication if needed ────────────────
SYNC_STATE="$(printf '%s' "${SYNC_STATE}" | tr '[:upper:]' '[:lower:]')"

case "${SYNC_STATE}" in
  async)
    echo "SYNC_STATE=async — no synchronous_standby_names needed."
    ;;
  sync|potential)
    echo "SYNC_STATE=${SYNC_STATE} — setting synchronous_standby_names = 'FIRST 1 (*)'."
    psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" <<-SQL
      ALTER SYSTEM SET synchronous_standby_names = 'FIRST 1 (*)';
      ALTER SYSTEM SET synchronous_commit = 'on';
SQL
    ;;
  quorum)
    echo "SYNC_STATE=quorum — setting synchronous_standby_names = 'ANY 1 (*)'."
    psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" <<-SQL
      ALTER SYSTEM SET synchronous_standby_names = 'ANY 1 (*)';
      ALTER SYSTEM SET synchronous_commit = 'on';
SQL
    ;;
  *)
    echo "WARNING: Unknown SYNC_STATE='${SYNC_STATE}'. Falling back to async." >&2
    ;;
esac

echo "Primary init-replication: replication user '${REPLICATION_USER}' created, pg_hba.conf updated, SYNC_STATE=${SYNC_STATE}."
