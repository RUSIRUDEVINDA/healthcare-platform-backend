# ─────────────────────────────────────────────
# Stage 1: Build
# ─────────────────────────────────────────────
# syntax=docker/dockerfile:1.7
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Disable Go Workspaces to ensure this service builds independently 
ENV GOWORK=off

# 1. Copy module manifests first to cache dependency downloads
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/appointment-service/go.mod services/appointment-service/go.sum ./services/appointment-service/

WORKDIR /app/services/appointment-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /app

# 2. Copy shared libraries (pkg)
COPY pkg/ ./pkg/

# 3. Copy only this service's code
COPY services/appointment-service/ ./services/appointment-service/

# 4. Compile binary
WORKDIR /app/services/appointment-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /appointment-service \
    ./cmd/main.go

# ─────────────────────────────────────────────
# Stage 2: Production image
# ─────────────────────────────────────────────
FROM alpine:3.20

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /appointment-service .

# Copy migration files (optional, as main.go has them embedded currently, but good for future)
COPY services/appointment-service/migrations/ ./migrations/

# Run as non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8004

CMD ["./appointment-service"]
