# 🏥 Healthcare Platform — Auth Service

SE3020 Distributed Systems — Assignment 1  
Group Member: **Rusiru** (Auth Service, Patient Service, Payment Service)

---

## 📁 Project Structure

```
healthcare-platform/
├── auth-service/               ← This service (Rusiru)
│   ├── cmd/main.go             ← Entry point
│   ├── internal/
│   │   ├── config/             ← Env var config
│   │   ├── handler/            ← HTTP handlers (Gin)
│   │   ├── middleware/         ← JWT auth, CORS, Logger
│   │   ├── model/              ← Domain models + DTOs
│   │   ├── repository/         ← PostgreSQL queries
│   │   └── service/            ← Business logic
│   ├── pkg/
│   │   ├── jwt/                ← JWT helpers (reusable)
│   │   ├── rabbitmq/           ← RabbitMQ client (reusable)
│   │   └── logger/             ← Structured logging (reusable)
│   ├── migrations/             ← SQL migration files
│   ├── Dockerfile
│   ├── Makefile
│   └── .env
├── k8s/auth-service/           ← Kubernetes manifests
├── nginx/nginx.conf            ← Reverse proxy config
├── docker-compose.yml          ← Local dev setup
└── .github/workflows/          ← CI/CD pipelines
```

---

## 🚀 Quick Start (Local Development)

### Prerequisites
- Docker Desktop installed
- Go 1.21+ (for local development without Docker)

### Step 1: Clone and setup
```bash
git clone <your-repo-url>
cd healthcare-platform
```

### Step 2: Start all services with Docker Compose
```bash
docker-compose up --build
```

This starts:
- PostgreSQL on port 5432
- RabbitMQ on port 5672 (UI: http://localhost:15672)
- Auth Service on port 8001
- Nginx on port 80

### Step 3: Test the Auth Service

**Register a new patient:**
```bash
curl -X POST http://localhost/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rusiru@example.com",
    "password": "password123",
    "first_name": "Rusiru",
    "last_name": "Test",
    "role": "patient"
  }'
```

**Login:**
```bash
curl -X POST http://localhost/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rusiru@example.com",
    "password": "password123"
  }'
```

**Refresh token:**
```bash
curl -X POST http://localhost/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'
```

**Logout:**
```bash
curl -X POST http://localhost/api/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'
```

---

## 🔑 API Endpoints

| Method | Endpoint              | Description                          | Auth Required |
|--------|-----------------------|--------------------------------------|---------------|
| POST   | /api/auth/register    | Register new patient or doctor       | No            |
| POST   | /api/auth/login       | Login and get tokens                 | No            |
| POST   | /api/auth/logout      | Invalidate refresh token             | No            |
| POST   | /api/auth/refresh     | Get new access token                 | No            |
| GET    | /api/auth/validate    | Validate JWT (internal use only)     | Bearer JWT    |
| GET    | /health               | Service health check                 | No            |

---

## 🐰 RabbitMQ Events Published

| Routing Key      | Exchange    | When                        | Subscribers               |
|------------------|-------------|-----------------------------|---------------------------|
| user.registered  | user_events | New user registers          | patient-service, notification-service |

---

## 🐳 Docker Commands

```bash
# Start all services
docker-compose up --build

# Start in background
docker-compose up -d --build

# View logs
docker-compose logs -f auth-service

# Stop all services
docker-compose down

# Stop and remove volumes (wipe database)
docker-compose down -v

# Rebuild just auth service
docker-compose up --build auth-service
```

---

## ☸️ Kubernetes Deployment

```bash
# Create namespace
kubectl create namespace healthcare

# Apply all K8s manifests
kubectl apply -f k8s/auth-service/

# Check deployment status
kubectl get pods -n healthcare -l app=auth-service

# View logs
kubectl logs -n healthcare -l app=auth-service -f

# Port forward for testing
kubectl port-forward -n healthcare svc/auth-service 8001:8001
```

---

## 🔒 Security Notes

- Passwords hashed with **bcrypt** (cost=12)
- **Refresh tokens** stored as SHA-256 hashes (never plaintext)
- **Token rotation**: refresh token is replaced on every use
- JWT access tokens expire in **15 minutes**
- **Rate limiting** applied at Nginx level

---

## ⚙️ Environment Variables

| Variable                   | Required | Default | Description                   |
|----------------------------|----------|---------|-------------------------------|
| APP_ENV                    | No       | development | Environment name         |
| PORT                       | No       | 8001    | Server port                   |
| DATABASE_URL               | Yes      | -       | PostgreSQL connection string  |
| RABBITMQ_URL               | No       | -       | RabbitMQ AMQP URL             |
| JWT_SECRET                 | Yes      | -       | Access token signing secret   |
| JWT_REFRESH_SECRET         | Yes      | -       | Refresh token signing secret  |
| ACCESS_TOKEN_TTL_MINUTES   | No       | 15      | Access token lifetime         |
| REFRESH_TOKEN_TTL_DAYS     | No       | 7       | Refresh token lifetime        |
