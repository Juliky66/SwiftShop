-include .env
export

DB_DSN ?= postgres://shopkuber:shopkuber@localhost:5432/shopkuber?sslmode=disable
MIGRATE := migrate -path ./migrations -database "$(DB_DSN)"

.PHONY: migrate-up migrate-down migrate-create seed keys \
        build-auth build-catalog build-orders build-seller build-reviews \
        run-auth run-catalog run-orders run-seller run-reviews \
        start stop test lint up down

## ── Migrations ──────────────────────────────────────────────────────────────

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

seed:
	psql "$(DB_DSN)" -f migrations/seed.sql

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir ./migrations -seq $$name

## ── RSA Key generation ───────────────────────────────────────────────────────

keys:
	mkdir -p keys
	openssl genrsa -out keys/private.pem 4096
	openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@echo "Keys generated in ./keys/ (DO NOT COMMIT private.pem)"

## ── Build ────────────────────────────────────────────────────────────────────

build-auth:
	go build -o bin/auth ./services/auth/cmd

build-catalog:
	go build -o bin/catalog ./services/catalog/cmd

build-orders:
	go build -o bin/orders ./services/orders/cmd

build-seller:
	go build -o bin/seller ./services/seller/cmd

build-reviews:
	go build -o bin/reviews ./services/reviews/cmd

build-all: build-auth build-catalog build-orders build-seller build-reviews

## ── Run (local, requires .env sourced) ──────────────────────────────────────

run-auth:
	go run ./services/auth/cmd

run-catalog:
	go run ./services/catalog/cmd

run-orders:
	go run ./services/orders/cmd

run-seller:
	go run ./services/seller/cmd

run-reviews:
	go run ./services/reviews/cmd

## ── Start / Stop everything ──────────────────────────────────────────────────

start: ## Generate keys if missing, then start backend
	@if [ ! -f keys/private.pem ]; then \
		echo "==> Generating RSA keys..."; \
		mkdir -p keys && sudo chown -R $$USER:$$USER keys/ 2>/dev/null || true; \
		openssl genrsa -out keys/private.pem 4096; \
		openssl rsa -in keys/private.pem -pubout -out keys/public.pem; \
		echo "Keys generated."; \
	fi
	@echo "==> Starting backend (docker-compose)..."
	docker compose -f deploy/docker-compose.dev.yaml up -d --build

stop:
	docker compose -f deploy/docker-compose.dev.yaml down

## ── Docker Compose ───────────────────────────────────────────────────────────

up:
	docker compose -f deploy/docker-compose.dev.yaml up -d

down:
	docker compose -f deploy/docker-compose.dev.yaml down

## ── Test ─────────────────────────────────────────────────────────────────────

test:
	go test ./...

## ── Lint ─────────────────────────────────────────────────────────────────────

lint:
	golangci-lint run ./...

# ── Mobile ────────────────────────────────────────────────────────────────────

mobile-install: ## Install mobile app dependencies
	cd mobile && npm install

mobile-web: ## Start Expo web server (http://localhost:8086, port 8081 is used by auth service)
	cd mobile && NO_PROXY="*" HTTP_PROXY="" HTTPS_PROXY="" http_proxy="" https_proxy="" npx expo start --web --port 8086

# ── Playwright Recording ──────────────────────────────────────────────────────

playwright-install: ## Install Playwright and Chromium
	cd playwright && npm install && npx playwright install chromium

record: ## Record buyer flow (requires: make up + make mobile-web in separate terminals)
	cd playwright && npx playwright test

record-headed: ## Record buyer flow with visible browser
	cd playwright && npx playwright test --headed

playwright-report: ## Open last recording report
	cd playwright && npx playwright show-report
