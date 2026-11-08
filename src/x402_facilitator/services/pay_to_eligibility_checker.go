package services

import (
	"context"
	"strconv"
	"strings"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	eligibleValue    = "1"
	notEligibleValue = "0"
)

// PayToEligibilityChecker caches lookups, refusals included; a cache failure falls through to the registry.
type PayToEligibilityChecker struct {
	accessSettings         *x402_facilitator_settings.AccessSettings
	agentRegistryRequester x402_facilitator_interfaces.AgentRegistryRequesterInterface
	cacheStorage           x402_facilitator_interfaces.CacheStorageInterface
	logger                 *zap.Logger
}

func NewPayToEligibilityChecker(
	accessSettings *x402_facilitator_settings.AccessSettings,
	agentRegistryRequester x402_facilitator_interfaces.AgentRegistryRequesterInterface,
	cacheStorage x402_facilitator_interfaces.CacheStorageInterface,
	logger *zap.Logger,
) *PayToEligibilityChecker {
	return &PayToEligibilityChecker{
		accessSettings:         accessSettings,
		agentRegistryRequester: agentRegistryRequester,
		cacheStorage:           cacheStorage,
		logger:                 logger,
	}
}

func (c *PayToEligibilityChecker) Check(ctx context.Context, payTo common.Address) error {
	if !c.accessSettings.AgentsOnly {
		return nil
	}

	eligible, cached := c.cached(ctx, payTo)
	if !cached {
		isAgent, err := c.agentRegistryRequester.IsAgentAddress(ctx, payTo)
		if err != nil {
			return apperrors.NewUnavailableError("registry_unavailable", "the agent registry is unavailable", err)
		}
		c.store(ctx, payTo, isAgent)
		eligible = isAgent
	}

	if !eligible {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPaymentRequirements,
			"payTo is not a registered 6022 agent address",
		)
	}
	return nil
}

func (c *PayToEligibilityChecker) cached(ctx context.Context, payTo common.Address) (bool, bool) {
	value, err := c.cacheStorage.ReadString(ctx, c.cacheKey(payTo))
	if err != nil {
		c.logger.Warn("eligibility cache read failed", zap.String("payTo", payTo.Hex()), zap.Error(err))
		return false, false
	}
	if value == nil {
		return false, false
	}
	return *value == eligibleValue, true
}

func (c *PayToEligibilityChecker) store(ctx context.Context, payTo common.Address, eligible bool) {
	value := lo.Ternary(eligible, eligibleValue, notEligibleValue)
	if err := c.cacheStorage.Store(ctx, c.cacheKey(payTo), value, c.accessSettings.CacheTtl); err != nil {
		c.logger.Warn("eligibility cache store failed", zap.String("payTo", payTo.Hex()), zap.Error(err))
	}
}

// cacheKey is scoped to the registries read: a testnet deployment sharing the Redis must not vouch for mainnet.
func (c *PayToEligibilityChecker) cacheKey(payTo common.Address) string {
	chainIds := lo.Map(c.agentRegistryRequester.RegistryChainIds(), func(chainId uint64, _ int) string {
		return strconv.FormatUint(chainId, 10)
	})
	return "x402:eligibility:" + strings.Join(chainIds, ",") + ":" + strings.ToLower(payTo.Hex())
}
