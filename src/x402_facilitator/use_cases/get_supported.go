package use_cases

import (
	"context"
	"fmt"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/responses"
)

type GetSupported struct {
	tokensSettings *x402_facilitator_settings.TokensSettings
	relayerAccount x402_facilitator_interfaces.RelayerAccountInterface
}

func NewGetSupported(
	tokensSettings *x402_facilitator_settings.TokensSettings,
	relayerAccount x402_facilitator_interfaces.RelayerAccountInterface,
) *GetSupported {
	return &GetSupported{tokensSettings: tokensSettings, relayerAccount: relayerAccount}
}

func (u *GetSupported) Execute(ctx context.Context) (*responses.GetSupportedResponse, error) {
	relayer, err := u.relayerAccount.Address(ctx)
	if err != nil {
		return nil, apperrors.NewUnavailableError(
			"relayer_unavailable",
			"the relayer wallet is unavailable",
			fmt.Errorf("load relayer: %w", err),
		)
	}
	return responses.NewGetSupportedResponse(u.tokensSettings.ChainIds(), relayer), nil
}
