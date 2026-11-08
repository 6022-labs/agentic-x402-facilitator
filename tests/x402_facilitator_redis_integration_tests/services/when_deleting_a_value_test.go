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

type WhenDeletingAValueTestingSuite struct {
	redisClient *redis.Client
	sut         *redis_services.RedisStorage
}

func WhenDeletingAValueBeforeEach(t *testing.T) *WhenDeletingAValueTestingSuite {
	redisClient := container_builders.GetRedisConnection()

	redisMetrics := metrics_mocks.NewMockRedisMetricsInterface(gomock.NewController(t))
	redisMetrics.EXPECT().RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &WhenDeletingAValueTestingSuite{
		redisClient: redisClient,
		sut:         redis_services.NewRedisStorage(redisClient, redisMetrics),
	}
}

func WhenDeletingAValueAfterEach(suite *WhenDeletingAValueTestingSuite) {
	container_builders.CloseRedisConnection(suite.redisClient)
}

func TestWhenDeletingAValue(t *testing.T) {
	t.Parallel()

	t.Run("Given the key is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should succeed", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingAValueBeforeEach(t)
			defer WhenDeletingAValueAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())

			assert.NoError(t, suite.sut.Delete(context.Background(), key))
		})
	})

	t.Run("Given the key is stored", func(t *testing.T) {
		t.Parallel()

		t.Run("Should remove it", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingAValueBeforeEach(t)
			defer WhenDeletingAValueAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			require.NoError(t, suite.redisClient.Set(context.Background(), key, "42", 0).Err())

			err := suite.sut.Delete(context.Background(), key)

			require.NoError(t, err)
			assert.Equal(t, int64(0), suite.redisClient.Exists(context.Background(), key).Val())
		})
	})
}
