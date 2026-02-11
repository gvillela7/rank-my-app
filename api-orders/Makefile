.PHONY: help build run test clean wire docker-build docker-run swagger

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building application..."
	go build -o bin/api ./cmd/main.go
	@echo "Build completed: bin/api"

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -cover ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	@echo "Clean completed"

wire: ## Generate Wire dependency injection code
	@echo "Generating Wire code..."
	wire ./wire
	@echo "Wire generation completed"

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	swag init -g cmd/main.go -o docs
	@echo "Swagger documentation generated in docs/"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	@echo "Dependencies downloaded"

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t api-orders:latest .
	@echo "Docker image built: api-orders:latest"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8000:8000 api-orders:latest

lint: ## Run golangci-lint
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
