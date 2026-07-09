# DevInfra Implementation Plan

This plan follows the order I want to implement next. The goal is to keep the project moving fast while still understanding and owning the code myself.

## Current State

DevInfra already has the core MVP working:

- Go API with Chi
- PostgreSQL persistence through sqlc
- Project creation, listing, lookup, and deletion
- Deployment creation and async worker queue
- Git clone support
- Docker image build and container run
- Deployment status tracking
- Dynamic port assignment
- Running/stopped deployment detection
- Restart flow for stopped containers
- Dockerized DevInfra backend on EC2
- Next.js frontend dashboard and deployment detail pages

The project is now past the first MVP stage. The next work should make it more like a real deployment platform.

## Phase 1: Auth

Auth should come first because it changes the data model and API ownership rules. It is better to add this before production hardening, Docker Compose support, Kubernetes, or Redis.

### Backend Goals

- Add users table.
- Add password-based signup/login.
- Hash passwords with bcrypt.
- Issue JWT access tokens.
- Add auth middleware.
- Add `user_id` ownership to projects.
- Make project and deployment queries user-scoped.
- Prevent users from reading/deleting/deploying other users' projects.

### Suggested Tables

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

Later, add:

```sql
ALTER TABLE projects
ADD COLUMN user_id INT REFERENCES users(id);
```

For existing local data, either backfill to a test user or reset the database during development.

### Backend Endpoints

- `POST /auth/signup`
- `POST /auth/login`
- `GET /me`

Existing protected endpoints:

- `POST /projects`
- `GET /projects`
- `GET /projects/{id}`
- `DELETE /projects/{id}`
- `POST /projects/{id}/deploy`
- `GET /deployments`
- `POST /deployments/{id}/restart`

### JWT Shape

Keep it simple:

```json
{
  "user_id": 1,
  "email": "user@example.com",
  "exp": 1234567890
}
```

Use an environment variable:

```env
JWT_SECRET=some-long-random-secret
```

### Frontend Goals

- Add login page.
- Add signup page.
- Store JWT in localStorage at first.
- Attach `Authorization: Bearer <token>` to API calls.
- Redirect unauthenticated users to login.
- Add logout button.

Later, this can be improved with HttpOnly cookies.

### Done Criteria

- A new user can sign up.
- A user can log in.
- Projects are only visible to their owner.
- Deployments are only visible through owned projects.
- Logout removes the token.
- Refreshing the frontend keeps the session if token is still valid.

## Phase 2: Deploy Frontend To Vercel

Once auth is working locally, deploy the frontend.

### Steps

- Push frontend repo to GitHub.
- Import frontend repo in Vercel.
- Set environment variable:

```env
NEXT_PUBLIC_API_URL=http://13.206.173.255:8080
```

- Confirm backend CORS allows the Vercel domain.
- Test signup/login from Vercel.
- Test project creation.
- Test deployment.
- Test project detail page.
- Test restart/delete flows.

### Backend CORS Note

Current CORS is permissive with:

```http
Access-Control-Allow-Origin: *
```

That is fine for MVP. Later, restrict it to:

```text
http://localhost:3000
https://your-vercel-domain.vercel.app
```

### Done Criteria

- Frontend works from Vercel.
- Backend still works from curl.
- Browser network tab shows successful API calls to EC2.
- No hydration errors in frontend console.

## Phase 3: Production Hardening

Keep this phase small and practical. Do not disappear into observability tooling too early.

### Backend Health Endpoint

Add:

```http
GET /health
```

Return:

```json
{
  "status": "ok"
}
```

Optional later:

```json
{
  "status": "ok",
  "database": "ok",
  "docker": "ok"
}
```

### Deployment Logs

Currently Docker logs exist on the instance, but the app cannot return them.

Add backend endpoint:

```http
GET /deployments/{id}/logs
```

Implementation idea:

- Check deployment exists.
- Check user owns the deployment's project once auth exists.
- Build container name:

```go
deployment-%d
```

- Run:

```bash
docker logs --tail 200 deployment-{id}
```

- Return plain text or JSON.

Plain text is simpler:

```http
Content-Type: text/plain
```

### Frontend Logs

On project detail page:

- Add a "Logs" button or tab per deployment.
- Fetch `/deployments/{id}/logs`.
- Show output in a monospace panel.
- Add refresh button.

### Stop Button

Add:

```http
POST /deployments/{id}/stop
```

Behavior:

- Only allow successful/running deployments.
- Run `docker stop deployment-{id}`.
- Keep deployment row as `success`.
- Frontend already detects `ContainerRunning=false`, so restart button appears.

This matches the current lifecycle model.

### Error Logging

Improve handler logs where useful:

- log the internal error
- return a safe user-facing message
- include HTTP status in middleware logs

### Migrations

Add a clear migration process before changing schema heavily for auth.

Options:

- Keep using goose manually.
- Add a migration container/job later.
- Add documented EC2 commands for applying migrations.

### Done Criteria

- `/health` works.
- Deployment logs can be viewed from frontend.
- Running deployment can be stopped.
- Stopped deployment shows restart button.
- Backend errors are visible in container logs.

## Phase 4: Docker Compose Deployment Support

Right now DevInfra deploys single-container repositories with a Dockerfile. This phase adds support for repos with `docker-compose.yml`.

### Detection

After cloning repo:

- If `docker-compose.yml` or `compose.yml` exists, use Compose flow.
- Else use existing Dockerfile flow.

### Compose Deployment Strategy

For each deployment:

- Clone repo to `./tmp/deployment-{id}`.
- Use a unique project name:

```bash
docker compose -p deployment-{id} up -d --build
```

This prevents container/network name collisions.

### Important Questions

- How does DevInfra expose ports for compose apps?
- Should user define ports in compose file?
- Should DevInfra require one public service?
- Should DevInfra parse compose YAML?

MVP approach:

- Let compose file define its own ports.
- Store deployment status only.
- Logs endpoint can call:

```bash
docker compose -p deployment-{id} logs --tail 200
```

Later approach:

- Parse compose file.
- Enforce a public service.
- Allocate ports dynamically.

### Stop/Delete

For compose deployments:

```bash
docker compose -p deployment-{id} stop
docker compose -p deployment-{id} down --rmi local
```

### Data Model

Add deployment type:

```text
dockerfile
compose
```

Possible column:

```sql
ALTER TABLE deployments
ADD COLUMN deployment_type TEXT NOT NULL DEFAULT 'dockerfile';
```

### Done Criteria

- Dockerfile repos still deploy.
- Compose repos deploy.
- Compose deployments can be stopped.
- Compose deployments can be deleted.
- Compose logs can be fetched.

## Phase 5: Kubernetes

Do this after Docker and Compose support are stable. Kubernetes should become another deployment runtime, not a rewrite of everything at once.

### Learning Goal

Replace direct `docker run` with real Kubernetes primitives:

- Deployment
- Pod
- Service
- Ingress
- ConfigMap
- Secret

### Architecture Direction

Introduce an interface like:

```go
type Deployer interface {
    Deploy(...)
    Stop(...)
    Restart(...)
    Logs(...)
    Delete(...)
}
```

Implement:

- DockerDeployer
- ComposeDeployer
- KubernetesDeployer

Do not start with this abstraction too early. Add it when Docker Compose support makes duplication obvious.

### Kubernetes MVP

- Build image.
- Push image to registry.
- Create Kubernetes Deployment.
- Create Service.
- Expose through Ingress or NodePort.
- Track status.
- Fetch pod logs.

### New Requirements

Kubernetes likely requires:

- Container registry
- kubeconfig
- namespace management
- image naming strategy
- cleanup strategy

### Done Criteria

- One project can deploy to Kubernetes.
- Logs are visible.
- Restart works.
- Delete cleans up K8s resources.
- Docker-based deployment still works.

## Phase 6: Redis

Redis comes last because it becomes most useful once auth and production flows exist.

### Rate Limiting

Use Redis for:

- login attempts
- signup attempts
- deploy endpoint limits

Example limits:

- login: 5 attempts per minute per IP/email
- deploy: 3 deployments per project per 5 minutes

### Queue

Current worker queue is in-memory:

```go
Queue: make(chan int32, 100)
```

This is fine for MVP, but jobs disappear if the backend restarts.

Redis-backed queue options:

- Simple Redis list
- Streams
- Asynq library

For learning, start with Redis list or streams before adding a library.

### Caching

Useful but lower priority:

- cache project list per user briefly
- cache deployment status briefly

Avoid caching too early. The database is not the bottleneck yet.

### Done Criteria

- Login is rate-limited.
- Deploy endpoint is rate-limited.
- Deployment queue survives backend restart.
- Worker can pick pending jobs after restart.

## Suggested Final README Story

When complete, the README should tell this story:

DevInfra is a self-hosted deployment platform built with Go, PostgreSQL, Docker, and Next.js. Users can connect GitHub repositories, trigger deployments, inspect logs, stop/restart containers, and manage deployment history. The platform supports Dockerfile-based apps, Docker Compose apps, and later Kubernetes-backed deployments.

## Personal Rule For The Rest Of The Project

I should write as much implementation code myself as possible. AI help should be used for:

- explaining concepts
- debugging specific errors
- reviewing design decisions
- generating small examples
- pointing out risks

The goal is not just to finish DevInfra. The goal is to understand it well enough to confidently explain, debug, and extend every major subsystem.
