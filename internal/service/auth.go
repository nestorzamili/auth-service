package service

import (
	"context"
	"fmt"

	"auth-service/internal/domain"
	"auth-service/internal/repository"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtService       *JWTService
	logger           *logger.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtService *JWTService,
	log *logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		logger:           log,
	}
}

func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	log := s.logger.WithContext(ctx)

	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		log.Warn("registration failed: username already exists")
		return nil, apperrors.AlreadyExists("username")
	}

	existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		log.Warn("registration failed: email already exists")
		return nil, apperrors.AlreadyExists("email")
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.WithError(err).Error("failed to hash password")
		return nil, apperrors.Internal("failed to process password")
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		log.WithError(err).Error("failed to create user")
		return nil, apperrors.Internal("failed to create user")
	}

	tokens, err := s.generateAndStoreTokens(ctx, user)
	if err != nil {
		log.WithError(err).Error("failed to generate tokens after registration")
		return nil, err
	}

	return &domain.AuthResponse{
		User: &domain.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
		},
		Tokens: tokens,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	log := s.logger.WithContext(ctx)

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		log.Warn("login failed: user not found")
		return nil, apperrors.InvalidCredentials()
	}

	if !user.IsActive {
		log.WithField("user_id", user.ID).Warn("login failed: user is inactive")
		return nil, apperrors.Unauthorized("account is inactive")
	}

	if err := verifyPassword(user.PasswordHash, req.Password); err != nil {
		log.WithField("user_id", user.ID).Warn("login failed: invalid password")
		return nil, apperrors.InvalidCredentials()
	}

	tokens, err := s.generateAndStoreTokens(ctx, user)
	if err != nil {
		log.WithError(err).Error("failed to generate tokens after login")
		return nil, err
	}

	return &domain.AuthResponse{
		User: &domain.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
		},
		Tokens: tokens,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*domain.TokenPair, error) {
	log := s.logger.WithContext(ctx)

	claims, err := s.jwtService.ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		log.WithError(err).Warn("refresh token validation failed")
		return nil, err
	}

	storedToken, err := s.refreshTokenRepo.GetByToken(ctx, refreshTokenStr)
	if err != nil {
		log.WithError(err).Warn("refresh token not found in database")
		return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason": "token not found or revoked",
		})
	}

	if !storedToken.IsValid() {
		log.WithField("token_id", storedToken.ID).Warn("refresh token is invalid")
		return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason": "token expired or revoked",
		})
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		log.WithError(err).Error("failed to get user for refresh token")
		return nil, apperrors.NotFound("user")
	}

	if !user.IsActive {
		log.WithField("user_id", user.ID).Warn("refresh token rejected: user is inactive")
		return nil, apperrors.Unauthorized("account is inactive")
	}

	if err := s.refreshTokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		log.WithError(err).Error("failed to revoke old refresh token")
	}

	log.WithField("user_id", user.ID).Info("tokens refreshed successfully")

	return s.generateAndStoreTokens(ctx, user)
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*domain.Claims, error) {
	log := s.logger.WithContext(ctx)

	claims, err := s.jwtService.ValidateAccessToken(tokenStr)
	if err != nil {
		log.WithError(err).Debug("token validation failed")
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		log.WithError(err).Warn("user not found for valid token")
		return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason": "user not found",
		})
	}

	if !user.IsActive {
		log.WithField("user_id", user.ID).Warn("token rejected: user is inactive")
		return nil, apperrors.Unauthorized("account is inactive")
	}

	return claims, nil
}

func (s *AuthService) Logout(ctx context.Context, userID int64) error {
	log := s.logger.WithContext(ctx).WithField("user_id", userID)

	if err := s.refreshTokenRepo.RevokeAllByUserID(ctx, userID); err != nil {
		log.WithError(err).Error("failed to revoke refresh tokens")
		return apperrors.Internal("failed to logout")
	}

	log.Info("user logged out successfully")
	return nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.NotFound("user")
	}

	if !user.IsActive {
		return nil, apperrors.Unauthorized("account is inactive")
	}

	return user, nil
}

func (s *AuthService) generateAndStoreTokens(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	tokens, refreshExpiresAt, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	refreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     tokens.RefreshToken,
		ExpiresAt: refreshExpiresAt,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		s.logger.WithError(err).Error("failed to store refresh token")
	}

	return tokens, nil
}

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
