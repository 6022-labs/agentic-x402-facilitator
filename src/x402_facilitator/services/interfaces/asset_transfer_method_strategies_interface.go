package interfaces

import "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"

// AssetTransferMethodStrategiesInterface resolves the strategy of an asset transfer method.
type AssetTransferMethodStrategiesInterface interface {
	// For returns the method's strategy, or false when the facilitator does not settle it.
	For(method enums.AssetTransferMethod) (AssetTransferMethodStrategyInterface, bool)
}
