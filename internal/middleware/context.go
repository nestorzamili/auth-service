package middleware

import (
	"fmt"
	"net/http"

	apperrors "auth-service/pkg/errors"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	ClaimsKey    contextKey = "claims"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	userID     interface{}
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) SetUserID(userID interface{}) {
	rw.userID = userID
}

func GetResponseWriter(w http.ResponseWriter) *responseWriter {
	if rw, ok := w.(*responseWriter); ok {
		return rw
	}
	return nil
}

func writeJSONError(w http.ResponseWriter, appErr *apperrors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)

	response := appErr.ToErrorResponse()
	fmt.Fprintf(w, `{"error":{"code":"%s","message":"%s"}}`, response.Error.Code, response.Error.Message)
}
