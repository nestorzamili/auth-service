package logger

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	UserIDKey    ContextKey = "user_id"
	TimeFormat              = "2006-01-02 15:04:05"
)

type Logger struct {
	logger zerolog.Logger
}

func New(level, format string) *Logger {
	zerolog.TimeFieldFormat = TimeFormat

	var writer io.Writer = os.Stdout

	if format == "text" {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: TimeFormat,
		}
	}

	logLevel := parseLevel(level)
	zerolog.SetGlobalLevel(logLevel)

	logger := zerolog.New(writer).
		With().
		Timestamp().
		Logger()

	return &Logger{logger: logger}
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.logger.With()

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.Str("request_id", requestID)
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		logger = logger.Str("user_id", userID)
	}

	return &Logger{logger: logger.Logger()}
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{logger: l.logger.With().Interface(key, value).Logger()}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	logger := l.logger.With()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	return &Logger{logger: logger.Logger()}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{logger: l.logger.With().Err(err).Logger()}
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

var global *Logger

func Init(level, format string) {
	global = New(level, format)
	log.Logger = global.logger
}

func Global() *Logger {
	if global == nil {
		global = New("info", "json")
	}
	return global
}

func Debug(msg string) {
	Global().Debug(msg)
}

func Info(msg string) {
	Global().Info(msg)
}

func Warn(msg string) {
	Global().Warn(msg)
}

func Error(msg string) {
	Global().Error(msg)
}

func Fatal(msg string) {
	Global().Fatal(msg)
}

func WithContext(ctx context.Context) *Logger {
	return Global().WithContext(ctx)
}

func WithField(key string, value interface{}) *Logger {
	return Global().WithField(key, value)
}

func WithFields(fields map[string]interface{}) *Logger {
	return Global().WithFields(fields)
}

func WithError(err error) *Logger {
	return Global().WithError(err)
}
