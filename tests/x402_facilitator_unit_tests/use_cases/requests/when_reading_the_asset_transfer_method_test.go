package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingTheAssetTransferMethod(t *testing.T) {
	t.Parallel()

	t.Run("Given requirements without an asset transfer method", func(t *testing.T) {
		t.Parallel()

		requirements := &requests.PaymentRequirements{}

		t.Run("Should default to eip3009", func(t *testing.T) {
			t.Parallel()

			method, known := requirements.AssetTransferMethod()

			assert.True(t, known)
			assert.Equal(t, enums.AssetTransferMethodEip3009, method)
		})
	})

	t.Run("Given an asset transfer method that is not a string", func(t *testing.T) {
		t.Parallel()

		requirements := &requests.PaymentRequirements{Extra: map[string]any{"assetTransferMethod": 2}}

		t.Run("Should report it unknown", func(t *testing.T) {
			t.Parallel()

			_, known := requirements.AssetTransferMethod()

			assert.False(t, known)
		})
	})

	t.Run("Given the permit2 asset transfer method", func(t *testing.T) {
		t.Parallel()

		requirements := &requests.PaymentRequirements{Extra: map[string]any{"assetTransferMethod": "permit2"}}

		t.Run("Should read permit2", func(t *testing.T) {
			t.Parallel()

			method, known := requirements.AssetTransferMethod()

			assert.True(t, known)
			assert.Equal(t, enums.AssetTransferMethodPermit2, method)
		})
	})
}
