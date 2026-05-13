.PHONY: dev-frontend dev-backend dev build-frontend build-backend build test lint sqlc-generate migrate-up migrate-down migrate-create bootstrap-roles docker-up docker-down

# Frontend
dev-frontend:
	cd frontend && npm run dev

build-frontend:
	cd frontend && npm run build

# Backend
dev-backend:
	set -a && . ./.env && set +a && cd backend && go run ./cmd/heimdall

build-backend:
	cd backend && go build -o bin/heimdall ./cmd/heimdall

# Both
dev:
	$(MAKE) -j2 dev-frontend dev-backend

build:
	$(MAKE) build-frontend build-backend

# Testing
test:
	cd backend && go test ./...
	cd frontend && npm run test

# Linting
lint:
	cd backend && go vet ./...
	cd frontend && npm run lint

# sqlc
sqlc-generate:
	cd backend && sqlc generate

# Migrations (requires DIRECT_URL — falls back to DATABASE_URL when unset).
# Phase 6 of the RLS role-split downgrades DATABASE_URL to a non-superuser;
# DIRECT_URL stays on the superuser so DDL keeps working. Until Phase 6,
# DIRECT_URL is empty and migrations transparently use DATABASE_URL.
migrate-up:
	cd backend && migrate -database "$${DIRECT_URL:-$$DATABASE_URL}" -path migrations up

migrate-down:
	cd backend && migrate -database "$${DIRECT_URL:-$$DATABASE_URL}" -path migrations down 1

migrate-create:
	@read -p "Migration name: " name; \
	cd backend && migrate create -ext sql -dir migrations -seq $$name

# Bootstrap the runtime roles (Phase 4 of the RLS role split).
#
# Migration 039 expects app_user and cron_user to already exist; it raises an
# EXCEPTION otherwise. Roles are NOT created in migrations to keep passwords
# out of the repo — operators run this target once per fresh DB, supplying
# passwords via env vars. Skip this target in dev where the local Postgres
# user is a superuser and can run all three URL paths directly.
#
# Required env vars: APP_USER_PASSWORD, CRON_USER_PASSWORD.
# Connection: uses DIRECT_URL (superuser) if set, else DATABASE_URL.
bootstrap-roles:
	@if [ -z "$$APP_USER_PASSWORD" ] || [ -z "$$CRON_USER_PASSWORD" ]; then \
		echo "error: APP_USER_PASSWORD and CRON_USER_PASSWORD must both be set."; \
		echo "  generate with: openssl rand -base64 36"; \
		echo "  then store in your secrets manager and re-run: APP_USER_PASSWORD=... CRON_USER_PASSWORD=... make bootstrap-roles"; \
		exit 1; \
	fi
	@psql "$${DIRECT_URL:-$$DATABASE_URL}" -v ON_ERROR_STOP=1 \
		-v app_pw="$$APP_USER_PASSWORD" -v cron_pw="$$CRON_USER_PASSWORD" \
		-c "CREATE ROLE app_user LOGIN PASSWORD :'app_pw';" \
		-c "CREATE ROLE cron_user LOGIN PASSWORD :'cron_pw' BYPASSRLS;" \
		-c "SELECT rolname, rolsuper, rolbypassrls FROM pg_roles WHERE rolname IN ('app_user', 'cron_user') ORDER BY rolname;"
	@echo "bootstrap complete. Verify the table above shows app_user(super=f, bypass=f) and cron_user(super=f, bypass=t), then run: make migrate-up"

# Docker
docker-up:
	docker compose up -d

docker-down:
	docker compose down
