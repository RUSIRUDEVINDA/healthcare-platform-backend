# ─────────────────────────────────────────────
# Stage 1: Build
# ─────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app
ENV GOWORK=off

COPY pkg/ ./pkg/
COPY services/admin-service/ ./services/admin-service/

WORKDIR /app/services/admin-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /admin-service \
    ./cmd/main.go

# ─────────────────────────────────────────────
# Stage 2: Production image
# ─────────────────────────────────────────────
FROM alpine:3.18

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata wget
COPY --from=builder /admin-service .
COPY services/admin-service/migrations/ ./migrations/
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8007
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8007/health || exit 1

CMD ["./admin-service"]
