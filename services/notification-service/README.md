# Notification Service

Port: 8006

This service consumes RabbitMQ events and sends notifications.

## Consumed Events

- appointment.booked
- appointment.cancelled
- consultation.completed

## Run

```powershell
docker compose up --build -d notification-service
```

## Health

- /health
- /ready
