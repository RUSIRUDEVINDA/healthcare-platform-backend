# API Gateway

Go reverse proxy on **port 8000** that routes **`/api/*`** to platform services, aligned with real backend paths (same prefixes the SPA uses under `/api` + `v1/…`).

## Routes

| Incoming prefix | Upstream | Backend path prefix |
|-----------------|----------|---------------------|
| `/api/auth/` | `AUTH_SERVICE_URL` | `/auth/` |
| `/api/doctors/` | `DOCTOR_SERVICE_URL` | `/doctors/` |
| `/api/v1/doctors/` | `DOCTOR_SERVICE_URL` | `/doctors/` |
| `/api/ai/symptom/` | `AI_SYMPTOM_SERVICE_URL` | `/symptoms/` |
| `/api/v1/patient/` | `PATIENT_SERVICE_URL` | `/api/v1/patient/` |
| `/api/v1/appointments/` | `APPOINTMENT_SERVICE_URL` | `/api/v1/appointments/` |
| `/api/v1/slots/` | `APPOINTMENT_SERVICE_URL` | `/api/v1/slots/` |
| `/api/v1/payments/` | `PAYMENT_SERVICE_URL` | `/api/v1/payments/` |
| `/api/v1/telemedicine/` | `TELEMEDICINE_SERVICE_URL` | `/api/v1/telemedicine/` |
| `/api/v1/files/` | `FILE_STORAGE_SERVICE_URL` | `/api/v1/files/` |
| `/api/admin/` | `ADMIN_SERVICE_URL` | `/admin/` |
| `/api/notifications/` | `NOTIFICATION_SERVICE_URL` | `/notifications/` |

Upstream base URLs default to the Docker service names and ports in `team-guide.md` (override via env or `.env`).

On startup the gateway logs **`API gateway route prefixes`** with every registered prefix—use it to confirm `/api/v1/patient` (and others) are active.

`GET /api/auth/validate` is restricted to private networks (same intent as nginx).

## Local run

**`.env` is loaded from** `services/api-gateway/.env` even if you start the process from the **repo root** (`go run ./services/api-gateway/cmd/main.go`). It also loads `.env` from the current working directory when present.

If **api-gateway runs on your host** (not inside Docker) but backends are on `localhost`, set URLs like `http://localhost:8002` in `.env`.

```bash
cd services/api-gateway
go run ./cmd/main.go
```

## Docker image (repo root)

```bash
docker build -f services/api-gateway/Dockerfile -t api-gateway:latest .
```
