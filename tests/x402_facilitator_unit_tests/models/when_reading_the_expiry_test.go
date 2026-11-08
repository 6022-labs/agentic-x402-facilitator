package models_test

import (
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingTheExpiry(t *testing.T) {
	t.Parallel()

	t.Run("Given a validBefore past int64 seconds", func(t *testing.T) {
		t.Parallel()

		authorization := model_builders.NewTransferAuthorizationBuilder().
			WithValidBefore(new(big.Int).Lsh(big.NewInt(1), 100)).
			Build()

		t.Run("Should read it as never expiring", func(t *testing.T) {
			t.Parallel()

			assert.True(t, authorization.ExpiresAt().After(time.Now().AddDate(1000, 0, 0)))
		})
	})

	t.Run("Given a validBefore near the int64 limit", func(t *testing.T) {
		t.Parallel()

		authorization := model_builders.NewTransferAuthorizationBuilder().
			WithValidBefore(big.NewInt(math.MaxInt64)).
			Build()

		t.Run("Should read it as never expiring", func(t *testing.T) {
			t.Parallel()

			assert.True(t, authorization.ExpiresAt().After(time.Now().AddDate(1000, 0, 0)))
		})
	})

	t.Run("Given a validBefore in seconds", func(t *testing.T) {
		t.Parallel()

		authorization := model_builders.NewTransferAuthorizationBuilder().
			WithValidBefore(big.NewInt(1_700_000_000)).
			Build()

		t.Run("Should return that time", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, time.Unix(1_700_000_000, 0), authorization.ExpiresAt())
		})
	})
}
