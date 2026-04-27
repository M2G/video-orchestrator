package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "fmt"
	"log"
	_ "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"video-orchestrator/db"
	"video-orchestrator/internal/application"
	"video-orchestrator/internal/infrastructure/repository"
	"video-orchestrator/internal/infrastructure/storage"
	"video-orchestrator/internal/interfaces"

	"github.com/urfave/cli/v3"

	_ "github.com/lib/pq"
)

func Run(context.Context, *cli.Command) error {

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		"admin", "admin", "postgres", "5432", "postgres",
	)

	// DB
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("db open:", err)
	}
	defer conn.Close()

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(time.Hour)

	queries := db.New(conn)

	repo := repository.New(queries)

	handler := interfaces.DefaultHandler{}

	orchestrator := application.NewOrchestrator(repo, handler, 5)

	scheduler := interfaces.NewScheduler(
		orchestrator,
		1*time.Second, // interval
		10,            // batch size
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage.StartFakeS3()

	s3 := storage.NewFakeS3("videos")
	s3.CreateBucket(ctx)

	watcher := interfaces.NewWatcher(
		"cmd/video-orchestrator/tmp/videos/done",
		"cmd/video-orchestrator/var/www/html/streams/video",
		s3,
		repo,
	)

	go watcher.Start(ctx)

	go scheduler.Start(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	log.Println("orchestrator started")

	<-sig

	log.Println("shutdown signal received")
	cancel()

	time.Sleep(2 * time.Second)

	log.Println("shutdown complete")

	return nil
}

func main() {
	cmd := &cli.Command{
		Name:    "boom",
		Version: "v1.0.0",
		Usage:   "Golang init",
		Action:  Run,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

	return
}
