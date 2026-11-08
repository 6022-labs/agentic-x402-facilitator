package model_builders

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/ethereum/go-ethereum/common"
)

// TransferAuthorizationBuilder builds an authorization whose signature is placeholder bytes, not a real signature.
type TransferAuthorizationBuilder struct {
	from        common.Address
	to          common.Address
	value       *big.Int
	validAfter  *big.Int
	validBefore *big.Int
	nonce       [32]byte
	signature   []byte
}

func NewTransferAuthorizationBuilder() *TransferAuthorizationBuilder {
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		panic(err)
	}
	now := time.Now().Unix()
	return &TransferAuthorizationBuilder{
		from:        common_tests.RandomEthAddress(),
		to:          common_tests.RandomEthAddress(),
		value:       big.NewInt(1000),
		validAfter:  big.NewInt(now - 60),
		validBefore: big.NewInt(now + 300),
		nonce:       nonce,
		signature:   make([]byte, 65),
	}
}

func (b *TransferAuthorizationBuilder) WithTo(to common.Address) *TransferAuthorizationBuilder {
	b.to = to
	return b
}

func (b *TransferAuthorizationBuilder) WithValue(value *big.Int) *TransferAuthorizationBuilder {
	b.value = value
	return b
}

func (b *TransferAuthorizationBuilder) WithValidAfter(validAfter *big.Int) *TransferAuthorizationBuilder {
	b.validAfter = validAfter
	return b
}

func (b *TransferAuthorizationBuilder) WithValidBefore(validBefore *big.Int) *TransferAuthorizationBuilder {
	b.validBefore = validBefore
	return b
}

func (b *TransferAuthorizationBuilder) WithSignature(signature []byte) *TransferAuthorizationBuilder {
	b.signature = signature
	return b
}

func (b *TransferAuthorizationBuilder) Build() *models.TransferAuthorization {
	return &models.TransferAuthorization{
		From:        b.from,
		To:          b.to,
		Value:       new(big.Int).Set(b.value),
		ValidAfter:  new(big.Int).Set(b.validAfter),
		ValidBefore: new(big.Int).Set(b.validBefore),
		Nonce:       b.nonce,
		Signature:   append([]byte(nil), b.signature...),
	}
}
