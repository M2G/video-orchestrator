package application

import (
	"context"
	"testing"
	"time"
	"video-orchestrator/internal/domain"
)

func TestConcurrentProcessing(t *testing.T) {

	repo := &fakeRepoMulti{}
	handler := &slowHandler{}

	orch := NewOrchestrator(repo, handler)

	start := time.Now()

	orch.RunOnce(context.Background(), 5)

	elapsed := time.Since(start)

	if elapsed > 3*time.Second {
		t.Fatal("expected concurrent execution, took too long")
	}
}

type fakeRepoMulti struct{}

func (f *fakeRepoMulti) LockAndMarkProcessing(ctx context.Context, limit int32) ([]domain.VideoJob, error) {

	jobs := make([]domain.VideoJob, 5)
	for i := range jobs {
		jobs[i] = domain.VideoJob{ID: int64(i)}
	}

	return jobs, nil
}

func (f *fakeRepoMulti) MarkDone(ctx context.Context, id int64)             {}
func (f *fakeRepoMulti) MarkRetry(ctx context.Context, id int64, delay int) {}
func (f *fakeRepoMulti) MarkFailed(ctx context.Context, id int64)           {}

type slowHandler struct{}

func (s *slowHandler) Handle(ctx context.Context, job domain.VideoJob) error {
	time.Sleep(500 * time.Millisecond)
	return nil
}
