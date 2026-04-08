# ─────────────────────────────────────────────
# Stage 1: Build (repo root context; independent service module)
# ─────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV GOWORK=off

COPY pkg/ ./pkg/

COPY services/ai-symptom-service/ ./services/ai-symptom-service/

WORKDIR /app/services/ai-symptom-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /ai-symptom-service \
    ./cmd/main.go

# ─────────────────────────────────────────────
# Stage 2: Production image
# ─────────────────────────────────────────────
FROM alpine:3.18

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /ai-symptom-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8008

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8008/health || exit 1

CMD ["./ai-symptom-service"]
