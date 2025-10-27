package logger

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ContextKey string

const (
	TimeFormat = "2006-01-02 15:04:05.000"
)

type Logger struct {
	logger zerolog.Logger
}

type orderedWriter struct {
	writer io.Writer
}

func (w *orderedWriter) Write(p []byte) (n int, err error) {
	var logData map[string]interface{}
	if err := json.Unmarshal(p, &logData); err != nil {
		return w.writer.Write(p)
	}

	orderedJSON := "{"
	first := true

	fieldOrder := []string{
		"time", "level", "request_id", "method", "path", "status",
		"duration_ms", "remote_addr", "user_agent", "user_id",
		"message", "environment", "port", "address", "error",
	}

	for _, key := range fieldOrder {
		if val, ok := logData[key]; ok {
			if !first {
				orderedJSON += ","
			}
			jsonVal, _ := json.Marshal(val)
			orderedJSON += `"` + key + `":` + string(jsonVal)
			first = false
			delete(logData, key)
		}
	}

	for key, val := range logData {
		if !first {
			orderedJSON += ","
		}
		jsonVal, _ := json.Marshal(val)
		orderedJSON += `"` + key + `":` + string(jsonVal)
		first = false
	}

	orderedJSON += "}\n"
	return w.writer.Write([]byte(orderedJSON))
}

func New(level, format, filePath string) *Logger {
	zerolog.TimeFieldFormat = TimeFormat

	var writers []io.Writer

	writers = append(writers, &orderedWriter{writer: os.Stdout})

	if filePath != "" {
		logDir := filepath.Dir(filePath)
		if err := os.MkdirAll(logDir, 0755); err == nil {
			if file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
				writers = append(writers, &orderedWriter{writer: file})
			}
		}
	}

	multiWriter := io.MultiWriter(writers...)
	logLevel := parseLevel(level)
	zerolog.SetGlobalLevel(logLevel)

	logger := zerolog.New(multiWriter).
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
	return l
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

func (l *Logger) HTTPRequest(method, path string, status int, durationMs int64, remoteAddr, userAgent, requestID string, userID interface{}) {
	event := l.logger.Info().
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Int64("duration_ms", durationMs).
		Str("remote_addr", remoteAddr).
		Str("user_agent", userAgent)

	if requestID != "" {
		event = event.Str("request_id", requestID)
	}

	if userID != nil {
		event = event.Interface("user_id", userID)
	}

	event.Msg("")
}

var global *Logger

func Init(level, format, filePath string) {
	global = New(level, format, filePath)
	log.Logger = global.logger
}

func Global() *Logger {
	if global == nil {
		global = New("info", "text", "logs/app.log")
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
