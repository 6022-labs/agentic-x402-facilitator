package model_builders

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"strconv"
	"time"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// PaymentRequestBuilder builds an x402 v2 exact payment, EIP-3009 unless WithPermit2, signed by a fresh payer key.
type PaymentRequestBuilder struct {
	payerKey      *ecdsa.PrivateKey
	chainId       uint64
	asset         common.Address
	payTo         common.Address
	amount        string
	validAfter    int64
	validBefore   int64
	nonce         [32]byte
	eip712Name    string
	eip712Version string
	legacy        bool
	rawRecoveryId bool
	method        enums.AssetTransferMethod
}

func NewPaymentRequestBuilder() *PaymentRequestBuilder {
	payerKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		panic(err)
	}
	now := time.Now().Unix()
	return &PaymentRequestBuilder{
		payerKey:      payerKey,
		chainId:       80002,
		asset:         common_tests.RandomEthAddress(),
		payTo:         common_tests.RandomEthAddress(),
		amount:        "1000",
		validAfter:    now - 60,
		validBefore:   now + 300,
		nonce:         nonce,
		eip712Name:    "USDC",
		eip712Version: "2",
		method:        enums.AssetTransferMethodEip3009,
	}
}

func (b *PaymentRequestBuilder) WithPermit2() *PaymentRequestBuilder {
	b.method = enums.AssetTransferMethodPermit2
	return b
}

func (b *PaymentRequestBuilder) WithChainId(chainId uint64) *PaymentRequestBuilder {
	b.chainId = chainId
	return b
}

func (b *PaymentRequestBuilder) WithAsset(asset common.Address) *PaymentRequestBuilder {
	b.asset = asset
	return b
}

func (b *PaymentRequestBuilder) WithAmount(amount string) *PaymentRequestBuilder {
	b.amount = amount
	return b
}

func (b *PaymentRequestBuilder) WithValidAfter(validAfter int64) *PaymentRequestBuilder {
	b.validAfter = validAfter
	return b
}

func (b *PaymentRequestBuilder) WithValidBefore(validBefore int64) *PaymentRequestBuilder {
	b.validBefore = validBefore
	return b
}

func (b *PaymentRequestBuilder) WithEip712Domain(name, version string) *PaymentRequestBuilder {
	b.eip712Name = name
	b.eip712Version = version
	return b
}

// WithLegacyPayload puts scheme and network at the payload's top level instead of in accepted.
func (b *PaymentRequestBuilder) WithLegacyPayload() *PaymentRequestBuilder {
	b.legacy = true
	return b
}

// WithRawRecoveryId signs with a v of 0 or 1, as some wallets do, instead of 27 or 28.
func (b *PaymentRequestBuilder) WithRawRecoveryId() *PaymentRequestBuilder {
	b.rawRecoveryId = true
	return b
}

func (b *PaymentRequestBuilder) Payer() common.Address {
	return crypto.PubkeyToAddress(b.payerKey.PublicKey)
}

func (b *PaymentRequestBuilder) Build() *requests.PaymentRequest {
	network := helpers.FormatCaip2(b.chainId)
	requirements := &requests.PaymentRequirements{
		Scheme:            enums.PaymentSchemeExact.String(),
		Network:           network,
		Amount:            b.amount,
		Asset:             b.asset.Hex(),
		PayTo:             b.payTo.Hex(),
		MaxTimeoutSeconds: 60,
		Extra:             map[string]any{"name": b.eip712Name, "version": b.eip712Version},
	}
	payload := &requests.PaymentPayload{
		X402Version: enums.X402Version,
		Payload:     requests.ExactEvmPayload{Signature: b.sign()},
	}
	if b.method == enums.AssetTransferMethodPermit2 {
		requirements.Extra["assetTransferMethod"] = enums.AssetTransferMethodPermit2.String()
		payload.Payload.Permit2Authorization = &requests.Permit2Authorization{
			Permitted: requests.Permit2TokenPermissions{Token: b.asset.Hex(), Amount: b.amount},
			From:      b.Payer().Hex(),
			Spender:   helpers.X402ExactPermit2ProxyAddress,
			Nonce:     new(big.Int).SetBytes(b.nonce[:]).String(),
			Deadline:  strconv.FormatInt(b.validBefore, 10),
			Witness: requests.Permit2Witness{
				To:         b.payTo.Hex(),
				ValidAfter: strconv.FormatInt(b.validAfter, 10),
			},
		}
	} else {
		payload.Payload.Eip3009Authorization = &requests.Eip3009Authorization{
			From:        b.Payer().Hex(),
			To:          b.payTo.Hex(),
			Value:       b.amount,
			ValidAfter:  strconv.FormatInt(b.validAfter, 10),
			ValidBefore: strconv.FormatInt(b.validBefore, 10),
			Nonce:       hexutil.Encode(b.nonce[:]),
		}
	}
	if b.legacy {
		payload.Scheme = enums.PaymentSchemeExact.String()
		payload.Network = network
	} else {
		accepted := *requirements
		payload.Accepted = &accepted
	}

	return &requests.PaymentRequest{
		X402Version:         enums.X402Version,
		PaymentPayload:      payload,
		PaymentRequirements: requirements,
	}
}

func (b *PaymentRequestBuilder) BuildVerifyRequest() *requests.VerifyPaymentRequest {
	return &requests.VerifyPaymentRequest{PaymentRequest: *b.Build()}
}

func (b *PaymentRequestBuilder) BuildSettleRequest() *requests.SettlePaymentRequest {
	return &requests.SettlePaymentRequest{PaymentRequest: *b.Build()}
}

// BuildPayment returns the domain payment the built request carries.
func (b *PaymentRequestBuilder) BuildPayment() *models.Payment {
	authorization := b.authorization()
	authorization.Signature = common.FromHex(b.sign())
	return &models.Payment{
		ChainId:       b.chainId,
		Asset:         b.asset,
		PayTo:         b.payTo,
		Amount:        b.Amount(),
		Method:        b.method,
		Authorization: authorization,
	}
}

func (b *PaymentRequestBuilder) authorization() *models.TransferAuthorization {
	return &models.TransferAuthorization{
		From:        b.Payer(),
		To:          b.payTo,
		Value:       b.Amount(),
		ValidAfter:  big.NewInt(b.validAfter),
		ValidBefore: big.NewInt(b.validBefore),
		Nonce:       b.nonce,
	}
}

func (b *PaymentRequestBuilder) sign() string {
	digest, err := helpers.TransferWithAuthorizationDigest(
		b.authorization(),
		b.asset,
		b.eip712Name,
		b.eip712Version,
		b.chainId,
	)
	if b.method == enums.AssetTransferMethodPermit2 {
		digest, err = helpers.PermitWitnessTransferFromDigest(b.authorization(), b.asset, b.chainId)
	}
	if err != nil {
		panic(err)
	}
	signature, err := crypto.Sign(digest, b.payerKey)
	if err != nil {
		panic(err)
	}
	if !b.rawRecoveryId {
		signature[64] += 27
	}
	return hexutil.Encode(signature)
}

func (b *PaymentRequestBuilder) Amount() *big.Int {
	amount, _ := new(big.Int).SetString(b.amount, 10)
	return amount
}
