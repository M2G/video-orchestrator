package domain

import "context"

type JobStatus string

const (
	StatusPending JobStatus = "PENDING"
	StatusDone    JobStatus = "DONE"
)

type VideoJobDetails struct {
	VideoJob
	Status JobStatus
}

type Repository interface {
	CreateJob(ctx context.Context, filename string) (VideoJob, error)
	MarkDone(ctx context.Context, id int64) error
	Exists(ctx context.Context, id int64) (bool, error)
	GetJobByID(ctx context.Context, id int64) (VideoJobDetails, error)
}
