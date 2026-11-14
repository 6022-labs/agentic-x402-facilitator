package services_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	redis_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/services"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/container_builders"
	metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_redis_mocks/metrics_mocks/interfaces_mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenStoringAValueTestingSuite struct {
	redisClient *redis.Client
	sut         *redis_services.RedisStorage
}

func WhenStoringAValueBeforeEach(t *testing.T) *WhenStoringAValueTestingSuite {
	redisClient := container_builders.GetRedisContainerBuilder().Create()

	redisMetrics := metrics_mocks.NewMockRedisMetricsInterface(gomock.NewController(t))
	redisMetrics.EXPECT().RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &WhenStoringAValueTestingSuite{
		redisClient: redisClient,
		sut:         redis_services.NewRedisStorage(redisClient, redisMetrics),
	}
}

func WhenStoringAValueAfterEach(suite *WhenStoringAValueTestingSuite) {
	container_builders.GetRedisContainerBuilder().DropConnection(suite.redisClient)
}

func TestWhenStoringAValue(t *testing.T) {
	t.Parallel()

	t.Run("Given a ttl", func(t *testing.T) {
		t.Parallel()

		t.Run("Should store the value with that expiry", func(t *testing.T) {
			t.Parallel()

			suite := WhenStoringAValueBeforeEach(t)
			defer WhenStoringAValueAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())

			err := suite.sut.Store(context.Background(), key, "1", time.Minute)

			require.NoError(t, err)
			assert.Equal(t, "1", suite.redisClient.Get(context.Background(), key).Val())
			ttl := suite.redisClient.TTL(context.Background(), key).Val()
			assert.True(t, ttl > 0 && ttl <= time.Minute)
		})
	})

	t.Run("Given a zero ttl", func(t *testing.T) {
		t.Parallel()

		t.Run("Should store the value without expiry", func(t *testing.T) {
			t.Parallel()

			suite := WhenStoringAValueBeforeEach(t)
			defer WhenStoringAValueAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())

			err := suite.sut.Store(context.Background(), key, "1", 0)

			require.NoError(t, err)
			assert.Equal(t, time.Duration(-1), suite.redisClient.TTL(context.Background(), key).Val())
		})
	})
}
