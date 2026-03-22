## Session 05 Feedback

Use this as an implementation-oriented follow-up based on the first live VPS deployment of CareerDock on Hostinger with Nginx in front of the app. The main focus here is making the dev deployment path reproducible, self-contained, and stable when the app is exposed through a public hostname.

## Requested Updates

### 1. Make `scripts/dev.sh` self-sufficient for Go and Air

Issue:
- The dev script currently assumes `go` and `air` are already available in `PATH`.
- On the VPS, the script failed under non-login execution because `/usr/local/go/bin` and `$HOME/go/bin` were not available in the environment inherited by the script.

Expected result:
- Bootstrap the required `PATH` entries inside `scripts/dev.sh` itself.
- Validate `go` and `air` before starting API and worker processes.
- Fail fast with a clear actionable error if either tool is missing.

### 2. Generate frontend env correctly for Next.js

Issue:
- The repo creates and uses a root `.env`, but the frontend is started from the `frontend/` directory.
- In that setup, Next.js did not reliably read the repo-root `.env`, so `NEXT_PUBLIC_API_URL` fell back to `http://localhost:8080`.
- On the VPS, that caused the browser to call `localhost:8080` instead of the public hostname behind Nginx.

Expected result:
- Generate or sync `frontend/.env.local` from the root `.env` for the required `NEXT_PUBLIC_*` values.
- Do this during `./scripts/dev.sh setup` and before frontend start or restart.
- Document clearly that frontend runtime env must exist in the frontend app directory when the app is launched from there.

### 3. Fix frontend restart and process cleanup in `scripts/dev.sh`

Issue:
- Restarting the frontend leaked orphan `next-server` processes.
- Multiple stale frontend processes stayed alive on ports `3000`, `3001`, and `3002`.
- That caused inconsistent asset serving, stale bundles, missing CSS, and broken restarts.

Expected result:
- Start managed processes in their own process groups or sessions.
- Stop the full frontend process tree, not just the parent PID.
- Remove stale PID files only after confirming the process tree is gone.
- Verify that the expected port is free before starting a replacement process.

### 4. Allow public dev hostname access in Next.js config

Issue:
- The app is being served in dev mode behind Nginx on a public hostname.
- Next.js dev emitted cross-origin warnings because the frontend was accessed through `careerdock.skriptvalley.com` rather than raw `localhost`.

Expected result:
- Add the proxied public dev hostname to `allowedDevOrigins` in the Next.js config.
- Keep the config explicit so public dev hosting behind Nginx works without asset or dev-server warnings.

### 5. Add a VPS-safe Docker Compose override

Issue:
- The base `docker-compose.yml` publishes Postgres, Redis, MinIO, and Mailhog on all interfaces.
- That is acceptable for local laptop development but not for a public VPS.

Expected result:
- Add a dedicated `docker-compose.vps.yml` override for VPS/dev-hosted setups.
- Bind internal data services to `127.0.0.1` only.
- Keep the base compose file optimized for normal local development.

### 6. Tighten deployment documentation for the public dev setup

Issue:
- Several of the live issues came from gaps between the intended local workflow and the actual VPS-hosted dev workflow.
- The current setup path does not make the env split, Nginx routing, and public-hostname assumptions explicit enough.

Expected result:
- Document the canonical public host and Nginx routing model clearly.
- Document that `/api/*` should proxy to the Go API and all other routes should proxy to the Next.js frontend.
- Document the difference between backend env and frontend env for VPS-hosted dev mode.

## Implementation Issues

### 1. Service worker should not interfere with public dev deployment

Issue:
- The frontend currently registers a service worker during development.
- On a public dev deployment, that makes debugging much harder because stale assets and old API behavior can survive reloads.

Expected result:
- Disable or gate service worker registration in dev mode.
- Only enable it in environments where asset caching behavior is intentional and stable.

### 2. Public dev mode should avoid stale-client failure modes

Issue:
- When the frontend was restarted or misconfigured, the browser could keep old cached assets while the server had moved on.
- That amplified otherwise recoverable deployment mistakes into full UI failures.

Expected result:
- Treat public dev hosting as a supported development mode with explicit safeguards.
- Reduce stale-client behavior by combining correct env generation, reliable process cleanup, and dev-only service worker gating.
