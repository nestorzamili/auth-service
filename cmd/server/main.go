package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Logger.Level, cfg.Logger.Format, cfg.Logger.FilePath)
	logger.Init(cfg.Logger.Level, cfg.Logger.Format, cfg.Logger.FilePath)

	log.WithFields(map[string]interface{}{
		"environment": cfg.Server.Environment,
		"port":        cfg.Server.Port,
	}).Info("starting auth microservice")

	log.Info("connecting to PostgreSQL database")
	db, err := config.NewPostgresConnection(&cfg.Database)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database - service cannot start")
	}
	defer db.Close()

	log.Info("successfully connected to PostgreSQL database")

	userRepo := repository.NewPostgresUserRepository(db)
	refreshTokenRepo := repository.NewPostgresRefreshTokenRepository(db)

	jwtService := service.NewJWTService(&cfg.JWT)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtService, log)

	authHandler := handler.NewAuthHandler(authService, log)

	router := setupRouter(authHandler, cfg, log)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.WithField("port", cfg.Server.Port).Info("server is listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("server forced to shutdown")
	}

	log.Info("server stopped")
}

func setupRouter(authHandler *handler.AuthHandler, cfg *config.Config, log *logger.Logger) http.Handler {
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	apiMux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	apiMux.HandleFunc("POST /api/v1/auth/refresh", authHandler.RefreshToken)
	apiMux.HandleFunc("POST /api/v1/auth/validate", authHandler.ValidateToken)
	apiMux.HandleFunc("GET /health", handler.HealthCheck)

	authMiddleware := middleware.Auth(log, cfg.JWT.AccessTokenSecret)
	apiMux.Handle("POST /api/v1/auth/logout", authMiddleware(http.HandlerFunc(authHandler.Logout)))
	apiMux.Handle("GET /api/v1/auth/me", authMiddleware(http.HandlerFunc(authHandler.Me)))

	var apiHandler http.Handler = apiMux

	apiHandler = middleware.RateLimit(log, cfg.Server.RateLimit)(apiHandler)

	apiHandler = middleware.Timeout(log, 30*time.Second)(apiHandler)

	apiHandler = middleware.MaxBodySize(log, 1<<20)(apiHandler)

	apiHandler = middleware.ValidateContentType(log, "application/json")(apiHandler)

	apiHandler = middleware.CORS(cfg.Server.AllowedOrigins)(apiHandler)

	apiHandler = middleware.SecurityHeaders(apiHandler)

	apiHandler = middleware.Recovery(log)(apiHandler)

	apiHandler = middleware.Logger(log)(apiHandler)

	apiHandler = middleware.RequestID(apiHandler)

	rootMux := http.NewServeMux()
	rootMux.Handle("/api/", apiHandler)
	rootMux.Handle("/health", apiHandler)

	return rootMux
}
