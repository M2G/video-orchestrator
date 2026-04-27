package application

import (
	"context"
	"log"
	"sync"
	"time"

	"video-orchestrator/internal/domain"
)

type Repository interface {
	LockAndMarkProcessing(ctx context.Context, limit int32) ([]domain.VideoJob, error)
	MarkDone(ctx context.Context, id int64) error
	MarkRetry(ctx context.Context, id int64, delay int) error
	MarkFailed(ctx context.Context, id int64) error
}

type Handler interface {
	Handle(ctx context.Context, job domain.VideoJob) error
}

type Orchestrator struct {
	repo    Repository
	handler Handler

	workers int
	breaker *CircuitBreaker
}

func NewOrchestrator(repo Repository, handler Handler, workers int) *Orchestrator {
	return &Orchestrator{
		repo:    repo,
		handler: handler,
		workers: workers,
		breaker: NewCircuitBreaker(5, 10*time.Second),
	}
}

func (o *Orchestrator) handleFailure(ctx context.Context, job domain.VideoJob) {

	maxRetries := 5

	if job.RetryCount+1 >= maxRetries {
		if err := o.repo.MarkFailed(ctx, job.ID); err != nil {
			log.Println("failed to mark failed:", err)
		}
		return
	}

	delay := domain.NextDelay(job.RetryCount)

	if err := o.repo.MarkRetry(ctx, job.ID, delay); err != nil {
		log.Println("failed to mark retry:", err)
	}
}

func (o *Orchestrator) RunOnce(ctx context.Context, limit int32) {

	jobs, err := o.repo.LockAndMarkProcessing(ctx, limit)
	if err != nil {
		log.Println(`{"event":"lock_failed"}`)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, o.workers) // worker pool

	for _, job := range jobs {

		sem <- struct{}{}
		wg.Add(1)

		go func(j domain.VideoJob) {
			defer wg.Done()
			defer func() { <-sem }()

			start := time.Now()

			// circuit breaker
			if !o.breaker.Allow() {
				log.Println(`{"event":"breaker_open"}`)
				return
			}

			err := o.handler.Handle(ctx, j)

			if err != nil {
				o.breaker.Fail()
				o.handleFailure(ctx, j)
				return
			}

			o.breaker.Success()

			if err := o.repo.MarkDone(ctx, j.ID); err != nil {
				log.Println("mark_done_error:", err)
			}

			log.Printf(`{"event":"job_done","job_id":%d,"duration_ms":%d}`,
				j.ID,
				time.Since(start).Milliseconds(),
			)

		}(job)
	}

	wg.Wait()
}
