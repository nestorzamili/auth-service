package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"
)

type client struct {
	tokens    int
	lastReset time.Time
}

var (
	clients = make(map[string]*client)
	mu      sync.Mutex
)

func RateLimit(log *logger.Logger, requestsPerMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			mu.Lock()
			c, exists := clients[ip]
			if !exists {
				c = &client{
					tokens:    requestsPerMinute,
					lastReset: time.Now(),
				}
				clients[ip] = c
			}

			if time.Since(c.lastReset) > time.Minute {
				c.tokens = requestsPerMinute
				c.lastReset = time.Now()
			}

			if c.tokens <= 0 {
				mu.Unlock()
				log.WithContext(r.Context()).WithFields(map[string]interface{}{
					"ip": ip,
				}).Warn("rate limit exceeded")
				appErr := apperrors.RateLimitExceeded()
				writeJSONError(w, appErr)
				return
			}

			c.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := splitAndTrim(xff, ",")
		if len(ips) > 0 {
			return ips[0]
		}
	}

	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
