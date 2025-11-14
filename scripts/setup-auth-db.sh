#!/bin/bash
# Dogelytics Dogelytics Database Setup Script
# Creates the dogelytics database and user for local development

set -e

echo "==================================="
echo "Dogelytics Dogelytics Database Setup"
echo "==================================="
echo ""

# Default values
DEFAULT_HOST="localhost"
DEFAULT_PORT="5432"
DEFAULT_ADMIN_USER="postgres"
DEFAULT_DB_USER="dogelytics"
DEFAULT_DB_NAME="dogelytics"
DEFAULT_DB_PASSWORD="changeme"

# Prompt for PostgreSQL connection details
read -p "PostgreSQL host [${DEFAULT_HOST}]: " POSTGRES_HOST
POSTGRES_HOST=${POSTGRES_HOST:-$DEFAULT_HOST}

read -p "PostgreSQL port [${DEFAULT_PORT}]: " POSTGRES_PORT
POSTGRES_PORT=${POSTGRES_PORT:-$DEFAULT_PORT}

read -p "PostgreSQL admin user [${DEFAULT_ADMIN_USER}]: " POSTGRES_ADMIN
POSTGRES_ADMIN=${POSTGRES_ADMIN:-$DEFAULT_ADMIN_USER}

read -sp "PostgreSQL admin password: " POSTGRES_PASSWORD
echo ""
echo ""

# Database details
read -p "Dogelytics database name [${DEFAULT_DB_NAME}]: " DB_NAME
DB_NAME=${DB_NAME:-$DEFAULT_DB_NAME}

read -p "Dogelytics database user [${DEFAULT_DB_USER}]: " DB_USER
DB_USER=${DB_USER:-$DEFAULT_DB_USER}

read -sp "Dogelytics database password [${DEFAULT_DB_PASSWORD}]: " DB_PASSWORD
DB_PASSWORD=${DB_PASSWORD:-$DEFAULT_DB_PASSWORD}
echo ""
echo ""

echo "Setting up database with the following configuration:"
echo "  Host: ${POSTGRES_HOST}:${POSTGRES_PORT}"
echo "  Database: ${DB_NAME}"
echo "  User: ${DB_USER}"
echo ""

# Test PostgreSQL connection
echo "Testing PostgreSQL connection..."
export PGPASSWORD="${POSTGRES_PASSWORD}"
if ! psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_ADMIN}" -d postgres -c '\q' 2>/dev/null; then
    echo "❌ Error: Could not connect to PostgreSQL"
    echo "   Please check your connection details and try again"
    exit 1
fi
echo "✓ Connected to PostgreSQL"
echo ""

# Check if database exists
echo "Checking if database '${DB_NAME}' exists..."
DB_EXISTS=$(psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_ADMIN}" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'")

if [ "$DB_EXISTS" = "1" ]; then
    echo "⚠ Database '${DB_NAME}' already exists"
    read -p "Do you want to drop and recreate it? (y/N): " RECREATE
    if [[ "$RECREATE" =~ ^[Yy]$ ]]; then
        echo "Dropping database '${DB_NAME}'..."
        psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_ADMIN}" -d postgres -c "DROP DATABASE IF EXISTS ${DB_NAME};"
        echo "✓ Database dropped"
    else
        echo "Keeping existing database"
    fi
fi

# Create user if it doesn't exist
echo ""
echo "Creating user '${DB_USER}' if it doesn't exist..."
psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_ADMIN}" -d postgres <<EOF
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '${DB_USER}') THEN
    CREATE USER ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';
    RAISE NOTICE 'User ${DB_USER} created';
  ELSE
    RAISE NOTICE 'User ${DB_USER} already exists';
  END IF;
END
\$\$;
EOF
echo "✓ User ready"

# Create database if it doesn't exist
echo ""
echo "Creating database '${DB_NAME}' if it doesn't exist..."
psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_ADMIN}" -d postgres -c "CREATE DATABASE ${DB_NAME};" 2>/dev/null || echo "Database already exists"
echo "✓ Database ready"

# Grant privileges
echo ""
echo "Granting privileges..."
psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_ADMIN}" -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${DB_USER};"

# Connect to dogelytics database and grant schema permissions
psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_ADMIN}" -d "${DB_NAME}" <<EOF
GRANT ALL ON SCHEMA public TO ${DB_USER};
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ${DB_USER};
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ${DB_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ${DB_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ${DB_USER};
EOF
echo "✓ Privileges granted"

# Generate connection string
DOGELYTICS_INDEXER_DBURL="postgres://${DB_USER}:${DB_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${DB_NAME}?sslmode=disable"

echo ""
echo "==================================="
echo "✓ Setup Complete!"
echo "==================================="
echo ""
echo "Add this to your .env file:"
echo ""
echo "DOGELYTICS_INDEXER_DBURL=${DOGELYTICS_DBURL}"
echo ""
echo "You can now run dogelytics:"
echo "  export DOGELYTICS_INDEXER_DBURL='${DOGELYTICS_DBURL}'"
echo "  go run ./cmd/dogelytics"
echo ""
echo "Or use the admin CLI:"
echo "  export DOGELYTICS_INDEXER_DBURL='${DOGELYTICS_DBURL}'"
echo "  go run ./cmd/admin create-user --email admin@example.com --password yourpassword123"
echo ""

