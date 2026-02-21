docker compose -f docker/postgres/docker-compose.postgres-replica.yml down
docker compose -f docker/postgres/docker-compose.postgres-replica.yml up -d
docker compose -f docker/postgres/docker-compose.postgres-replica.yml logs -f   

   