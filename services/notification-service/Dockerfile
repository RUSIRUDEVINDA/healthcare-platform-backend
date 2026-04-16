FROM golang:1.25-alpine AS builder

WORKDIR /app
ENV GOWORK=off

COPY pkg/ ./pkg/
COPY services/notification-service/ ./services/notification-service/

WORKDIR /app/services/notification-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /notification-service \
    ./cmd/main.go

FROM alpine:3.18

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata wget
COPY --from=builder /notification-service .
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8006
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8006/health || exit 1

CMD ["./notification-service"]
