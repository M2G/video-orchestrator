package interfaces

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type fakeStorage struct {
	uploaded []string
}

func (f *fakeStorage) Upload(ctx context.Context, path, key string) error {
	f.uploaded = append(f.uploaded, key)
	return nil
}

func TestWatcherUploadsFiles(t *testing.T) {

	tmpDir := t.TempDir()

	videoDir := filepath.Join(tmpDir, "video")
	doneDir := filepath.Join(tmpDir, "done")

	os.MkdirAll(videoDir, 0755)
	os.MkdirAll(doneDir, 0755)

	// fichiers HLS
	os.WriteFile(filepath.Join(videoDir, "seg1.ts"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(videoDir, "playlist.m3u8"), []byte("data"), 0644)

	// trigger
	os.WriteFile(filepath.Join(doneDir, "video.mp4"), []byte("done"), 0644)

	storage := &fakeStorage{}

	w := NewWatcher(doneDir, videoDir, storage)

	// appeler directement handle
	w.Handle(filepath.Join(doneDir, "video.mp4"))

	if len(storage.uploaded) == 0 {
		t.Fatal("expected uploads, got none")
	}
}
