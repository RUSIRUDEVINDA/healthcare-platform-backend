.PHONY: run build docker-build docker-run tidy

run:
	go run ./cmd/main.go

build:
	go build -o appointment-service ./cmd/main.go

tidy:
	go mod tidy

docker-build:
	docker build -t appointment-service:latest .

docker-run:
	docker run --env-file .env -p 8082:8082 appointment-service:latest