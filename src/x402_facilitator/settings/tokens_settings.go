package settings

import (
	"slices"
	"strings"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/ethereum/go-ethereum/common"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const TokensSettingsKey = "tokens"

const defaultAssetTransferMethods = "eip3009"

// TokenConfig lists how a token may be paid; the EIP-712 domain, needed by eip3009 only, never comes from the request.
type TokenConfig struct {
	AssetTransferMethods []enums.AssetTransferMethod
	Eip712Name           string
	Eip712Version        string
}

type tokenConfigRaw struct {
	AssetTransferMethods string `koanf:"asset_transfer_methods"`
	Eip712Name           string `koanf:"eip712_name"`
	Eip712Version        string `koanf:"eip712_version"`
}

// TokensSettings lists the settleable tokens per chain id, keyed by token address.
type TokensSettings map[uint64]map[common.Address]*TokenConfig

func NewTokensSettings(logger *zap.Logger, k *koanf.Koanf) *TokensSettings {
	raw := map[uint64]map[string]*tokenConfigRaw{}
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
			settings[chainId][common.HexToAddress(address)] = token.toTokenConfig(logger, address)
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

func (r *tokenConfigRaw) toTokenConfig(logger *zap.Logger, address string) *TokenConfig {
	token := &TokenConfig{
		Eip712Name:    strings.TrimSpace(r.Eip712Name),
		Eip712Version: strings.TrimSpace(r.Eip712Version),
	}
	methods := r.AssetTransferMethods
	if strings.TrimSpace(methods) == "" {
		methods = defaultAssetTransferMethods
	}
	for _, value := range strings.Split(methods, ",") {
		var method enums.AssetTransferMethod
		if err := method.Set(value); err != nil {
			logger.Fatal(
				"token asset_transfer_methods must list supported methods",
				zap.String("address", address),
				zap.Any("supported", enums.GetAllAssetTransferMethods()),
			)
		}
		token.AssetTransferMethods = append(token.AssetTransferMethods, method)
	}
	if lo.Contains(token.AssetTransferMethods, enums.AssetTransferMethodEip3009) &&
		(token.Eip712Name == "" || token.Eip712Version == "") {
		logger.Fatal("an eip3009 token needs eip712_name and eip712_version", zap.String("address", address))
	}
	return token
}
