CREATE TABLE IF NOT EXISTS users.sessions (
    session_id UUID PRIMARY KEY DEFAULT users.uuid_generate_v4(),
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_user_id_active ON users.sessions(user_id) 
WHERE is_revoked = false;