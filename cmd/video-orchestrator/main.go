package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"video-orchestrator/internal/application"
	"video-orchestrator/internal/infrastructure/db"
	"video-orchestrator/internal/infrastructure/repository"
	"video-orchestrator/internal/infrastructure/storage"
	"video-orchestrator/internal/interfaces"

	logrus "video-orchestrator/internal/infrastructure/logger"

	"github.com/urfave/cli/v3"
)

func handleShutdown(cancel context.CancelFunc) {

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-sigChan

	log.Println("shutdown signal received")
	cancel()
}

func Run(context.Context, *cli.Command) error {

	log := logrus.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleShutdown(cancel)

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		"admin", "admin", "postgres", "5432", "postgres",
	)

	pool := db.NewPool(dsn)
	defer pool.Close()

	repo := repository.New(pool)

	storage.StartFakeS3()

	s3 := storage.NewFakeS3("videos")
	s3.CreateBucket(ctx)

	handler := interfaces.NewVideoHandler()

	orchestrator := application.NewOrchestrator(repo, handler, 5, log)

	watcher := interfaces.NewWatcher(
		"cmd/video-orchestrator/tmp/videos/done",
		"cmd/video-orchestrator/var/www/html/streams/video",
		s3,
		repo,
		log,
	)

	go watcher.Start(ctx)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	log.Info("application started")

	for {
		select {
		case <-ctx.Done():
			log.Info("shutting down")
			return nil

		case <-ticker.C:
			orchestrator.RunOnce(ctx, 10)
		}
	}
}

func main() {
	cmd := &cli.Command{
		Name:    "boom",
		Version: "v1.0.0",
		Usage:   "Video orchestrator",
		Action:  Run,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

	return
}
