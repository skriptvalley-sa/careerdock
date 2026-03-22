# VPS Deployment Guide (Dev Mode)

This documents the canonical setup for running CareerDock in **development mode** on a public VPS behind Nginx.

## Architecture

```
Internet → Nginx (:443/:80)
              ├── /api/*        → Go API     (127.0.0.1:8080)
              └── everything    → Next.js    (127.0.0.1:3000)

Docker Compose (localhost only):
  ├── PostgreSQL  (127.0.0.1:5432)
  ├── Redis       (127.0.0.1:6379)
  ├── MinIO       (127.0.0.1:9000/9001)
  └── Mailhog     (127.0.0.1:1025/8025)
```

## Prerequisites

- Ubuntu/Debian VPS with Docker, Go >= 1.25, Node.js >= 20
- Nginx installed and configured for the domain
- DNS A record pointing `careerdock.skriptvalley.com` to the VPS IP

## Setup

### 1. Clone and configure

```bash
git clone https://github.com/skriptvalley/careerdock.git
cd careerdock
cp .env.example .env
```

Edit `.env` for the VPS:

```env
# Change these for VPS deployment:
ALLOWED_ORIGINS=https://careerdock.skriptvalley.com
NEXT_PUBLIC_API_URL=https://careerdock.skriptvalley.com
```

### 2. Start infrastructure with VPS-safe ports

Use the VPS compose override to bind services to localhost only:

```bash
docker compose -f docker-compose.yml -f docker-compose.vps.yml up -d
```

### 3. Run dev setup and start

```bash
./scripts/dev.sh setup
./scripts/dev.sh start
```

The setup step will:
- Check prerequisites (go, air, node, docker)
- Install dev tools (air, golangci-lint)
- Sync `NEXT_PUBLIC_*` vars from root `.env` to `frontend/.env.local`
- Install frontend dependencies

### 4. Configure Nginx

Example Nginx config for `/etc/nginx/sites-available/careerdock`:

```nginx
server {
    listen 80;
    server_name careerdock.skriptvalley.com;

    # API routes → Go backend
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Webhook routes → Go backend (no auth)
    location /api/webhooks/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Everything else → Next.js frontend
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support for Next.js HMR
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

Enable and reload:

```bash
ln -sf /etc/nginx/sites-available/careerdock /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx
```

## Environment Split

| Variable | Where it's used | Where it must be set |
|----------|----------------|---------------------|
| `DATABASE_URL`, `REDIS_URL`, etc. | Go backend | Root `.env` |
| `NEXT_PUBLIC_API_URL` | Next.js (browser) | `frontend/.env.local` |
| `ALLOWED_ORIGINS` | Go backend (CORS) | Root `.env` |

The `scripts/dev.sh` automatically syncs `NEXT_PUBLIC_*` vars from the root `.env` into `frontend/.env.local` during setup and before each frontend start/restart. **Do not edit `frontend/.env.local` manually** — it will be overwritten.

## Docker Compose: Local vs VPS

| Setup | Command |
|-------|---------|
| **Local dev** (laptop) | `docker compose up -d` |
| **VPS dev** (public) | `docker compose -f docker-compose.yml -f docker-compose.vps.yml up -d` |

The VPS override binds Postgres, Redis, MinIO, and Mailhog to `127.0.0.1` only, preventing external access to internal data services.

## Troubleshooting

### Stale frontend processes after restart
`scripts/dev.sh` kills the entire process group on restart. If orphan `next-server` processes persist:

```bash
./scripts/dev.sh restart frontend
# or manually:
lsof -i :3000 -t | xargs kill -9
```

### Frontend shows localhost URLs
Check that `NEXT_PUBLIC_API_URL` in `frontend/.env.local` points to the public hostname, not `localhost:8080`. Run `./scripts/dev.sh restart frontend` to re-sync.

### Service worker caching stale assets
Service worker is disabled in development mode. If you see stale assets, hard-refresh the browser (Ctrl+Shift+R) or clear browser storage.
