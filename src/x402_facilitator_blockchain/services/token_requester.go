package services

import (
	"context"
	"math/big"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/abi"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type TokenRequester struct {
	ethClientFactory *factory.EthClientFactory
}

func NewTokenRequester(ethClientFactory *factory.EthClientFactory) *TokenRequester {
	return &TokenRequester{ethClientFactory: ethClientFactory}
}

func (r *TokenRequester) AuthorizationState(
	ctx context.Context,
	chainId uint64,
	token common.Address,
	authorizer common.Address,
	nonce [32]byte,
) (bool, error) {
	caller, err := r.caller(chainId, token)
	if err != nil {
		return false, err
	}
	return caller.AuthorizationState(&bind.CallOpts{Context: ctx}, authorizer, nonce)
}

func (r *TokenRequester) BalanceOf(
	ctx context.Context,
	chainId uint64,
	token common.Address,
	owner common.Address,
) (*big.Int, error) {
	caller, err := r.caller(chainId, token)
	if err != nil {
		return nil, err
	}
	return caller.BalanceOf(&bind.CallOpts{Context: ctx}, owner)
}

func (r *TokenRequester) Allowance(
	ctx context.Context,
	chainId uint64,
	token common.Address,
	owner common.Address,
	spender common.Address,
) (*big.Int, error) {
	caller, err := r.caller(chainId, token)
	if err != nil {
		return nil, err
	}
	return caller.Allowance(&bind.CallOpts{Context: ctx}, owner, spender)
}

func (r *TokenRequester) caller(chainId uint64, token common.Address) (*abi.TokenCaller, error) {
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return nil, err
	}
	return abi.NewTokenCaller(token, client)
}
