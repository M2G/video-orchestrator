package interfaces

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"video-orchestrator/internal/domain"

	"github.com/sirupsen/logrus"
)

type JobPurger interface {
	DeleteDoneOlderThan(ctx context.Context, days int) ([]int64, error)
}

type Cleaner struct {
	repo          JobPurger
	locker        domain.Locker
	streamsDir    string
	interval      time.Duration
	retentionDays int
}

func NewCleaner(repo JobPurger, locker domain.Locker, streamsDir string, interval time.Duration, retentionDays int) *Cleaner {
	return &Cleaner{
		repo:          repo,
		locker:        locker,
		streamsDir:    streamsDir,
		interval:      interval,
		retentionDays: retentionDays,
	}
}

func (c *Cleaner) Start(ctx context.Context, log *logrus.Logger) {
	log.Info("cleaner_started")

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("cleaner_stopped")
			return
		case <-ticker.C:
			c.runOnce(ctx, log)
		}
	}
}

func (c *Cleaner) runOnce(ctx context.Context, log *logrus.Logger) {
	release, err := c.locker.Acquire(ctx, "lock:cleanup", 5*time.Minute)
	if errors.Is(err, domain.ErrLockNotAcquired) {
		log.Debug("cleanup_skipped_lock_held")
		return
	}
	if err != nil {
		log.WithError(err).Error("cleanup_lock_failed")
		return
	}
	defer func() {
		if err := release(context.WithoutCancel(ctx)); err != nil {
			log.WithError(err).Warn("cleanup_unlock_failed")
		}
	}()

	ids, err := c.repo.DeleteDoneOlderThan(ctx, c.retentionDays)
	if err != nil {
		log.WithError(err).Error("cleanup_db_failed")
		return
	}

	for _, id := range ids {
		dir := filepath.Join(c.streamsDir, strconv.FormatInt(id, 10))
		if err := os.RemoveAll(dir); err != nil {
			log.WithError(err).WithField("dir", dir).Warn("cleanup_remove_failed")
		}
	}

	log.WithField("deleted", len(ids)).Info("cleanup_done")
}
