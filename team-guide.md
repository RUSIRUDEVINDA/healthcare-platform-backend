# 🏥 Healthcare Platform — Team Git Guide

## 👥 Service Ownership

| # | Service | Owner | Branch |
|---|---------|-------|--------|
| 1 | Auth Service | Rusiru | `feature/auth-service` |
| 2 | Patient Management Service | Rusiru | `feature/patient-service` |
| 3 | Doctor Management Service | Ravindu | `feature/doctor-service` |
| 4 | Appointment Service | Ovindi | `feature/appointment-service` |
| 5 | Telemedicine Service | Ovindi | `feature/telemedicine-service` |
| 6 | Payment Service | Rusiru | `feature/payment-service` |
| 7 | Notification Service | Sandali | `feature/notification-service` |
| 8 | AI Symptom Checker Service | Ravindu | `feature/ai-symptom-service` |
| 9 | Admin Service | Sandali | `feature/admin-service` |
| 10 | API Gateway | Ravindu | `feature/api-gateway` |
| 11 | File Storage Service | Ovindi | `feature/file-storage-service` |

---

## 🌿 Branching Strategy

```
main                        ← Production-ready, protected. Only merged via PR.
│
└── develop                 ← Integration branch. All feature branches merge here.
    │
    ├── feature/auth-service          (Rusiru)
    ├── feature/patient-service       (Rusiru)
    ├── feature/payment-service       (Rusiru)
    ├── feature/doctor-service        (Ravindu)
    ├── feature/ai-symptom-service    (Ravindu)
    ├── feature/api-gateway           (Ravindu)
    ├── feature/appointment-service   (Ovindi)
    ├── feature/telemedicine-service  (Ovindi)
    ├── feature/file-storage-service  (Ovindi)
    ├── feature/notification-service  (Sandali)
    └── feature/admin-service         (Sandali)
```

### Rules
- ✅ **Never push directly to [main](file:///d:/healthcare-platform-backend/services/auth-service/cmd/main.go#28-110)**
- ✅ All work goes into your `feature/` branch
- ✅ Merge `feature/` → `develop` via **Pull Request** (team reviews)
- ✅ Merge `develop` → [main](file:///d:/healthcare-platform-backend/services/auth-service/cmd/main.go#28-110) only when a milestone is complete and everything works

---

## 📦 Step 1 — Rusiru: Push Initial Codebase (First Push)

> Run these commands from `d:\healthcare-platform-backend`

```bash
# 1. Initialize and set remote (if not done already)
git init
git remote add origin https://github.com/<your-org>/healthcare-platform-backend.git

# 2. Create and switch to develop branch
git checkout -b develop

# 3. Stage all current files
git add .

# 4. Commit with message below
git commit -m "chore: initial project scaffold with auth service and infrastructure

- services/auth-service/cmd/main.go           → App entrypoint, HTTP server setup
- services/auth-service/internal/config/      → Environment config loader
- services/auth-service/internal/handler/     → HTTP route handlers (register, login, refresh, validate)
- services/auth-service/internal/middleware/  → JWT authentication middleware
- services/auth-service/internal/model/       → User, token domain models
- services/auth-service/internal/repository/  → PostgreSQL data access layer
- services/auth-service/internal/service/     → Business logic (auth, token management)
- services/auth-service/migrations/           → DB migration SQL (up/down)
- services/auth-service/Dockerfile            → Multi-stage Go build image
- services/auth-service/Makefile              → Dev helper commands
- services/auth-service/go.mod                → Go module definition
- pkg/jwt/jwt.go                              → Shared JWT token generation/validation
- pkg/logger/                                 → Shared structured logger
- pkg/rabbitmq/                               → Shared RabbitMQ connection utility
- nginx/nginx.conf                            → Nginx reverse proxy with rate limiting and route config
- docker-compose.yml                          → Local dev orchestration (auth-db, rabbitmq, nginx)
- k8s/auth-service/deployment.yaml           → K8s deployment manifest
- k8s/auth-service/service.yaml              → K8s service manifest
- k8s/auth-service/configmap.yaml            → K8s environment config
- k8s/auth-service/secret.yaml               → K8s secrets manifest
- .github/workflows/auth-service.yml         → CI/CD pipeline (test → build → push → deploy)"

# 5. Push develop branch
git push -u origin develop

# 6. Create your feature branch for auth service
git checkout -b feature/auth-service
git push -u origin feature/auth-service
```

---

## 🛠️ Step 2 — All Members: Create Your Feature Branches

> After Rusiru pushes `develop`, every team member runs:

```bash
# Clone the repo
git clone https://github.com/<your-org>/healthcare-platform-backend.git
cd healthcare-platform-backend

# Switch to develop (base branch)
git checkout develop
git pull origin develop

# Create YOUR feature branch (replace with your branch name)
git checkout -b feature/doctor-service    # Ravindu example
git push -u origin feature/doctor-service
```

---

## 📝 Commit Message Convention

Use the format: `type(scope): short description`

| Type | When to use |
|------|-------------|
| `feat` | New feature or endpoint |
| `fix` | Bug fix |
| `chore` | Config, build files, no logic |
| `docs` | README, comments |
| `test` | Adding/fixing tests |
| `refactor` | Code cleanup without behavior change |

### Examples per service:

```bash
# Starting a new service
git commit -m "feat(doctor-service): initialize service scaffold with folder structure"

# Adding DB models
git commit -m "feat(doctor-service): add Doctor and Availability domain models"

# Adding an endpoint
git commit -m "feat(patient-service): implement POST /patients/register endpoint"

# Adding K8s config
git commit -m "chore(k8s): add doctor-service deployment and service manifests"

# Docker
git commit -m "chore(doctor-service): add multi-stage Dockerfile"

# Bug fix
git commit -m "fix(appointment-service): resolve double-booking race condition"

# Notification
git commit -m "feat(notification-service): integrate SendGrid email on appointment confirmation"
```

---

## 🚀 How Each Member Should Develop Their Service

### 📁 Service Template Structure (copy from auth-service)

Every new service must follow this structure:

```
services/<your-service>/
├── cmd/
│   └── main.go              ← HTTP server start, dependency wiring
├── internal/
│   ├── config/
│   │   └── config.go        ← Load env vars into a Config struct
│   ├── handler/
│   │   └── *.go             ← HTTP handlers (one file per resource group)
│   ├── middleware/
│   │   └── auth.go          ← Call auth-service /auth/validate to check JWT
│   ├── model/
│   │   └── *.go             ← Domain structs (request/response/DB models)
│   ├── repository/
│   │   └── *.go             ← DB queries (PostgreSQL with sqlx or pgx)
│   └── service/
│       └── *.go             ← Business logic
├── migrations/
│   ├── 0001_init.up.sql     ← Create tables
│   └── 0001_init.down.sql   ← Drop tables
├── Dockerfile               ← Copy from auth-service, change binary name
└── Makefile                 ← Copy from auth-service, update service name
```


---

## 👤 Ravindu — Doctor Service, AI Symptom Checker, API Gateway

### Doctor Service (`feature/doctor-service`)
- Port: **8003**
- DB: `doctor_db` (PostgreSQL)
- Key tables: `doctors`, `schedules`, `prescriptions`
- Key endpoints:
  - `GET /doctors` — list doctors with specialty filter
  - `GET /doctors/:id` — doctor profile
  - `PUT /doctors/:id/profile` — update profile (Doctor role)
  - `POST /doctors/:id/schedule` — set availability
  - `POST /doctors/:id/prescriptions` — issue prescription
- **Validate JWT** by calling `http://auth-service:8001/auth/validate` in middleware
- Publish to RabbitMQ queue `doctor.events` when profile is approved

### AI Symptom Checker (`feature/ai-symptom-service`)
- Port: **8008**
- Call OpenAI / Gemini API with patient symptoms
- Return: suggested specialty + preliminary notes
- **No DB needed** — stateless service
- Protect endpoint with JWT middleware (patient role only)

### API Gateway (`feature/api-gateway`)
- This can be implemented as an **enhanced Nginx config** OR a Go service
- If Go: Port **8000**, routes all `/api/*` to correct services
- Update [nginx/nginx.conf](file:///d:/healthcare-platform-backend/nginx/nginx.conf) to add upstreams for each new service (uncomment the prepared blocks)
- Add CI/CD workflow: `.github/workflows/api-gateway.yml`

---

## 👤 Ovindi — Appointment, Telemedicine, File Storage

### Appointment Service (`feature/appointment-service`)
- Port: **8004**
- DB: `appointment_db`
- Key tables: `appointments`, `slots`
- Key endpoints:
  - `POST /appointments` — book appointment
  - `GET /appointments/:id` — get status
  - `PUT /appointments/:id/cancel` — cancel
  - `GET /appointments/doctor/:doctor_id` — slots for a doctor
- After booking → publish `appointment.booked` to RabbitMQ (Notification service listens)

### Telemedicine Service (`feature/telemedicine-service`)
- Port: **8009**
- Integrate **Agora** or **Jitsi Meet**
- Key endpoints:
  - `POST /sessions` — create a session, return room token/URL
  - `GET /sessions/:id` — get session info
- Store session records in `telemedicine_db`

### File Storage Service (`feature/file-storage-service`)
- Port: **8010**
- Use **AWS S3** or **MinIO** (local S3-compatible)
- Key endpoints:
  - `POST /files/upload` — upload medical report (multipart/form-data)
  - `GET /files/:id` — get file URL
  - `DELETE /files/:id` — soft delete
- Max file size: 50MB (already configured in nginx.conf)

---

## 👤 Sandali — Notification Service, Admin Service

### Notification Service (`feature/notification-service`)
- Port: **8006**
- **Listens on RabbitMQ** (no direct HTTP from patients/doctors)
- Queues to consume: `appointment.booked`, `appointment.cancelled`, `consultation.completed`
- Email: integrate **SendGrid** or **SMTP (Gmail)**
- SMS: integrate **Twilio** or **Vonage**
- Key pattern:
```go
// In main.go — subscribe to queue
rabbitmq.Subscribe("appointment.booked", handleAppointmentBooked)
```

### Admin Service (`feature/admin-service`)
- Port: **8007**
- Requires **Admin role JWT** (check role claim in token)
- Key endpoints:
  - `GET /admin/users` — list all users
  - `PUT /admin/doctors/:id/verify` — approve doctor registration
  - `GET /admin/appointments` — view all appointments
  - `GET /admin/transactions` — view payment history
  - `DELETE /admin/users/:id` — deactivate user

---

## 👤 Rusiru — Patient Service, Payment Service (+ Auth done ✅)

### Patient Service (`feature/patient-service`)
- Port: **8002**
- DB: `patient_db`
- Key tables: `patients`, `medical_history`, `prescriptions`
- Key endpoints:
  - `POST /patients/register` — register patient (public)
  - `GET /patients/:id/profile` — view profile
  - `PUT /patients/:id/profile` — update profile
  - `GET /patients/:id/history` — medical history
  - `GET /patients/:id/prescriptions` — view prescriptions

### Payment Service (`feature/payment-service`)
- Port: **8005**
- Integrate **Stripe** (sandbox) or **PayHere**
- DB: `payment_db` → `transactions` table
- Key endpoints:
  - `POST /payments/checkout` — create payment session
  - `POST /payments/webhook` — receive payment gateway callback
  - `GET /payments/:id` — payment status
- After success → publish `payment.completed` to RabbitMQ

---

## 🔄 Workflow: Feature → Develop → Main

```bash
# 1. Work on your feature branch daily
git add services/doctor-service/
git commit -m "feat(doctor-service): add availability schedule endpoint"
git push origin feature/doctor-service

# 2. When your service is ready, open a Pull Request:
#    feature/doctor-service → develop
#    (at least 1 team member reviews and approves)

# 3. After review, merge to develop
# 4. Test integration (docker-compose up --build)

# 5. When all services pass → merge develop → main
```

---

## ✅ Checklist Before Opening a PR

- [ ] Service runs with `docker-compose up`
- [ ] All endpoints return correct HTTP status codes
- [ ] JWT middleware is applied to protected routes
- [ ] K8s manifests added in `k8s/<service-name>/`
- [ ] [nginx/nginx.conf](file:///d:/healthcare-platform-backend/nginx/nginx.conf) upstream uncommented for your service
- [ ] [docker-compose.yml](file:///d:/healthcare-platform-backend/docker-compose.yml) updated with your service and its DB
- [ ] CI/CD workflow added `.github/workflows/<service-name>.yml`
- [ ] No secrets committed (use [.env](file:///d:/healthcare-platform-backend/services/auth-service/.env) file, add to [.gitignore](file:///d:/healthcare-platform-backend/services/auth-service/.gitignore))
