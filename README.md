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

```bash
# Clone repository
git clone <repository-url>
cd auth-service

# Install dependencies
go mod download

# Run the application
go run cmd/server/main.go
```

---

## 🔧 Development Commands

### Running the Application

```bash
# Development mode (direct run)
go run cmd/server/main.go

# With hot reload (requires Air)
air

# Build binary
go build -o auth-service.exe cmd/server/main.go  # Windows
go build -o auth-service cmd/server/main.go      # Linux/Mac

# Run binary
./auth-service.exe  # Windows
./auth-service      # Linux/Mac
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run static analysis
go vet ./...

# Run linter (if golangci-lint installed)
golangci-lint run

# Tidy dependencies
go mod tidy
go mod verify
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

### Development Tools

```bash
# Install Air (hot reload)
go install github.com/air-verse/air@latest

# Install golangci-lint (linter)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
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
