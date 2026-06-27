DATABASE_URL ?= postgres://wealth_lens:wealth_lens@localhost:5432/wealth_lens?sslmode=disable
MIGRATIONS := backend/migrations

.PHONY: db-up db-down migrate-up migrate-down migrate-version test run snapshot-daily

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

run:
	cd backend && go run ./cmd/api

snapshot-daily:
	cd backend && go run ./cmd/snapshots -date "$(DATE)"
