.PHONY: help run build test clean migrate-up migrate-down swagger docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application
	@echo "🚀 Starting Loanee..."
	go run main.go

dev: ## Run with hot reload (requires air)
	@echo "🔥 Starting with hot reload..."
	air

build: ## Build the application
	@echo "🔨 Building..."
	go build -o bin/loanee main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out

lint: ## Run linter
	@echo "🔍 Linting..."
	golangci-lint run

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out

migrate-up: ## Run migrations up
	@echo "Running migrations..."
	migrate -path db/migrations -database "postgresql://postgres:postgres@localhost:5432/loanee?sslmode=disable" up

migrate-down: ## Run migrations down
	@echo "⬇Rolling back migrations..."
	migrate -path db/migrations -database "postgresql://postgres:postgres@localhost:5432/loanee?sslmode=disable" down

migrate-create: ## Create a new migration (usage: make migrate-create name=create_users_table)
	@echo "Creating migration: $(name)"
	migrate create -ext sql -dir db/migrations -seq $(name)

swagger: ## Generate swagger docs
	@echo "Generating Swagger docs..."
	swag init -g main.go -o docs

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	docker-compose up -d

docker-down: ## Stop Docker containers
	@echo "clearStopping Docker containers..."
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

tidy: ## Tidy go modules
	go mod tidy

install-tools: ## Install development tools
	@echo "🔧 Installing tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest