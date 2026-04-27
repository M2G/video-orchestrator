package interfaces

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Storage interface {
	Upload(ctx context.Context, path string, key string) error
}

type JobRepository interface {
	Exists(ctx context.Context, id int64) (bool, error)
	MarkDone(ctx context.Context, id int64) error
}

type Watcher struct {
	doneDir  string
	videoDir string
	storage  Storage
	repo     JobRepository
}

func NewWatcher(doneDir, videoDir string, storage Storage, repo JobRepository) *Watcher {
	return &Watcher{
		doneDir:  doneDir,
		videoDir: videoDir,
		storage:  storage,
		repo:     repo,
	}
}

func (w *Watcher) Start(ctx context.Context) {

	log.Println("watcher started")

	seen := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			log.Println("watcher stopped")
			return

		default:
			files, _ := filepath.Glob(filepath.Join(w.doneDir, "*.mp4"))

			for _, f := range files {

				if seen[f] {
					continue
				}

				go w.Handle(ctx, f)
				seen[f] = true
			}

			time.Sleep(2 * time.Second)
		}
	}
}

func (w *Watcher) Handle(ctx context.Context, mp4Path string) {

	videoID := strings.TrimSuffix(filepath.Base(mp4Path), ".mp4")

	id, err := strconv.ParseInt(videoID, 10, 64)
	if err != nil {
		log.Println("invalid video ID:", videoID)
		return
	}

	exists, err := w.repo.Exists(ctx, id)
	if err != nil || !exists {
		log.Println("job not found:", id)
		return
	}

	files, _ := filepath.Glob(filepath.Join(w.videoDir, "*"))

	for _, f := range files {

		ext := filepath.Ext(f)
		if ext != ".ts" && ext != ".m3u8" && ext != ".txt" {
			continue
		}

		key := fmt.Sprintf("%s/%s", videoID, filepath.Base(f))

		err := w.storage.Upload(ctx, f, key)
		if err != nil {
			log.Println("upload error:", err)
			return
		}
	}

	mp4Key := fmt.Sprintf("%s/video.mp4", videoID)

	err = w.storage.Upload(ctx, mp4Path, mp4Key)
	if err != nil {
		log.Println("upload mp4 error:", err)
		return
	}

	err = w.repo.MarkDone(ctx, id)
	if err != nil {
		log.Println("failed to mark job done:", err)
		return
	}

	log.Println("job completed:", id)
}
