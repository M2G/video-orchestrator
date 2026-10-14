package lock_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"video-orchestrator/internal/domain"
	"video-orchestrator/internal/infrastructure/lock"

	"github.com/redis/go-redis/v9"
)

func newLocker(t *testing.T) *lock.RedisLocker {
	t.Helper()

	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR not set, skipping redis tests")
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { client.Close() })

	return lock.NewRedisLocker(client)
}

func uniqueKey(t *testing.T) string {
	return fmt.Sprintf("test:lock:%s:%d", t.Name(), time.Now().UnixNano())
}

func TestAcquire_SecondCallFailsUntilReleased(t *testing.T) {
	locker := newLocker(t)
	ctx := context.Background()
	key := uniqueKey(t)

	release, err := locker.Acquire(ctx, key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := locker.Acquire(ctx, key, time.Minute); !errors.Is(err, domain.ErrLockNotAcquired) {
		t.Fatalf("expected ErrLockNotAcquired, got %v", err)
	}

	if err := release(ctx); err != nil {
		t.Fatal(err)
	}

	release2, err := locker.Acquire(ctx, key, time.Minute)
	if err != nil {
		t.Fatalf("expected the lock to be free after release, got %v", err)
	}
	_ = release2(ctx)
}

func TestRelease_DoesNotDeleteLockOfAnotherOwner(t *testing.T) {
	locker := newLocker(t)
	ctx := context.Background()
	key := uniqueKey(t)

	staleRelease, err := locker.Acquire(ctx, key, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(150 * time.Millisecond) // le TTL expire

	freshRelease, err := locker.Acquire(ctx, key, time.Minute) // un autre propriétaire prend le verrou
	if err != nil {
		t.Fatal(err)
	}
	defer freshRelease(ctx)

	// l'ancien propriétaire libère : le verrou du nouveau doit rester en place
	if err := staleRelease(ctx); err != nil {
		t.Fatal(err)
	}

	if _, err := locker.Acquire(ctx, key, time.Minute); !errors.Is(err, domain.ErrLockNotAcquired) {
		t.Fatalf("the new owner's lock must survive a stale release, got %v", err)
	}
}
