# ── Stage 1: Build ──────────────────────────────────────────────
FROM golang:1.26-alpine AS builder
WORKDIR /app

# Copy dependency files first — Docker caches this layer.
# Only re-downloads modules when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .

# CGO_ENABLED=0 → fully static binary (no glibc dependency)
# GOOS=linux    → cross-compile if you're on Windows
RUN CGO_ENABLED=0 GOOS=linux go build -o devinfra ./cmd/api

# ── Stage 2: Run ────────────────────────────────────────────────
FROM alpine:3.20

# Your worker calls exec.Command("git") and exec.Command("docker")
# Both must exist in the container
RUN apk add --no-cache git docker-cli

WORKDIR /app

# Copy only the compiled binary from stage 1
COPY --from=builder /app/devinfra .

EXPOSE 8080

CMD ["./devinfra"]