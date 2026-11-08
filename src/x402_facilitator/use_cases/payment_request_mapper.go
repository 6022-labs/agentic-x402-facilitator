package use_cases

import (
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// paymentRequestMapper maps a payment request that passed ValidateAndSanitize onto the domain payment.
type paymentRequestMapper struct{}

func (m paymentRequestMapper) toPayment(request *requests.PaymentRequest) *models.Payment {
	requirements := request.PaymentRequirements
	raw := request.PaymentPayload.Payload.Authorization
	chainId, _ := helpers.ParseCaip2(requirements.Network)
	nonce, _ := hexutil.Decode(raw.Nonce)
	signature, _ := hexutil.Decode(request.PaymentPayload.Payload.Signature)

	authorization := &models.TransferAuthorization{
		From:        common.HexToAddress(raw.From),
		To:          common.HexToAddress(raw.To),
		Value:       m.uint(raw.Value),
		ValidAfter:  m.uint(raw.ValidAfter),
		ValidBefore: m.uint(raw.ValidBefore),
		Signature:   signature,
	}
	copy(authorization.Nonce[:], nonce)

	return &models.Payment{
		ChainId:       chainId,
		Asset:         common.HexToAddress(requirements.Asset),
		PayTo:         common.HexToAddress(requirements.PayTo),
		Amount:        m.uint(requirements.Amount),
		Authorization: authorization,
	}
}

func (m paymentRequestMapper) uint(value string) *big.Int {
	parsed, _ := new(big.Int).SetString(strings.TrimSpace(value), 10)
	return parsed
}
