package configurations

import (
	"context"

	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/metrics/interfaces"
	redis_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/services"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/settings"
	"github.com/redis/go-redis/v9"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func AddX402FacilitatorRedisConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewRedisSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewRedisMetrics,
		dig.As(new(metrics_interfaces.RedisMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		redis_services.NewRedisStorage,
		dig.As(new(x402_facilitator_interfaces.CacheStorageInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		redis_services.NewRedisLockStorage,
		dig.As(new(x402_facilitator_interfaces.LockStorageInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(func(redisSettings *settings.RedisSettings, logger *zap.Logger) *redis.Client {
		client := redis.NewClient(&redis.Options{
			Addr:     redisSettings.Host,
			Password: redisSettings.Password,
			DB:       int(redisSettings.DB),
		})

		if _, err := client.Ping(context.Background()).Result(); err != nil {
			logger.Fatal("Error connecting to Redis: ", zap.Error(err))
		}

		return client
	}); err != nil {
		panic(err)
	}
}
