package settings_test

import (
	"testing"
	"time"

	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingAChainConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given a 10s request timeout and a 120s receipt timeout", func(t *testing.T) {
		t.Parallel()

		sut := blockchain_settings.NewChainConfig("http://rpc", 10*time.Second, 1, 120*time.Second, false)

		t.Run("Should hold the send lock for six request timeouts", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, 60*time.Second, sut.SendLockTtl)
		})

		t.Run("Should end signing after four request timeouts", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, 40*time.Second, sut.SignDeadline)
		})

		t.Run("Should bound a settlement by two locked sends and the receipt wait", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, 2*(60+40+10)*time.Second+120*time.Second, sut.MaxSettlementDuration)
		})
	})
}
