package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"auth-service/internal/domain"
	"auth-service/internal/middleware"
	"auth-service/internal/service"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"
	"auth-service/pkg/validator"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *logger.Logger
}

func NewAuthHandler(authService *service.AuthService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      log,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Warn("failed to decode registration request")
		writeAppError(w, apperrors.InvalidInput("invalid request body"))
		return
	}

	if err := validator.Validate(&req); err != nil {
		log.WithError(err).Warn("registration validation failed")
		writeAppError(w, apperrors.ValidationFailed(err.Error()))
		return
	}

	response, err := h.authService.Register(ctx, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			writeAppError(w, appErr)
		} else {
			log.WithError(err).Error("registration failed")
			writeAppError(w, apperrors.Internal("registration failed"))
		}
		return
	}

	if rw := middleware.GetResponseWriter(w); rw != nil {
		rw.SetUserID(response.User.UserID)
	}

	writeJSendSuccess(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Warn("failed to decode login request")
		writeAppError(w, apperrors.InvalidInput("invalid request body"))
		return
	}

	if err := validator.Validate(&req); err != nil {
		log.WithError(err).Warn("login validation failed")
		writeAppError(w, apperrors.ValidationFailed(err.Error()))
		return
	}

	response, err := h.authService.Login(ctx, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			writeAppError(w, appErr)
		} else {
			log.WithError(err).Error("login failed")
			writeAppError(w, apperrors.Internal("login failed"))
		}
		return
	}

	if rw := middleware.GetResponseWriter(w); rw != nil {
		rw.SetUserID(response.User.UserID)
	}

	writeJSendSuccess(w, http.StatusOK, response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	var req domain.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Warn("failed to decode refresh token request")
		writeAppError(w, apperrors.InvalidInput("invalid request body"))
		return
	}

	if err := validator.Validate(&req); err != nil {
		log.WithError(err).Warn("refresh token validation failed")
		writeAppError(w, apperrors.ValidationFailed(err.Error()))
		return
	}

	tokens, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			writeAppError(w, appErr)
		} else {
			log.WithError(err).Error("token refresh failed")
			writeAppError(w, apperrors.Internal("token refresh failed"))
		}
		return
	}

	writeJSendSuccess(w, http.StatusOK, tokens)
}

func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	var req domain.ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Warn("failed to decode validate token request")
		writeAppError(w, apperrors.InvalidInput("invalid request body"))
		return
	}

	if err := validator.Validate(&req); err != nil {
		log.WithError(err).Warn("validate token request validation failed")
		writeAppError(w, apperrors.ValidationFailed(err.Error()))
		return
	}

	claims, err := h.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		response := &domain.ValidateTokenResponse{
			Valid:  false,
			Claims: nil,
		}
		writeJSendSuccess(w, http.StatusOK, response)
		return
	}

	response := &domain.ValidateTokenResponse{
		Valid:  true,
		Claims: claims,
	}
	writeJSendSuccess(w, http.StatusOK, response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	claims, ok := ctx.Value(middleware.ClaimsKey).(*domain.Claims)
	if !ok {
		log.Error("failed to get claims from context")
		writeAppError(w, apperrors.Unauthorized("unauthorized"))
		return
	}

	if err := h.authService.Logout(ctx, claims.UserID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			writeAppError(w, appErr)
		} else {
			log.WithError(err).Error("logout failed")
			writeAppError(w, apperrors.Internal("logout failed"))
		}
		return
	}

	writeJSendSuccess(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.WithContext(ctx)

	claims, ok := ctx.Value(middleware.ClaimsKey).(*domain.Claims)
	if !ok {
		log.Error("failed to get claims from context")
		writeAppError(w, apperrors.Unauthorized("unauthorized"))
		return
	}

	user, err := h.authService.GetUserByID(ctx, claims.UserID)
	if err != nil {
		log.WithError(err).Error("failed to get user data")
		writeAppError(w, apperrors.Internal("failed to get user data"))
		return
	}

	writeJSendSuccess(w, http.StatusOK, &domain.UserResponse{
		UserID:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
	})
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSendSuccess(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(logger.TimeFormat),
	})
}
