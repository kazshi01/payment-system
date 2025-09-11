#!/usr/bin/env bash
set -euo pipefail

# compose を起動（未起動なら）
docker compose up -d

# プロジェクトのデフォルトネットワーク名（例: <dir>_default）
NET="$(basename "$PWD")_default"

DB_URL="postgres://app:password@db:5432/payment?sslmode=disable"

docker run --rm \
  -v "$(pwd)/db/migrations:/migrations" \
  --network "$NET" \
  migrate/migrate:4 \
  -path=/migrations -database "$DB_URL" up
