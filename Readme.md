# DevInfra

DevInfra is a self-hosted deployment platform built with Go and Next.js. It turns public GitHub repositories into managed Docker or constrained Docker Compose deployments with worker-based builds, JWT ownership, encrypted environment variables, logs, lifecycle controls, PostgreSQL, and Redis.

The project is inspired by platforms such as Render and Heroku, but focuses on exposing the infrastructure behind a deployment platform in a compact, understandable codebase.

## Highlights

- Deploy public GitHub repositories through a REST API or web dashboard.
- Build and run root-level Dockerfiles automatically.
- Detect and deploy supported Docker Compose applications.
- Process deployments asynchronously with a bounded queue and three Go workers.
- Track queued, running, successful, and failed deployment states in PostgreSQL.
- Isolate projects and deployments by authenticated user ownership.
- Store project environment variables using AES-256-GCM encryption.
- Inject variables at deployment time without returning their values to the frontend.
- View deployment logs and stop, restart, or delete running workloads.
- Cache per-user project lists in Redis with automatic invalidation.
- Apply request, resource, build-time, and deployment-count limits.
- Clean temporary repositories, environment files, containers, images, and Compose resources.

## Architecture

```text
Next.js dashboard
       |
       v
Go REST API (Chi)
       |
       +--> CORS, logging, rate limits, JWT authentication
       |
       +--> Service layer --> PostgreSQL (pgx + sqlc)
       |                         |
       |                         +--> Redis project cache
       |
       +--> Deployment queue --> Worker pool
                                  |
                                  +--> Git clone
                                  +--> Dockerfile or Compose validation
                                  +--> Docker build and run
                                  +--> Status, logs, and cleanup
```

## Deployment Flow

1. A user signs up or logs in and receives a 24-hour JWT.
2. The user registers a public GitHub repository as a project.
3. Optional project environment variables are encrypted before storage.
4. A deployment record is created with the `queued` status.
5. A worker marks it `running`, clones the repository, and prepares its environment.
6. DevInfra detects either a Dockerfile or a supported Compose file.
7. The workload is built with time and resource limits and exposed on a dynamic host port.
8. The deployment is marked `success` or `failed`.
9. Successful workloads can be inspected, stopped, restarted, or deleted from the dashboard.

## Supported Repositories

### Dockerfile

- Place a `Dockerfile` in the repository root.
- Expose one TCP application port with `EXPOSE`.
- Port `80` is preferred when multiple ports exist.
- If no port is exposed, DevInfra defaults to container port `80`.
- The resulting container receives a 512 MB memory limit, 0.5 CPU, and a 100 PID limit.

### Docker Compose

DevInfra detects `compose.yaml`, `compose.yml`, `docker-compose.yaml`, or `docker-compose.yml` in the repository root.

- A deployment may contain up to four services.
- Exactly one service must be publicly exposed.
- Services named `frontend`, `web`, `app`, or `client` are preferred automatically.
- Alternatively, set the `devinfra.public=true` service label.
- Use `devinfra.port` when the selected service publishes multiple ports.
- Project variables should be referenced as `${VARIABLE_NAME}` in the Compose file.
- External resources, host bind mounts, host networking, privileged mode, devices, Compose secrets, and Compose configs are intentionally rejected.

Example public-service labels:

```yaml
services:
  web:
    labels:
      devinfra.public: "true"
      devinfra.port: "3000"
```

## Technology

| Area | Technology |
| --- | --- |
| Backend | Go, Chi, slog |
| Database | PostgreSQL 17, pgx, sqlc, Goose migrations |
| Authentication | JWT HS256, bcrypt |
| Secrets | AES-256-GCM |
| Cache | Redis |
| Runtime | Docker Engine, Docker Compose v2, Git CLI |
| Frontend | Next.js 16, React 19, TypeScript, Tailwind CSS |
| Hosting | AWS EC2 backend, Vercel frontend |

## Local Setup

### Requirements

- Go 1.26 or newer
- Docker Engine or Docker Desktop with Compose v2
- PostgreSQL and Redis through the included Compose stack
- [Goose](https://github.com/pressly/goose) for migrations
- [sqlc](https://sqlc.dev/) when changing SQL queries

### 1. Clone and configure

```bash
git clone https://github.com/Rohan-Saxena644/Dev_Infra.git
cd Dev_Infra
cp .env.example .env
```

Replace the placeholder values in `.env`. Generate strong values with:

```bash
openssl rand -hex 24
openssl rand -hex 32
```

Use the first value for `POSTGRES_PASSWORD`. Generate separate 64-character hexadecimal values for `SECRET` and `ENV_ENCRYPTION_KEY`. Keep `ENV_ENCRYPTION_KEY` permanently; changing it makes stored project variables unreadable.

### 2. Start dependencies and migrate

```bash
docker compose up -d postgres redis

set -a
source .env
set +a

goose -dir sql/schema postgres \
  "postgres://admin:${POSTGRES_PASSWORD}@127.0.0.1:15432/devinfra?sslmode=disable" up
```

### 3. Start DevInfra

```bash
docker compose up -d --build devinfra
docker compose ps
docker compose logs -f devinfra
```

The API listens on `http://localhost:8080`. An unauthenticated request to `/projects` should return `401 Unauthorized`.

### 4. Start the frontend

The dashboard is maintained in [frontend_Dev_Infra](https://github.com/Rohan-Saxena644/frontend_Dev_Infra).

```bash
git clone https://github.com/Rohan-Saxena644/frontend_Dev_Infra.git
cd frontend_Dev_Infra
npm install
```

Create `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

Then run:

```bash
npm run dev
```

Open `http://localhost:3000`, create an account, register a repository, add any required environment variables, and start a deployment.

## API Overview

Public routes:

| Method | Route | Purpose |
| --- | --- | --- |
| POST | `/auth/signup` | Create a Gmail-based account |
| POST | `/auth/login` | Receive a JWT |

Protected routes require `Authorization: Bearer <token>`:

| Method | Route | Purpose |
| --- | --- | --- |
| GET / POST | `/projects` | List or create projects |
| GET / DELETE | `/projects/{id}` | Read or delete an owned project |
| POST | `/projects/{id}/deploy` | Queue a deployment |
| PUT / DELETE | `/projects/{id}/environment/{name}` | Set or delete an encrypted variable |
| GET | `/projects/{id}/environment` | List variable names only |
| GET | `/deployments` | List owned deployments |
| GET | `/deployments/{id}/logs` | Read the latest deployment logs |
| POST | `/deployments/{id}/stop` | Stop a successful workload |
| POST | `/deployments/{id}/restart` | Restart a stopped workload |

## Operational Limits

- Three active deployments per user.
- One active deployment per project.
- Ten retained deployment records per project using FIFO cleanup.
- Fifty environment variables per project.
- Git clone timeout: 2 minutes.
- Dockerfile build timeout: 10 minutes.
- Compose deployment timeout: 15 minutes.
- Deployment logs: last 200 lines, capped at 256 KiB.
- API, authentication, project, and deployment actions have separate rate limits.
- HTTP body, header, read, write, idle, and graceful-shutdown limits are configured.

## Development

Run backend checks:

```bash
go test ./...
go vet ./...
```

After editing files in `sql/queries`, regenerate database code instead of changing generated files manually:

```bash
sqlc generate
```

Create every database schema change as a new numbered migration in `sql/schema`, then apply it with Goose.

Frontend checks:

```bash
npm run lint
npm run build
```

## Security Notes

- Passwords are hashed with bcrypt and never stored directly.
- JWT signing secrets must be at least 32 characters.
- Environment values are encrypted at rest and are never returned by the API.
- Git repository URLs are restricted to public `https://github.com/owner/repository` paths.
- Compose input is normalized and dangerous host-level options are rejected.
- Workloads receive CPU, memory, and PID limits.
- PostgreSQL and Redis bind to loopback only in the included production Compose configuration.
- Configure `ALLOWED_ORIGINS` explicitly for the deployed frontend.
- Serve the backend over HTTPS before connecting it to an HTTPS Vercel frontend.

## Current Scope

DevInfra is a strong portfolio and learning platform, but it should not be treated as a hardened public multi-tenant PaaS yet.

- The deployment queue is in memory, so queued work is not durable across API restarts.
- Workloads are exposed through dynamic ports rather than a domain-based reverse proxy.
- The API container controls the host Docker daemon through its socket, which is a significant trust boundary.
- Docker builds execute code from submitted repositories and should only accept trusted users and repositories.
- Authentication currently supports Gmail and passwords; OAuth and email verification are not implemented.
- Broader integration tests, metrics, tracing, health checks, and automated CI/CD remain future improvements.

## Project Repositories

- Backend: [Rohan-Saxena644/Dev_Infra](https://github.com/Rohan-Saxena644/Dev_Infra)
- Frontend: [Rohan-Saxena644/frontend_Dev_Infra](https://github.com/Rohan-Saxena644/frontend_Dev_Infra)

## Project Status

The core platform is implemented and end-to-end tested with authentication, encrypted configuration, Docker deployment, logs, stop, restart, and cleanup. Docker Compose support is implemented for repositories that satisfy the documented safety contract.
