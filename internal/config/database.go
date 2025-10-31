package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresConnection(cfg *DatabaseConfig) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=users",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func RunMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	// First, ensure the extension is created in the public schema
	if _, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public;`); err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	migrations := []string{
		`CREATE SCHEMA IF NOT EXISTS users;
		
		CREATE TABLE IF NOT EXISTS users.users (
			user_id UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
			username VARCHAR(30) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			full_name VARCHAR(100) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_users_username ON users.users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users.users(email);
		CREATE INDEX IF NOT EXISTS idx_users_is_active ON users.users(is_active);`,

		`CREATE TABLE IF NOT EXISTS users.sessions (
			session_id UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users.users(user_id) ON DELETE CASCADE,
			refresh_token TEXT NOT NULL UNIQUE,
			device_info TEXT,
			ip_address INET,
			user_agent TEXT,
			last_activity_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_revoked BOOLEAN NOT NULL DEFAULT false,
			revoked_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON users.sessions(user_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON users.sessions(refresh_token);
		CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON users.sessions(expires_at);
		CREATE INDEX IF NOT EXISTS idx_sessions_last_activity ON users.sessions(last_activity_at);
		CREATE INDEX IF NOT EXISTS idx_sessions_is_revoked ON users.sessions(is_revoked);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_user_id_active ON users.sessions(user_id) WHERE is_revoked = false;`,
	}

	for i, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", i+1, err)
		}
	}

	return nil
}
