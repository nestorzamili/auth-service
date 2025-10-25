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

	log := logger.New(cfg.Logger.Level, cfg.Logger.Format)
	logger.Init(cfg.Logger.Level, cfg.Logger.Format)

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

	log.Info("running database migrations")
	if err := config.RunMigrations(db); err != nil {
		log.WithError(err).Fatal("failed to run migrations")
	}
	log.Info("migrations completed successfully")

	userRepo := repository.NewPostgresUserRepository(db)
	refreshTokenRepo := repository.NewPostgresRefreshTokenRepository(db)

	jwtService := service.NewJWTService(&cfg.JWT)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtService, log)

	authHandler := handler.NewAuthHandler(authService, log)

	router := setupRouter(authHandler, jwtService, cfg, log)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.WithField("address", server.Addr).Info("server is listening")
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

func setupRouter(authHandler *handler.AuthHandler, jwtService *service.JWTService, cfg *config.Config, log *logger.Logger) http.Handler {
	apiMux := http.NewServeMux()

	// Public endpoints
	apiMux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	apiMux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	apiMux.HandleFunc("POST /api/v1/auth/refresh", authHandler.RefreshToken)
	apiMux.HandleFunc("POST /api/v1/auth/validate", authHandler.ValidateToken)
	apiMux.HandleFunc("GET /health", handler.HealthCheck)

	// Protected endpoints (require authentication)
	authMiddleware := middleware.Auth(jwtService, log)
	apiMux.Handle("POST /api/v1/auth/logout", authMiddleware(http.HandlerFunc(authHandler.Logout)))
	apiMux.Handle("GET /api/v1/auth/me", authMiddleware(http.HandlerFunc(authHandler.Me)))

	// Apply middleware chain (order matters!)
	var apiHandler http.Handler = apiMux

	// 1. Security headers (applied to all responses)
	apiHandler = middleware.SecurityHeaders(apiHandler)

	// 2. CORS configuration
	allowedOrigins := []string{"*"}
	if cfg.IsProduction() {
		// In production, specify your frontend domains
		allowedOrigins = []string{"https://yourdomain.com", "https://app.yourdomain.com"}
	}
	apiHandler = middleware.CORS(allowedOrigins)(apiHandler)

	// 3. Content-Type validation for POST/PUT/PATCH
	apiHandler = middleware.ValidateContentType(apiHandler)

	// 4. Request body size limit (1MB)
	apiHandler = middleware.MaxBodySize(1 << 20)(apiHandler)

	// 5. Request timeout (30 seconds)
	apiHandler = middleware.Timeout(30 * time.Second)(apiHandler)

	// 6. Rate limiting (per IP)
	if cfg.IsProduction() {
		apiHandler = middleware.RateLimit(100)(apiHandler) // 100 requests/minute
	} else {
		apiHandler = middleware.RateLimit(1000)(apiHandler) // 1000 requests/minute for dev
	}

	// 7. Structured logging
	apiHandler = middleware.Logger(log)(apiHandler)

	// 8. Panic recovery (should be near the top of chain)
	apiHandler = middleware.Recovery(log)(apiHandler)

	// 9. Request ID (should be first to track entire request)
	apiHandler = middleware.RequestID(apiHandler)

	// Mount API handler
	rootMux := http.NewServeMux()
	rootMux.Handle("/api/", apiHandler)
	rootMux.Handle("/health", apiHandler)

	return rootMux
}
