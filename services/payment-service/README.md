# payment-service (PayHere)

## PayHere Sandbox + Local Development

PayHere works on localhost for browser redirects (return/cancel), but **PayHere cannot reach your localhost for `notify_url` (webhook)**.
Use a public tunnel (e.g., ngrok) for the webhook.

### 1) Configure env

Set these in `services/payment-service/.env` (or as environment variables when using docker-compose):

- `PAYHERE_ENV=sandbox`
- `PAYHERE_MERCHANT_ID=...`
- `PAYHERE_MERCHANT_SECRET=...`
- `PAYHERE_RETURN_URL=http://localhost:3000/payment/success`
- `PAYHERE_CANCEL_URL=http://localhost:3000/payment/cancel`
- `PAYHERE_NOTIFY_URL=https://<public-host>/api/v1/payments/webhook/payhere`

### 2) Run the service

If using docker-compose, expose the service to the host (already mapped in `docker-compose.yml`):

`docker-compose up --build payment-service payment-db rabbitmq`

### 3) Create a public webhook URL with ngrok

Run:

`ngrok http 8005`

Copy the HTTPS forwarding URL and set:

`PAYHERE_NOTIFY_URL=https://<id>.ngrok-free.app/api/v1/payments/webhook/payhere`

Restart `payment-service` after changing env vars.

### 4) API flow

1. Create a pending payment:

`POST /api/v1/payments/`

2. Create a PayHere checkout payload:

`POST /api/v1/payments/checkout`

This returns:

- `checkout_url` (PayHere endpoint)
- `fields` (key/value pairs your frontend must POST as an HTML form)

3. After payment, PayHere calls:

`POST /api/v1/payments/webhook/payhere`

The service verifies the signature, updates `payments.status`, and publishes `payment.completed` when it transitions to completed.

