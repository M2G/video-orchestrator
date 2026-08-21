package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"video-orchestrator/internal/infrastructure/logger"
	"video-orchestrator/internal/infrastructure/repository"
	"video-orchestrator/internal/interfaces"

	"video-orchestrator/internal/infrastructure/db"

	"github.com/sirupsen/logrus"
)

func main() {
	log := logger.New(logLevel())

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	connString := requireEnv("DATABASE_URL", log)
	uploadDir := requireEnv("UPLOAD_DIR", log)
	streamsDir := requireEnv("STREAMS_DIR", log)
	addr := getEnv("HTTP_ADDR", ":8181")

	pool, err := db.NewPool(ctx, connString)
	if err != nil {
		log.WithError(err).Fatal("database init failed")
	}
	defer pool.Close()

	repo := repository.New(pool)

	watcher := interfaces.NewWatcher(streamsDir, repo)
	go watcher.Start(ctx, log)

	httpServer := interfaces.NewHTTPServer(repo, uploadDir, streamsDir, log)

	go func() {
		<-ctx.Done()
		log.Info("shutting_down")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := httpServer.Stop(shutdownCtx); err != nil {
			log.WithError(err).Error("http_shutdown_error")
		}

		log.Info("shutdown_complete")
	}()

	log.Info("video_orchestrator_started")

	if err := httpServer.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.WithError(err).Fatal("http_server_error")
	}
}

func requireEnv(key string, log *logrus.Logger) string {
	val := os.Getenv(key)
	if val == "" {
		log.WithField("key", key).Fatal("missing required env var")
	}
	return val
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func logLevel() logrus.Level {
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		return logrus.InfoLevel
	}
	return level
}
