package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/domain"
	apperrors "auth-service/pkg/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	err := r.db.QueryRow(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.IsActive,
		now,
		now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_username_key\" (SQLSTATE 23505)" {
			return apperrors.AlreadyExists("username")
		}
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
			return apperrors.AlreadyExists("email")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("user")
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, is_active, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("user")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("user")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, password_hash = $3, full_name = $4, is_active = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.Exec(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.IsActive,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("user")
	}

	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("user")
	}

	return nil
}

type PostgresRefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRefreshTokenRepository(db *pgxpool.Pool) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (r *PostgresRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at, is_revoked)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	now := time.Now()
	err := r.db.QueryRow(
		ctx,
		query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		now,
		false,
	).Scan(&token.ID, &token.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *PostgresRefreshTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked_at, is_revoked
		FROM refresh_tokens
		WHERE token = $1
	`

	token := &domain.RefreshToken{}
	err := r.db.QueryRow(ctx, query, tokenStr).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.RevokedAt,
		&token.IsRevoked,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("refresh token")
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return token, nil
}

func (r *PostgresRefreshTokenRepository) GetByUserID(ctx context.Context, userID int64) ([]*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked_at, is_revoked
		FROM refresh_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*domain.RefreshToken
	for rows.Next() {
		token := &domain.RefreshToken{}
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Token,
			&token.ExpiresAt,
			&token.CreatedAt,
			&token.RevokedAt,
			&token.IsRevoked,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan refresh token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, id int64) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = true, revoked_at = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("refresh token")
	}

	return nil
}

func (r *PostgresRefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID int64) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = true, revoked_at = $1
		WHERE user_id = $2 AND is_revoked = false
	`

	_, err := r.db.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	return nil
}

func (r *PostgresRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < $1
	`

	_, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return nil
}
