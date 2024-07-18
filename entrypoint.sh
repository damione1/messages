#!/bin/sh
set -e

# Run database migrations
make db-up

# Start the application
exec "$@"
