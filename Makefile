# Campus Backend Management System - Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop install-deps

# Default target
help:
	@echo "Campus Backend Management System"
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application locally"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  docker-stop    - Stop Docker containers"
	@echo "  install-deps   - Install Go dependencies"
	@echo "  migrate        - Run database migrations"
	@echo "  lint           - Run linter"
	@echo "  format         - Format Go code"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/campus-backend cmd/server/main.go
	@echo "Build complete: bin/campus-backend"

# Run the application
run:
	@echo "Starting application..."
	go run cmd/server/main.go

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t campus-backend:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d
	@echo "Services started. API available at http://localhost:8080"

# Stop Docker containers
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# Run database migrations
migrate:
	@echo "Running database migrations..."
	go run cmd/server/main.go --migrate

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format Go code
format:
	@echo "Formatting Go code..."
	go fmt ./...
	goimports -w .

# Development setup
dev-setup: install-deps
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "Created .env file from template"; fi
	@echo "Development setup complete!"

# Production build
prod-build:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/campus-backend cmd/server/main.go
	@echo "Production build complete: bin/campus-backend"

# Check for security vulnerabilities
security-check:
	@echo "Checking for security vulnerabilities..."
	gosec ./...

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go; \
		echo "API documentation generated in docs/"; \
	else \
		echo "swag not installed. Run: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# Database operations
db-reset:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres
	sleep 5
	docker-compose up -d campus-backend

# Monitor logs
logs:
	@echo "Monitoring application logs..."
	docker-compose logs -f campus-backend

# Health check
health:
	@echo "Checking application health..."
	@curl -f http://localhost:8080/api/v1/health || echo "Application not responding"
