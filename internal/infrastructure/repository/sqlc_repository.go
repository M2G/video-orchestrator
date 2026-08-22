package repository

import (
	"context"
	"errors"
	"fmt"

	"video-orchestrator/db"
	"video-orchestrator/internal/domain"

	"github.com/jackc/pgx/v5"
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

func (r *SQLCRepository) CreateJob(ctx context.Context, filename string) (domain.VideoJob, error) {
	row, err := r.q.CreateJob(ctx, filename)
	if err != nil {
		return domain.VideoJob{}, fmt.Errorf("create job: %w", err)
	}

	return domain.VideoJob{
		ID:       row.ID,
		Filename: row.Filename,
	}, nil
}

func (r *SQLCRepository) MarkDone(ctx context.Context, id int64) error {
	return r.q.MarkDone(ctx, id)
}

func (r *SQLCRepository) Exists(ctx context.Context, id int64) (bool, error) {
	exists, err := r.q.JobExists(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return exists, nil
}

func (r *SQLCRepository) GetJobByID(ctx context.Context, id int64) (domain.VideoJobDetails, error) {
	row, err := r.q.GetJobByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VideoJobDetails{}, fmt.Errorf("job not found: %d", id)
		}
		return domain.VideoJobDetails{}, err
	}

	return domain.VideoJobDetails{
		VideoJob: domain.VideoJob{
			ID:       row.ID,
			Filename: row.Filename,
		},
		Status: domain.JobStatus(fmt.Sprintf("%v", row.Status)),
	}, nil
}
