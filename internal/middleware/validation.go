package middleware

import (
	"context"
	"net/http"
	"time"

	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"
)

func ValidateContentType(log *logger.Logger, contentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				ct := r.Header.Get("Content-Type")
				if ct != contentType {
					log.WithContext(r.Context()).WithFields(map[string]interface{}{
						"received": ct,
						"expected": contentType,
					}).Warn("invalid content type")
					appErr := apperrors.InvalidInput("Content-Type must be " + contentType)
					writeJSONError(w, appErr)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func MaxBodySize(log *logger.Logger, maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func Timeout(log *logger.Logger, timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})

			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					log.WithContext(r.Context()).Warn("request timeout")
					appErr := apperrors.Internal("request timeout")
					writeJSONError(w, appErr)
				}
			}
		})
	}
}
