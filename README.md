# shortener

[![CI](https://github.com/ertugruldasgin/shortener/actions/workflows/ci.yml/badge.svg)](https://github.com/ertugruldasgin/shortener/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ertugruldasgin/shortener.svg)](https://pkg.go.dev/github.com/ertugruldasgin/shortener)
[![License: GPL v2](https://img.shields.io/badge/license-GPLv2-blue.svg)](LICENSE)

A self-hosted URL shortener in Go, with custom aliases, link expiry, and click
analytics. Built around a hexagonal core: the domain package knows nothing about
HTTP, PostgreSQL, or Redis.

## Features

- Random slugs (CSPRNG + base64url, 36 bits of entropy) or custom aliases
- `307 Temporary Redirect` on every hit, so clicks are always counted
- Optional link expiry, expired links return `410 Gone`
- Asynchronous click tracking: referrer, user agent, timestamp
- Redis cache-aside on the redirect path
- Per-client rate limiting, also backed by Redis
- Prometheus metrics and provisioned Grafana dashboards
- Migrations embedded in the binary, applied on startup

## Quick start

```bash
git clone https://github.com/ertugruldasgin/shortener.git
cd shortener
cp .env.example .env     # set POSTGRES_PASSWORD, API_TOKEN, ADMIN_TOKEN, GRAFANA_PASSWORD
make up
```

The stack (app, Postgres, Redis, Prometheus, Grafana) comes up on
`localhost:8080`, with Grafana on `localhost:3000`.

For local development against a `go run` binary instead of a container:

```bash
make dev    # starts Postgres + Redis, runs tests, then the app
```
