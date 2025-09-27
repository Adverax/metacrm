#!/bin/bash
set -e
set -u

# Prepare list of databases: default POSTGRES_DB and optional POSTGRES_DBs (comma-separated)
databases="${POSTGRES_DB:-postgres}"
if [ -n "${POSTGRES_DBs:-}" ]; then
  # Replace commas with spaces to iterate
  databases+=" ${POSTGRES_DBs//,/ }"
fi

for db in $databases; do
  if [ -z "$db" ]; then
    continue
  fi
  echo "Creating pgTAP extension in $db"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$db" -c "CREATE EXTENSION IF NOT EXISTS pgtap;"
done
