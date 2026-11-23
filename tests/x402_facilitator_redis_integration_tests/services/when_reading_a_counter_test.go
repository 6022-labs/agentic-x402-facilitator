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

type WhenReadingACounterTestingSuite struct {
	redisClient *redis.Client
	sut         *redis_services.RedisCounterStorage
}

func WhenReadingACounterBeforeEach(t *testing.T) *WhenReadingACounterTestingSuite {
	redisClient := container_builders.GetRedisContainerBuilder().Create()

	redisMetrics := metrics_mocks.NewMockRedisMetricsInterface(gomock.NewController(t))
	redisMetrics.EXPECT().RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &WhenReadingACounterTestingSuite{
		redisClient: redisClient,
		sut:         redis_services.NewRedisCounterStorage(redisClient, redisMetrics),
	}
}

func WhenReadingACounterAfterEach(suite *WhenReadingACounterTestingSuite) {
	container_builders.GetRedisContainerBuilder().DropConnection(suite.redisClient)
}

func TestWhenReadingACounter(t *testing.T) {
	t.Parallel()

	t.Run("Given a missing counter", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return zero", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingACounterBeforeEach(t)
			defer WhenReadingACounterAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())

			value, err := suite.sut.Read(context.Background(), key)

			require.NoError(t, err)
			assert.Zero(t, value)
		})
	})

	t.Run("Given an existing counter", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its value", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingACounterBeforeEach(t)
			defer WhenReadingACounterAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			require.NoError(t, suite.redisClient.Set(context.Background(), key, 8, time.Minute).Err())

			value, err := suite.sut.Read(context.Background(), key)

			require.NoError(t, err)
			assert.Equal(t, int64(8), value)
		})
	})
}
