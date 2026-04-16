# ─────────────────────────────────────────────
# Stage 1: Build
# ─────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 1. Cache dependencies
# Copy only go.mod and go.sum first to leverage Docker cache
COPY services/support-service/go.mod ./services/support-service/
WORKDIR /app/services/support-service
RUN go mod download

# 2. Copy source code
WORKDIR /app
COPY services/support-service/ ./services/support-service/

# 3. Build binary
WORKDIR /app/services/support-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /support-service \
    ./cmd/main.go

# ─────────────────────────────────────────────
# Stage 2: Production
# ─────────────────────────────────────────────
FROM alpine:3.18

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /support-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8086

CMD ["./support-service"]
