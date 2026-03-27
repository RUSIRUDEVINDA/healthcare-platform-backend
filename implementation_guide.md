# Healthcare Platform — Backend Implementation Guide

Welcome to the **Healthcare Platform Backend** repository. This document outlines our microservices architecture, folder structure, coding patterns, and contribution workflow to ensure consistency across our 4-person development team.

---

## 🏗️ Architecture Overview

The system is built using a **Go-based Microservices Architecture** designed for high scalability and clear separation of concerns.

- **Frontend Connectivity**: All requests pass through an **Nginx Inverse Proxy** (acting as an API Gateway on port 80).
- **Service Communication**: Primarily asynchronous through **RabbitMQ**.
- **Data Persistence**: Each service manages its own **PostgreSQL** database (Database-per-Service pattern).
- **Deployment**: Configured for **Docker Compose** (local dev) and **Kubernetes** (staging/prod).

---

## 📁 Full System Folder Structure

Our project follows a **Multi-Module Workspace pattern** (using `go.work`). Below is the complete structure for our 11 current and planned services.

```text
healthcare-platform-backend/
├── services/                    # Autonomous Microservices (Each has its own go.mod)
│   ├── auth-service/            # Authentication & Identity Management
│   │   ├── cmd/main.go          # Entry point (Main application setup)
│   │   ├── internal/            # Service-specific logic
│   │   │   ├── config/          # .env configuration loading
│   │   │   ├── handler/         # HTTP Handlers (Controllers)
│   │   │   ├── middleware/      # Service-local middleware
│   │   │   ├── model/           # DTOs and database entities
│   │   │   ├── repository/      # Database interactions (SQL)
│   │   │   └── service/         # Core business logic
│   │   ├── migrations/          # SQL database migration files
│   │   ├── Dockerfile           # Multi-stage Docker build
│   │   ├── Makefile            # Service-level automation (make run, migrate, etc.)
│   │   ├── .env                 # Local secrets (ignored by git)
│   │   └── go.mod               # Service-specific Go dependencies
│   │
│   ├── patient-service/         # Patient profiles & health records
│   ├── doctor-service/          # Doctor profiles & availability
│   ├── appointment-service/     # Scheduling & booking engine
│   ├── telemedicine-service/    # WebRTC/Video conferencing logic
│   ├── payment-service/         # Stripe/PayPal integration & invoicing
│   ├── notification-service/    # Email, SMS, and Push notifications
│   ├── ai-symptom-service/      # ML logic & symptom analysis
│   │   └── internal/integrations/ # External ML API clients (OpenAI/Ollama)
│   ├── admin-service/           # Back-office platform management
│   ├── api-gateway/             # Custom routing & aggregate logic (Go-based)
│   └── file-storage-service/    # S3/Cloudinary/Local storage management
│
├── pkg/                         # Global Shared Utilities (Root-level module)
│   ├── jwt/                     # Shared JWT token generation & verification
│   ├── logger/                  # Global logger (Zap/Logrus) wrapper
│   ├── rabbitmq/                # Reusable RabbitMQ connection & pub/sub logic
│   ├── middleware/              # Global middleware (Auth verification, Cors)
│   ├── utils/                   # Shared helpers (Hash, UUID, Pagination)
│   └── go.mod                   # Shared package dependencies
│
├── nginx/                       # Reverse Proxy & Load Balancer
│   └── nginx.conf              # API Gateway routing configuration
│
├── deployments/                 # Orchestration & Infrastructure
│   ├── docker/                  # Dockerfiles for shared dependencies
│   └── k8s/                     # Kubernetes manifests (Deployments, Services)
│
├── scripts/                     # Shell scripts for automation
│   ├── migrate.sh               # Global migration runner
│   └── seed.sh                  # Global database seeder
│
├── docker-compose.yml           # Local dev environment orchestrator
├── go.work                      # Go workspace (Connects all modules)
├── README.md                    # Project high-level documentation
└── implementation_guide.md      # This technical documentation
```

---

## 🚀 Local Development Setup

To get started locally, follow these steps:

1. **Environment Variables**:
   Navigate to the service directory (e.g., `services/auth-service`) and copy `.env.example` to `.env` (if not already done).
   ```bash
   cp .env.example .env
   ```

2. **Infrastructure**:
   Run the backend infrastructure (PostgreSQL, RabbitMQ, Nginx) from the repository root:
   ```bash
   docker-compose up -d
   ```
   > [!IMPORTANT]
   > Our local PostgreSQL for `auth-service` is mapped to host port **5433** to avoid conflicts with native Windows/Linux Postgres services on port 5432.

3. **Running a Service**:
   You can run the service directly via Go (for development speed) or via an independent Docker container.
   ```bash
   cd services/auth-service
   make run
   ```

---

## 🛠️ Service Internal Pattern

Each service follows a **Clean Architecture** pattern to isolate business logic from database/transport layers.

| Layer | Responsibility | Location |
| :--- | :--- | :--- |
| **Handler** | HTTP/gRPC transport, request parsing, response encoding. | `internal/handler/` |
| **Service** | Core business logic, validation, orchestration. | `internal/service/` |
| **Repository** | Data access, database queries (SQL). | `internal/repository/` |
| **Model** | Plain Go structs representing entities. | `internal/model/` |
| **Config** | Environment variable loading and validation. | `internal/config/` |

---

## 🧪 Testing & Quality

- **Unit Testing**: Place `_test.go` files alongside the functions they test.
- **Run Tests**: Use `make test` within a service folder to run all nested tests.
- **Linting**: We use `golangci-lint`. Ensure it passes before pushing code.
- **Migrations**: Never modify existing migrations. Always create a new `up`/`down` pair in the `migrations/` folder.

---

## 🌿 Contribution Workflow (Git)

To maintain a clean main branch, follow this branching strategy:

1. **Main Branch**: `main` (Locked — requires Pull Request + Review).
2. **Branch Naming**:
   - Features: `feature/auth-password-reset`
   - Bugfixes: `bugfix/fix-jwt-expiration`
   - Chore: `chore/update-readme`

### Typical Pushing Flow:
1. Create a branch from `main`.
2. Implement logic + **Tests**.
3. Run `make tidy` and `make test`.
4. Push and open a Merge Request targeting `main`.
5. Address review feedback and merge.

---

## 📜 Shared Utilities (`pkg/`)

If you build a utility (e.g., a custom logger, unique hash generator) that could benefit other services, place it in the `pkg/` directory.

> [!TIP]
> Always check `pkg/` BEFORE implementing a generic utility to avoid code duplication!

---

## 💡 Common Commands (Makefile)

Within each service folder, helpful shortcuts are available:
- `make run`: Run locally with auto-env loading.
- `make build`: Build a production binary.
- `make migrate-up`: Apply database migrations.
- `make tidy`: Clean up go modules.
- `make help`: List all available commands.
