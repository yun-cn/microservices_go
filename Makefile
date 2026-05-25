# Configuration
APP_NAME := loan-engine
DOCKER_IMAGE := $(APP_NAME):latest
GRPC_PORT := 50051

# Database configuration
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= loan_engine_dev
DB_USER ?= postgres
DB_PASSWORD ?= postgres
GOOSE := CGO_ENABLED=0 go run github.com/pressly/goose/v3/cmd/goose@v3.27.1
MIGRATIONS_DIR := db/migrations

proto: ## Generate Go code from protobuf definitions
	@echo "Generating protobuf code..."
	@./scripts/generate-proto.sh


build: ## Build the Go binary
	@echo "Building $(APP_NAME)..."
	@go build -o server ./cmd/server/main.go
	@echo "Build complete!"

run: ## Run the server locally (requires PostgreSQL)
	@echo "Starting server on port $(GRPC_PORT)..."
	@set -a && . ./.env && set +a && go run ./cmd/server/main.go

db-create: ## Create a new migration file (usage: CGO_ENABLED=0 make db-create NAME=add_user_id)
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make db-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	@$(GOOSE) -dir $(MIGRATIONS_DIR) create $(NAME) sql
	@echo "Migration created in $(MIGRATIONS_DIR)"

db-status: ## Show migration status
	@echo "Checking migration status..."
	@set -a && . ./.env && set +a && $(GOOSE) -dir $(MIGRATIONS_DIR) status

db-up: ## Apply all pending migrations
	@echo "Applying migrations..."
	@set -a && . ./.env && set +a && $(GOOSE) -dir $(MIGRATIONS_DIR) up
	@echo "Migrations applied!"

db-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@set -a && . ./.env && set +a && $(GOOSE) -dir $(MIGRATIONS_DIR) down
	@echo "Rollback complete!"

mod-tidy: ## Tidy Go modules
	@echo "Tidying Go modules..."
	@go mod tidy
	@echo "Modules tidied!"

lint: ## Run golangci-lint
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		CGO_ENABLED=0 go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8; \
	fi
	@echo "Running lint..."
	@golangci-lint run ./...

k8s-deploy: ## Deploy to local Kubernetes cluster
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Applying Kubernetes manifests..."
	@kubectl apply -f k8s/config.yaml
	@kubectl apply -f k8s/postgres.yaml
	@kubectl apply -f k8s/app.yaml
	@echo "Waiting for deployment..."
	@kubectl rollout status deployment/loan-engine

k8s-delete: ## Remove all resources from Kubernetes cluster
	@echo "Deleting Kubernetes resources..."
	@kubectl delete -f k8s/app.yaml --ignore-not-found
	@kubectl delete -f k8s/postgres.yaml --ignore-not-found
	@kubectl delete -f k8s/config.yaml --ignore-not-found
	@kubectl delete pvc postgres-pvc --ignore-not-found
	@echo "Done."

k8s-logs: ## Tail logs from the loan-engine pod
	@kubectl logs -l app=loan-engine,component!=migration -f --tail=50
