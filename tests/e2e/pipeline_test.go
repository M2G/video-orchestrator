package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"video-orchestrator/internal/infrastructure/db"
	"video-orchestrator/internal/infrastructure/repository"
	"video-orchestrator/internal/infrastructure/storage"
	"video-orchestrator/internal/interfaces"
)

func TestPipelineE2E(t *testing.T) {

	ctx := context.Background()

	// =========================
	// DB
	// =========================
	pool := db.NewPool("postgresql://admin:admin@127.0.0.1:5432/postgres?sslmode=disable")
	defer pool.Close()

	_, _ = pool.Exec(ctx, `DELETE FROM video_jobs`)

	_, err := pool.Exec(ctx, `
		INSERT INTO video_jobs (id, filename, status, retry_count)
		VALUES (42, 'video.mp4', 'PENDING', 0)
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := repository.New(pool)

	// Fake S3
	storage.StartFakeS3()

	s3 := storage.NewFakeS3("videos")
	s3.CreateBucket(ctx)

	// Filesystem
	tmp := t.TempDir()

	videoDir := filepath.Join(tmp, "video")
	doneDir := filepath.Join(tmp, "done")

	os.MkdirAll(videoDir, 0755)
	os.MkdirAll(doneDir, 0755)

	// HLS files
	os.WriteFile(filepath.Join(videoDir, "seg.ts"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(videoDir, "playlist.m3u8"), []byte("data"), 0644)

	// trigger
	mp4 := filepath.Join(doneDir, "42.mp4")
	os.WriteFile(mp4, []byte("done"), 0644)

	// Watcher
	w := interfaces.NewWatcher(doneDir, videoDir, s3, repo)

	w.Handle(ctx, mp4)

	// VERIFY DB
	var status string
	err = pool.QueryRow(ctx, `SELECT status FROM video_jobs WHERE id=42`).Scan(&status)
	if err != nil {
		t.Fatal(err)
	}

	if status != "DONE" {
		t.Fatalf("expected DONE, got %s", status)
	}
}
