package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	JWT      JWTConfig
	Database DatabaseConfig
	Logger   LoggerConfig
}

type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Environment     string
	AllowedOrigins  []string
	RateLimit       int
}

type JWTConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
	AllowedAlgorithm   string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type LoggerConfig struct {
	Level    string
	Format   string // json or text
	FilePath string // path to log file
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 5 * time.Second,
			Environment:     getEnv("ENVIRONMENT", "development"),
			AllowedOrigins:  getEnvAsSlice("ALLOWED_ORIGINS", []string{"*"}),
			RateLimit:       getEnvAsInt("RATE_LIMIT", 100),
		},
		JWT: JWTConfig{
			AccessTokenSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshTokenSecret: getEnv("JWT_REFRESH_SECRET", ""),
			AccessTokenExpiry:  getEnvAsDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			Issuer:             "auth-service",
			AllowedAlgorithm:   "HS256",
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "authdb"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Logger: LoggerConfig{
			Level:    getEnv("LOG_LEVEL", "info"),
			Format:   "json",
			FilePath: "logs/app.log",
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.JWT.AccessTokenSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required - must be set in environment")
	}
	if c.JWT.RefreshTokenSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required - must be set in environment")
	}
	if len(c.JWT.AccessTokenSecret) < 32 {
		return fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 characters for security (current: %d)", len(c.JWT.AccessTokenSecret))
	}
	if len(c.JWT.RefreshTokenSecret) < 32 {
		return fmt.Errorf("JWT_REFRESH_SECRET must be at least 32 characters for security (current: %d)", len(c.JWT.RefreshTokenSecret))
	}
	if c.JWT.AccessTokenSecret == c.JWT.RefreshTokenSecret {
		return fmt.Errorf("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must be different")
	}
	if c.JWT.AccessTokenExpiry < 1*time.Minute {
		return fmt.Errorf("JWT_ACCESS_EXPIRY must be at least 1 minute")
	}
	if c.JWT.RefreshTokenExpiry < 1*time.Hour {
		return fmt.Errorf("JWT_REFRESH_EXPIRY must be at least 1 hour")
	}
	if c.JWT.AccessTokenExpiry >= c.JWT.RefreshTokenExpiry {
		return fmt.Errorf("JWT_REFRESH_EXPIRY must be longer than JWT_ACCESS_EXPIRY")
	}

	validEnvs := map[string]bool{"development": true, "staging": true, "production": true}
	if !validEnvs[c.Server.Environment] {
		return fmt.Errorf("invalid environment: %s (must be development, staging, or production)", c.Server.Environment)
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be between 1-65535)", c.Server.Port)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid DB_PORT: %d (must be between 1-65535)", c.Database.Port)
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required - database credentials must be set")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	validSSLModes := map[string]bool{"disable": true, "require": true, "verify-ca": true, "verify-full": true}
	if !validSSLModes[c.Database.SSLMode] {
		return fmt.Errorf("invalid DB_SSL_MODE: %s (must be disable, require, verify-ca, or verify-full)", c.Database.SSLMode)
	}

	if c.IsProduction() {
		if c.Database.SSLMode == "disable" {
			return fmt.Errorf("SSL must be enabled in production (DB_SSL_MODE cannot be 'disable')")
		}
		if c.JWT.AccessTokenExpiry > 1*time.Hour {
			return fmt.Errorf("in production, JWT_ACCESS_EXPIRY should not exceed 1 hour for security")
		}
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logger.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.Logger.Level)
	}
	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[c.Logger.Format] {
		return fmt.Errorf("invalid log format: %s (must be json or text)", c.Logger.Format)
	}

	return nil
}

func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		var result []string
		for _, v := range splitAndTrim(value, ",") {
			if v != "" {
				result = append(result, v)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, v := range splitString(s, sep) {
		trimmed := trimSpace(v)
		result = append(result, trimmed)
	}
	return result
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	var current string
	for _, char := range s {
		if string(char) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	result = append(result, current)
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
