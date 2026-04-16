# syntax=docker/dockerfile:1.7
FROM golang:1.25-alpine AS builder

WORKDIR /app
ENV GOWORK=off

COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/telemedicine-service/go.mod services/telemedicine-service/go.sum ./services/telemedicine-service/

WORKDIR /app/services/telemedicine-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /app
COPY pkg/ ./pkg/
COPY services/telemedicine-service/ ./services/telemedicine-service/

WORKDIR /app/services/telemedicine-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /telemedicine-service \
    ./cmd/main.go

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /telemedicine-service .
COPY services/telemedicine-service/migrations/ ./migrations/

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8009

CMD ["./telemedicine-service"]
