# AI Symptom Checker Service

Stateless service (team guide: **no database**). Accepts patient-reported symptoms, calls **OpenAI** or **Google Gemini**, and returns a **suggested specialty** plus **preliminary notes** (non-diagnostic). Port **8008**.

- **JWT**: Validates tokens via **auth-service** `GET /auth/validate` (same pattern as doctor-service).
- **Authorization**: `POST /symptoms/check` requires role **`patient`** only.
- **Independence**: Own `go.mod`; only shared dependency is `healthcare-platform/pkg` (logger). No imports from other services.

## Environment

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8008` | HTTP port |
| `AUTH_SERVICE_URL` | Yes | `http://localhost:8001` | Auth service base URL |
| `AI_PROVIDER` | Yes | `openai` | `openai` or `gemini` |
| `OPENAI_API_KEY` | If OpenAI | — | API key |
| `OPENAI_MODEL` | No | `gpt-4o-mini` | Chat model |
| `GEMINI_API_KEY` | If Gemini | — | API key |
| `GEMINI_MODEL` | No | `gemini-2.5-flash` | Model id ([rate limits](https://ai.google.dev/gemini-api/docs/rate-limits)) |

Copy `.env.example` to `.env`.

## Run locally

1. Start **auth-service** (for `/auth/validate`).
2. Register/login as a **patient** and obtain a Bearer token.
3. From this directory:

```bash
make run
```

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Liveness |
| GET | `/ready` | No | Readiness |
| POST | `/symptoms/check` | Bearer JWT, role `patient` | Symptom analysis |

**Request body:**

```json
{
  "symptoms": "I have had a pressing chest discomfort for two hours with shortness of breath.",
  "optional_context": "Age 55, history of hypertension"
}
```

**Success:** `200` with `data.suggested_specialty`, `data.preliminary_notes`, `data.disclaimer`.

Through **Nginx** (compose): `http://localhost/api/ai/symptom/check` (proxies to this service’s `/symptoms/check`).

## Troubleshooting

- **Google quota / `limit: 0` on free tier** for an old model: set **`GEMINI_MODEL=gemini-2.5-flash`** (or `gemini-2.5-flash-lite`), recreate the container, and check [rate limits](https://ai.google.dev/gemini-api/docs/rate-limits). The API may return **429** with a clear message when quota is hit.

## Docker

Build from repository root:

```bash
docker build -f services/ai-symptom-service/Dockerfile -t ai-symptom-service:latest .
```
