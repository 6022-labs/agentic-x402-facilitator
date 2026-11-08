package contracts

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/container_builders"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Deployer sends deployments and calls signed by the test's funded anvil key.
type Deployer struct {
	Address common.Address

	ctx          context.Context
	client       *ethclient.Client
	transactOpts *bind.TransactOpts
}

func NewDeployer(
	ctx context.Context,
	client *ethclient.Client,
	anvilAccountSettings container_builders.AnvilAccountSettings,
) (*Deployer, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(anvilAccountSettings.PrivateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid anvil private key: %w", err)
	}
	transactOpts, err := bind.NewKeyedTransactorWithChainID(
		key,
		new(big.Int).SetUint64(anvilAccountSettings.ChainIdUint),
	)
	if err != nil {
		return nil, err
	}

	return &Deployer{
		Address:      anvilAccountSettings.Address,
		ctx:          ctx,
		client:       client,
		transactOpts: transactOpts,
	}, nil
}

func (d *Deployer) Deploy(
	name string,
	contract *CompiledContract,
	args ...any,
) (common.Address, *bind.BoundContract, error) {
	address, tx, boundContract, err := bind.DeployContract(
		d.transactOpts,
		contract.Abi,
		contract.Bytecode,
		d.client,
		args...)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to deploy %s: %w", name, err)
	}
	if _, err := bind.WaitDeployed(d.ctx, d.client, tx); err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to mine %s deployment: %w", name, err)
	}
	return address, boundContract, nil
}

func (d *Deployer) Call(contract *bind.BoundContract, result any, method string, args ...any) error {
	results := []any{result}
	if err := contract.Call(&bind.CallOpts{Context: d.ctx}, &results, method, args...); err != nil {
		return fmt.Errorf("failed to call %s: %w", method, err)
	}
	return nil
}

func (d *Deployer) Transact(contract *bind.BoundContract, method string, args ...any) error {
	tx, err := contract.Transact(d.transactOpts, method, args...)
	if err != nil {
		return fmt.Errorf("failed to send %s: %w", method, err)
	}
	receipt, err := bind.WaitMined(d.ctx, d.client, tx)
	if err != nil {
		return fmt.Errorf("failed to mine %s: %w", method, err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("%s reverted (tx %s)", method, tx.Hash().Hex())
	}
	return nil
}
