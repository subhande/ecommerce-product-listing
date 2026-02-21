#!/usr/bin/env bash
set -euo pipefail

# Configurable defaults; override with environment variables if needed.
PRIMARY_CONTAINER="${PRIMARY_CONTAINER:-ecommerce-postgres}"
REPLICA_CONTAINER="${REPLICA_CONTAINER:-ecommerce-postgres-replica}"
NETWORK_NAME="${NETWORK_NAME:-ecommerce-postgres-net}"
PRIMARY_VOLUME="${PRIMARY_VOLUME:-ecommerce-postgres-data}"
REPLICA_VOLUME="${REPLICA_VOLUME:-ecommerce-postgres-replica-data}"
PRIMARY_DATA_DIR="${PRIMARY_DATA_DIR:-postgres-data}"
REPLICA_DATA_DIR="${REPLICA_DATA_DIR:-postgres-replica-data}"
PRIMARY_PORT="${PRIMARY_PORT:-5432}"
REPLICA_PORT="${REPLICA_PORT:-5433}"
POSTGRES_IMAGE="${POSTGRES_IMAGE:-postgres:18.2}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
POSTGRES_DB="${POSTGRES_DB:-ecommerce}"
REPLICATION_USER="${REPLICATION_USER:-replicator}"
REPLICATION_PASSWORD="${REPLICATION_PASSWORD:-replicator_pass}"

# Wait until Postgres responds to health checks in a given container.
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

# 1) Reset local state and Docker resources used by primary/replica.
echo "Preparing local data directories..."
mkdir -p "${PRIMARY_DATA_DIR}" "${REPLICA_DATA_DIR}"

echo "Removing existing containers..."
docker rm -f "${REPLICA_CONTAINER}" >/dev/null 2>&1 || true
docker rm -f "${PRIMARY_CONTAINER}" >/dev/null 2>&1 || true

echo "Removing existing Docker volumes..."
docker volume rm -f "${REPLICA_VOLUME}" >/dev/null 2>&1 || true
docker volume rm -f "${PRIMARY_VOLUME}" >/dev/null 2>&1 || true

echo "Cleaning local data directories..."
find "${PRIMARY_DATA_DIR}" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
find "${REPLICA_DATA_DIR}" -mindepth 1 -maxdepth 1 -exec rm -rf {} +

# 2) Recreate bind-mounted volumes and shared network.
echo "Creating primary and replica bind volumes..."
docker volume create \
  --name "${PRIMARY_VOLUME}" \
  --opt type=none \
  --opt device="$(pwd)/${PRIMARY_DATA_DIR}" \
  --opt o=bind >/dev/null

docker volume create \
  --name "${REPLICA_VOLUME}" \
  --opt type=none \
  --opt device="$(pwd)/${REPLICA_DATA_DIR}" \
  --opt o=bind >/dev/null

echo "Creating network (if needed): ${NETWORK_NAME}"
docker network inspect "${NETWORK_NAME}" >/dev/null 2>&1 || docker network create "${NETWORK_NAME}" >/dev/null

# 3) Start primary and wait for readiness.
echo "Starting primary container..."
docker run --name "${PRIMARY_CONTAINER}" \
  --network "${NETWORK_NAME}" \
  -e POSTGRES_PASSWORD="${POSTGRES_PASSWORD}" \
  -e POSTGRES_USER="${POSTGRES_USER}" \
  -e POSTGRES_DB="${POSTGRES_DB}" \
  -p "${PRIMARY_PORT}:5432" \
  -v "${PRIMARY_VOLUME}:/var/lib/postgresql" \
  -d "${POSTGRES_IMAGE}" >/dev/null

echo "Waiting for primary to become ready..."
wait_for_postgres "${PRIMARY_CONTAINER}"

# 4) Configure primary for replication connections.
echo "Creating replication user on primary..."
docker exec "${PRIMARY_CONTAINER}" psql -U "${POSTGRES_USER}" -v ON_ERROR_STOP=1 -c \
  "DO \$\$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='${REPLICATION_USER}') THEN CREATE ROLE ${REPLICATION_USER} WITH REPLICATION LOGIN PASSWORD '${REPLICATION_PASSWORD}'; END IF; END \$\$;"

PRIMARY_DB_DATA_DIR="$(docker exec "${PRIMARY_CONTAINER}" psql -U "${POSTGRES_USER}" -tAc "show data_directory;" | tr -d '[:space:]')"
if [ -z "${PRIMARY_DB_DATA_DIR}" ]; then
  echo "Unable to detect primary data_directory." >&2
  exit 1
fi

echo "Ensuring replication rule exists in pg_hba.conf..."
docker exec "${PRIMARY_CONTAINER}" sh -lc \
  "grep -q \"host replication ${REPLICATION_USER} 0.0.0.0/0 md5\" ${PRIMARY_DB_DATA_DIR}/pg_hba.conf || echo \"host replication ${REPLICATION_USER} 0.0.0.0/0 md5\" >> ${PRIMARY_DB_DATA_DIR}/pg_hba.conf"

docker exec "${PRIMARY_CONTAINER}" psql -U "${POSTGRES_USER}" -c "SELECT pg_reload_conf();" >/dev/null

# 5) Bootstrap replica from a base backup and start it.
echo "Bootstrapping replica with pg_basebackup..."
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

# 6) Print replication health checks.
echo "Replica recovery mode:"
docker exec "${REPLICA_CONTAINER}" psql -U "${POSTGRES_USER}" -tAc "SELECT pg_is_in_recovery();"

echo "Primary replication status:"
docker exec "${PRIMARY_CONTAINER}" psql -U "${POSTGRES_USER}" -x -c \
  "SELECT client_addr, state, sync_state, write_lsn, flush_lsn, replay_lsn FROM pg_stat_replication;"

echo "Primary and replica setup complete."
echo "Primary: ${PRIMARY_CONTAINER} on port ${PRIMARY_PORT}"
echo "Replica: ${REPLICA_CONTAINER} on port ${REPLICA_PORT}"
