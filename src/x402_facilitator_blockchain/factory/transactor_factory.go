package factory

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/zap"
)

const (
	// relayerKeyReadTimeout bounds the Vault read every waiting request queues behind.
	relayerKeyReadTimeout = 10 * time.Second
	// relayerKeyRetryDelay lets the requests queued behind a failed read fail at once instead of reading in turn.
	relayerKeyRetryDelay = 5 * time.Second
)

// TransactorFactory loads the relayer key from the secret store once, then signs every settlement with it.
type TransactorFactory struct {
	relayerKeyStorage x402_facilitator_interfaces.RelayerKeyStorageInterface
	logger            *zap.Logger
	mu                sync.Mutex
	privateKey        *ecdsa.PrivateKey
	failedAt          time.Time
	failure           error
}

func NewTransactorFactory(
	relayerKeyStorage x402_facilitator_interfaces.RelayerKeyStorageInterface,
	logger *zap.Logger,
) *TransactorFactory {
	return &TransactorFactory{relayerKeyStorage: relayerKeyStorage, logger: logger}
}

func (f *TransactorFactory) Create(ctx context.Context, chainId uint64) (*bind.TransactOpts, error) {
	privateKey, err := f.load(ctx)
	if err != nil {
		return nil, err
	}
	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, new(big.Int).SetUint64(chainId))
	if err != nil {
		return nil, fmt.Errorf("build relayer transactor: %w", err)
	}
	opts.Context = ctx
	return opts, nil
}

func (f *TransactorFactory) Address(ctx context.Context) (common.Address, error) {
	privateKey, err := f.load(ctx)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(privateKey.PublicKey), nil
}

// load holds mu across the Vault read so one call serves every waiter; a failure is retried after a short delay.
func (f *TransactorFactory) load(ctx context.Context) (*ecdsa.PrivateKey, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.privateKey != nil {
		return f.privateKey, nil
	}
	if f.failure != nil && time.Since(f.failedAt) < relayerKeyRetryDelay {
		return nil, f.failure
	}

	readCtx, cancel := context.WithTimeout(ctx, relayerKeyReadTimeout)
	defer cancel()

	privateKeyHex, err := f.relayerKeyStorage.ReadPrivateKey(readCtx)
	if err != nil {
		return nil, f.fail(fmt.Errorf("%w: read relayer key: %w", apperrors.ErrRelayerUnavailable, err))
	}
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(strings.TrimSpace(privateKeyHex), "0x"))
	if err != nil {
		return nil, f.fail(fmt.Errorf("%w: parse relayer key: %w", apperrors.ErrRelayerUnavailable, err))
	}

	f.privateKey = privateKey
	f.logger.Info("relayer key loaded", zap.String("address", crypto.PubkeyToAddress(privateKey.PublicKey).Hex()))
	return privateKey, nil
}

func (f *TransactorFactory) fail(err error) error {
	f.failedAt = time.Now()
	f.failure = err
	return err
}
