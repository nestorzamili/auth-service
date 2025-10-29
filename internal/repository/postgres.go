package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/domain"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
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
		RETURNING user_id, created_at, updated_at
	`

	now := time.Now()
	user.UserID = uuid.New()
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
	).Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt)

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

func (r *PostgresUserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password_hash, full_name, is_active, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
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
		SELECT user_id, username, email, password_hash, full_name, is_active, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.UserID,
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
		SELECT user_id, username, email, password_hash, full_name, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.UserID,
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
		WHERE user_id = $7
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
		user.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("user")
	}

	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE user_id = $1`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("user")
	}

	return nil
}

type PostgresSessionRepository struct {
	db *pgxpool.Pool
}

func NewPostgresSessionRepository(db *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{db: db}
}

func (r *PostgresSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (
			user_id, refresh_token, device_info, 
			ip_address, user_agent, last_activity_at, expires_at, 
			created_at, updated_at, is_revoked
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING session_id, created_at, updated_at
	`

	now := time.Now()

	// Convert empty strings to nil for nullable fields
	var ipAddress interface{} = session.IPAddress
	if session.IPAddress == "" {
		ipAddress = nil
	}

	var deviceInfo interface{} = session.DeviceInfo
	if session.DeviceInfo == "" {
		deviceInfo = nil
	}

	var userAgent interface{} = session.UserAgent
	if session.UserAgent == "" {
		userAgent = nil
	}

	err := r.db.QueryRow(
		ctx,
		query,
		session.UserID,
		session.RefreshToken,
		deviceInfo,
		ipAddress,
		userAgent,
		now,
		session.ExpiresAt,
		now,
		now,
		false,
	).Scan(&session.SessionID, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r *PostgresSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	query := `
		SELECT session_id, user_id, refresh_token, device_info, 
		       ip_address, user_agent, last_activity_at, expires_at, 
		       created_at, updated_at, is_revoked, revoked_at
		FROM sessions
		WHERE refresh_token = $1 AND expires_at > $2
	`

	session := &domain.Session{}
	err := r.db.QueryRow(ctx, query, refreshToken, time.Now()).Scan(
		&session.SessionID,
		&session.UserID,
		&session.RefreshToken,
		&session.DeviceInfo,
		&session.IPAddress,
		&session.UserAgent,
		&session.LastActivityAt,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.IsRevoked,
		&session.RevokedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("session")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (r *PostgresSessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Session, error) {
	query := `
		SELECT session_id, user_id, refresh_token, device_info, 
		       ip_address, user_agent, last_activity_at, expires_at, 
		       created_at, updated_at, is_revoked, revoked_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	session := &domain.Session{}
	err := r.db.QueryRow(ctx, query, userID, time.Now()).Scan(
		&session.SessionID,
		&session.UserID,
		&session.RefreshToken,
		&session.DeviceInfo,
		&session.IPAddress,
		&session.UserAgent,
		&session.LastActivityAt,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.IsRevoked,
		&session.RevokedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("session")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (r *PostgresSessionRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	query := `
		SELECT session_id, user_id, refresh_token, device_info, 
		       ip_address, user_agent, last_activity_at, expires_at, 
		       created_at, updated_at, is_revoked, revoked_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		session := &domain.Session{}
		err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.RefreshToken,
			&session.DeviceInfo,
			&session.IPAddress,
			&session.UserAgent,
			&session.LastActivityAt,
			&session.ExpiresAt,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.IsRevoked,
			&session.RevokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *PostgresSessionRepository) UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE sessions
		SET last_activity_at = $1, updated_at = $2
		WHERE session_id = $3
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, now, now, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("session")
	}

	return nil
}

func (r *PostgresSessionRepository) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE sessions
		SET is_revoked = true, revoked_at = $1, updated_at = $2
		WHERE session_id = $3
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, now, now, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("session")
	}

	return nil
}

func (r *PostgresSessionRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE sessions
		SET is_revoked = true, revoked_at = $1, updated_at = $2
		WHERE user_id = $3 AND is_revoked = false
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, now, now, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	return nil
}

func (r *PostgresSessionRepository) DeleteByID(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE session_id = $1`

	result, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("session")
	}

	return nil
}

func (r *PostgresSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

func (r *PostgresSessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < $1`

	_, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

func (r *PostgresSessionRepository) ReplaceUserSession(ctx context.Context, session *domain.Session) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	deleteQuery := `DELETE FROM sessions WHERE user_id = $1`
	_, err = tx.Exec(ctx, deleteQuery, session.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete existing sessions: %w", err)
	}

	insertQuery := `
		INSERT INTO sessions (
			user_id, refresh_token, device_info, 
			ip_address, user_agent, last_activity_at, expires_at, 
			created_at, updated_at, is_revoked
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING session_id, created_at, updated_at
	`

	now := time.Now()

	// Convert empty strings to nil for nullable fields
	var ipAddress interface{} = session.IPAddress
	if session.IPAddress == "" {
		ipAddress = nil
	}

	var deviceInfo interface{} = session.DeviceInfo
	if session.DeviceInfo == "" {
		deviceInfo = nil
	}

	var userAgent interface{} = session.UserAgent
	if session.UserAgent == "" {
		userAgent = nil
	}

	err = tx.QueryRow(
		ctx,
		insertQuery,
		session.UserID,
		session.RefreshToken,
		deviceInfo,
		ipAddress,
		userAgent,
		now,
		session.ExpiresAt,
		now,
		now,
		false,
	).Scan(&session.SessionID, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create new session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
