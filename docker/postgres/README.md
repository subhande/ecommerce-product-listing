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

or 

```bash
sh docker/postgres/postgres-primary-replica.sh
```

or 
```bash
sh docker/postgres/postgres-compose.sh
```