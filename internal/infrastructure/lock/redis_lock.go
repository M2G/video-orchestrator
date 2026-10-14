package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"video-orchestrator/internal/domain"

	"github.com/redis/go-redis/v9"
)

var releaseScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
end
return 0
`)

type RedisLocker struct {
	client *redis.Client
}

func NewRedisLocker(client *redis.Client) *RedisLocker {
	return &RedisLocker{client: client}
}

func (l *RedisLocker) Acquire(ctx context.Context, key string, ttl time.Duration) (func(context.Context) error, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("generate lock token: %w", err)
	}
	token := hex.EncodeToString(buf)

	ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("acquire lock %s: %w", key, err)
	}
	if !ok {
		return nil, domain.ErrLockNotAcquired
	}

	return func(ctx context.Context) error {
		return releaseScript.Run(ctx, l.client, []string{key}, token).Err()
	}, nil
}
