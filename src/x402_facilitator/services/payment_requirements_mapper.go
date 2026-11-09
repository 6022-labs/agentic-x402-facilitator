package services

import (
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
	"github.com/ethereum/go-ethereum/common"
)

// paymentRequirementsMapper maps the requirements every method shares, from a request that passed validation.
type paymentRequirementsMapper struct{}

func (m paymentRequirementsMapper) payment(
	request *requests.PaymentRequest,
	method enums.AssetTransferMethod,
) *models.Payment {
	requirements := request.PaymentRequirements
	chainId, _ := helpers.ParseCaip2(requirements.Network)
	amount, _ := helpers.ParseUint256(requirements.Amount)
	return &models.Payment{
		ChainId: chainId,
		Asset:   common.HexToAddress(requirements.Asset),
		PayTo:   common.HexToAddress(requirements.PayTo),
		Amount:  amount,
		Method:  method,
	}
}
