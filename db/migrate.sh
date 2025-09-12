#!/usr/bin/env bash
set -euo pipefail

# スクリプト基準の絶対パス
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MIGRATIONS_PATH="${SCRIPT_DIR}/migrations"

# .env を読み込む
ENV_FILE="${SCRIPT_DIR}/../.env"
if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi


# DBを起動（単体で起動でOK）
docker compose up -d db

# Composeプロジェクト名 → ネットワーク名
PROJ_NAME="${COMPOSE_PROJECT_NAME:-$(basename "$(cd "${SCRIPT_DIR}/.."; pwd)")}"
NET="${PROJ_NAME}_default"

DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB}?sslmode=disable"

echo "Waiting for Postgres to be ready on ${NET}..."
# 最大60秒待機（1秒ポーリング）
for i in $(seq 1 60); do
  if docker run --rm --network "${NET}" postgres:16-alpine \
      pg_isready -h db -p 5432 -U ${POSTGRES_USER} -d ${POSTGRES_DB} >/dev/null 2>&1; then
    echo "Postgres is ready."
    break
  fi
  sleep 1
  if [ "$i" = "60" ]; then
    echo "Postgres not ready after 60s. Aborting." >&2
    docker compose logs db | tail -n 100 || true
    exit 1
  fi
done

# マイグレーション実行
docker run --rm \
  -v "${MIGRATIONS_PATH}:/migrations" \
  --network "${NET}" \
  migrate/migrate:4 \
  -path=/migrations \
  -database "${DB_URL}" up
