package fixtures

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/factory"
	blockchain_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/settings"
	redis_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_redis/services"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/model_builders"
	blockchain_metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_blockchain_mocks/metrics_mocks/interfaces_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_mocks/services_mocks/interfaces_mocks"
	redis_metrics_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_redis_mocks/metrics_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const (
	TokenEip712Name    = "USD Coin"
	TokenEip712Version = "2"
)

// IntegrationChain is a fresh relayer and token on the shared anvil chain and redis, wired like the host.
type IntegrationChain struct {
	ChainId           uint64
	Client            *ethclient.Client
	ChainsSettings    *blockchain_settings.ChainsSettings
	EthClientFactory  *factory.EthClientFactory
	TransactorFactory *factory.TransactorFactory
	RedisStorage      *redis_services.RedisStorage
	RedisLockStorage  *redis_services.RedisLockStorage
	Relayer           container_builders.AnvilAccountSettings
	Token             *Eip3009Token
}

func NewIntegrationChain(t *testing.T, chainConfig blockchain_settings.ChainConfig) *IntegrationChain {
	relayer := container_builders.GetAnvilContainerBuilder().Create()
	client, err := ethclient.Dial(relayer.HttpUrl)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	chainsSettings := &blockchain_settings.ChainsSettings{relayer.ChainIdUint: blockchain_settings.NewChainConfig(
		relayer.HttpUrl,
		10*time.Second,
		chainConfig.Confirmations,
		chainConfig.ReceiptTimeout,
		false,
	)}

	mockController := gomock.NewController(t)
	mockBlockchainRpcMetrics := blockchain_metrics_mocks.NewMockBlockchainRpcMetricsInterface(mockController)
	mockBlockchainRpcMetrics.EXPECT().
		RecordRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()
	mockRelayerKeyStorage := interfaces_mocks.NewMockRelayerKeyStorageInterface(mockController)
	mockRelayerKeyStorage.EXPECT().ReadPrivateKey(gomock.Any()).Return(relayer.PrivateKeyHex, nil).AnyTimes()
	mockRedisMetrics := redis_metrics_mocks.NewMockRedisMetricsInterface(mockController)
	mockRedisMetrics.EXPECT().RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	redisClient := container_builders.GetRedisConnection()
	t.Cleanup(func() { container_builders.CloseRedisConnection(redisClient) })

	token, err := DeployEip3009Token(
		context.Background(),
		client,
		container_builders.GetAnvilContainerBuilder().Create(),
		TokenEip712Name,
		TokenEip712Version,
	)
	require.NoError(t, err)

	return &IntegrationChain{
		ChainId:           relayer.ChainIdUint,
		Client:            client,
		ChainsSettings:    chainsSettings,
		EthClientFactory:  factory.NewEthClientFactory(chainsSettings, mockBlockchainRpcMetrics),
		TransactorFactory: factory.NewTransactorFactory(mockRelayerKeyStorage, zap.NewNop()),
		RedisStorage:      redis_services.NewRedisStorage(redisClient, mockRedisMetrics),
		RedisLockStorage:  redis_services.NewRedisLockStorage(redisClient, mockRedisMetrics),
		Relayer:           relayer,
		Token:             token,
	}
}

// FundedPayment builds a 1000-token payment whose payer holds balance tokens.
func (c *IntegrationChain) FundedPayment(t *testing.T, balance int64) *model_builders.PaymentRequestBuilder {
	builder := model_builders.NewPaymentRequestBuilder().
		WithChainId(c.ChainId).
		WithAsset(c.Token.Address).
		WithEip712Domain(TokenEip712Name, TokenEip712Version).
		WithAmount("1000")
	require.NoError(t, c.Token.Mint(builder.Payer(), big.NewInt(balance)))
	return builder
}

// FundedPermit2Payment is FundedPayment paid through Permit2, which the payer approved for the whole balance.
func (c *IntegrationChain) FundedPermit2Payment(t *testing.T, balance int64) *model_builders.PaymentRequestBuilder {
	require.NoError(t, InstallPermit2(context.Background(), c.Client))
	builder := c.FundedPayment(t, balance).WithPermit2()
	require.NoError(t, c.Token.SetAllowance(
		builder.Payer(),
		common.HexToAddress(helpers.Permit2Address),
		big.NewInt(balance),
	))
	return builder
}

// TransferCall is the settlement call the facilitator's strategy builds for the payment.
func (c *IntegrationChain) TransferCall(t *testing.T, payment *models.Payment) *models.TransferCall {
	var strategy x402_facilitator_interfaces.AssetTransferMethodStrategyInterface = x402_facilitator_services.
		NewEip3009AssetTransferMethodStrategy(zap.NewNop(), nil)
	if payment.Method == enums.AssetTransferMethodPermit2 {
		strategy = x402_facilitator_services.NewPermit2AssetTransferMethodStrategy(zap.NewNop(), nil, nil)
	}
	call, err := strategy.TransferCall(payment)
	require.NoError(t, err)
	return call
}
