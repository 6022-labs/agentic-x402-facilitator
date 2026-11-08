package factory

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/metrics/interfaces"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/transports"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// EthClientFactory dials and caches one RPC client per chain for this adapter.
type EthClientFactory struct {
	chainsSettings       *blockchain_settings.ChainsSettings
	blockchainRpcMetrics metrics_interfaces.BlockchainRpcMetricsInterface
	mu                   sync.Mutex
	clients              map[uint64]*ethclient.Client
}

func NewEthClientFactory(
	chainsSettings *blockchain_settings.ChainsSettings,
	blockchainRpcMetrics metrics_interfaces.BlockchainRpcMetricsInterface,
) *EthClientFactory {
	return &EthClientFactory{
		chainsSettings:       chainsSettings,
		blockchainRpcMetrics: blockchainRpcMetrics,
		clients:              map[uint64]*ethclient.Client{},
	}
}

// For returns the client for a chain, dialing it on first use.
func (f *EthClientFactory) For(chainId uint64) (*ethclient.Client, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if client, ok := f.clients[chainId]; ok {
		return client, nil
	}

	chain := f.chainsSettings.Get(chainId)
	if chain == nil {
		return nil, fmt.Errorf("no RPC configured for chain %d", chainId)
	}

	rpcRoundTripper, err := transports.NewRpcRoundTripper(
		http.DefaultTransport,
		chain.RpcHttpUrl,
		f.blockchainRpcMetrics,
		strconv.FormatUint(chainId, 10),
	)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{Timeout: chain.RequestTimeout, Transport: rpcRoundTripper}

	rpcClient, err := rpc.DialOptions(context.Background(), rpcRoundTripper.Origin(), rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("dial chain %d: %w", chainId, err)
	}

	client := ethclient.NewClient(rpcClient)
	f.clients[chainId] = client
	return client, nil
}
