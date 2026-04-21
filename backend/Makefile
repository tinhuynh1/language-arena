DB_URL ?= postgres://lingouser:lingopass@localhost:5432/lingodb?sslmode=disable

.PHONY: migrate-up migrate-status build run dev

## Run all pending migrations
migrate-up:
	DB_URL=$(DB_URL) go run ./cmd/migrate up

## Show migration status
migrate-status:
	DB_URL=$(DB_URL) go run ./cmd/migrate status

## Build server and migrate binaries
build:
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate

## Run the server locally
run:
	DB_URL=$(DB_URL) go run ./cmd/server

## Run with live reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air
