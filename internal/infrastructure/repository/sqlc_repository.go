package repository

import (
	"context"

	"video-orchestrator/db"
	"video-orchestrator/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLCRepository struct {
	q *db.Queries
}

func New(pool *pgxpool.Pool) *SQLCRepository {
	return &SQLCRepository{
		q: db.New(pool),
	}
}

func (r *SQLCRepository) LockAndMarkProcessing(ctx context.Context, limit int32) ([]domain.VideoJob, error) {

	rows, err := r.q.LockAndMarkProcessing(ctx, int64(limit))
	if err != nil {
		return nil, err
	}

	jobs := make([]domain.VideoJob, 0, len(rows))

	for _, row := range rows {
		jobs = append(jobs, domain.VideoJob{
			ID:         row.ID,
			Filename:   row.Filename,
			RetryCount: int(row.RetryCount),
		})
	}

	return jobs, nil
}

func (r *SQLCRepository) MarkDone(ctx context.Context, id int64) error {
	return r.q.MarkDone(ctx, id)
}

func (r *SQLCRepository) MarkRetry(ctx context.Context, id int64, delay int) error {
	return r.q.MarkRetry(ctx, db.MarkRetryParams{
		Column1: int32(delay),
		ID:      id,
	})
}

func (r *SQLCRepository) MarkFailed(ctx context.Context, id int64) error {
	return r.q.MarkFailed(ctx, id)
}

func (r *SQLCRepository) Exists(ctx context.Context, id int64) (bool, error) {
	row, err := r.q.GetJobByID(ctx, id)
	if err != nil {
		return false, err
	}
	return row != 0, nil
}
