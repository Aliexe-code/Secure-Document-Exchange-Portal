#!/bin/bash

# Script to start required services for SDEP

echo "Starting services for Secure Document Exchange Portal..."

# Check if docker or podman is available
if command -v docker &> /dev/null; then
    CONTAINER_CMD="docker"
    COMPOSE_CMD="docker compose"
elif command -v podman &> /dev/null; then
    CONTAINER_CMD="podman"
    COMPOSE_CMD="podman-compose"
else
    echo "Error: Neither docker nor podman is installed."
    echo "Please install one of them to continue."
    exit 1
fi

echo "Using $CONTAINER_CMD..."

# Check if compose is available
if ! command -v $COMPOSE_CMD &> /dev/null; then
    echo "Error: $COMPOSE_CMD is not available."
    echo "Please install docker-compose or podman-compose."
    exit 1
fi

# Start services
echo "Starting services..."
$COMPOSE_CMD up -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 5

# Create MinIO bucket
echo "Creating MinIO bucket..."
$CONTAINER_CMD exec sdep_minio mc alias set local http://localhost:9000 minioadmin minioadmin 2>/dev/null || true
$CONTAINER_CMD exec sdep_minio mc mb local/documents 2>/dev/null || echo "Bucket already exists"

echo ""
echo "Services started successfully!"
echo ""
echo "MinIO Console: http://localhost:9001 (minioadmin/minioadmin)"
echo "MinIO API: http://localhost:9000"
echo "Redis: localhost:6379"
echo "PostgreSQL: localhost:5432"
echo ""
echo "To stop services, run: $COMPOSE_CMD down"
