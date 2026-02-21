# Run Postgres in Docker


# Container without replica

```bash
sh docker/postgres/postgress.sh
```

# Container with replica

```bash
sh docker/postgres/postgress.sh
sh docker/postgres/postgres-replica.sh
```

Filter replication status by sync mode (optional):
```bash
SYNC_STATE=sync sh docker/postgres/postgres-replica.sh
```

or 

```bash
sh docker/postgres/postgres-primary-replica.sh
```

or with filter:
```bash
SYNC_STATE=async sh docker/postgres/postgres-primary-replica.sh
```

or 
```bash
sh docker/postgres/postgres-compose.sh
```

with synchronous replication mode:
```bash
REPLICA_SYNC_MODE=sync sh docker/postgres/postgres-compose.sh
```

with quorum sync mode:
```bash
REPLICA_SYNC_MODE=quorum SYNC_REPLICA_COUNT=1 sh docker/postgres/postgres-compose.sh
```

Available `REPLICA_SYNC_MODE` values:
- `async` (default)
- `sync`
- `quorum`
- `remote_write`
- `remote_apply`
