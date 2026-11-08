package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"

	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/abi"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	evmAddressType = "evm"
	// collectionsPageSize bounds each findAgentByAddress call, which loops over collections on-chain.
	collectionsPageSize = 50
)

type agentRegistry struct {
	chainId uint64
	caller  *abi.AgentCollectionsManagerCaller
}

// AgentRegistryRequester searches the AgentCollectionsManager of every configured chain.
type AgentRegistryRequester struct {
	registries []agentRegistry
}

// NewAgentRegistryRequester binds nothing when access is open: nobody asks for a lookup then.
func NewAgentRegistryRequester(
	logger *zap.Logger,
	accessSettings *x402_facilitator_settings.AccessSettings,
	chainsSettings *blockchain_settings.ChainsSettings,
	agentCollectionsManagersSettings *blockchain_settings.AgentCollectionsManagersSettings,
	ethClientFactory *factory.EthClientFactory,
) *AgentRegistryRequester {
	if !accessSettings.AgentsOnly {
		return &AgentRegistryRequester{}
	}

	registries := []agentRegistry{}
	// Every chain counts, registry or not: a mainnet chain would pay for testnet agents for free.
	networks := map[bool]bool{}
	for _, chainId := range chainsSettings.ChainIds() {
		networks[chainsSettings.Get(chainId).Testnet] = true
		address, found := agentCollectionsManagersSettings.Get(chainId)
		if !found {
			continue
		}
		client, err := ethClientFactory.For(chainId)
		if err != nil {
			logger.Fatal("failed to dial a registry chain", zap.Uint64("chainId", chainId), zap.Error(err))
		}
		caller, err := abi.NewAgentCollectionsManagerCaller(address, client)
		if err != nil {
			logger.Fatal("failed to bind AgentCollectionsManager", zap.Uint64("chainId", chainId), zap.Error(err))
		}
		registries = append(registries, agentRegistry{chainId: chainId, caller: caller})
		logger.Info(
			"bound AgentCollectionsManager",
			zap.Uint64("chainId", chainId),
			zap.String("address", address.Hex()),
		)
	}

	if len(registries) == 0 {
		logger.Fatal("access.agents_only needs a configured chain with a known AgentCollectionsManager")
	}
	if len(networks) > 1 {
		logger.Fatal("chains mix testnet and mainnet: a testnet agent would settle on mainnet for free")
	}
	return &AgentRegistryRequester{registries: registries}
}

func (r *AgentRegistryRequester) RegistryChainIds() []uint64 {
	chainIds := lo.Map(r.registries, func(registry agentRegistry, _ int) uint64 { return registry.chainId })
	slices.Sort(chainIds)
	return chainIds
}

func (r *AgentRegistryRequester) IsAgentAddress(ctx context.Context, address common.Address) (bool, error) {
	if len(r.registries) == 0 {
		return false, errors.New("agent registry is not bound")
	}

	// A chain that cannot be read only matters when no other chain knows the address.
	var unreadable error
	for _, registry := range r.registries {
		isAgent, err := r.isAgentAddressOn(ctx, registry, address)
		if err != nil {
			unreadable = errors.Join(unreadable, fmt.Errorf("chain %d: %w", registry.chainId, err))
			continue
		}
		if isAgent {
			return true, nil
		}
	}
	return false, unreadable
}

func (r *AgentRegistryRequester) isAgentAddressOn(
	ctx context.Context,
	registry agentRegistry,
	address common.Address,
) (bool, error) {
	value := strings.ToLower(address.Hex())
	offset := big.NewInt(0)
	limit := big.NewInt(collectionsPageSize)
	for {
		found, err := registry.caller.FindAgentByAddress(
			&bind.CallOpts{Context: ctx},
			evmAddressType,
			value,
			offset,
			limit,
		)
		if err != nil {
			return false, err
		}
		if found.Collection != (common.Address{}) {
			return true, nil
		}
		if found.Next == nil || found.Next.Sign() == 0 {
			return false, nil
		}
		offset = found.Next
	}
}
