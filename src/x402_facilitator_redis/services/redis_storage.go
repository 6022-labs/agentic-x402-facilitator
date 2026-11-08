package services

import (
	"context"
	"errors"
	"time"

	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/metrics/interfaces"
	"github.com/redis/go-redis/v9"
)

// RedisStorage is a thin, atomic adapter over the go-redis client.
type RedisStorage struct {
	redisClient  *redis.Client
	redisMetrics metrics_interfaces.RedisMetricsInterface
}

func NewRedisStorage(
	redisClient *redis.Client,
	redisMetrics metrics_interfaces.RedisMetricsInterface,
) *RedisStorage {
	return &RedisStorage{redisClient: redisClient, redisMetrics: redisMetrics}
}

func (rs *RedisStorage) Store(ctx context.Context, key string, value string, ttl time.Duration) error {
	start := time.Now()
	err := rs.redisClient.Set(ctx, key, value, ttl).Err()
	rs.redisMetrics.RecordOperation(ctx, "SET", time.Since(start), err)
	return err
}

func (rs *RedisStorage) ReadString(ctx context.Context, key string) (*string, error) {
	start := time.Now()
	value, err := rs.redisClient.Get(ctx, key).Result()
	rs.redisMetrics.RecordOperation(ctx, "GET", time.Since(start), rs.operationError(err))
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (rs *RedisStorage) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := rs.redisClient.Del(ctx, key).Err()
	rs.redisMetrics.RecordOperation(ctx, "DEL", time.Since(start), err)
	return err
}

// operationError treats a missing key as a miss rather than an operational failure.
func (rs *RedisStorage) operationError(err error) error {
	if errors.Is(err, redis.Nil) {
		return nil
	}
	return err
}
