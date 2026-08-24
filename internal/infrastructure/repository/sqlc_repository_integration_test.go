package repository_test

import (
	"context"
	"os"
	"testing"

	"video-orchestrator/internal/infrastructure/db"
	"video-orchestrator/internal/infrastructure/repository"
)

func setupDB(t *testing.T) *repository.SQLCRepository {
	t.Helper()

	connString := os.Getenv("TEST_DATABASE_URL")
	if connString == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, connString)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return repository.New(pool)
}

func TestCreateJob(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	job, err := repo.CreateJob(ctx, "test.mp4")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job.ID == 0 {
		t.Error("expected non-zero job ID")
	}

	if job.Filename != "test.mp4" {
		t.Errorf("expected filename test.mp4, got %s", job.Filename)
	}
}

func TestExists_True(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	job, err := repo.CreateJob(ctx, "exists.mp4")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	exists, err := repo.Exists(ctx, job.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !exists {
		t.Error("expected job to exist")
	}
}

func TestExists_False(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	exists, err := repo.Exists(ctx, 999999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if exists {
		t.Error("expected job to not exist")
	}
}

func TestMarkDone(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	job, err := repo.CreateJob(ctx, "done.mp4")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := repo.MarkDone(ctx, job.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	details, err := repo.GetJobByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if details.Status != "DONE" {
		t.Errorf("expected status DONE, got %s", details.Status)
	}
}

func TestGetJobByID_NotFound(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	_, err := repo.GetJobByID(ctx, 999999)
	if err == nil {
		t.Error("expected error for non-existent job")
	}
}
