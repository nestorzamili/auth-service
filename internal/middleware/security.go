package middleware

import (
	"context"
	"net/http"
	"strings"

	"auth-service/internal/domain"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func Auth(log *logger.Logger, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.WithContext(r.Context()).Warn("missing authorization header")
				appErr := apperrors.Unauthorized("missing authorization header")
				writeJSONError(w, appErr)
				return
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
				log.WithContext(r.Context()).Warn("invalid authorization header format")
				appErr := apperrors.Unauthorized("invalid authorization header format")
				writeJSONError(w, appErr)
				return
			}

			tokenString := bearerToken[1]
			claims := jwt.MapClaims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, apperrors.Unauthorized("invalid signing method")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				log.WithContext(r.Context()).Warn("invalid or expired token")
				appErr := apperrors.Unauthorized("invalid or expired token")
				writeJSONError(w, appErr)
				return
			}

			userIDStr, ok := claims["user_id"].(string)
			if !ok {
				log.WithContext(r.Context()).Warn("user_id not found in token claims")
				appErr := apperrors.Unauthorized("invalid token claims")
				writeJSONError(w, appErr)
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				log.WithContext(r.Context()).Warn("invalid user_id format in token")
				appErr := apperrors.Unauthorized("invalid token claims")
				writeJSONError(w, appErr)
				return
			}

			username, _ := claims["username"].(string)
			email, _ := claims["email"].(string)
			tokenType, _ := claims["type"].(string)

			domainClaims := &domain.Claims{
				UserID:   userID,
				Username: username,
				Email:    email,
				Type:     tokenType,
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, domainClaims)
			ctx = context.WithValue(ctx, UserIDKey, userID)

			if rw := GetResponseWriter(w); rw != nil {
				rw.SetUserID(userID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")

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

		next.ServeHTTP(w, r)
	})
}
