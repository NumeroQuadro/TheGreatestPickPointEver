#!/bin/sh
set -e

#echo "Database is up. Running migrations..."
goose -dir "$MIGRATION_FOLDER" postgres "$POSTGRES_SETUP_MAIN" up

echo "Starting the application..."
exec /hw-4
