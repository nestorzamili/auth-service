# Auth Service API Documentation

## Overview

Enterprise-grade authentication microservice built with Go. This service provides JWT-based authentication with access and refresh tokens, user registration, login, and token management.

**Base URL**: `http://localhost:8080`

**Version**: 1.0.0

**Content-Type**: `application/json`

---

## Table of Contents

1. [Authentication Flow](#authentication-flow)
2. [API Endpoints](#api-endpoints)
3. [Request & Response Format](#request--response-format)
4. [Error Handling](#error-handling)
5. [Rate Limiting](#rate-limiting)
6. [Security](#security)

---

## Authentication Flow

```
1. User Registration
   POST /api/v1/auth/register
   ↓
   Returns: User data (no tokens)

2. User Login
   POST /api/v1/auth/login
   ↓
   Returns: Access Token + Refresh Token

3. Access Protected Resources
   Add Header: Authorization: Bearer <access_token>
   ↓
   GET /api/v1/auth/me

4. Refresh Access Token (when expired)
   POST /api/v1/auth/refresh
   Body: { "refresh_token": "..." }
   ↓
   Returns: New Access Token + New Refresh Token

5. Logout
   POST /api/v1/auth/logout
   Header: Authorization: Bearer <access_token>
```

---

## API Endpoints

### 1. Health Check

Check if the service is running.

**Endpoint**: `GET /health`

**Authentication**: Not required

**Response**:
```json
{
  "status": "success",
  "data": {
    "status": "healthy",
    "timestamp": "2025-10-25T10:30:00Z"
  }
}
```

---

### 2. Register User

Create a new user account.

**Endpoint**: `POST /api/v1/auth/register`

**Authentication**: Not required

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "full_name": "John Doe"
}
```

**Validation Rules**:
- `email`: Valid email format, unique
- `password`: Minimum 8 characters, must contain uppercase, lowercase, number, and special character
- `full_name`: 2-100 characters

**Success Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "full_name": "John Doe",
    "created_at": "2025-10-25T10:30:00Z"
  }
}
```

**Error Response** (400 Bad Request):
```json
{
  "status": "fail",
  "data": {
    "email": "Email already exists",
    "password": "Password must be at least 8 characters"
  }
}
```

---

### 3. Login

Authenticate user and receive tokens.

**Endpoint**: `POST /api/v1/auth/login`

**Authentication**: Not required

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Success Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "full_name": "John Doe"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900
  }
}
```

**Error Response** (401 Unauthorized):
```json
{
  "status": "fail",
  "data": {
    "message": "Invalid email or password"
  }
}
```

---

### 4. Refresh Token

Get new access token using refresh token.

**Endpoint**: `POST /api/v1/auth/refresh`

**Authentication**: Not required (uses refresh token in body)

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900
  }
}
```

**Error Response** (401 Unauthorized):
```json
{
  "status": "error",
  "message": "Invalid or expired refresh token"
}
```

---

### 5. Validate Token

Validate an access token (useful for other services).

**Endpoint**: `POST /api/v1/auth/validate`

**Authentication**: Not required (token in body)

**Request Body**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "valid": true,
    "user_id": 1,
    "email": "user@example.com",
    "expires_at": "2025-10-25T11:30:00Z"
  }
}
```

---

### 6. Get Current User (Protected)

Get currently authenticated user information.

**Endpoint**: `GET /api/v1/auth/me`

**Authentication**: Required (Bearer token)

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "full_name": "John Doe",
    "created_at": "2025-10-25T10:30:00Z"
  }
}
```

**Error Response** (401 Unauthorized):
```json
{
  "status": "error",
  "message": "Invalid or expired token"
}
```

---

### 7. Logout (Protected)

Invalidate refresh token and logout user.

**Endpoint**: `POST /api/v1/auth/logout`

**Authentication**: Required (Bearer token)

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "message": "Logged out successfully"
  }
}
```

---

## Request & Response Format

### JSend Standard

All responses follow the [JSend](https://github.com/omniti-labs/jsend) specification:

**Success Response**:
```json
{
  "status": "success",
  "data": { }
}
```

**Fail Response** (Client Error):
```json
{
  "status": "fail",
  "data": { }
}
```

**Error Response** (Server Error):
```json
{
  "status": "error",
  "message": "Error message",
  "code": "ERROR_CODE"
}
```

---

## Error Handling

### HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request (validation error) |
| 401 | Unauthorized (authentication failed) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not Found |
| 409 | Conflict (duplicate resource) |
| 429 | Too Many Requests (rate limit) |
| 500 | Internal Server Error |

### Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Input validation failed |
| `AUTH_INVALID_CREDENTIALS` | Invalid email or password |
| `AUTH_TOKEN_INVALID` | Invalid or malformed token |
| `AUTH_TOKEN_EXPIRED` | Token has expired |
| `AUTH_TOKEN_MISSING` | Authorization header missing |
| `USER_NOT_FOUND` | User does not exist |
| `USER_ALREADY_EXISTS` | Email already registered |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `INTERNAL_ERROR` | Server error |

---

## Rate Limiting

The API implements rate limiting to prevent abuse:

**Development**: 1000 requests per minute per IP
**Production**: 100 requests per minute per IP

### Rate Limit Headers

Every response includes rate limit information:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1635174000
```

When rate limit is exceeded, you'll receive a `429 Too Many Requests` response:

```json
{
  "status": "error",
  "message": "Rate limit exceeded. Please try again later.",
  "code": "RATE_LIMIT_EXCEEDED"
}
```

---

## Security

### Security Features

1. **HTTPS Only** (Production)
2. **CORS Configuration**
3. **Security Headers**:
   - `X-Content-Type-Options: nosniff`
   - `X-Frame-Options: DENY`
   - `X-XSS-Protection: 1; mode=block`
   - `Strict-Transport-Security: max-age=31536000`
4. **Password Hashing**: bcrypt with cost factor 12
5. **JWT Tokens**:
   - Access Token: 15 minutes expiry
   - Refresh Token: 7 days expiry
6. **Rate Limiting**
7. **Request ID Tracking**
8. **Structured Logging**

### Best Practices

1. **Always use HTTPS in production**
2. **Store tokens securely** (httpOnly cookies or secure storage)
3. **Implement token refresh** before access token expires
4. **Logout on client side** by removing tokens
5. **Never log sensitive data** (passwords, tokens)
6. **Use strong passwords**
7. **Rotate secrets regularly**

---

## Example Requests

### cURL Examples

**Register**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "full_name": "John Doe"
  }'
```

**Login**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

**Get Current User**:
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Refresh Token**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

**Logout**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

---

## Testing

Run the included Postman collection for easy API testing:
- Import `docs/postman/Auth-Service.postman_collection.json`
- Set environment variable `base_url` to `http://localhost:8080`

---

## Support

For issues and questions:
- GitHub Issues: [Your Repository]
- Email: support@example.com

---

## License

MIT License - See LICENSE file for details
