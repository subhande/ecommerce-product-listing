# Remove existing container if it exists
docker rm -f ecommerce-postgres

# Make data directory if it doesn't exist
mkdir -p postgres-data

# Remove existing volume if it exists
docker volume rm -f ecommerce-postgres-data

# Use Docker volume for data persistence in current directory/data
docker volume create \
    --name ecommerce-postgres-data \
    --opt type=none \
    --opt device=$(pwd)/postgres-data \
    --opt o=bind
docker run --name ecommerce-postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_DB=ecommerce \
    -p 5432:5432 \
    -v ecommerce-postgres-data:/var/lib/postgresql \
    -d postgres:18.2
