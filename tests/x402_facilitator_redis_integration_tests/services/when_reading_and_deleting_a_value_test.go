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

type WhenReadingAndDeletingAValueTestingSuite struct {
	redisClient *redis.Client
	sut         *redis_services.RedisStorage
}

func WhenReadingAndDeletingAValueBeforeEach(t *testing.T) *WhenReadingAndDeletingAValueTestingSuite {
	redisClient := container_builders.GetRedisContainerBuilder().Create()

	redisMetrics := metrics_mocks.NewMockRedisMetricsInterface(gomock.NewController(t))
	redisMetrics.EXPECT().RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &WhenReadingAndDeletingAValueTestingSuite{
		redisClient: redisClient,
		sut:         redis_services.NewRedisStorage(redisClient, redisMetrics),
	}
}

func WhenReadingAndDeletingAValueAfterEach(suite *WhenReadingAndDeletingAValueTestingSuite) {
	container_builders.GetRedisContainerBuilder().DropConnection(suite.redisClient)
}

func TestWhenReadingAndDeletingAValue(t *testing.T) {
	t.Parallel()

	t.Run("Given the key is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nil", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAndDeletingAValueBeforeEach(t)
			defer WhenReadingAndDeletingAValueAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())

			value, err := suite.sut.ReadAndDelete(context.Background(), key)

			assert.NoError(t, err)
			assert.Nil(t, value)
		})
	})

	t.Run("Given the key is stored", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its value and delete it", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAndDeletingAValueBeforeEach(t)
			defer WhenReadingAndDeletingAValueAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			require.NoError(t, suite.redisClient.Set(context.Background(), key, "42", 0).Err())
			value, err := suite.sut.ReadAndDelete(context.Background(), key)

			require.NoError(t, err)
			require.NotNil(t, value)
			assert.Equal(t, "42", *value)
			assert.Zero(t, suite.redisClient.Exists(context.Background(), key).Val())
		})
	})

	t.Run("Given several callers read the same key at once", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the value to exactly one of them", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAndDeletingAValueBeforeEach(t)
			defer WhenReadingAndDeletingAValueAfterEach(suite)
			key := fmt.Sprintf("test:%s:%d", t.Name(), time.Now().UnixNano())
			require.NoError(t, suite.redisClient.Set(context.Background(), key, "42", 0).Err())
			results := make(chan bool, 20)

			for range 20 {
				go func() {
					value, err := suite.sut.ReadAndDelete(context.Background(), key)
					results <- err == nil && value != nil
				}()
			}

			received := 0
			for range 20 {
				if <-results {
					received++
				}
			}
			assert.Equal(t, 1, received)
		})
	})
}
