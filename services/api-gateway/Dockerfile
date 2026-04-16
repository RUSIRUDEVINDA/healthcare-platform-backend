# API Gateway — same build layout as doctor-service (repo root as context).
FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV GOWORK=off

COPY pkg/ ./pkg/

COPY services/api-gateway/ ./services/api-gateway/

WORKDIR /app/services/api-gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /api-gateway \
    ./cmd/main.go

FROM alpine:3.18

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /api-gateway .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8000/health || exit 1

CMD ["./api-gateway"]
