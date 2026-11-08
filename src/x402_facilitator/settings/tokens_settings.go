package settings

import (
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const TokensSettingsKey = "tokens"

// TokenConfig is an EIP-3009 token's EIP-712 domain; the facilitator hashes with it, never with the request's extra.
type TokenConfig struct {
	Eip712Name    string `koanf:"eip712_name"`
	Eip712Version string `koanf:"eip712_version"`
}

// TokensSettings lists the settleable tokens per chain id, keyed by token address.
type TokensSettings map[uint64]map[common.Address]*TokenConfig

func NewTokensSettings(logger *zap.Logger, k *koanf.Koanf) *TokensSettings {
	raw := map[uint64]map[string]*TokenConfig{}
	if err := k.UnmarshalWithConf(TokensSettingsKey, &raw, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal tokens settings", zap.Error(err))
	}

	if len(raw) == 0 {
		logger.Fatal("tokens settings is empty: configure at least one token via TOKENS__<chain_id>__<address>__*")
	}

	settings := TokensSettings{}
	for chainId, tokens := range raw {
		if len(tokens) == 0 {
			logger.Fatal("chain has no token", zap.Uint64("chainId", chainId))
		}
		settings[chainId] = map[common.Address]*TokenConfig{}
		for address, token := range tokens {
			if !common.IsHexAddress(address) {
				logger.Fatal("token address is not a valid hex address", zap.String("address", address))
			}
			if token == nil {
				logger.Fatal("token entry is nil", zap.String("address", address))
			}
			// Env keys arrive lowercased: a differently cased JSON key would race them for the same token.
			if _, duplicate := settings[chainId][common.HexToAddress(address)]; duplicate {
				logger.Fatal(
					"token configured twice; write token addresses in lowercase",
					zap.String("address", address),
				)
			}
			token.Eip712Name = strings.TrimSpace(token.Eip712Name)
			token.Eip712Version = strings.TrimSpace(token.Eip712Version)
			if token.Eip712Name == "" || token.Eip712Version == "" {
				logger.Fatal("token eip712_name and eip712_version are required", zap.String("address", address))
			}
			settings[chainId][common.HexToAddress(address)] = token
		}
	}

	return &settings
}

func (s *TokensSettings) Get(chainId uint64, address common.Address) *TokenConfig {
	if s == nil {
		return nil
	}
	return (*s)[chainId][address]
}

func (s *TokensSettings) HasChain(chainId uint64) bool {
	if s == nil {
		return false
	}
	_, found := (*s)[chainId]
	return found
}

func (s *TokensSettings) ChainIds() []uint64 {
	if s == nil {
		return nil
	}
	chainIds := lo.Keys(*s)
	slices.Sort(chainIds)
	return chainIds
}
