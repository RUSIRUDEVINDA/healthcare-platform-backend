APP_NAME=file-storage-service

.PHONY: run build test tidy help

run:
	go run ./cmd/main.go

build:
	go build -o $(APP_NAME) ./cmd

test:
	go test ./...

tidy:
	go mod tidy

help:
	@echo "Available targets: run, build, test, tidy"
