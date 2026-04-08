# ─────────────────────────────────────────────
# Stage 1: Build (same layout as auth-service Dockerfile)
# ─────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV GOWORK=off

COPY pkg/ ./pkg/

COPY services/doctor-service/ ./services/doctor-service/

WORKDIR /app/services/doctor-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /doctor-service \
    ./cmd/main.go

# ─────────────────────────────────────────────
# Stage 2: Production image
# ─────────────────────────────────────────────
FROM alpine:3.18

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /doctor-service .

COPY services/doctor-service/migrations/ ./migrations/

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8003

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8003/health || exit 1

CMD ["./doctor-service"]
