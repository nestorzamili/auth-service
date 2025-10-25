# Security Guide

## Overview

This document outlines the security features, best practices, and recommendations for deploying and maintaining the Auth Service.

---

## Security Features

### 1. Authentication & Authorization

#### JWT Tokens
- **Algorithm**: HMAC-SHA256 (HS256)
- **Access Token Expiry**: 15 minutes (configurable)
- **Refresh Token Expiry**: 7 days (configurable)
- **Token Rotation**: New refresh token on every refresh
- **Signature Verification**: All tokens verified before use

#### Password Security
- **Hashing Algorithm**: bcrypt with cost factor 12
- **Password Requirements**:
  - Minimum 8 characters
  - At least one uppercase letter
  - At least one lowercase letter
  - At least one number
  - At least one special character
- **No Password Storage**: Only hashed passwords stored
- **Timing-Safe Comparison**: Prevents timing attacks

### 2. API Security

#### Rate Limiting
- **Development**: 1000 requests/minute per IP
- **Production**: 100 requests/minute per IP
- **Headers**: `X-RateLimit-*` headers in response
- **Cleanup**: Automatic cleanup of old entries

#### CORS (Cross-Origin Resource Sharing)
- **Development**: Allows all origins (`*`)
- **Production**: Whitelist specific domains
- **Credentials**: Support for cookies and auth headers
- **Methods**: GET, POST, PUT, DELETE, OPTIONS
- **Headers**: Content-Type, Authorization, X-Request-ID

#### Security Headers
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
```

### 3. Input Validation

#### Request Validation
- **Content-Type**: Must be `application/json` for POST/PUT/PATCH
- **Max Body Size**: 1MB limit (configurable)
- **Field Validation**: Using go-playground/validator
- **SQL Injection**: Prevented by parameterized queries

#### Email Validation
- RFC 5322 compliant
- DNS validation (optional)
- Lowercase normalization

### 4. Database Security

#### Connection Security
- **SSL Mode**: Required in production
- **Credentials**: Environment variables only
- **Connection Pool**: Limited connections (25 max)
- **Prepared Statements**: All queries parameterized

#### Data Protection
- **Password Hashing**: bcrypt before storage
- **Refresh Token Hashing**: SHA-256 (recommended)
- **No Sensitive Logs**: Passwords never logged
- **SQL Injection**: Prevented by pgx driver

### 5. Network Security

#### HTTPS
- **Production**: HTTPS only (enforced by HSTS)
- **TLS Version**: TLS 1.2 minimum recommended
- **Certificate**: Valid SSL certificate required

#### Request Tracking
- **Request ID**: UUID for each request
- **X-Request-ID**: Header for tracing
- **Structured Logging**: All requests logged

---

## Environment Configuration

### Required Environment Variables

```bash
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
SERVER_SHUTDOWN_TIMEOUT=5s
ENVIRONMENT=production

# JWT Configuration - MUST BE UNIQUE AND SECRET
JWT_ACCESS_SECRET=your-super-secret-access-key-min-32-chars
JWT_REFRESH_SECRET=your-super-secret-refresh-key-min-32-chars
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
JWT_ISSUER=auth-service
JWT_ALGORITHM=HS256

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=authuser
DB_PASSWORD=secure-database-password
DB_NAME=authdb
DB_SSL_MODE=require  # MUST be 'require' or higher in production

# Logger Configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

### Secret Generation

Generate secure secrets using:

```bash
# On Linux/Mac
openssl rand -base64 48

# On Windows (PowerShell)
[Convert]::ToBase64String((1..48 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))

# Using Go
go run -c "package main; import (\"crypto/rand\"; \"encoding/base64\"; \"fmt\"); func main() { b := make([]byte, 48); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }"
```

---

## Production Checklist

### Before Deployment

- [ ] Change all default passwords and secrets
- [ ] Set `ENVIRONMENT=production`
- [ ] Enable SSL for database (`DB_SSL_MODE=require`)
- [ ] Configure CORS with specific domains (not `*`)
- [ ] Set appropriate rate limits
- [ ] Use HTTPS with valid SSL certificate
- [ ] Set secure JWT secrets (min 32 characters)
- [ ] Review and update allowed origins
- [ ] Set up monitoring and alerting
- [ ] Configure log aggregation
- [ ] Test all endpoints with production config
- [ ] Review database indexes
- [ ] Set up automated backups
- [ ] Configure firewall rules
- [ ] Implement reverse proxy (nginx/traefik)

### Configuration Validation

The application validates all configuration on startup:

```
✓ JWT secrets at least 32 characters
✓ JWT secrets are different
✓ Database credentials provided
✓ SSL enabled in production
✓ Valid environment (development/staging/production)
✓ Access token expiry < refresh token expiry
✓ All required variables set
```

---

## Security Best Practices

### 1. Token Management

#### Storage
- **Never** store tokens in localStorage (XSS vulnerable)
- **Recommended**: httpOnly cookies with secure flag
- **Alternative**: Memory only (lost on refresh)
- **Mobile**: Secure storage (Keychain/KeyStore)

#### Usage
```javascript
// Good: httpOnly cookie
Set-Cookie: access_token=...; HttpOnly; Secure; SameSite=Strict

// Bad: localStorage
localStorage.setItem('token', '...'); // DON'T DO THIS
```

#### Refresh Strategy
- Refresh access token before expiry
- Implement automatic refresh in client
- Clear tokens on logout
- Invalidate refresh token on logout

### 2. Password Policy

Enforce strong passwords:
- Minimum 8 characters (12+ recommended)
- Mix of uppercase, lowercase, numbers, symbols
- No common passwords (check against dictionary)
- Password history (prevent reuse)
- Consider password strength meter

### 3. API Security

#### Authentication
```bash
# Always use Bearer token format
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Never send credentials in URL
# BAD: /api/login?password=secret
# GOOD: POST /api/login with JSON body
```

#### Request Validation
- Validate all inputs
- Sanitize user data
- Use Content-Type validation
- Implement request size limits
- Set request timeouts

### 4. Database Security

#### Connection
- Use connection pooling
- Enable SSL/TLS
- Rotate credentials regularly
- Use read-only users where possible
- Implement query timeouts

#### Data Protection
- Hash passwords with bcrypt
- Don't store sensitive data in plain text
- Use database encryption at rest
- Regular backups
- Test restore procedures

### 5. Logging & Monitoring

#### What to Log
- All authentication attempts
- Failed login attempts
- Token generation/validation
- Rate limit violations
- Database connection issues
- Configuration errors

#### What NOT to Log
- Passwords (plain or hashed)
- Access tokens
- Refresh tokens
- Personal identifiable information (mask if needed)
- Full credit card numbers

#### Log Example
```json
{
  "level": "info",
  "timestamp": "2025-10-25T10:30:00Z",
  "request_id": "abc-123",
  "method": "POST",
  "path": "/api/v1/auth/login",
  "status": 200,
  "duration_ms": 45,
  "email": "u***@example.com"  // Masked
}
```

---

## Attack Prevention

### 1. Brute Force Attacks

**Prevention**:
- Rate limiting (100 requests/minute)
- Account lockout after N failed attempts
- Progressive delays
- CAPTCHA after failures
- Monitor failed login patterns

### 2. SQL Injection

**Prevention**:
- Parameterized queries (pgx driver)
- Input validation
- Prepared statements
- ORM/Query builder
- Regular security audits

### 3. XSS (Cross-Site Scripting)

**Prevention**:
- JSON API only (no HTML rendering)
- Content-Type validation
- X-Content-Type-Options: nosniff
- Content-Security-Policy header
- Input sanitization

### 4. CSRF (Cross-Site Request Forgery)

**Prevention**:
- CORS configuration
- SameSite cookies
- Token-based auth (not cookies for auth)
- Validate Origin/Referer headers

### 5. JWT Attacks

**Prevention**:
- Strong secret keys (32+ characters)
- Token expiration (15 minutes)
- Token rotation
- Signature verification
- Algorithm whitelist (HS256 only)
- No sensitive data in payload

### 6. DDoS Attacks

**Prevention**:
- Rate limiting per IP
- Connection timeouts
- Max body size limits
- Use reverse proxy/CDN
- Geographic restrictions if applicable

---

## Incident Response

### If Credentials Compromised

1. **Immediate Actions**:
   - Rotate JWT secrets immediately
   - Invalidate all active tokens
   - Force all users to re-login
   - Review access logs
   - Notify affected users

2. **Investigation**:
   - Check for unauthorized access
   - Review recent changes
   - Analyze logs for patterns
   - Document timeline

3. **Recovery**:
   - Deploy new secrets
   - Update documentation
   - Implement additional monitoring
   - Post-incident review

### Security Monitoring

**Alerts to Set Up**:
- High rate of failed logins
- Unusual access patterns
- Database connection errors
- Rate limit violations
- Certificate expiry warnings
- Disk space alerts

---

## Compliance

### GDPR Considerations
- User data encryption
- Right to deletion
- Data export capability
- Privacy policy
- Consent management

### Best Practices
- Regular security audits
- Dependency updates
- Penetration testing
- Security training
- Incident response plan

---

## Resources

### Tools
- **OWASP ZAP**: Security testing
- **SQLMap**: SQL injection testing
- **Burp Suite**: Penetration testing
- **gosec**: Go security scanner

### Commands
```bash
# Security audit
make audit

# Dependency check
go list -json -m all

# Run security scanner (gosec)
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...
```

### References
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP API Security](https://owasp.org/www-project-api-security/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [Go Security](https://golang.org/doc/security/)

---

## Support

For security issues, please report to: security@example.com

**Do not** open public issues for security vulnerabilities.
