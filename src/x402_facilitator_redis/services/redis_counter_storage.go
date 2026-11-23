package services

import (
	"context"
	"errors"
	"time"

	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/metrics/interfaces"
	"github.com/redis/go-redis/v9"
)

// RedisCounterStorage runs each change as a script, so replicas never pass a limit together.
type RedisCounterStorage struct {
	redisClient    *redis.Client
	redisMetrics   metrics_interfaces.RedisMetricsInterface
	tryAddScript   *redis.Script
	subtractScript *redis.Script
}

func NewRedisCounterStorage(
	redisClient *redis.Client,
	redisMetrics metrics_interfaces.RedisMetricsInterface,
) *RedisCounterStorage {
	return &RedisCounterStorage{
		redisClient:  redisClient,
		redisMetrics: redisMetrics,
		tryAddScript: redis.NewScript(`
local current = redis.call("GET", KEYS[1])
if tonumber(current or "0") + tonumber(ARGV[1]) > tonumber(ARGV[2]) then return 0 end
redis.call("INCRBY", KEYS[1], ARGV[1])
if not current then redis.call("PEXPIRE", KEYS[1], ARGV[3]) end
return 1`),
		subtractScript: redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then redis.call("DECRBY", KEYS[1], ARGV[1]) end
return 1`),
	}
}

func (s *RedisCounterStorage) TryAdd(
	ctx context.Context,
	key string,
	amount int64,
	limit int64,
	ttl time.Duration,
) (bool, error) {
	start := time.Now()
	added, err := s.tryAddScript.Run(ctx, s.redisClient, []string{key}, amount, limit, ttl.Milliseconds()).Int()
	s.redisMetrics.RecordOperation(ctx, "EVALSHA", time.Since(start), err)
	return added == 1, err
}

func (s *RedisCounterStorage) Subtract(ctx context.Context, key string, amount int64) error {
	start := time.Now()
	err := s.subtractScript.Run(ctx, s.redisClient, []string{key}, amount).Err()
	s.redisMetrics.RecordOperation(ctx, "EVALSHA", time.Since(start), err)
	return err
}

func (s *RedisCounterStorage) Read(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	value, err := s.redisClient.Get(ctx, key).Int64()
	// A missing counter is zero, not a failure.
	if errors.Is(err, redis.Nil) {
		value, err = 0, nil
	}
	s.redisMetrics.RecordOperation(ctx, "GET", time.Since(start), err)
	return value, err
}
