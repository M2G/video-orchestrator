package interfaces_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
	"video-orchestrator/internal/domain"

	"video-orchestrator/internal/interfaces"
)

type mockRepoWatcher struct {
	existsFn     func(ctx context.Context, id int64) (bool, error)
	markDoneFn   func(ctx context.Context, id int64) error
	createJobFn  func(ctx context.Context, filename string) (domain.VideoJob, error)
	getJobByIDFn func(ctx context.Context, id int64) (domain.VideoJobDetails, error)
}

func (m *mockRepoWatcher) Exists(ctx context.Context, id int64) (bool, error) {
	return m.existsFn(ctx, id)
}

func (m *mockRepoWatcher) MarkDone(ctx context.Context, id int64) error {
	return m.markDoneFn(ctx, id)
}

func (m *mockRepoWatcher) CreateJob(ctx context.Context, filename string) (domain.VideoJob, error) {
	return domain.VideoJob{}, nil
}

func (m *mockRepoWatcher) GetJobByID(ctx context.Context, id int64) (domain.VideoJobDetails, error) {
	return domain.VideoJobDetails{}, nil
}

func TestWatcher_DetectsM3U8AndMarksDone(t *testing.T) {
	streamsDir := t.TempDir()
	videoDir := filepath.Join(streamsDir, "42")
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		t.Fatal(err)
	}

	marked := make(chan int64, 1)

	repo := &mockRepoWatcher{
		existsFn: func(_ context.Context, id int64) (bool, error) {
			return true, nil
		},
		markDoneFn: func(_ context.Context, id int64) error {
			marked <- id
			return nil
		},
	}

	watcher := interfaces.NewWatcher(streamsDir, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go watcher.Start(ctx, noopLogger())

	// simule la fin du traitement par video_segmenter
	m3u8Path := filepath.Join(videoDir, "index.m3u8")
	if err := os.WriteFile(m3u8Path, []byte("#EXTM3U"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case id := <-marked:
		if id != 42 {
			t.Errorf("expected job id 42, got %d", id)
		}
	case <-ctx.Done():
		t.Fatal("timeout: watcher did not detect index.m3u8")
	}
}

func TestWatcher_IgnoresInvalidVideoID(t *testing.T) {
	streamsDir := t.TempDir()
	videoDir := filepath.Join(streamsDir, "not-a-number")
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		t.Fatal(err)
	}

	marked := make(chan int64, 1)

	repo := &mockRepoWatcher{
		existsFn:   func(_ context.Context, id int64) (bool, error) { return true, nil },
		markDoneFn: func(_ context.Context, id int64) error { marked <- id; return nil },
	}

	watcher := interfaces.NewWatcher(streamsDir, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go watcher.Start(ctx, noopLogger())

	m3u8Path := filepath.Join(videoDir, "index.m3u8")
	if err := os.WriteFile(m3u8Path, []byte("#EXTM3U"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case id := <-marked:
		t.Errorf("expected no markDone call, got id %d", id)
	case <-ctx.Done():
	}
}

func TestWatcher_IgnoresAlreadySeen(t *testing.T) {
	streamsDir := t.TempDir()
	videoDir := filepath.Join(streamsDir, "99")
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		t.Fatal(err)
	}

	m3u8Path := filepath.Join(videoDir, "index.m3u8")
	if err := os.WriteFile(m3u8Path, []byte("#EXTM3U"), 0644); err != nil {
		t.Fatal(err)
	}

	count := 0
	repo := &mockRepoWatcher{
		existsFn: func(_ context.Context, id int64) (bool, error) { return true, nil },
		markDoneFn: func(_ context.Context, id int64) error {
			count++
			return nil
		},
	}

	watcher := interfaces.NewWatcher(streamsDir, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go watcher.Start(ctx, noopLogger())

	time.Sleep(3 * time.Second)

	if count > 1 {
		t.Errorf("expected markDone called once, got %d", count)
	}
}

func TestWatcher_JobNotFound(t *testing.T) {
	streamsDir := t.TempDir()
	videoDir := filepath.Join(streamsDir, "77")
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		t.Fatal(err)
	}

	marked := make(chan int64, 1)

	repo := &mockRepoWatcher{
		existsFn:   func(_ context.Context, id int64) (bool, error) { return false, nil },
		markDoneFn: func(_ context.Context, id int64) error { marked <- id; return nil },
	}

	watcher := interfaces.NewWatcher(streamsDir, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go watcher.Start(ctx, noopLogger())

	m3u8Path := filepath.Join(videoDir, "index.m3u8")
	if err := os.WriteFile(m3u8Path, []byte("#EXTM3U"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case id := <-marked:
		t.Errorf("expected no markDone call for non-existent job, got id %d", id)
	case <-ctx.Done():
	}
}
