# Doctor Management Service

Doctor profiles and related APIs. Follows the same layering as **auth-service**: `handler` → `service` → `repository` → `model`, PostgreSQL via `database/sql`, shared **`pkg/logger`** and **`pkg/rabbitmq`**.

Per **team-guide.md**: listens on port **8003**, database **`doctor_db`**, validates JWTs by calling **auth-service** `GET /auth/validate`, and publishes **`doctor.created`** on the **`doctor_events`** exchange when a doctor is created.

## Layout

```
services/doctor-service/
├── cmd/main.go
├── internal/config/
├── internal/handler/
├── internal/middleware/
├── internal/model/
├── internal/repository/
├── internal/service/
├── migrations/
├── Dockerfile
├── Makefile
└── go.mod
```

## Environment

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8003` | HTTP port |
| `DATABASE_URL` | Yes | — | PostgreSQL DSN (`doctor_db`) |
| `RABBITMQ_URL` | No | — | If empty, doctor events are not published |
| `AUTH_SERVICE_URL` | Yes | `http://localhost:8001` | Base URL for auth-service (e.g. `http://auth-service:8001` in Docker) |

Copy `.env.example` to `.env` for local runs.

## Run locally

1. Start infra (e.g. `docker compose up -d doctor-db rabbitmq`) so **`doctor_db`** is on **`localhost:5439`** (dedicated **`doctor-db`** container).
2. Start **auth-service** on **8001** (JWT validation).
3. From this directory:

```bash
make run
# or: go run ./cmd/main.go with env vars set
```

## API summary

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Liveness |
| GET | `/ready` | No | Readiness |
| GET | `/doctors` | No | List doctors (`?specialization=` optional filter) |
| GET | `/doctors/:id` | No | Get one doctor |
| POST | `/doctors` | Bearer JWT + **admin** | Create doctor |
| PUT | `/doctors` | Bearer JWT + **doctor** or **admin** | Update doctor (`id` + fields in JSON body) |
| PUT | `/doctors/:id` | Bearer JWT + **doctor** or **admin** | Update doctor |
| PUT | `/doctors/:id/profile` | Bearer JWT + **doctor** or **admin** | Update profile (same payload as PUT `/:id`) |
| DELETE | `/doctors/:id` | Bearer JWT + **admin** | Delete doctor |

Use a **numeric** id in the path: **`/doctors/4`**. Avoid sending template text literally: `/doctors/:4`, `/doctors/{4}`, or `/doctors/{{id}}` unless Postman variables resolve to plain digits (the service strips `:`, `{`, `}` if sent by mistake).

Through **Nginx** (repo root `docker compose up`): `http://localhost/doctor/...` and `http://localhost/api/doctors/...` proxy to this service’s `/doctors/` routes.

## Dev admin login (local only)

`POST /auth/register` cannot create `role: admin`. For testing **POST /doctors**, seed an admin once, then log in:

| Field | Value |
|-------|--------|
| **Email** | `admin@healthcare.local` |
| **Password** | `Admin123!` |

**Seed the user** (Postgres container `auth-db` must be running):

```powershell
Get-Content scripts/seed-dev-admin.sql | docker exec -i auth-db psql -U postgres -d auth_db
```

**Login:** `POST http://localhost:8001/auth/login` with JSON `{"email":"admin@healthcare.local","password":"Admin123!"}`. Use `data.access_token` as Bearer token on `POST /doctors`.

Do not use these credentials in production.

## Docker Compose

From repository root:

```bash
docker compose up --build
```

- **`doctor-db`**: Postgres **only** for `doctor_db`, host port **`5439`**
- **`auth-db`**: Postgres for `auth_db`, host port **`5433`**
- Doctor API: `http://localhost:8003`
