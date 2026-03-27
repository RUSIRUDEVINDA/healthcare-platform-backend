# ─────────────────────────────────────────────
# Stage 1: Build
# Uses full Go image to compile the binary
# ─────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 1. Copy root dependency files
# Docker caches this layer — only re-runs go mod download when go.mod changes
COPY go.mod go.sum ./
RUN go mod download

# 2. Copy shared libraries (pkg)
COPY pkg/ ./pkg/

# 3. Copy service-specific code
COPY services/auth-service/ ./services/auth-service/

# 4. Compile binary from within the service directory
# CGO_ENABLED=0 = static binary
# GOOS=linux = compile for Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /auth-service \
    ./services/auth-service/cmd/main.go

# ─────────────────────────────────────────────
# Stage 2: Production image
# Uses tiny Alpine image
# ─────────────────────────────────────────────
FROM alpine:3.18

WORKDIR /app

# SSL certs and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder stage
COPY --from=builder /auth-service .

# Copy migration files specifically from the service folder
COPY services/auth-service/migrations/ ./migrations/

# Run as non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8001

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8001/health || exit 1

CMD ["./auth-service"]
