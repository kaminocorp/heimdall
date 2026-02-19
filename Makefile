.PHONY: dev-frontend dev-backend dev build-frontend build-backend build test lint sqlc-generate migrate-up migrate-down migrate-create docker-up docker-down

# Frontend
dev-frontend:
	cd frontend && npm run dev

build-frontend:
	cd frontend && npm run build

# Backend
dev-backend:
	cd backend && go run ./cmd/heimdall

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

# Migrations (requires DATABASE_URL env var)
migrate-up:
	cd backend && migrate -database "$$DATABASE_URL" -path migrations up

migrate-down:
	cd backend && migrate -database "$$DATABASE_URL" -path migrations down 1

migrate-create:
	@read -p "Migration name: " name; \
	cd backend && migrate create -ext sql -dir migrations -seq $$name

# Docker
docker-up:
	docker compose up -d

docker-down:
	docker compose down
