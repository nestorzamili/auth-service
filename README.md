# 🔐 Auth Service API

> Enterprise-grade authentication microservice built with Go, implementing clean architecture principles, JSend response format, and comprehensive security features.

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## ✨ Features

### Core Features
- 🔒 **JWT-based Authentication** - Access & refresh tokens with automatic rotation
- 🔐 **Password Security** - bcrypt hashing with configurable cost factor
- ✅ **Input Validation** - Comprehensive request validation
- 📝 **Structured Logging** - JSON logs with zerolog
- 🗄️ **PostgreSQL** - Database with connection pooling
- 🚀 **Production-Ready** - Fail-fast validation & graceful shutdown
- 🌐 **JSend Standard** - Consistent response format

### Security Features
- 🛡️ **Rate Limiting** - IP-based rate limiting (100-1000 req/min)
- 🔒 **Security Headers** - HSTS, CSP, X-Frame-Options, etc.
- 🚫 **CORS Protection** - Configurable origin whitelist
- ⏱️ **Request Timeout** - Automatic timeout handling
- 📏 **Body Size Limits** - Prevent payload attacks
- 🔍 **Request ID Tracking** - Full request traceability

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| 📖 [API Reference](docs/API.md) | Complete API documentation with examples |
| 🏗️ [Architecture](docs/ARCHITECTURE.md) | System design and architecture |
| 🔐 [Security Guide](docs/SECURITY.md) | Security features and best practices |

---

## 🚀 Quick Start

### Prerequisites

- **Go** 1.23+ 
- **Docker** & **Docker Compose** (recommended)
- **PostgreSQL** 15+ (if not using Docker)

### Option 1: Docker (Recommended)

```bash
# Clone repository
git clone <repository-url>
cd auth-service

# Copy and configure environment
copy .env.example .env
# Edit .env with your configuration

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f auth-service

# Check health
curl http://localhost:8080/health
```

### Option 2: Local Development

```bash
# Install dependencies
go mod download

# Start database
docker-compose up -d postgres

# Configure environment
copy .env.example .env

# Run application
go run cmd/server/main.go

# Or with hot reload (requires Air)
air
```

---

## 🔧 Available Commands

### Using PowerShell Script (Recommended for Windows)

```powershell
# Development
.\build.ps1 dev           # Hot reload with Air
.\build.ps1 run           # Run directly
.\build.ps1 build         # Build binary
.\build.ps1 clean         # Clean artifacts

# Testing
.\build.ps1 test          # Run tests
.\build.ps1 test-cover    # Generate coverage

# Quality
.\build.ps1 fmt           # Format code
.\build.ps1 vet           # Run go vet
.\build.ps1 tidy          # Tidy dependencies
.\build.ps1 lint          # Run linter (if installed)

# Docker
.\build.ps1 docker-up     # Start all services
.\build.ps1 docker-down   # Stop services
.\build.ps1 docker-logs   # View logs

# Utilities
.\build.ps1 install-tools # Install Air & golangci-lint
.\build.ps1 help          # Show all commands
```

### Using Makefile (If make is installed)

```bash
# Same commands but with make prefix
make dev
make test
make docker-up
make help
```

### Manual Commands (Without Scripts)

```powershell
# Development
go run cmd/server/main.go                           # Run
go build -o auth-service.exe cmd/server/main.go     # Build

# Testing
go test -v ./...                                     # Run tests
go test -v -coverprofile=coverage.out ./...         # Coverage

# Docker
docker-compose up -d                                 # Start
docker-compose logs -f auth-service                  # Logs
docker-compose down                                  # Stop

# Quality
go fmt ./...                                         # Format
go vet ./...                                         # Vet
```
---

## 🤝 Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

---

## 📄 License

MIT License - feel free to use this project for learning or production.

---

## 📞 Support

- 📖 **Documentation**: [docs/](docs/)
- 🐛 **Issues**: [GitHub Issues](issues)
- 💬 **Discussions**: [GitHub Discussions](discussions)

---

## ⭐ Show Your Support

If you find this project helpful, please give it a star! ⭐

---

**Made with ❤️ using Go**
