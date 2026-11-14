package services_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	redis_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/services"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/container_builders"
	metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_redis_mocks/metrics_mocks/interfaces_mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenReleasingALockTestingSuite struct {
	redisClient *redis.Client
	sut         *redis_services.RedisLockStorage
}

func WhenReleasingALockBeforeEach(t *testing.T) *WhenReleasingALockTestingSuite {
	redisClient := container_builders.GetRedisContainerBuilder().Create()

	redisMetrics := metrics_mocks.NewMockRedisMetricsInterface(gomock.NewController(t))
	redisMetrics.EXPECT().RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &WhenReleasingALockTestingSuite{
		redisClient: redisClient,
		sut:         redis_services.NewRedisLockStorage(redisClient, redisMetrics),
	}
}

func WhenReleasingALockAfterEach(suite *WhenReleasingALockTestingSuite) {
	container_builders.GetRedisContainerBuilder().DropConnection(suite.redisClient)
}

func TestWhenReleasingALock(t *testing.T) {
	t.Parallel()

	t.Run("Given another owner holds the lock", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report it not owned", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingALockBeforeEach(t)
			defer WhenReleasingALockAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			_, _, err := suite.sut.TryLock(context.Background(), key, time.Minute)
			require.NoError(t, err)
			err = suite.sut.Unlock(context.Background(), key, "someone-else")

			assert.ErrorIs(t, err, apperrors.ErrLockNotOwned)
		})

		t.Run("Should leave it held", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingALockBeforeEach(t)
			defer WhenReleasingALockAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			_, _, err := suite.sut.TryLock(context.Background(), key, time.Minute)
			require.NoError(t, err)
			_ = suite.sut.Unlock(context.Background(), key, "someone-else")

			assert.Equal(t, int64(1), suite.redisClient.Exists(context.Background(), key).Val())
		})
	})

	t.Run("Given the token owns the lock", func(t *testing.T) {
		t.Parallel()

		t.Run("Should free it", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingALockBeforeEach(t)
			defer WhenReleasingALockAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			token, _, err := suite.sut.TryLock(context.Background(), key, time.Minute)
			require.NoError(t, err)
			err = suite.sut.Unlock(context.Background(), key, token)

			require.NoError(t, err)
			assert.Equal(t, int64(0), suite.redisClient.Exists(context.Background(), key).Val())
		})
	})
}
