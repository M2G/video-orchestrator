package application

import (
	"context"
	"errors"
	"sync"
	"testing"

	"video-orchestrator/internal/domain"
)

// MOCK REPO

type mockRepo struct {
	jobs []domain.VideoJob

	doneCalled   bool
	retryCalled  bool
	failedCalled bool
	mu           sync.Mutex
}

func (m *mockRepo) LockAndMarkProcessing(ctx context.Context, limit int32) ([]domain.VideoJob, error) {
	return m.jobs, nil
}

func (m *mockRepo) MarkDone(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.doneCalled = true
	return nil
}

func (m *mockRepo) MarkRetry(ctx context.Context, id int64, delay int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retryCalled = true
	return nil
}

func (m *mockRepo) MarkFailed(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedCalled = true
	return nil
}

func (m *mockRepo) Exists(ctx context.Context, id int64) (bool, error) {
	return true, nil
}

// MOCK HANDLER

type successHandler struct{}

func (h *successHandler) Handle(ctx context.Context, job domain.VideoJob) error {
	return nil
}

type failingHandler struct{}

func (h *failingHandler) Handle(ctx context.Context, job domain.VideoJob) error {
	return errors.New("fail")
}

// TESTS

func TestRunOnce_Success(t *testing.T) {

	repo := &mockRepo{
		jobs: []domain.VideoJob{
			{ID: 1},
		},
	}

	handler := &successHandler{}

	orch := NewOrchestrator(repo, handler, 1)

	orch.RunOnce(context.Background(), 1)

	if !repo.doneCalled {
		t.Fatal("expected MarkDone to be called")
	}
}

func TestRunOnce_Retry(t *testing.T) {

	repo := &mockRepo{
		jobs: []domain.VideoJob{
			{ID: 1, RetryCount: 1},
		},
	}

	handler := &failingHandler{}

	orch := NewOrchestrator(repo, handler, 1)

	orch.RunOnce(context.Background(), 1)

	if !repo.retryCalled {
		t.Fatal("expected MarkRetry to be called")
	}
}

func TestRunOnce_Failed(t *testing.T) {

	repo := &mockRepo{
		jobs: []domain.VideoJob{
			{ID: 1, RetryCount: 5},
		},
	}

	handler := &failingHandler{}

	orch := NewOrchestrator(repo, handler, 1)

	orch.RunOnce(context.Background(), 1)

	if !repo.failedCalled {
		t.Fatal("expected MarkFailed to be called")
	}
}
