package interfaces

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type fakeStorage struct {
	count int
}

func (f *fakeStorage) Upload(ctx context.Context, path, key string) error {
	f.count++
	return nil
}

type fakeRepo struct{}

func (f *fakeRepo) Exists(ctx context.Context, id int64) (bool, error) {
	return true, nil
}

func (f *fakeRepo) MarkDone(ctx context.Context, id int64) error {
	return nil
}

func TestWatcher_Handle(t *testing.T) {

	tmp := t.TempDir()

	videoDir := filepath.Join(tmp, "video")
	doneDir := filepath.Join(tmp, "done")

	os.MkdirAll(videoDir, 0755)
	os.MkdirAll(doneDir, 0755)

	os.WriteFile(filepath.Join(videoDir, "seg.ts"), []byte("data"), 0644)

	mp4 := filepath.Join(doneDir, "42.mp4")
	os.WriteFile(mp4, []byte("done"), 0644)

	storage := &fakeStorage{}
	repo := &fakeRepo{}

	w := NewWatcher(doneDir, videoDir, storage, repo)

	w.Handle(context.Background(), mp4)

	if storage.count == 0 {
		t.Fatal("expected uploads")
	}
}
