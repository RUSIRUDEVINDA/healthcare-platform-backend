.PHONY: run build test tidy lint docker-build docker-run

run:
	@export $$(cat .env | xargs) && go run ./cmd/main.go

build:
	CGO_ENABLED=0 go build -o bin/notification-service ./cmd/main.go

test:
	go test -v -race ./...

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

docker-build:
	docker build -t notification-service:latest .

docker-run:
	docker run --rm -p 8006:8006 \
		--env-file .env \
		--name notification-service \
		notification-service:latest
