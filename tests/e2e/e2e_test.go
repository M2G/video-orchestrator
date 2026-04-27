package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"video-orchestrator/internal/infrastructure/storage"
	"video-orchestrator/internal/interfaces"
)

func TestPipelineE2E(t *testing.T) {

	storage.StartFakeS3()

	ctx := context.Background()

	s := storage.NewFakeS3("videos")
	s.CreateBucket(ctx)

	tmpDir := t.TempDir()

	videoDir := filepath.Join(tmpDir, "video")
	doneDir := filepath.Join(tmpDir, "done")

	os.MkdirAll(videoDir, 0755)
	os.MkdirAll(doneDir, 0755)

	// créer HLS
	os.WriteFile(filepath.Join(videoDir, "seg.ts"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(videoDir, "playlist.m3u8"), []byte("data"), 0644)

	// trigger
	mp4 := filepath.Join(doneDir, "video.mp4")
	os.WriteFile(mp4, []byte("done"), 0644)

	w := interfaces.NewWatcher(doneDir, videoDir, s)

	w.Handle(mp4)

	// pas d'erreur = OK
}
