DATABASE_URL ?= postgres://wealth_lens:wealth_lens@localhost:5432/wealth_lens?sslmode=disable
MIGRATIONS := backend/migrations

.PHONY: db-up db-down app-up app-down app-logs migrate-up migrate-down migrate-version test test-integration test-all run prices-nightly snapshot-daily snapshot-weekly snapshot-goals-monthly frontend-install frontend-dev frontend-lint frontend-test frontend-build frontend-check

db-up:
	docker compose up -d db

db-down:
	docker compose down

app-up:
	docker compose -f compose.yaml -f compose.app.yaml up --build -d

app-down:
	docker compose -f compose.yaml -f compose.app.yaml down

app-logs:
	docker compose -f compose.yaml -f compose.app.yaml logs -f api frontend

migrate-up:
	migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" down

migrate-version:
	migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" version

test:
	cd backend && GOCACHE="$(CURDIR)/backend/.gocache" go test ./...

test-integration:
	./scripts/test-backend-integration.sh

test-all: test frontend-check test-integration

run:
	cd backend && go run ./cmd/api

prices-nightly:
	cd backend && go run ./cmd/prices $(if $(FROM),-from "$(FROM)") $(if $(TO),-to "$(TO)")

snapshot-daily:
	cd backend && go run ./cmd/snapshots -date "$(DATE)"

snapshot-weekly:
	cd backend && go run ./cmd/snapshots -period weekly -date "$(DATE)"

snapshot-goals-monthly:
	cd backend && go run ./cmd/goals -date "$(DATE)"

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-lint:
	cd frontend && npm run lint

frontend-test:
	cd frontend && npm test

frontend-build:
	cd frontend && npm run build

frontend-check: frontend-lint frontend-test frontend-build
