package settings

import (
	"slices"
	"strings"
	"time"

	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const ChainsSettingsKey = "chains"

const (
	// sendLockTtlFactor and signDeadlineFactor scale the request timeout: signing makes up to six RPC calls.
	sendLockTtlFactor  = 6
	signDeadlineFactor = 4

	defaultRequestTimeoutSeconds = 10
	defaultConfirmations         = 1
	defaultReceiptTimeoutSeconds = 120
)

// ChainConfig is one EVM chain the relayer reads and settles on; Confirmations counts the receipt's own block.
type ChainConfig struct {
	RpcHttpUrl     string
	RequestTimeout time.Duration
	Confirmations  uint64
	ReceiptTimeout time.Duration
	Testnet        bool
	// SendLockTtl outlives a send, so a dead replica's lock frees itself.
	SendLockTtl time.Duration
	// SignDeadline ends signing early enough to leave the send a full request timeout inside the lock.
	SignDeadline time.Duration
	// MaxSettlementDuration bounds a settlement: a send, the receipt wait, then a possible rebroadcast.
	MaxSettlementDuration time.Duration
}

type chainConfigRaw struct {
	RpcHttpUrl            string `koanf:"rpc_http_url"`
	RequestTimeoutSeconds uint64 `koanf:"request_timeout_seconds"`
	Confirmations         uint64 `koanf:"confirmations"`
	ReceiptTimeoutSeconds uint64 `koanf:"receipt_timeout_seconds"`
	Testnet               *bool  `koanf:"testnet"`
}

type ChainsSettings map[uint64]*ChainConfig

func NewChainConfig(
	rpcHttpUrl string,
	requestTimeout time.Duration,
	confirmations uint64,
	receiptTimeout time.Duration,
	testnet bool,
) *ChainConfig {
	sendLockTtl := sendLockTtlFactor * requestTimeout
	signDeadline := signDeadlineFactor * requestTimeout
	return &ChainConfig{
		RpcHttpUrl:            rpcHttpUrl,
		RequestTimeout:        requestTimeout,
		Confirmations:         confirmations,
		ReceiptTimeout:        receiptTimeout,
		Testnet:               testnet,
		SendLockTtl:           sendLockTtl,
		SignDeadline:          signDeadline,
		MaxSettlementDuration: 2*(sendLockTtl+signDeadline+requestTimeout) + receiptTimeout,
	}
}

// NewChainsSettings fails when a token's chain has no RPC, or a settlement can outlive its lock.
func NewChainsSettings(
	logger *zap.Logger,
	k *koanf.Koanf,
	tokensSettings *x402_facilitator_settings.TokensSettings,
	settlementSettings *x402_facilitator_settings.SettlementSettings,
) *ChainsSettings {
	raw := map[uint64]*chainConfigRaw{}
	if err := k.UnmarshalWithConf(ChainsSettingsKey, &raw, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal chains settings", zap.Error(err))
	}

	settings := ChainsSettings{}
	for chainId, chain := range raw {
		if chain == nil {
			logger.Fatal("chain entry is nil", zap.Uint64("chainId", chainId))
		}
		chain.RpcHttpUrl = strings.TrimSpace(chain.RpcHttpUrl)
		if chain.RpcHttpUrl == "" {
			logger.Fatal("chain rpc_http_url is required", zap.Uint64("chainId", chainId))
		}
		// Required: defaulting to mainnet would let a forgotten testnet flag mix networks.
		if chain.Testnet == nil {
			logger.Fatal("chains.<id>.testnet is required", zap.Uint64("chainId", chainId))
		}
		if chain.RequestTimeoutSeconds == 0 {
			chain.RequestTimeoutSeconds = defaultRequestTimeoutSeconds
		}
		if chain.Confirmations == 0 {
			chain.Confirmations = defaultConfirmations
		}
		if chain.ReceiptTimeoutSeconds == 0 {
			chain.ReceiptTimeoutSeconds = defaultReceiptTimeoutSeconds
		}
		config := NewChainConfig(
			chain.RpcHttpUrl,
			time.Duration(chain.RequestTimeoutSeconds)*time.Second,
			chain.Confirmations,
			time.Duration(chain.ReceiptTimeoutSeconds)*time.Second,
			*chain.Testnet,
		)
		if config.MaxSettlementDuration >= settlementSettings.LockTtl {
			logger.Fatal(
				"settlement.lock_ttl_seconds must exceed the chain's longest settlement",
				zap.Uint64("chainId", chainId),
				zap.Duration("longestSettlement", config.MaxSettlementDuration),
			)
		}
		settings[chainId] = config
	}

	for _, chainId := range tokensSettings.ChainIds() {
		if settings.Get(chainId) == nil {
			logger.Fatal(
				"tokens are configured on a chain with no chains.<id>.rpc_http_url",
				zap.Uint64("chainId", chainId),
			)
		}
	}
	return &settings
}

func (s *ChainsSettings) ChainIds() []uint64 {
	if s == nil {
		return nil
	}
	chainIds := lo.Keys(*s)
	slices.Sort(chainIds)
	return chainIds
}

func (s *ChainsSettings) Get(chainId uint64) *ChainConfig {
	if s == nil {
		return nil
	}
	return (*s)[chainId]
}
