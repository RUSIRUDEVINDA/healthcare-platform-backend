# Admin Service

Port: `8007`

This service is the back-office management service for the healthcare platform.
It exposes HTTP endpoints protected by JWT and requires the `admin` role claim.

## Responsibilities

- List all users
- Approve doctor registrations
- View appointments
- View transactions
- Deactivate users
- Maintain local admin-side mirrors of platform events through RabbitMQ

## Protected Endpoints

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| GET | `/admin/users` | List all users | Yes, admin JWT |
| PUT | `/admin/doctors/:id/verify` | Approve doctor registration | Yes, admin JWT |
| GET | `/admin/appointments` | View all appointments | Yes, admin JWT |
| GET | `/admin/transactions` | View payment history | Yes, admin JWT |
| DELETE | `/admin/users/:id` | Deactivate a user | Yes, admin JWT |

## Health Checks

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Liveness check |
| GET | `/ready` | Readiness check |

## Environment Variables

| Variable | Required | Default | Purpose |
|---|---|---|---|
| `APP_ENV` | No | `development` | Environment name |
| `PORT` | No | `8007` | HTTP port |
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `RABBITMQ_URL` | Yes | - | RabbitMQ connection string |
| `JWT_SECRET` | Yes | - | JWT access token secret |

## Event Consumers

The service mirrors platform data from RabbitMQ so the admin UI can inspect it later.

| Routing Key | Exchange | Purpose |
|---|---|---|
| `user.registered` | `user_events` | Store new user records |
| `appointment.booked` | `appointment_events` | Store appointment records |
| `payment.completed` | `payment_events` | Store transaction records |

## Run Locally

From the repository root:

```powershell
docker compose up --build -d
```

Check logs:

```powershell
docker compose logs -f admin-service
```

## How to Test with Postman

### 1. Start the stack
Make sure Docker Desktop is running, then run:

```powershell
docker compose up --build -d
```

### 2. Create a user and get a JWT from auth-service
Use these requests first so you have a token.

#### Register user
- Method: `POST`
- URL: `http://localhost/api/auth/register`
- Headers:
  - `Content-Type: application/json`
- Body raw JSON:
```json
{
  "email": "sandali.admin@example.com",
  "password": "Password123",
  "first_name": "Sandali",
  "last_name": "Admin",
  "role": "doctor"
}
```

#### Login
- Method: `POST`
- URL: `http://localhost/api/auth/login`
- Headers:
  - `Content-Type: application/json`
- Body raw JSON:
```json
{
  "email": "sandali.admin@example.com",
  "password": "Password123"
}
```

Save the returned `access_token`.

### 3. Promote that user to admin for testing
Run this in PowerShell:

```powershell
docker exec -it auth-db psql -U postgres -d auth_db -c "UPDATE users SET role='admin', is_verified=TRUE WHERE email='sandali.admin@example.com';"
```

Then login again to get a fresh access token with `role: admin`.

### 4. Use this header on every admin request
- `Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN`
- `Content-Type: application/json` only when the request has a JSON body

### 5. Test each endpoint

#### List users
- Method: `GET`
- URL: `http://localhost/api/admin/users`
- Headers:
  - `Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN`

Expected: `200 OK`

#### Verify doctor
- Method: `PUT`
- URL: `http://localhost/api/admin/doctors/<DOCTOR_ID>/verify`
- Headers:
  - `Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN`
  - `Content-Type: application/json`
- Body raw JSON:
```json
{
  "notes": "verified for onboarding"
}
```

Expected: `200 OK`

#### List appointments
- Method: `GET`
- URL: `http://localhost/api/admin/appointments`
- Headers:
  - `Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN`

Expected: `200 OK`

#### List transactions
- Method: `GET`
- URL: `http://localhost/api/admin/transactions`
- Headers:
  - `Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN`

Expected: `200 OK`

#### Deactivate user
- Method: `DELETE`
- URL: `http://localhost/api/admin/users/<USER_ID>`
- Headers:
  - `Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN`

Expected: `200 OK`

### 6. Check whether data is being mirrored
If you register users or publish events, check:

```powershell
docker compose logs -f admin-service
```

You should see mirrored event processing logs.

## Notes for Demo

- The admin service will return `403 Forbidden` if the JWT role is not `admin`.
- The route prefix is `/api/admin/...` through Nginx, which forwards to the service’s `/admin/...` routes.
- This service depends on events from auth, appointment, and payment flows to populate its admin tables.
