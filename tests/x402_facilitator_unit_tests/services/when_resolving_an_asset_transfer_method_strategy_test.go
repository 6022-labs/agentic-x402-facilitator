package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenResolvingAnAssetTransferMethodStrategyTestingSuite struct {
	strategy *x402_facilitator_services.Eip3009AssetTransferMethodStrategy
	sut      *x402_facilitator_services.AssetTransferMethodStrategies
}

func WhenResolvingAnAssetTransferMethodStrategyBeforeEach() *WhenResolvingAnAssetTransferMethodStrategyTestingSuite {
	strategy := x402_facilitator_services.NewEip3009AssetTransferMethodStrategy(zap.NewNop(), nil)
	return &WhenResolvingAnAssetTransferMethodStrategyTestingSuite{
		strategy: strategy,
		sut: x402_facilitator_services.NewAssetTransferMethodStrategies(
			x402_facilitator_services.NewAssetTransferMethodStrategiesParams{
				Strategies: []x402_facilitator_interfaces.AssetTransferMethodStrategyInterface{strategy},
			},
		),
	}
}

func TestWhenResolvingAnAssetTransferMethodStrategy(t *testing.T) {
	t.Parallel()

	t.Run("Given a method no strategy settles", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report it missing", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAnAssetTransferMethodStrategyBeforeEach()

			_, found := suite.sut.For(enums.AssetTransferMethodPermit2)

			assert.False(t, found)
		})
	})

	t.Run("Given a method a strategy settles", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return that strategy", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAnAssetTransferMethodStrategyBeforeEach()

			strategy, found := suite.sut.For(enums.AssetTransferMethodEip3009)

			assert.True(t, found)
			assert.Same(t, suite.strategy, strategy)
		})
	})
}
