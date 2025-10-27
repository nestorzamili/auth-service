package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"

	"github.com/google/uuid"
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

			var requestID string
			if rid := r.Context().Value(RequestIDKey); rid != nil {
				if str, ok := rid.(string); ok {
					requestID = str
				}
			}

			log.HTTPRequest(
				r.Method,
				r.URL.Path,
				wrapped.statusCode,
				duration.Milliseconds(),
				r.RemoteAddr,
				r.UserAgent(),
				requestID,
				wrapped.userID,
			)
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
