package services

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
)

const permit2SettleAbi = `[{"type":"function","name":"settle","inputs":[
{"name":"permit","type":"tuple","components":[
{"name":"permitted","type":"tuple","components":[{"name":"token","type":"address"},{"name":"amount","type":"uint256"}]},
{"name":"nonce","type":"uint256"},{"name":"deadline","type":"uint256"}]},
{"name":"owner","type":"address"},
{"name":"witness","type":"tuple","components":[{"name":"to","type":"address"},{"name":"validAfter","type":"uint256"}]},
{"name":"signature","type":"bytes"}],"outputs":[]}]`

type permit2TokenPermissions struct {
	Token  common.Address
	Amount *big.Int
}

type permit2PermitTransferFrom struct {
	Permitted permit2TokenPermissions
	Nonce     *big.Int
	Deadline  *big.Int
}

type permit2Witness struct {
	To         common.Address
	ValidAfter *big.Int
}

// Permit2AssetTransferMethodStrategy settles through x402ExactPermit2Proxy.settle, for tokens without EIP-3009.
type Permit2AssetTransferMethodStrategy struct {
	paymentRequirementsMapper
	tokenRequester   x402_facilitator_interfaces.TokenRequesterInterface
	permit2Requester x402_facilitator_interfaces.Permit2RequesterInterface
	proxyAbi         abi.ABI
}

func NewPermit2AssetTransferMethodStrategy(
	logger *zap.Logger,
	tokenRequester x402_facilitator_interfaces.TokenRequesterInterface,
	permit2Requester x402_facilitator_interfaces.Permit2RequesterInterface,
) *Permit2AssetTransferMethodStrategy {
	proxyAbi, err := abi.JSON(strings.NewReader(permit2SettleAbi))
	if err != nil {
		logger.Fatal("failed to parse the x402ExactPermit2Proxy ABI", zap.Error(err))
	}
	return &Permit2AssetTransferMethodStrategy{
		tokenRequester:   tokenRequester,
		permit2Requester: permit2Requester,
		proxyAbi:         proxyAbi,
	}
}

func (s *Permit2AssetTransferMethodStrategy) Method() enums.AssetTransferMethod {
	return enums.AssetTransferMethodPermit2
}

func (s *Permit2AssetTransferMethodStrategy) Payment(request *requests.PaymentRequest) (*models.Payment, error) {
	raw := request.PaymentPayload.Payload.Permit2Authorization
	if raw == nil {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPayload,
			"permit2Authorization is required",
		)
	}
	if !common.IsHexAddress(raw.From) || !common.IsHexAddress(raw.Spender) ||
		!common.IsHexAddress(raw.Permitted.Token) || !common.IsHexAddress(raw.Witness.To) {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPayload,
			"permit2Authorization from, spender, token and witness to must be addresses",
		)
	}
	amount, amountOk := helpers.ParseUint256(raw.Permitted.Amount)
	nonce, nonceOk := helpers.ParseUint256(raw.Nonce)
	deadline, deadlineOk := helpers.ParseUint256(raw.Deadline)
	validAfter, validAfterOk := helpers.ParseUint256(raw.Witness.ValidAfter)
	if !amountOk || !nonceOk || !deadlineOk || !validAfterOk {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPayload,
			"permit2Authorization amount, nonce, deadline and validAfter must be decimal integers",
		)
	}
	signature, err := hexutil.Decode(request.PaymentPayload.Payload.Signature)
	if err != nil || len(signature) != 65 {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPermit2Signature,
			"signature must be 65 bytes of hex",
		)
	}

	payment := s.payment(request, enums.AssetTransferMethodPermit2)
	if common.HexToAddress(raw.Spender) != common.HexToAddress(helpers.X402ExactPermit2ProxyAddress) {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPermit2Spender,
			"spender must be x402ExactPermit2Proxy",
		)
	}
	if common.HexToAddress(raw.Witness.To) != payment.PayTo {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPermit2RecipientMismatch,
			"witness to does not match payTo",
		)
	}
	if common.HexToAddress(raw.Permitted.Token) != payment.Asset {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonPermit2TokenMismatch,
			"permitted token does not match asset",
		)
	}
	if amount.Cmp(payment.Amount) != 0 {
		return nil, apperrors.NewInvalidPaymentError(
			enums.PaymentReasonPermit2AmountMismatch,
			"permitted amount does not match amount",
		)
	}

	payment.Authorization = &models.TransferAuthorization{
		From:        common.HexToAddress(raw.From),
		To:          payment.PayTo,
		Value:       amount,
		ValidAfter:  validAfter,
		ValidBefore: deadline,
		Signature:   signature,
	}
	nonce.FillBytes(payment.Authorization.Nonce[:])
	return payment, nil
}

func (s *Permit2AssetTransferMethodStrategy) VerifySignature(
	payment *models.Payment,
	_ *x402_facilitator_settings.TokenConfig,
) error {
	digest, err := helpers.PermitWitnessTransferFromDigest(payment.Authorization, payment.Asset, payment.ChainId)
	if err != nil {
		return apperrors.NewInvalidPaymentError(enums.PaymentReasonInvalidPayload, "authorization cannot be hashed")
	}
	signer, err := helpers.RecoverSigner(digest, payment.Authorization.Signature)
	if err != nil || signer != payment.Authorization.From {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidPermit2Signature,
			"signature does not recover to from",
		)
	}
	return nil
}

func (s *Permit2AssetTransferMethodStrategy) VerifyOnChain(ctx context.Context, payment *models.Payment) error {
	deployed, err := s.permit2Requester.ProxyDeployed(ctx, payment.ChainId)
	if err != nil {
		return err
	}
	if !deployed {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonPermit2ProxyNotDeployed,
			"x402ExactPermit2Proxy is not deployed on this network",
		)
	}

	authorization := payment.Authorization
	used, err := s.permit2Requester.NonceUsed(ctx, payment.ChainId, authorization.From, authorization.NonceInt())
	if err != nil {
		return err
	}
	if used {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonInvalidTransactionState,
			"permit2 nonce is already used",
		)
	}

	allowance, err := s.tokenRequester.Allowance(
		ctx,
		payment.ChainId,
		payment.Asset,
		authorization.From,
		common.HexToAddress(helpers.Permit2Address),
	)
	if err != nil {
		return err
	}
	if allowance.Cmp(authorization.Value) < 0 {
		return apperrors.NewInvalidPaymentError(
			enums.PaymentReasonPermit2AllowanceRequired,
			"the payer has not approved Permit2 for this amount",
		)
	}
	return nil
}

func (s *Permit2AssetTransferMethodStrategy) TransferCall(payment *models.Payment) (*models.TransferCall, error) {
	authorization := payment.Authorization
	// Permit2 hands the signature to ecrecover, which only takes a v of 27/28.
	signature := append([]byte(nil), authorization.Signature...)
	if signature[64] < 27 {
		signature[64] += 27
	}
	data, err := s.proxyAbi.Pack(
		"settle",
		permit2PermitTransferFrom{
			Permitted: permit2TokenPermissions{Token: payment.Asset, Amount: authorization.Value},
			Nonce:     authorization.NonceInt(),
			Deadline:  authorization.ValidBefore,
		},
		authorization.From,
		permit2Witness{To: authorization.To, ValidAfter: authorization.ValidAfter},
		signature,
	)
	if err != nil {
		return nil, fmt.Errorf("pack settle: %w", err)
	}
	return &models.TransferCall{To: common.HexToAddress(helpers.X402ExactPermit2ProxyAddress), Data: data}, nil
}
