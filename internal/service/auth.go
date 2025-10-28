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
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	jwtService  *JWTService
	logger      *logger.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	jwtService *JWTService,
	log *logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtService:  jwtService,
		logger:      log,
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

	metadata := s.getSessionMetadataFromContext(ctx)

	tokens, err := s.generateAndStoreTokensWithSession(ctx, user, metadata)
	if err != nil {
		log.WithError(err).Error("failed to generate tokens after registration")
		return nil, err
	}

	return &domain.AuthResponse{
		User: &domain.UserResponse{
			UserID:   user.UserID,
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
		log.WithField("user_id", user.UserID).Warn("login failed: user is inactive")
		return nil, apperrors.Unauthorized("account is inactive")
	}

	if err := verifyPassword(user.PasswordHash, req.Password); err != nil {
		log.WithField("user_id", user.UserID).Warn("login failed: invalid password")
		return nil, apperrors.InvalidCredentials()
	}

	metadata := s.getSessionMetadataFromContext(ctx)

	tokens, err := s.generateAndStoreTokensWithSession(ctx, user, metadata)
	if err != nil {
		log.WithError(err).Error("failed to generate tokens after login")
		return nil, err
	}

	log.WithField("user_id", user.UserID).Info("user logged in successfully, previous session replaced")

	return &domain.AuthResponse{
		User: &domain.UserResponse{
			UserID:   user.UserID,
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

	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		log.WithError(err).Warn("session not found in database")
		return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason": "session not found or revoked",
		})
	}

	if !session.IsValid() {
		log.WithField("session_id", session.SessionID).Warn("session is invalid")
		return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason": "session expired or revoked",
		})
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		log.WithError(err).Error("failed to get user for refresh token")
		return nil, apperrors.NotFound("user")
	}

	if !user.IsActive {
		log.WithField("user_id", user.UserID).Warn("refresh token rejected: user is inactive")
		return nil, apperrors.Unauthorized("account is inactive")
	}

	if err := s.sessionRepo.Revoke(ctx, session.SessionID); err != nil {
		log.WithError(err).Error("failed to revoke old session")
	}

	metadata := s.getSessionMetadataFromContext(ctx)

	log.WithField("user_id", user.UserID).Info("tokens refreshed successfully")

	return s.generateAndStoreTokensWithSession(ctx, user, metadata)
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
		log.WithField("user_id", user.UserID).Warn("token rejected: user is inactive")
		return nil, apperrors.Unauthorized("account is inactive")
	}

	return claims, nil
}

func (s *AuthService) Logout(ctx context.Context, userID int64) error {
	log := s.logger.WithContext(ctx).WithField("user_id", userID)

	if err := s.sessionRepo.RevokeAllByUserID(ctx, userID); err != nil {
		log.WithError(err).Error("failed to revoke sessions")
		return apperrors.Internal("failed to logout")
	}

	log.Info("user logged out successfully, all sessions revoked")
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

func (s *AuthService) generateAndStoreTokensWithSession(ctx context.Context, user *domain.User, metadata *domain.SessionMetadata) (*domain.TokenPair, error) {
	tokens, refreshExpiresAt, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	session := &domain.Session{
		UserID:       user.UserID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    refreshExpiresAt,
	}

	if metadata != nil {
		session.DeviceInfo = metadata.DeviceInfo
		session.IPAddress = metadata.IPAddress
		session.UserAgent = metadata.UserAgent
	}

	if err := s.sessionRepo.ReplaceUserSession(ctx, session); err != nil {
		s.logger.WithError(err).Error("failed to replace user session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return tokens, nil
}

func (s *AuthService) getSessionMetadataFromContext(ctx context.Context) *domain.SessionMetadata {
	metadata := &domain.SessionMetadata{}

	if ipAddr, ok := ctx.Value("ip_address").(string); ok {
		metadata.IPAddress = ipAddr
	}

	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		metadata.UserAgent = userAgent
	}

	if deviceInfo, ok := ctx.Value("device_info").(string); ok {
		metadata.DeviceInfo = deviceInfo
	}

	return metadata
}

func (s *AuthService) ValidateSession(ctx context.Context, refreshToken string) (*domain.Session, error) {
	log := s.logger.WithContext(ctx)

	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		log.WithError(err).Debug("session not found")
		return nil, apperrors.NotFound("session")
	}

	if !session.IsValid() {
		log.WithField("session_id", session.SessionID).Warn("session is expired or revoked")
		return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason": "session expired or revoked",
		})
	}

	if err := s.sessionRepo.UpdateLastActivity(ctx, session.SessionID); err != nil {
		log.WithError(err).Warn("failed to update session last activity")
	}

	return session, nil
}

func (s *AuthService) GetUserSessions(ctx context.Context, userID int64) ([]*domain.Session, error) {
	return s.sessionRepo.GetAllByUserID(ctx, userID)
}

func (s *AuthService) CleanupExpiredSessions(ctx context.Context) error {
	log := s.logger.WithContext(ctx)

	if err := s.sessionRepo.DeleteExpired(ctx); err != nil {
		log.WithError(err).Error("failed to cleanup expired sessions")
		return err
	}

	log.Info("expired sessions cleaned up successfully")
	return nil
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
