package domain

import (
	"context"
	"errors"
	"time"
)

var ErrLockNotAcquired = errors.New("lock not acquired")

type Locker interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (release func(context.Context) error, err error)
}
