package domain

import (
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Type     string `json:"type"` // "access" or "refresh"
}

type RefreshToken struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	IsRevoked bool       `json:"is_revoked"`
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required,username"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,username"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type AuthResponse struct {
	User   *User      `json:"user"`
	Tokens *TokenPair `json:"tokens"`
}

type ValidateTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

type ValidateTokenResponse struct {
	Valid  bool    `json:"valid"`
	Claims *Claims `json:"claims,omitempty"`
}
