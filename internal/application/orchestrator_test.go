package application

import (
	"context"
	"errors"
	"testing"
	"video-orchestrator/internal/domain"
)

func TestRetryLogic(t *testing.T) {

	repo := &fakeRepo{}
	handler := &failingHandler{}

	orch := NewOrchestrator(repo, handler)

	job := domain.VideoJob{
		ID:         1,
		Filename:   "video.mp4",
		RetryCount: 4,
	}

	orch.RunOnce(context.Background(), 1)

	if !repo.failedCalled {
		t.Fatal("expected job to be marked as FAILED")
	}
}

// mocks
type fakeRepo struct {
	failedCalled bool
}

func (f *fakeRepo) LockAndMarkProcessing(ctx context.Context, limit int32) ([]domain.VideoJob, error) {
	return []domain.VideoJob{
		{ID: 1, RetryCount: 4},
	}, nil
}
func (f *fakeRepo) MarkDone(ctx context.Context, id int64)             {}
func (f *fakeRepo) MarkRetry(ctx context.Context, id int64, delay int) {}
func (f *fakeRepo) MarkFailed(ctx context.Context, id int64) {
	f.failedCalled = true
}

type failingHandler struct{}

func (f *failingHandler) Handle(ctx context.Context, job domain.VideoJob) error {
	return errors.New("fail")
}
