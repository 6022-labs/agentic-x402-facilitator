package settings

import (
	"math"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

// GasBudgetSettingsKey is shared with the blockchain adapter's chains settings, which ignore the budget.
const GasBudgetSettingsKey = "chains"

const gweiPerToken = 1e9

// Redis compares counters as doubles, exact up to 2^53.
const maxDailyGasBudgetGwei = 1 << 53

// GasBudgetSettings is the gas the relayer may pay per UTC day, in gwei by chain id; an absent chain has no budget.
type GasBudgetSettings map[uint64]int64

type gasBudgetChainRaw struct {
	DailyGasBudget float64 `koanf:"daily_gas_budget"`
}

func NewGasBudgetSettings(logger *zap.Logger, k *koanf.Koanf) *GasBudgetSettings {
	raw := map[uint64]*gasBudgetChainRaw{}
	if err := k.UnmarshalWithConf(GasBudgetSettingsKey, &raw, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal chains gas budgets", zap.Error(err))
	}

	settings := GasBudgetSettings{}
	for chainId, chain := range raw {
		if chain == nil || chain.DailyGasBudget == 0 {
			continue
		}
		gwei := math.Floor(chain.DailyGasBudget * gweiPerToken)
		// Fatal: falling back to no budget would silently remove the bound.
		if math.IsNaN(gwei) || gwei < 1 || gwei > maxDailyGasBudgetGwei {
			logger.Fatal(
				"chains.<id>.daily_gas_budget is out of range",
				zap.Uint64("chainId", chainId),
				zap.Float64("dailyGasBudget", chain.DailyGasBudget),
			)
		}
		settings[chainId] = int64(gwei)
	}
	return &settings
}

// Get returns the chain's daily budget in gwei, false when it has none.
func (s *GasBudgetSettings) Get(chainId uint64) (int64, bool) {
	if s == nil {
		return 0, false
	}
	budget, found := (*s)[chainId]
	return budget, found
}
