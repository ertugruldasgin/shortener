.PHONY: run build test check migrate up down

VERSION := $(shell git describe --tags --always --dirty)
ENV := set -a; . ./.env; set +a;

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

migrate:
	$(ENV) goose -dir migrations postgres "$$DATABASE_URL" up

dlint:
	hadolint Dockerfile

up:
	docker compose -f deploy/compose.yaml up -d

down:
	docker compose -f deploy/compose.yaml down
