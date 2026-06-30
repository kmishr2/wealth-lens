DATABASE_URL ?= postgres://wealth_lens:wealth_lens@localhost:5432/wealth_lens?sslmode=disable
MIGRATIONS := backend/migrations

.PHONY: db-up db-down migrate-up migrate-down migrate-version test test-integration test-all run snapshot-daily frontend-install frontend-dev frontend-lint frontend-build frontend-check

db-up:
	docker compose up -d db

db-down:
	docker compose down

migrate-up:
	migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" down

migrate-version:
	migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" version

test:
	cd backend && GOCACHE="$(CURDIR)/backend/.gocache" go test ./...

test-integration:
	./scripts/test-snapshot-integration.sh

test-all: test frontend-check test-integration

run:
	cd backend && go run ./cmd/api

snapshot-daily:
	cd backend && go run ./cmd/snapshots -date "$(DATE)"

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-lint:
	cd frontend && npm run lint

frontend-build:
	cd frontend && npm run build

frontend-check: frontend-lint frontend-build
