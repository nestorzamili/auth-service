package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	apperrors "auth-service/pkg/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	config *config.JWTConfig
}

func NewJWTService(cfg *config.JWTConfig) *JWTService {
	return &JWTService{
		config: cfg,
	}
}

type customClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Type     string    `json:"type"`
	jwt.RegisteredClaims
}

func (s *JWTService) GenerateTokenPair(user *domain.User) (*domain.TokenPair, time.Time, error) {
	accessToken, _, err := s.generateToken(user, "access", s.config.AccessTokenExpiry, s.config.AccessTokenSecret)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.generateToken(user, "refresh", s.config.RefreshTokenExpiry, s.config.RefreshTokenSecret)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, refreshExpiresAt, nil
}

func (s *JWTService) generateToken(user *domain.User, tokenType string, expiry time.Duration, secret string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	claims := customClaims{
		UserID:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   fmt.Sprintf("%d", user.UserID),
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, expiresAt, nil
}

func (s *JWTService) ValidateAccessToken(tokenString string) (*domain.Claims, error) {
	return s.validateToken(tokenString, "access", s.config.AccessTokenSecret)
}

func (s *JWTService) ValidateRefreshToken(tokenString string) (*domain.Claims, error) {
	return s.validateToken(tokenString, "refresh", s.config.RefreshTokenSecret)
}

func (s *JWTService) validateToken(tokenString, expectedType, secret string) (*domain.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
				"reason": "invalid signing method",
			})
		}
		return []byte(secret), nil
	})

	if err != nil {
		switch {
		case jwt.ErrTokenExpired.Error() == err.Error():
			return nil, apperrors.TokenExpired()
		case jwt.ErrTokenNotValidYet.Error() == err.Error():
			return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
				"reason": "token not valid yet",
			})
		default:
			return nil, apperrors.TokenInvalid().WithError(err)
		}
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason": "invalid token claims",
		})
	}

	if claims.Type != expectedType {
		return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason":   "wrong token type",
			"expected": expectedType,
			"got":      claims.Type,
		})
	}

	if claims.Issuer != s.config.Issuer {
		return nil, apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason": "invalid issuer",
		})
	}

	return &domain.Claims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Type:     claims.Type,
	}, nil
}

func generateJTI() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}

func ExtractTokenFromBearer(bearerToken string) (string, error) {
	if bearerToken == "" {
		return "", apperrors.TokenMissing()
	}

	const bearerPrefix = "Bearer "
	if len(bearerToken) < len(bearerPrefix) {
		return "", apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason": "invalid bearer token format",
		})
	}

	if bearerToken[:len(bearerPrefix)] != bearerPrefix {
		return "", apperrors.TokenInvalid().WithDetails(map[string]string{
			"reason": "missing bearer prefix",
		})
	}

	token := bearerToken[len(bearerPrefix):]
	if token == "" {
		return "", apperrors.TokenMissing()
	}

	return token, nil
}
