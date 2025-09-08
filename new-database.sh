#!/bin/bash

set -e

# Container name
POSTGRES_CONTAINER="postgres"

# Define color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

cleanup() {
    echo "Cleaning up..."
    docker stop $POSTGRES_CONTAINER > /dev/null
    docker rm $POSTGRES_CONTAINER > /dev/null
}

trap cleanup EXIT

echo "Starting PostgreSQL container..."
docker run -d \
    --name $POSTGRES_CONTAINER \
    -e POSTGRES_PASSWORD=password \
    -e POSTGRES_DB=scrabble_test \
    -p 5433:5432 \
    postgres:latest > /dev/null

echo "Waiting for services to be ready..."
until docker exec $POSTGRES_CONTAINER pg_isready -U postgres > /dev/null 2>&1; do
    sleep 3
done

echo "Running database migrations..."
cd migrations
if ! go run up/up.go -dir . -dsn postgres://postgres:password@localhost:5433/postgres?sslmode=disable; then
    echo -e "${RED}Database migration failed!${NC}"
    exit 1
fi
cd ..

read -p "Press [Enter] key to stop the database..."
