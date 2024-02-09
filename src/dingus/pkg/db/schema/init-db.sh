#!/bin/bash

# Ensure the script stops on error
set -e

# Apply the schema to the already created database
# Note: POSTGRES_DB, POSTGRES_USER, and related environment variables are set by the Docker Compose environment

# Check if the database schema file exists
SCHEMA_FILE="/docker-entrypoint-initdb.d/tagsdb.sql"

if [ -f "$SCHEMA_FILE" ]; then
    echo "Applying database schema from $SCHEMA_FILE..."
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -a -f "$SCHEMA_FILE"
else
    echo "Database schema file $SCHEMA_FILE not found."
    exit 1
fi

echo "Database initialization complete."
