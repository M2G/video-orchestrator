package interfaces

import (
	"context"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"video-orchestrator/internal/domain"

	"github.com/sirupsen/logrus"
)

type Watcher struct {
	streamsDir string
	repo       domain.Repository
}

func NewWatcher(streamsDir string, repo domain.Repository) *Watcher {
	return &Watcher{
		streamsDir: streamsDir,
		repo:       repo,
	}
}

func (w *Watcher) Start(ctx context.Context, log *logrus.Logger) {
	log.Info("watcher_started")

	var mu sync.Mutex
	seen := make(map[string]struct{})

	for {
		select {
		case <-ctx.Done():
			log.Info("watcher_stopped")
			return

		default:
			pattern := filepath.Join(w.streamsDir, "*", "*.m3u8")
			matches, err := filepath.Glob(pattern)
			if err != nil {
				log.WithError(err).Error("glob_failed")
				continue
			}

			for _, match := range matches {
				mu.Lock()
				_, already := seen[match]
				if !already {
					seen[match] = struct{}{}
				}
				mu.Unlock()
				if already {
					continue
				}

				dir := filepath.Dir(match)
				videoID := filepath.Base(dir)

				id, err := strconv.ParseInt(videoID, 10, 64)
				if err != nil {
					log.WithField("dir", dir).Warn("invalid_video_id")
					continue
				}

				go func(jobID int64, path string) {
					exists, err := w.repo.Exists(ctx, jobID)
					if err != nil {
						log.WithError(err).WithField("job_id", jobID).Error("exists_check_failed")
						return
					}
					if !exists {
						log.WithField("job_id", jobID).Warn("job_not_found")
						return
					}

					if err := w.repo.MarkDone(ctx, jobID); err != nil {
						log.WithError(err).WithField("job_id", jobID).Error("mark_done_failed")
						mu.Lock()
						delete(seen, path) // retry au prochain tick
						mu.Unlock()
						return
					}

					log.WithField("job_id", jobID).Info("job_complete")
				}(id, match)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}
