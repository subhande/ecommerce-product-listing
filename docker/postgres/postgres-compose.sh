
# Down docker/postgres/docker-compose.postgres.replica.yml
docker-compose -f docker-compose.postgres.replica.yml down

# Run the setup script to recreate volumes and start the primary/replica containers
docker-compose -f docker-compose.postgres.replica.yml up --build -d