# 🏥 Healthcare Platform - Microservices Ecosystem

[![Tech Stack](https://img.shields.io/badge/Stack-Go%20|%20React%20|%20PostgreSQL-0ea5e9?style=for-the-badge)](https://github.com/RUSIRUDEVINDA/healthcare-platform-backend)
[![Uptime](https://img.shields.io/badge/Architecture-Event--Driven-818cf8?style=for-the-badge)](https://rabbitmq.com)
[![Nginx](https://img.shields.io/badge/nginx-%23009639.svg?style=for-the-badge&logo=nginx&logoColor=white)](https://nginx.org)
[![Ngrok](https://img.shields.io/badge/ngrok-1F1E1E.svg?style=for-the-badge&logo=ngrok&logoColor=white)](https://ngrok.com)
[![Gin](https://img.shields.io/badge/Gin-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://gin-gonic.com)
[![RabbitMQ](https://img.shields.io/badge/RabbitMQ-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)](https://rabbitmq.com)
[![TypeScript](https://img.shields.io/badge/typescript-%23007ACC.svg?style=for-the-badge&logo=typescript&logoColor=white)](https://typescriptlang.org)

A premium, highly-distributed telemedicine and healthcare management platform built with **Go (Gin)**, **React (TypeScript)**, **Nginx**, and **RabbitMQ**. This project leverages a microservices architecture to provide scalable patient care, doctor channeling, and secure payment processing.

---

## 🛠️ Technical Stack & Infrastructure

The project utilizes a robust and modern tech stack selected for performance, security, and developer productivity.

#### Backend Technology Layer
- **Core Language**: **Go (Golang)** – Used for its legendary performance in concurrent microservices.
- **Web Engine**: **Gin** – A high-performance HTTP web framework (Gin Gonic).
- **Event Messaging**: **RabbitMQ 3.12** – Facilitates asynchronous, event-driven communication (Pub/Sub).
- **Relational Storage**: **PostgreSQL 17** – Utilized as the primary data store with specialized schemas for each service.
- **Reverse Proxy & Entry**: **NGINX & Custom API Gateway** – Handles path-based routing, TLS termination, and request scrubbing.

#### Frontend & DevOps
- **Frontend Core**: **TypeScript** & **React** – Ensuring type-safety and robust UI component architecture.
- **Tunneling**: **Ngrok** – Utilized for secure tunneling of local services for external webhook testing (e.g., PayHere callbacks).
- **Containerization**: **Docker** & **Kubernetes** – Providing environment parity and orchestration.

---

## 🏗️ System Architecture & Ports

The platform uses a **Database-per-Service** pattern with **RabbitMQ** orchestrating eventual consistency.

| Service | Port | Primary Responsibility | Essential .env Keys |
| :--- | :---: | :--- | :--- |
| **🔐 Auth** | `8001` | JWT & Identity | `JWT_SECRET`, `ACCESS_TOKEN_TTL` |
| **👤 Patient** | `8002` | Patient Profiles | `INTERNAL_API_KEY`, `DATABASE_URL` |
| **🩺 Doctor** | `8003` | Practitioner Data | `AUTH_SERVICE_URL`, `DATABASE_URL` |
| **📅 Appointment** | `8004` | Booking Engine | `JITSI_BASE_URL`, `INTERNAL_API_KEY` |
| **💳 Payment** | `8005` | PayHere Checkout | `PAYHERE_MERCHANT_ID`, `PAYHERE_SECRET` |
| **🔔 Notification**| `8006` | Email/SMS Comms | `SMTP_HOST`, `SMTP_PASS`, `SENDER_EMAIL` |
| **⚙️ Admin** | `8007` | Resource Control | `RABBITMQ_URL`, `DATABASE_URL` |
| **🤖 AI Symptom** | `8008` | Medical Triage | `OPENAI_API_KEY`, `AI_PROVIDER` |
| **📹 Telemedicine**| `8009` | Session Handling | `APPOINTMENT_SERVICE_URL` |
| **📂 File Storage** | `8010` | Media & Scans | `MAX_IMAGE_SIZE`, `DATABASE_URL` |
| **🛠️ Support** | `8011` | User Ticketing | `DATABASE_URL`, `PORT` |
| **🚪 API Gateway** | `8888` | Traffic Routing | `CORS_ALLOWED_ORIGINS` |

---

## 🛠️ REST API Reference

All services are accessible via the **API Gateway** on port `8888` or directly for internal development.

### Identity & Access (Auth)
- `POST /v1/auth/register` - Create new patient/doctor.
- `POST /v1/auth/login` - Authenticate and receive JWT.
- `GET /v1/auth/me` - Validate session.

### Clinical Management (Appointments)
- `GET /v1/appointments` - List all current appointments.
- `POST /v1/appointments` - Book a slot (supports `pay_now` or `pay_later`).
- `GET /v1/slots` - List doctor availability windows.
- `PUT /v1/appointments/:id/cancel` - Cancel a booking and release slot.

### Practitioner & Profile
- `GET /v1/doctors` - Search clinical practitioners.
- `GET /v1/patients/profile` - Fetch patient-specific history.
- `PUT /v1/doctor/profile` - Update doctor professional bio.

### Financials (Payment)
- `POST /v1/payments/checkout` - Generate PayHere checkout session.
- `POST /v1/payments/webhook` - PayHere IPN notification handler.

### Intelligent Triage (AI)
- `POST /v1/symptoms/analyze` - Process natural language symptoms for triage.

---

## 📡 Messaging (RabbitMQ Events)

The system utilizes an internal event-bus to maintain sync across 12 isolated databases.

| Event Exchange | Routing Key | Purpose |
| :--- | :--- | :--- |
| `appointment.events` | `appointment.booked` | Notifies Payment & Notification services. |
| `appointment.events` | `appointment.cancelled` | Releases held payments and updates capacity. |
| `payment.events` | `payment.completed` | Triggers final confirmation in Appointment service. |
| `user.events` | `patient.deleted` | Cascades data deletion across all 11 other services. |

---

## 🚀 Development Setup

### 1. Prerequisite
- Docker & Docker Compose
- Go 1.25+ (for local development)
- Node.js 20+ (for frontend development)

### 2. Startup
```bash
docker-compose up --build
```
*   **Frontend**: `http://localhost:80`
*   **RabbitMQ Portal**: `http://localhost:15672` (u: admin, p: password123)
*   **Health Check**: Every service exposes a `/health` endpoint.

### 3. Database Access
If using local Docker DBs, ports are mapped in the `5433` to `5443` range. See `docker-compose.yml` for specific service-to-port mapping.

---

## 🎨 Design Philosophy
The platform utilizes a **"Calm Operational Workspace"** aesthetic:
- **Typography**: Inter (Sans-Serif) for maximum legibility.
- **Visuals**: Glassmorphism, card-based layouts, and vibrant slate-teal-indigo color palettes.
- **Experience**: Premium minimalist email designs and human-readable formatting for all transactional data.

---

> [!IMPORTANT]  
> For production deployments, ensure all `.env` secrets are replaced with cryptographically secure keys and that Nginx is configured for SSL termination.

> [!TIP]
> **Pro-Tip**: For telemedicine video sessions, the `Telemedicine Service` generates short-lived Jitsi tokens. Ensure your `JWT_SECRET` is consistent across all services to allow for seamless token validation.
