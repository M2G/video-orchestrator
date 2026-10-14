package interfaces_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
	"time"
	"video-orchestrator/internal/interfaces"

	"video-orchestrator/internal/domain"

	"github.com/sirupsen/logrus"
)

type mockRepoHTTP struct {
	createJobFn  func(ctx context.Context, filename string) (domain.VideoJob, error)
	getJobByIDFn func(ctx context.Context, id int64) (domain.VideoJobDetails, error)
	existsFn     func(ctx context.Context, id int64) (bool, error)
	markDoneFn   func(ctx context.Context, id int64) error
}

func (m *mockRepoHTTP) CreateJob(ctx context.Context, filename string) (domain.VideoJob, error) {
	return m.createJobFn(ctx, filename)
}

func (m *mockRepoHTTP) GetJobByID(ctx context.Context, id int64) (domain.VideoJobDetails, error) {
	return m.getJobByIDFn(ctx, id)
}

func (m *mockRepoHTTP) Exists(ctx context.Context, id int64) (bool, error) {
	return m.existsFn(ctx, id)
}

func (m *mockRepoHTTP) MarkDone(ctx context.Context, id int64) error {
	return m.markDoneFn(ctx, id)
}

func defaultRepoHTTP() *mockRepoHTTP {
	return &mockRepoHTTP{
		createJobFn: func(_ context.Context, filename string) (domain.VideoJob, error) {
			return domain.VideoJob{ID: 1, Filename: filename}, nil
		},
		getJobByIDFn: func(_ context.Context, id int64) (domain.VideoJobDetails, error) {
			return domain.VideoJobDetails{
				VideoJob: domain.VideoJob{ID: id, Filename: "test.mp4"},
				Status:   domain.StatusPending,
			}, nil
		},
		existsFn:   func(_ context.Context, id int64) (bool, error) { return true, nil },
		markDoneFn: func(_ context.Context, id int64) error { return nil },
	}
}

func multipartBody(t *testing.T, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(content)
	writer.Close()
	return body, writer.FormDataContentType()
}

func noopLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

func TestWatcher_DetectsJobNamedM3U8(t *testing.T) {
	streamsDir := t.TempDir()
	videoDir := filepath.Join(streamsDir, "1")
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		t.Fatal(err)
	}

	marked := make(chan int64, 1)
	repo := defaultRepoHTTP()
	repo.markDoneFn = func(_ context.Context, id int64) error {
		marked <- id
		return nil
	}

	watcher := interfaces.NewWatcher(streamsDir, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go watcher.Start(ctx, noopLogger())

	if err := os.WriteFile(filepath.Join(videoDir, "1.m3u8"), []byte("#EXTM3U"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case id := <-marked:
		if id != 1 {
			t.Errorf("expected job id 1, got %d", id)
		}
	case <-ctx.Done():
		t.Fatal("timeout: watcher did not detect 1.m3u8")
	}
}
