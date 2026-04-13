FROM golang:1.25-alpine AS builder

WORKDIR /app
ENV GOWORK=off

COPY pkg/ ./pkg/
COPY services/file-storage-service/ ./services/file-storage-service/

WORKDIR /app/services/file-storage-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
	-ldflags="-w -s" \
	-o /file-storage-service \
	./cmd/main.go

FROM alpine:3.18

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /file-storage-service .
COPY services/file-storage-service/migrations/ ./migrations/

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8010

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
	CMD wget -qO- http://localhost:8010/health || exit 1

CMD ["./file-storage-service"]
