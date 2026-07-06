#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_DB_NAME="wealth_lens_test_backend_$$"
TEST_DATABASE_URL="postgres://wealth_lens:wealth_lens@localhost:5432/${TEST_DB_NAME}?sslmode=disable"

case "${TEST_DB_NAME}" in
  wealth_lens_test_*) ;;
  *)
    echo "refusing to manage non-test database: ${TEST_DB_NAME}" >&2
    exit 1
    ;;
esac

cleanup() {
  docker compose --project-directory "${ROOT_DIR}" exec -T db \
    dropdb --if-exists --force -U wealth_lens "${TEST_DB_NAME}" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

cd "${ROOT_DIR}"
docker compose up -d db

ready=false
for _ in $(seq 1 30); do
  if docker compose exec -T db pg_isready -U wealth_lens -d wealth_lens >/dev/null 2>&1; then
    ready=true
    break
  fi
  sleep 1
done
if [[ "${ready}" != "true" ]]; then
  echo "PostgreSQL did not become ready" >&2
  exit 1
fi

docker compose exec -T db createdb -U wealth_lens "${TEST_DB_NAME}"
migrate -path backend/migrations -database "${TEST_DATABASE_URL}" up

SNAPSHOT_TEST_DATABASE_URL="${TEST_DATABASE_URL}" \
BACKEND_TEST_DATABASE_URL="${TEST_DATABASE_URL}" \
GOCACHE="${ROOT_DIR}/backend/.gocache" \
  go test -C backend -tags=integration \
    ./internal/auth ./internal/fixeddeposits ./internal/notifications ./internal/prices ./internal/snapshots ./internal/transactions -count=1

migrate -path backend/migrations -database "${TEST_DATABASE_URL}" down 1
migrate -path backend/migrations -database "${TEST_DATABASE_URL}" up
