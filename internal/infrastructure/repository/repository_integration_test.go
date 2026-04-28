package repository

import (
	"context"
	"testing"

	"video-orchestrator/internal/infrastructure/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {

	conn := "postgresql://admin:admin@127.0.0.1:5432/postgres?sslmode=disable"

	pool := db.NewPool(conn)

	// clean table
	_, err := pool.Exec(context.Background(), `DELETE FROM video_jobs`)
	if err != nil {
		t.Fatal(err)
	}

	return pool
}

func TestRepository_MarkDone(t *testing.T) {

	ctx := context.Background()
	pool := setupTestDB(t)
	defer pool.Close()

	repo := New(pool)

	// insert job
	_, err := pool.Exec(ctx, `
		INSERT INTO video_jobs (id, filename, status, retry_count)
		VALUES (42, 'video.mp4', 'PROCESSING', 0)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// call
	err = repo.MarkDone(ctx, 42)
	if err != nil {
		t.Fatal(err)
	}

	// verify
	var status string
	err = pool.QueryRow(ctx, `SELECT status FROM video_jobs WHERE id=42`).Scan(&status)
	if err != nil {
		t.Fatal(err)
	}

	if status != "DONE" {
		t.Fatalf("expected DONE, got %s", status)
	}
}
