package interfaces_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"video-orchestrator/internal/domain"
	"video-orchestrator/internal/interfaces"
)

type fakeLocker struct {
	err      error
	released chan struct{}
}

func (f *fakeLocker) Acquire(context.Context, string, time.Duration) (func(context.Context) error, error) {
	if f.err != nil {
		return nil, f.err
	}
	return func(context.Context) error {
		select {
		case f.released <- struct{}{}:
		default:
		}
		return nil
	}, nil
}

type fakePurger struct {
	ids   []int64
	err   error
	calls atomic.Int32
}

func (f *fakePurger) DeleteDoneOlderThan(context.Context, int) ([]int64, error) {
	f.calls.Add(1)
	return f.ids, f.err
}

func TestCleaner_RemovesHLSDirsOfPurgedJobs(t *testing.T) {
	streamsDir := t.TempDir()
	for _, id := range []string{"1", "2", "3"} {
		if err := os.MkdirAll(filepath.Join(streamsDir, id), 0755); err != nil {
			t.Fatal(err)
		}
	}

	locker := &fakeLocker{released: make(chan struct{}, 1)}
	purger := &fakePurger{ids: []int64{1, 2}}
	cleaner := interfaces.NewCleaner(purger, locker, streamsDir, 10*time.Millisecond, 7)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go cleaner.Start(ctx, noopLogger())

	select {
	case <-locker.released:
	case <-ctx.Done():
		t.Fatal("timeout: cleanup did not run")
	}

	for _, id := range []string{"1", "2"} {
		if _, err := os.Stat(filepath.Join(streamsDir, id)); !os.IsNotExist(err) {
			t.Errorf("expected directory %s to be removed", id)
		}
	}
	if _, err := os.Stat(filepath.Join(streamsDir, "3")); err != nil {
		t.Errorf("directory 3 must be kept: %v", err)
	}
}

func TestCleaner_SkipsWhenLockHeld(t *testing.T) {
	locker := &fakeLocker{err: domain.ErrLockNotAcquired, released: make(chan struct{}, 1)}
	purger := &fakePurger{}
	cleaner := interfaces.NewCleaner(purger, locker, t.TempDir(), 10*time.Millisecond, 7)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	cleaner.Start(ctx, noopLogger()) // bloque jusqu'à l'expiration du contexte

	if got := purger.calls.Load(); got != 0 {
		t.Errorf("purge must not run when the lock is held, got %d calls", got)
	}
}

func TestCleaner_KeepsDirsWhenPurgeFails(t *testing.T) {
	streamsDir := t.TempDir()
	dir := filepath.Join(streamsDir, "1")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	locker := &fakeLocker{released: make(chan struct{}, 1)}
	purger := &fakePurger{ids: []int64{1}, err: errors.New("db down")}
	cleaner := interfaces.NewCleaner(purger, locker, streamsDir, 10*time.Millisecond, 7)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go cleaner.Start(ctx, noopLogger())

	select {
	case <-locker.released: // le verrou est libéré même en cas d'erreur
	case <-ctx.Done():
		t.Fatal("timeout: cleanup did not run")
	}

	if _, err := os.Stat(dir); err != nil {
		t.Errorf("directory must be kept when the purge fails: %v", err)
	}
}
