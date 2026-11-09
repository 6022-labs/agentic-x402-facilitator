package services

import (
	"context"
	"math/big"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/abi"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Permit2Requester struct {
	ethClientFactory *factory.EthClientFactory
}

func NewPermit2Requester(ethClientFactory *factory.EthClientFactory) *Permit2Requester {
	return &Permit2Requester{ethClientFactory: ethClientFactory}
}

func (r *Permit2Requester) NonceUsed(
	ctx context.Context,
	chainId uint64,
	owner common.Address,
	nonce *big.Int,
) (bool, error) {
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return false, err
	}
	caller, err := abi.NewPermit2Caller(common.HexToAddress(helpers.Permit2Address), client)
	if err != nil {
		return false, err
	}
	// Permit2 keeps unordered nonces as bits: the word at nonce >> 8, the bit at its low 8 bits.
	word, err := caller.NonceBitmap(&bind.CallOpts{Context: ctx}, owner, new(big.Int).Rsh(nonce, 8))
	if err != nil {
		return false, err
	}
	return word.Bit(int(new(big.Int).And(nonce, big.NewInt(0xff)).Int64())) == 1, nil
}

func (r *Permit2Requester) ProxyDeployed(ctx context.Context, chainId uint64) (bool, error) {
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return false, err
	}
	code, err := client.CodeAt(ctx, common.HexToAddress(helpers.X402ExactPermit2ProxyAddress), nil)
	if err != nil {
		return false, err
	}
	return len(code) > 0, nil
}
