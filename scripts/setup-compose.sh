#!/bin/sh
set -eu

repository_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
secrets_dir="$repository_dir/secrets"
environment_file="$repository_dir/.env"

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl is required to generate Compose secrets" >&2
  exit 1
fi

umask 077
mkdir -p "$secrets_dir"

create_hex_secret() {
  target=$1
  bytes=$2
  if [ -s "$target" ]; then
    echo "Keeping existing $(basename "$target")"
    return
  fi
  openssl rand -hex "$bytes" >"$target"
  echo "Created $(basename "$target")"
}

create_hex_secret "$secrets_dir/postgres_password" 24
create_hex_secret "$secrets_dir/session_secret" 32
create_hex_secret "$secrets_dir/analytics_secret" 32

if [ ! -s "$secrets_dir/database_url" ]; then
  postgres_password=$(tr -d '\r\n' <"$secrets_dir/postgres_password")
  printf 'postgres://dogelytics:%s@postgres:5432/dogelytics?sslmode=disable\n' "$postgres_password" >"$secrets_dir/database_url"
  echo "Created database_url"
else
  echo "Keeping existing database_url"
fi

if [ ! -e "$environment_file" ]; then
  cp "$repository_dir/compose.env.example" "$environment_file"
  chmod 600 "$environment_file"
  echo "Created .env from compose.env.example"
else
  echo "Keeping existing .env"
fi

echo "Compose configuration is ready. Review .env, then run: docker compose up -d --build"
