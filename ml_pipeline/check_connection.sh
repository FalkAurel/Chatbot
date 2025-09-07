#!/bin/bash

# Wait for SurrealDB to be ready
echo "Testing connection to surrealdb..."
until ping -c 1 surrealdb &> /dev/null; do
    echo "Waiting for surrealdb to be reachable..."
    sleep 1
done

echo "Testing port 8000..."
until nc -zv surrealdb 8000; do
    echo "Waiting for surrealdb port 8000..."
    sleep 1
done

echo "Connection established, starting application..."
exec "$@"

lshw -C cpu -C memory -C display >> hardware_specs.txt && ./target/debug/ml_pipeline
