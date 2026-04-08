# AI Symptom Checker Service Makefile
# Usage: make <target>

.PHONY: run build test tidy lint help

## run: Run the service locally (needs .env file)
run:
	@export $$(cat .env | xargs) && go run ./cmd/main.go

## build: Compile the Go binary
build:
	CGO_ENABLED=0 go build -o bin/ai-symptom-service ./cmd/main.go

## test: Run all tests
test:
	go test -v -race ./...

## tidy: Tidy go modules
tidy:
	go mod tidy

## lint: Run linter (needs golangci-lint installed)
lint:
	golangci-lint run ./...

## docker-build: Build Docker image (run from repo root)
docker-build:
	cd ../.. && docker build -f services/ai-symptom-service/Dockerfile -t ai-symptom-service:latest .

## help: Show this help
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
