# Admin Service Makefile
# Usage: make <target>

.PHONY: run build test docker-build docker-run migrate-up migrate-down tidy lint help

run:
	@export $$(cat .env | xargs) && go run ./cmd/main.go

build:
	CGO_ENABLED=0 go build -o bin/admin-service ./cmd/main.go

test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

docker-build:
	docker build -t admin-service:latest .

docker-run:
	docker run --rm -p 8007:8007 \
		--env-file .env \
		--name admin-service \
		admin-service:latest

migrate-up:
	migrate -path ./migrations -database "$${DATABASE_URL}" up

migrate-down:
	migrate -path ./migrations -database "$${DATABASE_URL}" down 1

help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
