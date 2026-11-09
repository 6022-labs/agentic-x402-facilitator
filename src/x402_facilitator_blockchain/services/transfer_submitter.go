package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/helpers"
	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/metrics/interfaces"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/avast/retry-go/v4"
	"github.com/ethereum/go-ethereum"
	gethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

const (
	confirmationPollInterval = 2 * time.Second
	sendLockPollInterval     = 50 * time.Millisecond
	// gasMarginPercent covers a payee emptying its balance between the estimate and inclusion (a zero-to-nonzero write).
	gasMarginPercent = 20
	// keptTransactionTtl lets a signed transaction kept to refill its nonce expire.
	keptTransactionTtl = time.Hour
)

var (
	errSendLockHeld = errors.New("relayer send lock held by another replica")
	errNotConfirmed = errors.New("settlement not confirmed yet")
)

// TransferSubmitter shares the relayer nonce and send lock in Redis, so replicas never reuse a nonce.
type TransferSubmitter struct {
	ethClientFactory  *factory.EthClientFactory
	transactorFactory *factory.TransactorFactory
	chainsSettings    *blockchain_settings.ChainsSettings
	lockStorage       x402_facilitator_interfaces.LockStorageInterface
	cacheStorage      x402_facilitator_interfaces.CacheStorageInterface
	relayerMetrics    metrics_interfaces.RelayerMetricsInterface
	logger            *zap.Logger
}

func NewTransferSubmitter(
	ethClientFactory *factory.EthClientFactory,
	transactorFactory *factory.TransactorFactory,
	chainsSettings *blockchain_settings.ChainsSettings,
	lockStorage x402_facilitator_interfaces.LockStorageInterface,
	cacheStorage x402_facilitator_interfaces.CacheStorageInterface,
	relayerMetrics metrics_interfaces.RelayerMetricsInterface,
	logger *zap.Logger,
) *TransferSubmitter {
	return &TransferSubmitter{
		ethClientFactory:  ethClientFactory,
		transactorFactory: transactorFactory,
		chainsSettings:    chainsSettings,
		lockStorage:       lockStorage,
		cacheStorage:      cacheStorage,
		relayerMetrics:    relayerMetrics,
		logger:            logger,
	}
}

func (s *TransferSubmitter) Simulate(ctx context.Context, chainId uint64, call *models.TransferCall) error {
	client, err := s.ethClientFactory.For(chainId)
	if err != nil {
		return err
	}
	relayer, err := s.transactorFactory.Address(ctx)
	if err != nil {
		return err
	}
	_, err = client.CallContract(ctx, ethereum.CallMsg{From: relayer, To: &call.To, Data: call.Data}, nil)
	if helpers.IsRevert(err) {
		return fmt.Errorf("%w: %v", apperrors.ErrTransferRejected, err)
	}
	return err
}

func (s *TransferSubmitter) Broadcast(
	ctx context.Context,
	chainId uint64,
	call *models.TransferCall,
) (common.Hash, error) {
	chain, client, err := s.chain(chainId)
	if err != nil {
		return common.Hash{}, err
	}
	opts, err := s.transactorFactory.Create(ctx, chainId)
	if err != nil {
		return common.Hash{}, err
	}
	contract := bind.NewBoundContract(call.To, gethabi.ABI{}, client, client, client)

	keys := s.relayerKeys(chainId, opts.From)
	lockToken, err := s.acquireSendLock(ctx, keys.lock, chain)
	if err != nil {
		return common.Hash{}, err
	}
	defer s.releaseSendLock(ctx, keys.lock, lockToken)

	// Signing and sending must end before the lock expires, or another replica could take the same nonce.
	signCtx, cancel := context.WithTimeout(ctx, chain.SignDeadline)
	defer cancel()
	opts.Context = signCtx

	nonce, err := s.nextNonce(signCtx, client, chain, keys, opts.From)
	if err != nil {
		return common.Hash{}, err
	}
	gasLimit, err := s.estimateGas(signCtx, client, opts.From, call)
	if err != nil {
		return common.Hash{}, err
	}
	opts.Nonce = new(big.Int).SetUint64(nonce)
	opts.GasLimit = gasLimit
	opts.NoSend = true

	tx, err := contract.RawTransact(opts, call.Data)
	if err != nil {
		return common.Hash{}, fmt.Errorf("sign settlement: %w", err)
	}

	// Kept before sending: a send that dies half-way must still leave a transaction to refill its nonce.
	previous, err := s.cacheStorage.ReadString(signCtx, keys.signed(nonce))
	if err != nil {
		return common.Hash{}, fmt.Errorf("read kept transaction: %w", err)
	}
	sendErr := s.keepSigned(signCtx, keys, tx)
	// Past the sign deadline, another replica may hold the send lock and sign at this nonce.
	if sendErr == nil {
		sendErr = signCtx.Err()
	}
	if sendErr == nil {
		sendErr = s.send(ctx, client, chain, tx)
	}

	// The writes after the send get the lock's last RequestTimeout.
	afterCtx, cancelAfter := context.WithTimeout(context.WithoutCancel(ctx), chain.RequestTimeout)
	defer cancelAfter()
	if sendErr != nil {
		// An unsent transaction left kept could be resent later, so it is then reported as sent, not as refused.
		if restoreErr := s.restoreSigned(afterCtx, keys, nonce, previous); restoreErr != nil {
			s.logger.Warn("refused transaction still kept, tracking it as sent", zap.String("tx", tx.Hash().Hex()))
			return tx.Hash(), nil
		}
		return common.Hash{}, fmt.Errorf("send settlement: %w", sendErr)
	}
	return tx.Hash(), nil
}

func (s *TransferSubmitter) estimateGas(
	ctx context.Context,
	client *ethclient.Client,
	relayer common.Address,
	call *models.TransferCall,
) (uint64, error) {
	estimate, err := client.EstimateGas(ctx, ethereum.CallMsg{From: relayer, To: &call.To, Data: call.Data})
	if helpers.IsRevert(err) {
		return 0, fmt.Errorf("%w: %v", apperrors.ErrTransferRejected, err)
	}
	if err != nil {
		return 0, fmt.Errorf("estimate settlement gas: %w", err)
	}
	return estimate + estimate*gasMarginPercent/100, nil
}

func (s *TransferSubmitter) AwaitTransfer(
	ctx context.Context,
	chainId uint64,
	txHash common.Hash,
) (*models.Transfer, error) {
	chain, client, err := s.chain(chainId)
	if err != nil {
		return nil, err
	}

	// The transaction is out: a caller hanging up must not stop us from reporting its outcome.
	waitCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), chain.ReceiptTimeout)
	defer cancel()

	status, receipt := s.waitForOutcome(waitCtx, client, chain, txHash)
	if status == models.TransferStatusPending {
		s.refillNonceGap(ctx, client, chain, chainId)
	} else {
		s.recordGasSpent(ctx, chainId, receipt, status)
	}
	return &models.Transfer{TxHash: txHash, Status: status}, nil
}

func (s *TransferSubmitter) TransferStatus(
	ctx context.Context,
	chainId uint64,
	txHash common.Hash,
) (*models.Transfer, error) {
	chain, client, err := s.chain(chainId)
	if err != nil {
		return nil, err
	}

	receipt, err := client.TransactionReceipt(ctx, txHash)
	if errors.Is(err, ethereum.NotFound) {
		return &models.Transfer{TxHash: txHash, Status: models.TransferStatusPending}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read settlement receipt: %w", err)
	}
	head, err := client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("read head block: %w", err)
	}
	if head < receipt.BlockNumber.Uint64()+chain.Confirmations-1 {
		return &models.Transfer{TxHash: txHash, Status: models.TransferStatusPending}, nil
	}
	return &models.Transfer{TxHash: txHash, Status: s.outcome(receipt)}, nil
}

func (s *TransferSubmitter) chain(chainId uint64) (*blockchain_settings.ChainConfig, *ethclient.Client, error) {
	chain := s.chainsSettings.Get(chainId)
	if chain == nil {
		return nil, nil, fmt.Errorf("no RPC configured for chain %d", chainId)
	}
	client, err := s.ethClientFactory.For(chainId)
	if err != nil {
		return nil, nil, err
	}
	return chain, client, nil
}

// send treats an ambiguous failure as sent: the node may hold the transaction, and its nonce is spent either way.
func (s *TransferSubmitter) send(
	ctx context.Context,
	client *ethclient.Client,
	chain *blockchain_settings.ChainConfig,
	tx *types.Transaction,
) error {
	sendCtx, cancel := context.WithTimeout(ctx, chain.RequestTimeout)
	defer cancel()

	err := client.SendTransaction(sendCtx, tx)
	if err == nil {
		return nil
	}
	switch helpers.ClassifySendError(err) {
	case helpers.SendErrorAlreadyKnown:
		return nil
	case helpers.SendErrorUnknown:
		s.logger.Warn(
			"relayer send failed ambiguously, tracking it as sent",
			zap.String("tx", tx.Hash().Hex()),
			zap.Error(err),
		)
		return nil
	default:
		return err
	}
}

// acquireSendLock waits at most one lock ttl, by which time even a dead holder's lock has expired.
func (s *TransferSubmitter) acquireSendLock(
	ctx context.Context,
	lockKey string,
	chain *blockchain_settings.ChainConfig,
) (string, error) {
	ttl := chain.SendLockTtl
	ctx, cancel := context.WithTimeout(ctx, ttl)
	defer cancel()

	token, err := retry.DoWithData(
		func() (string, error) {
			token, acquired, err := s.lockStorage.TryLock(ctx, lockKey, ttl)
			if err != nil {
				return "", retry.Unrecoverable(fmt.Errorf("take relayer send lock: %w", err))
			}
			if !acquired {
				return "", errSendLockHeld
			}
			return token, nil
		},
		retry.Context(ctx),
		retry.Attempts(0),
		retry.Delay(sendLockPollInterval),
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
	)
	if errors.Is(err, errSendLockHeld) || errors.Is(err, context.DeadlineExceeded) {
		return "", fmt.Errorf("wait for relayer send lock: %w", err)
	}
	return token, err
}

func (s *TransferSubmitter) releaseSendLock(ctx context.Context, lockKey string, token string) {
	if err := s.lockStorage.Unlock(context.WithoutCancel(ctx), lockKey, token); err != nil {
		s.logger.Warn("failed to release the relayer send lock", zap.String("key", lockKey), zap.Error(err))
	}
}

// nextNonce resends the transactions kept from the node's pending nonce on and takes the first nonce with none.
func (s *TransferSubmitter) nextNonce(
	ctx context.Context,
	client *ethclient.Client,
	chain *blockchain_settings.ChainConfig,
	keys relayerKeys,
	relayer common.Address,
) (uint64, error) {
	nonce, err := client.PendingNonceAt(ctx, relayer)
	if err != nil {
		return 0, fmt.Errorf("read relayer pending nonce: %w", err)
	}
	for {
		resent, err := s.resendSigned(ctx, client, chain, keys, nonce)
		if err != nil {
			return 0, err
		}
		if !resent {
			return nonce, nil
		}
		nonce++
	}
}

// refillNonceGap runs after a pending wait: with no later send, nothing else would notice a lost transaction.
func (s *TransferSubmitter) refillNonceGap(
	ctx context.Context,
	client *ethclient.Client,
	chain *blockchain_settings.ChainConfig,
	chainId uint64,
) {
	ctx = context.WithoutCancel(ctx)
	relayer, err := s.transactorFactory.Address(ctx)
	if err != nil {
		s.logger.Warn("nonce gap not checked: relayer unavailable", zap.Uint64("chainId", chainId), zap.Error(err))
		return
	}
	keys := s.relayerKeys(chainId, relayer)
	lockToken, err := s.acquireSendLock(ctx, keys.lock, chain)
	if err != nil {
		s.logger.Warn("nonce gap not checked: send lock unavailable", zap.Uint64("chainId", chainId), zap.Error(err))
		return
	}
	defer s.releaseSendLock(ctx, keys.lock, lockToken)

	checkCtx, cancel := context.WithTimeout(ctx, chain.SignDeadline)
	defer cancel()
	if _, err := s.nextNonce(checkCtx, client, chain, keys, relayer); err != nil {
		s.logger.Warn("failed to check the relayer nonce", zap.Uint64("chainId", chainId), zap.Error(err))
	}
}

// resendSigned sends the kept transaction at nonce again; false when the nonce is still free.
func (s *TransferSubmitter) resendSigned(
	ctx context.Context,
	client *ethclient.Client,
	chain *blockchain_settings.ChainConfig,
	keys relayerKeys,
	nonce uint64,
) (bool, error) {
	raw, err := s.cacheStorage.ReadString(ctx, keys.signed(nonce))
	if err != nil {
		return false, fmt.Errorf("read kept transaction: %w", err)
	}
	if raw == nil {
		return false, nil
	}
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(common.FromHex(*raw)); err != nil {
		s.logger.Warn("kept transaction unreadable", zap.Uint64("nonce", nonce), zap.Error(err))
		return false, nil
	}
	if err := s.send(ctx, client, chain, tx); err != nil {
		// A lagging node may not count a mined nonce yet: taking it again would only fail.
		if helpers.ClassifySendError(err) == helpers.SendErrorNonceTaken {
			return true, nil
		}
		s.logger.Warn("failed to resend a lost settlement", zap.String("tx", tx.Hash().Hex()), zap.Error(err))
		return false, nil
	}
	return true, nil
}

// restoreSigned puts back what was kept at nonce before a send the node refused.
func (s *TransferSubmitter) restoreSigned(
	ctx context.Context,
	keys relayerKeys,
	nonce uint64,
	previous *string,
) error {
	if previous == nil {
		return s.cacheStorage.Delete(ctx, keys.signed(nonce))
	}
	return s.cacheStorage.Store(ctx, keys.signed(nonce), *previous, keptTransactionTtl)
}

// keepSigned stores the signed transaction, so a later send can refill its nonce if the chain loses it.
func (s *TransferSubmitter) keepSigned(ctx context.Context, keys relayerKeys, tx *types.Transaction) error {
	raw, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	return s.cacheStorage.Store(ctx, keys.signed(tx.Nonce()), hexutil.Encode(raw), keptTransactionTtl)
}

// relayerKeys are the store keys of one relayer on one chain.
type relayerKeys struct {
	lock   string
	prefix string
}

func (k relayerKeys) signed(nonce uint64) string {
	return fmt.Sprintf("%s:tx:%d", k.prefix, nonce)
}

func (s *TransferSubmitter) relayerKeys(chainId uint64, relayer common.Address) relayerKeys {
	prefix := fmt.Sprintf("x402:relayer:%d:%s", chainId, strings.ToLower(relayer.Hex()))
	return relayerKeys{lock: prefix + ":lock", prefix: prefix}
}

// waitForOutcome returns the receipt the status rests on, nil when pending.
func (s *TransferSubmitter) waitForOutcome(
	ctx context.Context,
	client *ethclient.Client,
	chain *blockchain_settings.ChainConfig,
	txHash common.Hash,
) (models.TransferStatus, *types.Receipt) {
	receipt, err := bind.WaitMinedHash(ctx, client, txHash)
	if err != nil {
		s.logger.Warn("settlement receipt not seen in time", zap.String("tx", txHash.Hex()), zap.Error(err))
		return models.TransferStatusPending, nil
	}
	status := models.TransferStatusPending
	err = retry.Do(
		func() error {
			target := receipt.BlockNumber.Uint64() + chain.Confirmations - 1
			head, err := client.BlockNumber(ctx)
			if err != nil || head < target {
				return errNotConfirmed
			}
			// A reorg may have dropped the transaction or moved it, as a revert or into a later block.
			current, err := client.TransactionReceipt(ctx, txHash)
			switch {
			case err != nil:
				// A lagging node may not know the receipt yet: keep polling until the wait ends.
				return errNotConfirmed
			case current.BlockHash == receipt.BlockHash:
				status, receipt = s.outcome(current), current
			default:
				receipt = current
				return errNotConfirmed
			}
			return nil
		},
		retry.Context(ctx),
		retry.Attempts(0),
		retry.Delay(confirmationPollInterval),
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
	)
	if err != nil {
		s.logger.Warn("settlement confirmations not reached in time", zap.String("tx", txHash.Hex()))
		return models.TransferStatusPending, nil
	}
	return status, receipt
}

func (s *TransferSubmitter) outcome(receipt *types.Receipt) models.TransferStatus {
	if receipt.Status == types.ReceiptStatusFailed {
		return models.TransferStatusReverted
	}
	return models.TransferStatusConfirmed
}

// recordGasSpent prices the gas at the receipt's effective price, which every post-London node returns.
func (s *TransferSubmitter) recordGasSpent(
	ctx context.Context,
	chainId uint64,
	receipt *types.Receipt,
	status models.TransferStatus,
) {
	if receipt.EffectiveGasPrice == nil {
		return
	}
	cost := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receipt.EffectiveGasPrice)
	s.relayerMetrics.RecordGasSpent(ctx, chainId, receipt.GasUsed, cost, string(status))
}
