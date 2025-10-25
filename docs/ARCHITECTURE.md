# Architecture Documentation

## System Architecture

Auth Service follows Clean Architecture principles with clear separation of concerns.

```
┌─────────────────────────────────────────────────────────────┐
│                         HTTP Layer                          │
│                    (cmd/server/main.go)                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Middleware Layer                          │
│              (internal/middleware/*.go)                      │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────────┐   │
│  │Request  │ Logger  │Recovery │  CORS   │Rate Limiter │   │
│  │   ID    │         │         │         │             │   │
│  └─────────┴─────────┴─────────┴─────────┴─────────────┘   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Handler Layer                            │
│                (internal/handler/*.go)                       │
│           - Request validation                               │
│           - Response formatting (JSend)                      │
│           - HTTP status codes                                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                            │
│                (internal/service/*.go)                       │
│           - Business logic                                   │
│           - Password hashing                                 │
│           - JWT token generation/validation                  │
│           - Orchestration                                    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                           │
│              (internal/repository/*.go)                      │
│           - Database operations                              │
│           - Query building                                   │
│           - Transaction management                           │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Database Layer                            │
│                  PostgreSQL Database                         │
│           - Users table                                      │
│           - Refresh tokens table                             │
└─────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
auth-service/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/                     # Private application code
│   ├── config/                  # Configuration management
│   │   ├── config.go            # Config loading & validation
│   │   └── database.go          # Database connection
│   ├── domain/                  # Domain models
│   │   └── models.go            # User, RefreshToken models
│   ├── handler/                 # HTTP handlers
│   │   ├── auth.go              # Authentication endpoints
│   │   └── jsend.go             # JSend response helpers
│   ├── middleware/              # HTTP middlewares
│   │   └── middleware.go        # Request processing
│   ├── repository/              # Data access layer
│   │   ├── repository.go        # Interfaces
│   │   └── postgres.go          # PostgreSQL implementation
│   └── service/                 # Business logic
│       ├── auth.go              # Auth service
│       └── jwt.go               # JWT service
├── pkg/                         # Public/reusable packages
│   ├── errors/                  # Error handling
│   ├── logger/                  # Structured logging
│   └── validator/               # Input validation
├── migrations/                  # Database migrations
│   ├── 001_create_users_table.up.sql
│   └── ...
├── docs/                        # Documentation
│   ├── API.md                   # API documentation
│   ├── ARCHITECTURE.md          # This file
│   └── postman/                 # Postman collection
└── docker-compose.yml           # Docker setup
```

## Components

### 1. HTTP Layer (cmd/server)
- Application entry point
- Server configuration
- Router setup
- Middleware chain
- Graceful shutdown

### 2. Handler Layer (internal/handler)
- HTTP request/response handling
- Input validation
- JSend response formatting
- Status code management

### 3. Service Layer (internal/service)
- Core business logic
- Password hashing (bcrypt)
- JWT token operations
- User authentication
- Token refresh logic

### 4. Repository Layer (internal/repository)
- Database abstraction
- CRUD operations
- Query execution
- Transaction management

### 5. Domain Layer (internal/domain)
- Entity definitions
- Business models
- Value objects

### 6. Middleware Layer (internal/middleware)
- Request ID generation
- Structured logging
- Panic recovery
- CORS handling
- Rate limiting
- Authentication
- Security headers

### 7. Configuration (internal/config)
- Environment variable loading
- Configuration validation
- Database connection pooling
- Migration management

## Data Flow

### Registration Flow
```
1. POST /api/v1/auth/register
   ↓
2. Handler validates input (email, password, name)
   ↓
3. Service hashes password with bcrypt
   ↓
4. Repository checks if email exists
   ↓
5. Repository creates user in database
   ↓
6. Handler returns user data (no tokens)
```

### Login Flow
```
1. POST /api/v1/auth/login
   ↓
2. Handler validates credentials
   ↓
3. Repository finds user by email
   ↓
4. Service verifies password hash
   ↓
5. JWT Service generates access + refresh tokens
   ↓
6. Repository stores refresh token
   ↓
7. Handler returns tokens + user data
```

### Protected Endpoint Flow
```
1. Request with Authorization header
   ↓
2. Auth middleware extracts token
   ↓
3. JWT Service validates token signature
   ↓
4. JWT Service checks expiration
   ↓
5. Middleware adds claims to context
   ↓
6. Handler processes request
   ↓
7. Service executes business logic
   ↓
8. Repository queries database
   ↓
9. Handler returns response
```

## Security Layers

### 1. Network Security
- HTTPS only in production
- CORS configuration
- Rate limiting per IP

### 2. Authentication Security
- JWT with HMAC-SHA256
- Access token: 15 minutes
- Refresh token: 7 days
- Token rotation on refresh

### 3. Password Security
- bcrypt hashing (cost 12)
- Minimum 8 characters
- Complexity requirements
- No password in logs

### 4. Request Security
- Request ID tracking
- Content-Type validation
- Max body size limit
- Request timeout

### 5. Response Security
- Security headers
- No sensitive data exposure
- Consistent error messages

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Refresh Tokens Table
```sql
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP
);
```

## Scalability Considerations

### Horizontal Scaling
- Stateless design (JWT tokens)
- No session storage
- Database connection pooling
- Load balancer ready

### Vertical Scaling
- Efficient queries with indexes
- Connection pooling
- Rate limiting per instance

### Future Enhancements
- Redis for token blacklist
- Distributed rate limiting
- Database read replicas
- Caching layer (Redis)
- Message queue for async tasks

## Monitoring & Observability

### Logging
- Structured JSON logs
- Request/response logging
- Error tracking with stack traces
- Performance metrics (duration)

### Metrics (Planned)
- Request count
- Response time
- Error rate
- Active users

### Health Checks
- `/health` endpoint
- Database connectivity
- Dependency checks

## Development Workflow

1. **Local Development**
   ```bash
   make dev          # Hot reload with air
   make docker/up    # Start dependencies
   ```

2. **Testing**
   ```bash
   make test         # Unit tests
   make test/cover   # Coverage report
   ```

3. **Quality Checks**
   ```bash
   make lint         # Linting
   make audit        # Security audit
   ```

4. **Build & Deploy**
   ```bash
   make build        # Local build
   make docker/build # Docker image
   ```

## Technology Stack

- **Language**: Go 1.23
- **Web Framework**: net/http (stdlib)
- **Database**: PostgreSQL 15
- **JWT**: golang-jwt/jwt/v5
- **Password**: golang.org/x/crypto/bcrypt
- **Validation**: go-playground/validator/v10
- **Logging**: rs/zerolog
- **Database Driver**: jackc/pgx/v5
- **Container**: Docker & Docker Compose

## Design Principles

1. **Clean Architecture**: Separation of concerns with layers
2. **Dependency Injection**: Loose coupling between components
3. **Interface Segregation**: Small, focused interfaces
4. **Single Responsibility**: Each component has one job
5. **Fail-Fast**: Validate early, error handling everywhere
6. **Explicit over Implicit**: Clear, readable code
7. **Production-Ready**: Comprehensive error handling & logging
