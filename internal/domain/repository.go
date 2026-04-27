package domain

import "context"

type Repository interface {
	LockAndMarkProcessing(ctx context.Context, limit int32) ([]VideoJob, error)
	MarkDone(ctx context.Context, id int64) error
	MarkRetry(ctx context.Context, id int64, delay int) error
	MarkFailed(ctx context.Context, id int64) error
	Exists(ctx context.Context, id int64) (bool, error)
}
