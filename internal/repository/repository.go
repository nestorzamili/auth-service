package repository

import (
	"auth-service/internal/domain"
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Session, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error)
	UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error
	Revoke(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteByID(ctx context.Context, sessionID uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	ReplaceUserSession(ctx context.Context, session *domain.Session) error
}
