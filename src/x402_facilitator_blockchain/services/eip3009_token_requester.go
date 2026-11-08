package services

import (
	"context"
	"fmt"
	"math/big"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/abi"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/helpers"
	"github.com/ethereum/go-ethereum"
	gethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type Eip3009TokenRequester struct {
	ethClientFactory  *factory.EthClientFactory
	transactorFactory *factory.TransactorFactory
	tokenAbi          *gethabi.ABI
}

func NewEip3009TokenRequester(
	logger *zap.Logger,
	ethClientFactory *factory.EthClientFactory,
	transactorFactory *factory.TransactorFactory,
) *Eip3009TokenRequester {
	tokenAbi, err := abi.Eip3009TokenMetaData.GetAbi()
	if err != nil {
		logger.Fatal("failed to parse the EIP-3009 token ABI", zap.Error(err))
	}
	return &Eip3009TokenRequester{
		ethClientFactory:  ethClientFactory,
		transactorFactory: transactorFactory,
		tokenAbi:          tokenAbi,
	}
}

func (r *Eip3009TokenRequester) AuthorizationState(
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

func (r *Eip3009TokenRequester) BalanceOf(
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

func (r *Eip3009TokenRequester) SimulateTransferWithAuthorization(
	ctx context.Context,
	chainId uint64,
	token common.Address,
	authorization *models.TransferAuthorization,
) error {
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return err
	}
	relayer, err := r.transactorFactory.Address(ctx)
	if err != nil {
		return err
	}

	v, sigR, sigS := helpers.SplitSignature(authorization.Signature)
	data, err := r.tokenAbi.Pack(
		"transferWithAuthorization",
		authorization.From,
		authorization.To,
		authorization.Value,
		authorization.ValidAfter,
		authorization.ValidBefore,
		authorization.Nonce,
		v,
		sigR,
		sigS,
	)
	if err != nil {
		return fmt.Errorf("pack transferWithAuthorization: %w", err)
	}

	_, err = client.CallContract(ctx, ethereum.CallMsg{From: relayer, To: &token, Data: data}, nil)
	if helpers.IsRevert(err) {
		return fmt.Errorf("%w: %v", apperrors.ErrTransferRejected, err)
	}
	return err
}

func (r *Eip3009TokenRequester) caller(chainId uint64, token common.Address) (*abi.Eip3009TokenCaller, error) {
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return nil, err
	}
	return abi.NewEip3009TokenCaller(token, client)
}
