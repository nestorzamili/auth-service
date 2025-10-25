# ==================================================================================== #
# AUTH SERVICE - PowerShell Build Script
# Alternative to Makefile for Windows without Make installed
# ==================================================================================== #

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

$BINARY_NAME = "auth-service"
$BINARY_WINDOWS = "$BINARY_NAME.exe"
$MAIN_PATH = ".\cmd\server\main.go"

# Colors for output
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Show-Help {
    Write-Host ""
    Write-ColorOutput Green "Auth Service - PowerShell Build Script"
    Write-Host "========================================"
    Write-Host ""
    Write-Host "Usage: .\build.ps1 <command>"
    Write-Host ""
    Write-ColorOutput Cyan "Available Commands:"
    Write-Host ""
    Write-Host "  Development:"
    Write-Host "    dev              - Run with hot reload (Air)"
    Write-Host "    run              - Run application directly"
    Write-Host "    build            - Build binary"
    Write-Host "    clean            - Remove build artifacts"
    Write-Host ""
    Write-Host "  Testing:"
    Write-Host "    test             - Run all tests"
    Write-Host "    test-cover       - Run tests with coverage"
    Write-Host "    test-integration - Run integration tests"
    Write-Host ""
    Write-Host "  Quality:"
    Write-Host "    fmt              - Format code"
    Write-Host "    vet              - Run go vet"
    Write-Host "    tidy             - Tidy dependencies"
    Write-Host "    lint             - Run linter (if installed)"
    Write-Host ""
    Write-Host "  Docker:"
    Write-Host "    docker-build     - Build Docker image"
    Write-Host "    docker-up        - Start services"
    Write-Host "    docker-down      - Stop services"
    Write-Host "    docker-logs      - View logs"
    Write-Host "    docker-ps        - Show containers"
    Write-Host "    docker-clean     - Clean Docker resources"
    Write-Host ""
    Write-Host "  Database:"
    Write-Host "    db-up            - Start PostgreSQL"
    Write-Host "    db-down          - Stop PostgreSQL"
    Write-Host "    db-logs          - View database logs"
    Write-Host ""
    Write-Host "  Utilities:"
    Write-Host "    install-tools    - Install dev tools (Air, linter)"
    Write-Host "    version          - Show Go version"
    Write-Host "    help             - Show this help"
    Write-Host ""
}

# Development Commands
function Start-Dev {
    Write-ColorOutput Green "Starting development server with hot reload..."
    if (Get-Command air -ErrorAction SilentlyContinue) {
        air
    } else {
        Write-ColorOutput Red "Air not installed. Install with: .\build.ps1 install-tools"
        Write-ColorOutput Yellow "Or run without hot reload: .\build.ps1 run"
    }
}

function Start-Run {
    Write-ColorOutput Green "Running application..."
    go run $MAIN_PATH
}

function Start-Build {
    Write-ColorOutput Green "Building $BINARY_WINDOWS..."
    go build -ldflags="-s -w" -o $BINARY_WINDOWS $MAIN_PATH
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput Green "✓ Build complete: $BINARY_WINDOWS"
    } else {
        Write-ColorOutput Red "✗ Build failed"
    }
}

function Start-Clean {
    Write-ColorOutput Green "Cleaning build artifacts..."
    if (Test-Path $BINARY_WINDOWS) {
        Remove-Item $BINARY_WINDOWS -Force
        Write-ColorOutput Green "✓ Removed $BINARY_WINDOWS"
    }
    if (Test-Path "tmp") {
        Remove-Item "tmp" -Recurse -Force
        Write-ColorOutput Green "✓ Removed tmp/"
    }
    go clean
    Write-ColorOutput Green "✓ Clean complete"
}

# Testing Commands
function Start-Test {
    Write-ColorOutput Green "Running tests..."
    go test -v -race ./...
}

function Start-TestCover {
    Write-ColorOutput Green "Running tests with coverage..."
    go test -v -race -coverprofile=coverage.out ./...
    if ($LASTEXITCODE -eq 0) {
        go tool cover -html=coverage.out -o coverage.html
        Write-ColorOutput Green "✓ Coverage report: coverage.html"
    }
}

function Start-TestIntegration {
    Write-ColorOutput Green "Running integration tests..."
    go test -v -tags=integration ./...
}

# Quality Commands
function Start-Fmt {
    Write-ColorOutput Green "Formatting code..."
    go fmt ./...
    Write-ColorOutput Green "✓ Format complete"
}

function Start-Vet {
    Write-ColorOutput Green "Running go vet..."
    go vet ./...
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput Green "✓ Vet complete"
    }
}

function Start-Tidy {
    Write-ColorOutput Green "Tidying dependencies..."
    go mod tidy
    go mod verify
    Write-ColorOutput Green "✓ Dependencies tidied"
}

function Start-Lint {
    Write-ColorOutput Green "Running linter..."
    if (Get-Command golangci-lint -ErrorAction SilentlyContinue) {
        golangci-lint run
    } else {
        Write-ColorOutput Yellow "golangci-lint not installed"
        Write-ColorOutput Yellow "Install with: .\build.ps1 install-tools"
    }
}

# Docker Commands
function Start-DockerBuild {
    Write-ColorOutput Green "Building Docker image..."
    docker build -t "${BINARY_NAME}:latest" .
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput Green "✓ Docker image built: ${BINARY_NAME}:latest"
    }
}

function Start-DockerUp {
    Write-ColorOutput Green "Starting services with docker-compose..."
    docker-compose up -d
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput Green "✓ Services started"
        Write-ColorOutput Cyan "View logs: .\build.ps1 docker-logs"
    }
}

function Start-DockerDown {
    Write-ColorOutput Green "Stopping services..."
    docker-compose down
    Write-ColorOutput Green "✓ Services stopped"
}

function Start-DockerLogs {
    Write-ColorOutput Green "Viewing logs (Ctrl+C to exit)..."
    docker-compose logs -f auth-service
}

function Start-DockerPs {
    docker-compose ps
}

function Start-DockerClean {
    Write-ColorOutput Green "Cleaning Docker resources..."
    docker-compose down -v
    Write-ColorOutput Green "✓ Docker resources cleaned"
}

# Database Commands
function Start-DbUp {
    Write-ColorOutput Green "Starting PostgreSQL..."
    docker-compose up -d postgres
    Write-ColorOutput Green "✓ PostgreSQL started"
}

function Start-DbDown {
    Write-ColorOutput Green "Stopping PostgreSQL..."
    docker-compose stop postgres
    Write-ColorOutput Green "✓ PostgreSQL stopped"
}

function Start-DbLogs {
    Write-ColorOutput Green "Viewing database logs (Ctrl+C to exit)..."
    docker-compose logs -f postgres
}

# Utility Commands
function Install-Tools {
    Write-ColorOutput Green "Installing development tools..."
    
    Write-Host "Installing Air (hot reload)..."
    go install github.com/air-verse/air@latest
    
    Write-Host ""
    Write-Host "Installing golangci-lint (linter)..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    
    Write-ColorOutput Green "✓ Tools installed"
    Write-ColorOutput Cyan "Make sure %GOPATH%\bin is in your PATH"
}

function Show-Version {
    go version
}

# Main switch
switch ($Command.ToLower()) {
    "dev" { Start-Dev }
    "run" { Start-Run }
    "build" { Start-Build }
    "clean" { Start-Clean }
    
    "test" { Start-Test }
    "test-cover" { Start-TestCover }
    "test-integration" { Start-TestIntegration }
    
    "fmt" { Start-Fmt }
    "vet" { Start-Vet }
    "tidy" { Start-Tidy }
    "lint" { Start-Lint }
    
    "docker-build" { Start-DockerBuild }
    "docker-up" { Start-DockerUp }
    "docker-down" { Start-DockerDown }
    "docker-logs" { Start-DockerLogs }
    "docker-ps" { Start-DockerPs }
    "docker-clean" { Start-DockerClean }
    
    "db-up" { Start-DbUp }
    "db-down" { Start-DbDown }
    "db-logs" { Start-DbLogs }
    
    "install-tools" { Install-Tools }
    "version" { Show-Version }
    "help" { Show-Help }
    
    default {
        Write-ColorOutput Red "Unknown command: $Command"
        Write-Host ""
        Show-Help
        exit 1
    }
}
