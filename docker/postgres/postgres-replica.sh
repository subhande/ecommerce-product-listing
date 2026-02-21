#!/usr/bin/env bash
set -euo pipefail

PRIMARY_CONTAINER="${PRIMARY_CONTAINER:-ecommerce-postgres}"
REPLICA_CONTAINER="${REPLICA_CONTAINER:-ecommerce-postgres-replica}"
NETWORK_NAME="${NETWORK_NAME:-ecommerce-postgres-net}"
PRIMARY_VOLUME="${PRIMARY_VOLUME:-ecommerce-postgres-data}"
REPLICA_VOLUME="${REPLICA_VOLUME:-ecommerce-postgres-replica-data}"
PRIMARY_PORT="${PRIMARY_PORT:-5432}"
REPLICA_PORT="${REPLICA_PORT:-5433}"
POSTGRES_IMAGE="${POSTGRES_IMAGE:-postgres:18.2}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
POSTGRES_DB="${POSTGRES_DB:-ecommerce}"
REPLICATION_USER="${REPLICATION_USER:-replicator}"
REPLICATION_PASSWORD="${REPLICATION_PASSWORD:-replicator_pass}"
REPLICA_DATA_DIR="${REPLICA_DATA_DIR:-postgres-replica-data}"
SYNC_STATE="${SYNC_STATE:-async}"
PRIMARY_DATA_DIR_IN_CONTAINER="/var/lib/postgresql/18/docker"

if [ -n "${SYNC_STATE}" ]; then
  SYNC_STATE="$(printf '%s' "${SYNC_STATE}" | tr '[:upper:]' '[:lower:]')"
  case "${SYNC_STATE}" in
    async|sync|potential|quorum) ;;
    *)
      echo "Invalid SYNC_STATE='${SYNC_STATE}'. Allowed: async, sync, potential, quorum." >&2
      exit 1
      ;;
  esac
fi

wait_for_postgres() {
  local container="$1"
  local retries=60
  local i
  for ((i = 1; i <= retries; i++)); do
    if docker exec "${container}" pg_isready -U "${POSTGRES_USER}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "Postgres in container '${container}' did not become ready in time." >&2
  return 1
}

echo "Preparing replica data directory..."
mkdir -p "${REPLICA_DATA_DIR}"

echo "Creating network (if needed): ${NETWORK_NAME}"
docker network inspect "${NETWORK_NAME}" >/dev/null 2>&1 || docker network create "${NETWORK_NAME}"

echo "Connecting primary container to network (if needed): ${PRIMARY_CONTAINER}"
docker network connect "${NETWORK_NAME}" "${PRIMARY_CONTAINER}" >/dev/null 2>&1 || true

echo "Ensuring replication user exists on primary..."
docker exec "${PRIMARY_CONTAINER}" psql -U "${POSTGRES_USER}" -v ON_ERROR_STOP=1 -c \
  "DO \$\$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='${REPLICATION_USER}') THEN CREATE ROLE ${REPLICATION_USER} WITH REPLICATION LOGIN PASSWORD '${REPLICATION_PASSWORD}'; END IF; END \$\$;"

echo "Ensuring pg_hba replication rule exists on primary..."
docker exec "${PRIMARY_CONTAINER}" sh -lc \
  "grep -q \"host replication ${REPLICATION_USER} 0.0.0.0/0 md5\" ${PRIMARY_DATA_DIR_IN_CONTAINER}/pg_hba.conf || echo \"host replication ${REPLICATION_USER} 0.0.0.0/0 md5\" >> ${PRIMARY_DATA_DIR_IN_CONTAINER}/pg_hba.conf"

echo "Reloading primary config..."
docker exec "${PRIMARY_CONTAINER}" psql -U "${POSTGRES_USER}" -c "SELECT pg_reload_conf();"

echo "Removing existing replica container/volume..."
docker rm -f "${REPLICA_CONTAINER}" >/dev/null 2>&1 || true
docker volume rm -f "${REPLICA_VOLUME}" >/dev/null 2>&1 || true

echo "Cleaning local replica data directory..."
find "${REPLICA_DATA_DIR}" -mindepth 1 -maxdepth 1 -exec rm -rf {} +

echo "Creating replica volume..."
docker volume create \
  --name "${REPLICA_VOLUME}" \
  --opt type=none \
  --opt device="$(pwd)/${REPLICA_DATA_DIR}" \
  --opt o=bind >/dev/null

echo "Taking base backup from primary..."
docker run --rm \
  --network "${NETWORK_NAME}" \
  -e PGPASSWORD="${REPLICATION_PASSWORD}" \
  -v "${REPLICA_VOLUME}:/var/lib/postgresql" \
  "${POSTGRES_IMAGE}" \
  sh -lc "mkdir -p /var/lib/postgresql/18/docker && pg_basebackup -h ${PRIMARY_CONTAINER} -U ${REPLICATION_USER} -D /var/lib/postgresql/18/docker -Fp -Xs -P -R"

echo "Starting replica container..."
docker run --name "${REPLICA_CONTAINER}" \
  --network "${NETWORK_NAME}" \
  -e POSTGRES_PASSWORD="${POSTGRES_PASSWORD}" \
  -e POSTGRES_USER="${POSTGRES_USER}" \
  -e POSTGRES_DB="${POSTGRES_DB}" \
  -p "${REPLICA_PORT}:5432" \
  -v "${REPLICA_VOLUME}:/var/lib/postgresql" \
  -d "${POSTGRES_IMAGE}" >/dev/null

echo "Waiting for replica to become ready..."
wait_for_postgres "${REPLICA_CONTAINER}"

echo "Validating replica recovery mode..."
docker exec "${REPLICA_CONTAINER}" psql -U "${POSTGRES_USER}" -tAc "SELECT pg_is_in_recovery();"

echo "Validating replication on primary..."
REPLICATION_QUERY="SELECT client_addr, state, sync_state, write_lsn, flush_lsn, replay_lsn FROM pg_stat_replication"
if [ -n "${SYNC_STATE}" ]; then
  echo "Applying replication filter: sync_state='${SYNC_STATE}'"
  REPLICATION_QUERY="${REPLICATION_QUERY} WHERE sync_state = '${SYNC_STATE}'"
fi
REPLICATION_QUERY="${REPLICATION_QUERY};"

docker exec "${PRIMARY_CONTAINER}" psql -U "${POSTGRES_USER}" -x -c "${REPLICATION_QUERY}"

echo "Replica setup complete."
echo "Primary: ${PRIMARY_CONTAINER} on ${PRIMARY_PORT}"
echo "Replica: ${REPLICA_CONTAINER} on ${REPLICA_PORT}"
