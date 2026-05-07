# Cross-platform helper targets for local development and container workflows.

DEV_COMPOSE = docker compose -f docker-compose.dev.yml
PROD_COMPOSE = docker compose -f docker-compose.yml
PROD_BOOTSTRAP_COMPOSE = docker compose -f docker-compose.bootstrap.yml

.PHONY: frontend-install frontend-dev frontend-build backend-run dev-up dev-down prod-bootstrap-up prod-bootstrap-down prod-up prod-down migrate-up migrate-down migrate-version mockery-help

# Installs frontend dependencies for local work outside Docker.
frontend-install:
	cd frontend && npm install

# Starts the frontend dev server outside Docker.
frontend-dev:
	cd frontend && npm run dev

# Builds the frontend production bundle.
frontend-build:
	cd frontend && npm run build

# Runs the Go backend directly from the local machine.
backend-run:
	go run ./cmd/bot

# Starts the development docker stack.
dev-up:
	$(DEV_COMPOSE) up --build -d

# Stops the development docker stack.
dev-down:
	$(DEV_COMPOSE) down

# Starts the production docker stack.
prod-up:
	$(PROD_COMPOSE) up --build -d

# Starts the HTTP-only nginx stack for the first certbot challenge.
prod-bootstrap-up:
	$(PROD_BOOTSTRAP_COMPOSE) up --build -d

# Stops the HTTP-only nginx bootstrap stack.
prod-bootstrap-down:
	$(PROD_BOOTSTRAP_COMPOSE) down

# Stops the production docker stack.
prod-down:
	$(PROD_COMPOSE) down

# Applies all pending migrations in the development stack.
migrate-up:
	$(DEV_COMPOSE) --profile tools run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" up

# Rolls back the latest migration in the development stack.
migrate-down:
	$(DEV_COMPOSE) --profile tools run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" down 1

# Shows the current migration version in the development stack.
migrate-version:
	$(DEV_COMPOSE) --profile tools run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" version

# Explains the preferred mock generation tool.
mockery-help:
	@echo Install mockery: go install github.com/vektra/mockery/v2@latest
	@echo Generate mocks example: mockery --dir internal/services --name ProfileRepository --output internal/mocks
