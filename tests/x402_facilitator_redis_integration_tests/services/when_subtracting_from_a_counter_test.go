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

type WhenSubtractingFromACounterTestingSuite struct {
	redisClient *redis.Client
	sut         *redis_services.RedisCounterStorage
}

func WhenSubtractingFromACounterBeforeEach(t *testing.T) *WhenSubtractingFromACounterTestingSuite {
	redisClient := container_builders.GetRedisContainerBuilder().Create()

	redisMetrics := metrics_mocks.NewMockRedisMetricsInterface(gomock.NewController(t))
	redisMetrics.EXPECT().RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &WhenSubtractingFromACounterTestingSuite{
		redisClient: redisClient,
		sut:         redis_services.NewRedisCounterStorage(redisClient, redisMetrics),
	}
}

func WhenSubtractingFromACounterAfterEach(suite *WhenSubtractingFromACounterTestingSuite) {
	container_builders.GetRedisContainerBuilder().DropConnection(suite.redisClient)
}

func TestWhenSubtractingFromACounter(t *testing.T) {
	t.Parallel()

	t.Run("Given a missing counter", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not create it", func(t *testing.T) {
			t.Parallel()

			suite := WhenSubtractingFromACounterBeforeEach(t)
			defer WhenSubtractingFromACounterAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())

			err := suite.sut.Subtract(context.Background(), key, 3)

			require.NoError(t, err)
			assert.Zero(t, suite.redisClient.Exists(context.Background(), key).Val())
		})
	})

	t.Run("Given an existing counter", func(t *testing.T) {
		t.Parallel()

		t.Run("Should subtract the amount and keep its expiry", func(t *testing.T) {
			t.Parallel()

			suite := WhenSubtractingFromACounterBeforeEach(t)
			defer WhenSubtractingFromACounterAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			require.NoError(t, suite.redisClient.Set(context.Background(), key, 8, time.Minute).Err())

			err := suite.sut.Subtract(context.Background(), key, 3)

			require.NoError(t, err)
			assert.Equal(t, "5", suite.redisClient.Get(context.Background(), key).Val())
			assert.Positive(t, suite.redisClient.PTTL(context.Background(), key).Val())
		})
	})
}
