package services

import (
	"context"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/metrics/interfaces"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisLockStorage stores each lock's owner token; its script stops a former owner from freeing it.
type RedisLockStorage struct {
	redisClient  *redis.Client
	redisMetrics metrics_interfaces.RedisMetricsInterface
	unlockScript *redis.Script
}

func NewRedisLockStorage(
	redisClient *redis.Client,
	redisMetrics metrics_interfaces.RedisMetricsInterface,
) *RedisLockStorage {
	return &RedisLockStorage{
		redisClient:  redisClient,
		redisMetrics: redisMetrics,
		unlockScript: redis.NewScript(
			`if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) end return 0`,
		),
	}
}

func (s *RedisLockStorage) TryLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	token := uuid.NewString()

	start := time.Now()
	acquired, err := s.redisClient.SetNX(ctx, key, token, ttl).Result()
	s.redisMetrics.RecordOperation(ctx, "SET", time.Since(start), err)
	if err != nil || !acquired {
		return "", false, err
	}
	return token, true, nil
}

func (s *RedisLockStorage) Unlock(ctx context.Context, key string, token string) error {
	start := time.Now()
	owned, err := s.unlockScript.Run(ctx, s.redisClient, []string{key}, token).Int()
	s.redisMetrics.RecordOperation(ctx, "EVALSHA", time.Since(start), err)
	return s.ownership(owned, err)
}

// ownership reads a script's 0 as a lock the token no longer owns.
func (s *RedisLockStorage) ownership(owned int, err error) error {
	if err != nil {
		return err
	}
	if owned == 0 {
		return apperrors.ErrLockNotOwned
	}
	return nil
}
