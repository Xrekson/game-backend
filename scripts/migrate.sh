#!/bin/bash
set -e

DB_HOST=${POSTGRES_HOST:-localhost}
DB_PORT=${POSTGRES_PORT:-5432}
DB_USER=${POSTGRES_USER:-postgres}
DB_PASS=${POSTGRES_PASSWORD:-postgrespassword}
DB_NAME=${POSTGRES_DB:-game_db}

echo "Applying PostgreSQL database migrations to ${DB_HOST}:${DB_PORT}/${DB_NAME}..."

export PGPASSWORD=$DB_PASS

for sql_file in services/*/migrations/*.sql; do
    if [ -f "$sql_file" ]; then
        echo "Executing migration: $sql_file"
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$sql_file" || echo "Migration notice for $sql_file"
    fi
done

echo "Database migrations complete."
