# shortener

[![CI](https://github.com/ertugruldasgin/shortener/actions/workflows/ci.yml/badge.svg)](https://github.com/ertugruldasgin/shortener/actions/workflows/ci.yml)
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

To run the app directly instead of in a container:

```bash
make dev    # starts Postgres + Redis, runs tests, then the app
```

## API

Create a link:

```bash
curl -X POST localhost:8080/api/links \
  -H "Authorization: Bearer $API_TOKEN" \
  -d '{"target":"https://example.com"}'
```

```json
{ "slug": "exampl", "target": "https://example.com" }
```

With a custom alias and an expiry (any Go duration):

```bash
curl -X POST localhost:8080/api/links \
  -H "Authorization: Bearer $API_TOKEN" \
  -d '{"target":"https://example.com","alias":"docs","expires_in":"24h"}'
```

Delete a link (admin token, separate from the create token):

```bash
curl -X DELETE localhost:8080/api/links/docs \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

| Method   | Path                | Auth          | Response                   |
| -------- | ------------------- | ------------- | -------------------------- |
| `POST`   | `/api/links`        | `API_TOKEN`   | `201`, `400`, `409`, `429` |
| `DELETE` | `/api/links/{slug}` | `ADMIN_TOKEN` | `204`, `404`               |
| `GET`    | `/{slug}`           | —             | `307`, `404`, `410`, `429` |
| `GET`    | `/healthz`          | —             | `200`                      |
| `GET`    | `/metrics`          | —             | Prometheus exposition      |

## Configuration

| Variable              | Default                 | Purpose                                    |
| --------------------- | ----------------------- | ------------------------------------------ |
| `DATABASE_URL`        | —                       | Postgres connection string (required)      |
| `REDIS_URL`           | —                       | Redis connection string (required)         |
| `API_TOKEN`           | —                       | Bearer token for creating links (required) |
| `ADMIN_TOKEN`         | —                       | Bearer token for deleting links (required) |
| `ADDR`                | `:8080`                 | Listen address                             |
| `BASE_URL`            | `http://localhost:8080` | Public origin of short links               |
| `CLICK_BUFFER_SIZE`   | `2048`                  | Click queue depth                          |
| `RATE_LIMIT_CREATE`   | `60`                    | Creates per client per minute              |
| `RATE_LIMIT_REDIRECT` | `300`                   | Redirects per client per minute            |

## Architecture

`internal/link` is the domain core: entities, validation, use cases, and the
ports that everything else implements. It imports no transport, storage, or
cache package — dependencies point inward.

```
internal/
├── link/         domain + use cases + ports
├── postgres/     Repository adapter
├── rediscache/   Cache and Limiter adapters
├── memstore/     in-memory Repository, used in tests
├── slug/         Generator adapter
├── httpapi/      HTTP handlers, middleware, metrics
└── config/       environment loading
```

Clicks never block a redirect. The handler pushes an event onto a bounded
channel and returns immediately; a background writer drains it and inserts in
batches via `COPY`. If the queue is full the click is dropped rather than
delaying the user, and the drop is exported as a Prometheus counter.
