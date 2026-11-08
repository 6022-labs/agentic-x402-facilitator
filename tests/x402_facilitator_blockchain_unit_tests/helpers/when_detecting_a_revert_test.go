package helpers_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/helpers"
	"github.com/stretchr/testify/assert"
)

type revertRpcError struct{ code int }

func (e *revertRpcError) Error() string  { return "vm error" }
func (e *revertRpcError) ErrorCode() int { return e.code }

func TestWhenDetectingARevert(t *testing.T) {
	t.Parallel()

	t.Run("Given no error", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not report a revert", func(t *testing.T) {
			t.Parallel()

			assert.False(t, helpers.IsRevert(nil))
		})
	})

	t.Run("Given a JSON-RPC error with the execution reverted code", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report a revert", func(t *testing.T) {
			t.Parallel()

			assert.True(t, helpers.IsRevert(fmt.Errorf("estimate gas: %w", &revertRpcError{code: 3})))
		})
	})

	t.Run("Given an error saying execution reverted", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report a revert", func(t *testing.T) {
			t.Parallel()

			assert.True(t, helpers.IsRevert(errors.New("execution reverted: FiatTokenV2: invalid signature")))
		})
	})

	t.Run("Given a transport error", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not report a revert", func(t *testing.T) {
			t.Parallel()

			assert.False(t, helpers.IsRevert(errors.New("dial tcp: connection refused")))
		})
	})

	t.Run("Given a JSON-RPC error with another code", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not report a revert", func(t *testing.T) {
			t.Parallel()

			assert.False(t, helpers.IsRevert(&revertRpcError{code: -32005}))
		})
	})
}
