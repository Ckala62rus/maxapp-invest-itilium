# Production Deployment

The production stack is designed for an Ubuntu host with Docker Compose, host-issued Let's Encrypt certificates and nginx serving the Vue bundle over HTTPS.

## Services

- `web` builds the Vue app and runs nginx on ports `80` and `443`.
- `backend` runs the Go API on the internal compose network only.
- `redis` stores cache entries for repeated ITILIUM reads.
- `loki`, `prometheus` and `grafana` provide logs and metrics and are bound to localhost ports on the host.

Postgres is intentionally not started in production right now. The current backend uses `MemoryUserRepository` for temporary profile snapshots, Redis for cache and ITILIUM as the source of truth. The `migrations/` directory and dev migrate profile remain as a scaffold for a future persistent repository.

## Host Preparation

Install Docker Engine, the Compose plugin, the Loki Docker logging driver and certbot on the server. Make sure DNS for `CERT_DOMAIN` points to this host and ports `80` and `443` are open.

Create the certbot webroot once:

```bash
sudo mkdir -p /var/www/certbot
```

## Environment

Copy `.env.example` to `.env` on the server and fill production values:

```bash
CERT_DOMAIN=your.domain.example
WEB_HTTP_PORT=80
WEB_HTTPS_PORT=443

ITILIUM_BASE_URL=https://itilium.example/itilium/hs/Max
ITILIUM_LOGIN=...
ITILIUM_PASSWORD=...
ITILIUM_INSECURE_SKIP_VERIFY=false

MAX_BOT_TOKEN=...
AUTH_ACCESS_TOKEN_SECRET=replace-with-a-long-random-secret
AUTH_ALLOW_DEBUG_IDENTITY_HEADERS=false

GF_ADMIN_USER=admin
GF_ADMIN_PASSWORD=replace-me
GRAFANA_HTTP_PORT=3001
PROMETHEUS_HTTP_PORT=9090
LOKI_HTTP_PORT=3100
```

Do not commit `.env` or real credentials.

## First Certificate

The main production nginx config needs existing certificate files. After `.env` is filled, start the HTTP-only bootstrap nginx:

```bash
docker compose -f docker-compose.bootstrap.yml up --build -d
```

Then issue the certificate through certbot webroot:

```bash
sudo certbot certonly --webroot \
  -w /var/www/certbot \
  -d your.domain.example
```

Stop the bootstrap stack after the certificate is issued:

```bash
docker compose -f docker-compose.bootstrap.yml down
```

## Start And Check

Build and start the stack:

```bash
docker compose up --build -d
```

Check the public endpoints:

```bash
curl -fsS https://your.domain.example/healthz
curl -fsS https://your.domain.example/readyz
```

Check backend logs during first launch:

```bash
docker compose logs -f backend web
```

After the app is registered in MAX, set the mini app URL to `https://your.domain.example/` and verify `POST /api/v1/auth/max/validate`, `GET /api/v1/users/me` and a test ITILIUM ticket flow.

## Certificate Renewal

Renew certificates on the host with certbot, then reload the nginx container:

```bash
sudo certbot renew
docker compose exec web nginx -s reload
```

Use a cron/systemd timer for renewal on the server.
