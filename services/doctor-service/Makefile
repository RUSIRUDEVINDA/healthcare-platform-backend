# Doctor Service Makefile
# Usage: make <target>

.PHONY: run build test docker-build docker-run migrate-up migrate-down tidy

## run: Run the service locally (needs .env file)
run:
	@export $$(cat .env | xargs) && go run ./cmd/main.go

## build: Compile the Go binary
build:
	CGO_ENABLED=0 go build -o bin/doctor-service ./cmd/main.go

## test: Run all tests
test:
	go test -v -race ./...

## test-cover: Run tests with coverage report
test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## tidy: Tidy go modules
tidy:
	go mod tidy

## lint: Run linter (needs golangci-lint installed)
lint:
	golangci-lint run ./...

## docker-build: Build Docker image (run from repo root)
docker-build:
	cd ../.. && docker build -f services/doctor-service/Dockerfile -t doctor-service:latest .

## migrate-up: Run migrations (needs migrate CLI)
migrate-up:
	migrate -path ./migrations -database "$${DATABASE_URL}" up

## migrate-down: Rollback last migration
migrate-down:
	migrate -path ./migrations -database "$${DATABASE_URL}" down 1

## help: Show this help
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
