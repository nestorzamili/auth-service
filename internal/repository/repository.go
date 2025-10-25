package repository

import (
	"auth-service/internal/domain"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	GetByUserID(ctx context.Context, userID int64) ([]*domain.RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
	RevokeAllByUserID(ctx context.Context, userID int64) error
	DeleteExpired(ctx context.Context) error
}
