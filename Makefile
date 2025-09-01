.PHONY: help dev build test clean migrate-up migrate-down docker-up docker-down templ-generate css-build

# Variables
APP_NAME := should-i-get-it
BINARY_NAME := $(APP_NAME)
MAIN_PATH := ./cmd/server
DC          ?= docker compose
PGSERVICE   ?= postgres          # service name from docker-compose.yml
PGUSER      ?= postgres
PGDB        ?= should_i_get_it
PGPORT      ?= 5432
AIR_CONFIG := $(CURDIR)/.air.toml

# Default target
help:
	@echo "Available commands:"
	@echo "  dev          - Start development server with hot reload"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Run database migrations down"
	@echo "  docker-up    - Start docker services"
	@echo "  docker-down  - Stop docker services"
	@echo "  templ-gen    - Generate templ templates"
	@echo "  css-build    - Build CSS with Tailwind"
	@echo "  setup        - Initial project setup"

# Development server with hot reload
dev:
	@echo "Starting development server..."
	@test -f "$(AIR_CONFIG)" || (echo "Missing $(AIR_CONFIG)"; exit 1)
	@make templ-gen
	@make css-build &
	@air -c "$(AIR_CONFIG)"

# Build the application
build:
	@echo "Building application..."
	@make templ-gen
	@npm run build-css-prod
	@go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -rf node_modules/

# Generate templ templates
templ-gen:
	@echo "Generating templ templates..."
	@templ generate

# Build CSS
css-build:
	@echo "Building CSS..."
	@npm run build-css

# Database migrations
migrate-up:
	@echo "Running migrations (inside container) ..."
	@$(DC) exec -T $(PGSERVICE) sh -lc 'set -eu; \
	  for f in /docker-entrypoint-initdb.d/*.sql; do \
	    echo ">> applying $$f"; \
	    psql -v ON_ERROR_STOP=1 -U $(PGUSER) -d $(PGDB) -h localhost -p $(PGPORT) -f "$$f"; \
	  done'


migrate-file:
	@test -n "$(FILE)" || (echo "Usage: make migrate-file FILE=xxxx.sql"; exit 1)
	@echo "Applying $(FILE) ..."
	@$(DC) exec -T $(PGSERVICE) sh -lc 'psql -v ON_ERROR_STOP=1 -U $(PGUSER) -d $(PGDB) -h localhost -p $(PGPORT) -f "/docker-entrypoint-initdb.d/$(FILE)"'

# Drop & recreate public schema (dangerous; nukes data)
migrate-down:
	@echo "Dropping and recreating public schema (inside container) ..."
	@$(DC) exec -T $(PGSERVICE) sh -lc 'psql -U $(PGUSER) -d $(PGDB) -h localhost -p $(PGPORT) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"'

# Handy interactive shell to your DB
psql:
	@$(DC) exec -it $(PGSERVICE) psql -U $(PGUSER) -d $(PGDB)

# Docker services
docker-up:
	@echo "Starting docker services..."
	@$(DC) up -d

docker-down:
	@echo "Stopping docker services..."
	@$(DC) down


# Initial setup
setup:
	@echo "Setting up project..."
	@go mod tidy
	@npm install
	@make templ-gen
	@make css-build
	@echo "Setup complete!"