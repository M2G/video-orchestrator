package interfaces

import (
	"context"
	"log"
	"time"
)

type Orchestrator interface {
	RunOnce(ctx context.Context, limit int32)
}

type Scheduler struct {
	orch     Orchestrator
	interval time.Duration
	limit    int32
}

func NewScheduler(orch Orchestrator, interval time.Duration, limit int32) *Scheduler {
	return &Scheduler{
		orch:     orch,
		interval: interval,
		limit:    limit,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Println("scheduler started")

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler stopped")
			return

		case <-ticker.C:
			s.orch.RunOnce(ctx, s.limit)
		}
	}
}
