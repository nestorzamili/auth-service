package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"auth-service/internal/service"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"
	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	ClaimsKey    contextKey = "claims"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			log.WithContext(r.Context()).WithFields(map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      wrapped.statusCode,
				"duration_ms": duration.Milliseconds(),
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			}).Info("request completed")
		})
	}
}

func Recovery(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.WithContext(r.Context()).WithFields(map[string]interface{}{
						"error": err,
						"stack": string(debug.Stack()),
					}).Error("panic recovered")

					appErr := apperrors.Internal("internal server error")
					writeJSONError(w, appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else if len(allowedOrigins) > 0 {
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	type client struct {
		requests  int
		resetTime time.Time
	}

	clients := make(map[string]*client)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			now := time.Now()

			// Clean up old entries periodically
			if len(clients) > 10000 {
				for k, v := range clients {
					if now.After(v.resetTime.Add(5 * time.Minute)) {
						delete(clients, k)
					}
				}
			}

			c, exists := clients[ip]
			if !exists || now.After(c.resetTime) {
				clients[ip] = &client{
					requests:  1,
					resetTime: now.Add(time.Minute),
				}
				next.ServeHTTP(w, r)
				return
			}

			if c.requests >= requestsPerMinute {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", c.resetTime.Unix()))

				appErr := apperrors.RateLimitExceeded()
				writeJSONError(w, appErr)
				return
			}

			c.requests++
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", requestsPerMinute-c.requests))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", c.resetTime.Unix()))

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateContentType ensures only JSON content is accepted for POST/PUT/PATCH requests
func ValidateContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" && !contains(contentType, "application/json") {
				appErr := apperrors.ValidationFailed("Content-Type must be application/json")
				writeJSONError(w, appErr)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// MaxBodySize limits the size of request body
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout adds a timeout to the request context
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Auth(jwtService *service.JWTService, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				appErr := apperrors.TokenMissing()
				writeJSONError(w, appErr)
				return
			}

			token, err := service.ExtractTokenFromBearer(authHeader)
			if err != nil {
				if appErr, ok := err.(*apperrors.AppError); ok {
					writeJSONError(w, appErr)
				} else {
					writeJSONError(w, apperrors.TokenInvalid())
				}
				return
			}

			claims, err := jwtService.ValidateAccessToken(token)
			if err != nil {
				if appErr, ok := err.(*apperrors.AppError); ok {
					writeJSONError(w, appErr)
				} else {
					log.WithContext(r.Context()).WithError(err).Error("token validation failed")
					writeJSONError(w, apperrors.TokenInvalid())
				}
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			ctx = context.WithValue(ctx, UserIDKey, fmt.Sprintf("%d", claims.UserID))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func writeJSONError(w http.ResponseWriter, appErr *apperrors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)

	response := appErr.ToErrorResponse()
	fmt.Fprintf(w, `{"error":{"code":"%s","message":"%s"}}`, response.Error.Code, response.Error.Message)
}

// getClientIP extracts the real client IP from headers or RemoteAddr
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := splitAndTrim(xff, ",")
		if len(ips) > 0 {
			return ips[0]
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, item := range splitString(s, sep) {
		trimmed := trimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
