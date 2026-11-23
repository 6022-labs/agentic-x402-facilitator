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

type WhenAddingToACounterTestingSuite struct {
	redisClient *redis.Client
	sut         *redis_services.RedisCounterStorage
}

func WhenAddingToACounterBeforeEach(t *testing.T) *WhenAddingToACounterTestingSuite {
	redisClient := container_builders.GetRedisContainerBuilder().Create()

	redisMetrics := metrics_mocks.NewMockRedisMetricsInterface(gomock.NewController(t))
	redisMetrics.EXPECT().RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &WhenAddingToACounterTestingSuite{
		redisClient: redisClient,
		sut:         redis_services.NewRedisCounterStorage(redisClient, redisMetrics),
	}
}

func WhenAddingToACounterAfterEach(suite *WhenAddingToACounterTestingSuite) {
	container_builders.GetRedisContainerBuilder().DropConnection(suite.redisClient)
}

func TestWhenAddingToACounter(t *testing.T) {
	t.Parallel()

	t.Run("Given an amount that would push the counter above the limit", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it and leave the counter unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenAddingToACounterBeforeEach(t)
			defer WhenAddingToACounterAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			require.NoError(t, suite.redisClient.Set(context.Background(), key, 8, time.Minute).Err())

			added, err := suite.sut.TryAdd(context.Background(), key, 3, 10, time.Minute)

			require.NoError(t, err)
			assert.False(t, added)
			assert.Equal(t, "8", suite.redisClient.Get(context.Background(), key).Val())
		})
	})

	t.Run("Given a missing counter and an amount above the limit", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it without creating the counter", func(t *testing.T) {
			t.Parallel()

			suite := WhenAddingToACounterBeforeEach(t)
			defer WhenAddingToACounterAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())

			added, err := suite.sut.TryAdd(context.Background(), key, 11, 10, time.Minute)

			require.NoError(t, err)
			assert.False(t, added)
			assert.Zero(t, suite.redisClient.Exists(context.Background(), key).Val())
		})
	})

	t.Run("Given a missing counter and an amount within the limit", func(t *testing.T) {
		t.Parallel()

		t.Run("Should create it with the amount and the ttl", func(t *testing.T) {
			t.Parallel()

			suite := WhenAddingToACounterBeforeEach(t)
			defer WhenAddingToACounterAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())

			added, err := suite.sut.TryAdd(context.Background(), key, 3, 10, time.Minute)

			require.NoError(t, err)
			assert.True(t, added)
			assert.Equal(t, "3", suite.redisClient.Get(context.Background(), key).Val())
			assert.InDelta(
				t,
				time.Minute,
				suite.redisClient.PTTL(context.Background(), key).Val(),
				float64(10*time.Second),
			)
		})
	})

	t.Run("Given an amount that brings the counter exactly to the limit", func(t *testing.T) {
		t.Parallel()

		t.Run("Should add the amount and keep its expiry", func(t *testing.T) {
			t.Parallel()

			suite := WhenAddingToACounterBeforeEach(t)
			defer WhenAddingToACounterAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			require.NoError(t, suite.redisClient.Set(context.Background(), key, 7, time.Minute).Err())

			added, err := suite.sut.TryAdd(context.Background(), key, 3, 10, time.Hour)

			require.NoError(t, err)
			assert.True(t, added)
			assert.Equal(t, "10", suite.redisClient.Get(context.Background(), key).Val())
			assert.LessOrEqual(t, suite.redisClient.PTTL(context.Background(), key).Val(), time.Minute)
		})
	})

	t.Run("Given more concurrent additions than the limit allows", func(t *testing.T) {
		t.Parallel()

		t.Run("Should stop exactly at the limit", func(t *testing.T) {
			t.Parallel()

			suite := WhenAddingToACounterBeforeEach(t)
			defer WhenAddingToACounterAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			results := make(chan bool, 20)

			for range 20 {
				go func() {
					added, err := suite.sut.TryAdd(context.Background(), key, 1, 5, time.Minute)
					results <- err == nil && added
				}()
			}

			accepted := 0
			for range 20 {
				if <-results {
					accepted++
				}
			}
			require.Equal(t, 5, accepted)
			assert.Equal(t, "5", suite.redisClient.Get(context.Background(), key).Val())
		})
	})

	t.Run("Given a counter near 2^53", func(t *testing.T) {
		t.Parallel()

		t.Run("Should count it exactly", func(t *testing.T) {
			t.Parallel()

			suite := WhenAddingToACounterBeforeEach(t)
			defer WhenAddingToACounterAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			require.NoError(t, suite.redisClient.Set(context.Background(), key, int64(1<<53-2), time.Minute).Err())

			added, err := suite.sut.TryAdd(context.Background(), key, 2, 1<<53, time.Minute)

			require.NoError(t, err)
			assert.True(t, added)
			assert.Equal(t, "9007199254740992", suite.redisClient.Get(context.Background(), key).Val())
		})
	})
}
