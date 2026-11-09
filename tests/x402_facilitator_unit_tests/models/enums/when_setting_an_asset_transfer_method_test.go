package enums_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenSettingAnAssetTransferMethod(t *testing.T) {
	t.Parallel()

	t.Run("Given an unknown value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			var method enums.AssetTransferMethod

			err := method.Set("erc7710")

			assert.ErrorContains(t, err, "invalid AssetTransferMethod")
		})
	})

	t.Run("Given an uppercase and padded value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should set the method it names", func(t *testing.T) {
			t.Parallel()

			var method enums.AssetTransferMethod

			err := method.Set(" PERMIT2 ")

			require.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodPermit2, method)
		})
	})

	t.Run("Given a known value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should set it", func(t *testing.T) {
			t.Parallel()

			var method enums.AssetTransferMethod

			err := method.Set("eip3009")

			require.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodEip3009, method)
		})
	})
}
