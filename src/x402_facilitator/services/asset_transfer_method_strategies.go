package services

import (
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"go.uber.org/dig"
)

// AssetTransferMethodStrategies routes a payment to its method's strategy via the dig group.
type AssetTransferMethodStrategies struct {
	byMethod map[enums.AssetTransferMethod]x402_facilitator_interfaces.AssetTransferMethodStrategyInterface
}

type NewAssetTransferMethodStrategiesParams struct {
	dig.In

	Strategies []x402_facilitator_interfaces.AssetTransferMethodStrategyInterface `group:"asset_transfer_method_strategies"`
}

func NewAssetTransferMethodStrategies(params NewAssetTransferMethodStrategiesParams) *AssetTransferMethodStrategies {
	byMethod := map[enums.AssetTransferMethod]x402_facilitator_interfaces.AssetTransferMethodStrategyInterface{}
	for _, strategy := range params.Strategies {
		byMethod[strategy.Method()] = strategy
	}
	return &AssetTransferMethodStrategies{byMethod: byMethod}
}

func (r *AssetTransferMethodStrategies) For(
	method enums.AssetTransferMethod,
) (x402_facilitator_interfaces.AssetTransferMethodStrategyInterface, bool) {
	strategy, ok := r.byMethod[method]
	return strategy, ok
}
