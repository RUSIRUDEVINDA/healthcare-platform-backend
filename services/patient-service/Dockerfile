# ─────────────────────────────────────────────
# Stage 1: Build
# Uses full Go image to compile the binary
# ─────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Disable Go Workspaces to ensure this service builds independently 
# using only its own go.mod and the local 'pkg' directory.
ENV GOWORK=off

# 1. Copy go.mod files for caching
COPY services/patient-service/go.mod ./services/patient-service/
# If there's a pkg/go.mod, copy that too
COPY pkg/ ./pkg/

WORKDIR /app/services/patient-service
# Use a GOPROXY that might be more stable if proxy.golang.org fails
ENV GOPROXY=https://proxy.golang.org,direct
RUN go mod download

# 2. Copy the rest of the source code
WORKDIR /app
COPY services/patient-service/ ./services/patient-service/

# 3. Compile binary
WORKDIR /app/services/patient-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /patient-service \
    ./cmd/main.go

# ─────────────────────────────────────────────
# Stage 2: Production image
# Uses tiny Alpine image
# ─────────────────────────────────────────────
FROM alpine:3.18

WORKDIR /app

# SSL certs and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder stage
COPY --from=builder /patient-service .

# Copy migration files specifically from the service folder
COPY services/patient-service/migrations/ ./migrations/

# Run as non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8002

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8002/health || exit 1

CMD ["./patient-service"]
