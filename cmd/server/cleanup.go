package main

import (
	"context"
	"time"

	"auth-service/internal/service"
	"auth-service/pkg/logger"
)

func StartSessionCleanup(authService *service.AuthService, log *logger.Logger, interval time.Duration) {
	log.WithField("interval", interval).Info("starting session cleanup scheduler")

	ticker := time.NewTicker(interval)

	go func() {
		ctx := context.Background()
		log.Info("running initial session cleanup")
		if err := authService.CleanupExpiredSessions(ctx); err != nil {
			log.WithError(err).Error("initial session cleanup failed")
		} else {
			log.Info("initial session cleanup completed successfully")
		}
	}()

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			log.Info("running scheduled session cleanup")

			if err := authService.CleanupExpiredSessions(ctx); err != nil {
				log.WithError(err).Error("scheduled session cleanup failed")
			} else {
				log.Info("scheduled session cleanup completed successfully")
			}
		}
	}()
}
