package handler

import (
	"encoding/json"
	"net/http"

	apperrors "auth-service/pkg/errors"
)

type SuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type FailResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func writeJSendSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Status: "success",
		Data:   data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeJSendFail(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := FailResponse{
		Status: "fail",
		Data:   data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeJSendError(w http.ResponseWriter, statusCode int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Status:  "error",
		Message: message,
		Code:    code,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
	}
}

func writeAppError(w http.ResponseWriter, appErr *apperrors.AppError) {
	if appErr.HTTPStatus >= 500 {
		writeJSendError(w, appErr.HTTPStatus, appErr.Message, string(appErr.Code))
		return
	}

	failData := map[string]interface{}{
		"code":    appErr.Code,
		"message": appErr.Message,
	}

	if len(appErr.Details) > 0 {
		failData["details"] = appErr.Details
	}

	writeJSendFail(w, appErr.HTTPStatus, failData)
}
