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

type WhenTakingALockTestingSuite struct {
	redisClient *redis.Client
	sut         *redis_services.RedisLockStorage
}

func WhenTakingALockBeforeEach(t *testing.T) *WhenTakingALockTestingSuite {
	redisClient := container_builders.GetRedisContainerBuilder().Create()

	redisMetrics := metrics_mocks.NewMockRedisMetricsInterface(gomock.NewController(t))
	redisMetrics.EXPECT().RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &WhenTakingALockTestingSuite{
		redisClient: redisClient,
		sut:         redis_services.NewRedisLockStorage(redisClient, redisMetrics),
	}
}

func WhenTakingALockAfterEach(suite *WhenTakingALockTestingSuite) {
	container_builders.GetRedisContainerBuilder().DropConnection(suite.redisClient)
}

func TestWhenTakingALock(t *testing.T) {
	t.Parallel()

	t.Run("Given another owner holds the lock", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenTakingALockBeforeEach(t)
			defer WhenTakingALockAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			_, _, err := suite.sut.TryLock(context.Background(), key, time.Minute)
			require.NoError(t, err)
			_, acquired, err := suite.sut.TryLock(context.Background(), key, time.Minute)

			assert.NoError(t, err)
			assert.False(t, acquired)
		})
	})

	t.Run("Given the lock is free", func(t *testing.T) {
		t.Parallel()

		t.Run("Should take it and return its token", func(t *testing.T) {
			t.Parallel()

			suite := WhenTakingALockBeforeEach(t)
			defer WhenTakingALockAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())

			token, acquired, err := suite.sut.TryLock(context.Background(), key, time.Minute)

			require.NoError(t, err)
			assert.True(t, acquired)
			assert.Equal(t, token, suite.redisClient.Get(context.Background(), key).Val())
		})
	})
}
