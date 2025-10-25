# ==================================================================================== #
# VARIABLES
# ==================================================================================== #

BINARY_NAME=auth-service
BINARY_WINDOWS=$(BINARY_NAME).exe
MAIN_PATH=./cmd/server/main.go

## help: Display this help message
.PHONY: help
help:
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## dev: Run application in development mode with hot reload (requires air)
.PHONY: dev
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Error: air is not installed. Install it with: go install github.com/air-verse/air@latest"; \
		echo "Or run: make run"; \
	fi

## run: Run application directly
.PHONY: run
run:
	go run $(MAIN_PATH)

## build: Build the application for current platform
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="-s -w" -o $(BINARY_WINDOWS) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_WINDOWS)"

## build-linux: Build the application for Linux
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	set GOOS=linux& set GOARCH=amd64& set CGO_ENABLED=0& go build -ldflags="-s -w" -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_NAME)"

## clean: Remove build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@if exist $(BINARY_WINDOWS) del /F /Q $(BINARY_WINDOWS)
	@if exist $(BINARY_NAME) del /F /Q $(BINARY_NAME)
	go clean
	@echo "Clean complete"

# ==================================================================================== #
# TESTING
# ==================================================================================== #

## test: Run all tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -buildvcs ./...

## test/cover: Run tests with coverage
.PHONY: test/cover
test/cover:
	@echo "Running tests with coverage..."
	go test -v -race -buildvcs -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test/integration: Run integration tests
.PHONY: test/integration
test/integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## lint: Run linter (requires golangci-lint)
.PHONY: lint
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "Error: golangci-lint is not installed"; \
		echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format all Go files
.PHONY: fmt
fmt:
	@echo "Formatting Go files..."
	go fmt ./...
	@echo "Format complete"

## vet: Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete"

## tidy: Tidy and verify module dependencies
.PHONY: tidy
tidy:
	@echo "Tidying module dependencies..."
	go mod tidy
	go mod verify
	@echo "Dependencies tidied"

## audit: Run quality control checks
.PHONY: audit
audit: fmt vet tidy
	@echo "Running security audit..."
	go list -json -m all | go run github.com/sonatype-nexus-community/nancy@latest sleuth
	@echo "Audit complete"

# ==================================================================================== #
# DOCKER
# ==================================================================================== #

## docker/build: Build Docker image
.PHONY: docker/build
docker/build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .
	@echo "Docker image built: $(BINARY_NAME):latest"

## docker/run: Run Docker container
.PHONY: docker/run
docker/run:
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 --env-file .env $(BINARY_NAME):latest

## docker/up: Start all services with docker-compose
.PHONY: docker/up
docker/up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d
	@echo "Services started. View logs with: make docker/logs"

## docker/down: Stop all services
.PHONY: docker/down
docker/down:
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped"

## docker/logs: View docker-compose logs
.PHONY: docker/logs
docker/logs:
	docker-compose logs -f auth-service

## docker/ps: Show running containers
.PHONY: docker/ps
docker/ps:
	docker-compose ps

## docker/restart: Restart services
.PHONY: docker/restart
docker/restart:
	@echo "Restarting services..."
	docker-compose restart
	@echo "Services restarted"

## docker/clean: Remove all containers and volumes
.PHONY: docker/clean
docker/clean:
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	@echo "Docker resources cleaned"

# ==================================================================================== #
# DATABASE
# ==================================================================================== #

## db/up: Start PostgreSQL database
.PHONY: db/up
db/up:
	docker-compose up -d postgres
	@echo "PostgreSQL started"

## db/down: Stop PostgreSQL database
.PHONY: db/down
db/down:
	docker-compose stop postgres
	@echo "PostgreSQL stopped"

## db/psql: Connect to PostgreSQL with psql
.PHONY: db/psql
db/psql:
	docker-compose exec postgres psql -U authuser -d authdb

## db/logs: View database logs
.PHONY: db/logs
db/logs:
	docker-compose logs -f postgres

# ==================================================================================== #
# DEPENDENCIES
# ==================================================================================== #

## deps/install: Install development dependencies
.PHONY: deps/install
deps/install:
	@echo "Installing development dependencies..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Dependencies installed"

## deps/update: Update all dependencies
.PHONY: deps/update
deps/update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy
	@echo "Dependencies updated"

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

## prod/build: Build production binary
.PHONY: prod/build
prod/build:
	@echo "Building production binary..."
	set CGO_ENABLED=0& set GOOS=linux& set GOARCH=amd64& go build -ldflags="-s -w -X main.version=$(shell git describe --tags --always --dirty)" -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Production build complete: $(BINARY_NAME)"

## prod/deploy: Deploy to production (customize as needed)
.PHONY: prod/deploy
prod/deploy:
	@echo "Deployment process not configured"
	@echo "Configure this target for your production environment"

# ==================================================================================== #
# UTILITIES
# ==================================================================================== #

## version: Display Go version
.PHONY: version
version:
	@go version

## info: Display project information
.PHONY: info
info:
	@echo "Project: $(BINARY_NAME)"
	@echo "Main: $(MAIN_PATH)"
	@echo "Go version: $(shell go version)"
	@echo "Binary: $(BINARY_WINDOWS)"

.DEFAULT_GOAL := help
