#!/bin/sh
set -e

goose -dir "$MIGRATION_FOLDER" postgres "$POSTGRES_SETUP_MAIN" up

echo "Starting the application..."
exec /hw-4
