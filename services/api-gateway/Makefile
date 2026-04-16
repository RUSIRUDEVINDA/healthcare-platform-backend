# API Gateway Makefile
.PHONY: run build test tidy lint docker-build help

run:
	@export $$(cat .env 2>/dev/null | xargs) && go run ./cmd/main.go

build:
	CGO_ENABLED=0 go build -o bin/api-gateway ./cmd/main.go

test:
	go test -v -race ./...

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

docker-build:
	cd ../.. && docker build -f services/api-gateway/Dockerfile -t api-gateway:latest .

help:
	@echo "targets: run build test tidy lint docker-build"
