.PHONY: run build test check dlint up down db logs psql dev deploy

ENV := set -a; . ./.env; set +a;
VERSION := $(shell git describe --tags --always --dirty)
COMPOSE_PROD := docker compose --env-file .env -f deploy/compose.yaml
COMPOSE := $(COMPOSE_PROD) -f deploy/compose.override.yaml

run:
	$(ENV) go run ./cmd/shortener

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/shortener ./cmd/shortener

test:
	go test ./...

check:
	gofmt -l .
	go vet ./...
	go test ./...

dlint:
	hadolint Dockerfile

up:
	VERSION=$(VERSION) $(COMPOSE) up -d --build

down:
	$(COMPOSE) down

db:
	$(COMPOSE) up -d --wait postgres redis

logs:
	$(COMPOSE) logs -f

psql:
	$(COMPOSE) exec postgres psql -U shortener -d shortener

dev: check db run

deploy:
	VERSION=$(VERSION) $(COMPOSE_PROD) up -d --build
