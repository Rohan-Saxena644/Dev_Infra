# DevInfra - GitHub Deployment Platform

DevInfra is a lightweight deployment platform inspired by services like Render and Heroku. It automates the process of cloning GitHub repositories, building Docker images, and deploying applications inside isolated containers while tracking deployment status in PostgreSQL.

## Features

* Register GitHub repositories
* Trigger deployments through REST APIs
* Asynchronous deployment queue using Go workers
* Clone repositories directly from GitHub
* Build Docker images automatically
* Run containers with dynamically allocated ports
* Track deployment lifecycle (Queued → Running → Success / Failed)
* PostgreSQL database with sqlc generated queries
* Graceful shutdown using Context
* Dependency Injection for services and workers
* Temporary workspace cleanup after deployments

## Tech Stack

* Go
* PostgreSQL
* pgx
* sqlc
* Chi Router
* Docker
* Git CLI
* Worker Pools
* REST APIs

## Architecture

```
Client
   │
   ▼
REST API
   │
Middleware
   │
Service Layer
   │
PostgreSQL (sqlc)
   │
Deployment Queue
   │
Worker Pool
   │
Git Clone
   │
Docker Build
   │
Docker Run
```

## Current Deployment Flow

1. Create Project
2. Store GitHub Repository
3. Create Deployment
4. Queue Deployment
5. Worker Picks Job
6. Clone Repository
7. Build Docker Image
8. Run Docker Container
9. Save Dynamic Port
10. Mark Deployment Successful

## Future Improvements

* AWS EC2 Remote Deployment
* Docker Compose Support
* Authentication & Authorization
* Redis Caching
* Rate Limiting
* Deployment Logs
* CI/CD Pipelines
* React / Next.js Dashboard
* Metrics & Monitoring
* Multi-user Support

## Learning Goals

This project is built to explore backend systems engineering concepts including concurrency, deployment automation, Docker, cloud infrastructure, distributed systems, and production backend architecture.
