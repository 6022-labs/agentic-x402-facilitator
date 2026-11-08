package services_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	vault_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/services"
	vault_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/settings"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/container_builders"
	metrics_interfaces_mocks "github.com/6022-labs/agentic-x402-facilitator/tests/x402_facilitator_vault_mocks/metrics_mocks/interfaces_mocks"
	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const relayerVaultKey = "relayer"

type WhenReadingTheRelayerKeyTestingSuite struct {
	mountPath string
	client    *vault.Client
	sut       *vault_services.KvV1RelayerKeyStorage
}

func WhenReadingTheRelayerKeyBeforeEach(t *testing.T) *WhenReadingTheRelayerKeyTestingSuite {
	cfg := container_builders.GetVaultContainerBuilder().Create()

	vaultCfg := vault.DefaultConfig()
	vaultCfg.Address = cfg.Address
	client, err := vault.NewClient(vaultCfg)
	require.NoError(t, err)
	client.SetToken(cfg.Token)

	mockVaultMetrics := metrics_interfaces_mocks.NewMockVaultMetricsInterface(gomock.NewController(t))
	mockVaultMetrics.EXPECT().
		RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	return &WhenReadingTheRelayerKeyTestingSuite{
		mountPath: cfg.KvV1MountPath,
		client:    client,
		sut: vault_services.NewKvV1RelayerKeyStorage(
			client,
			&vault_settings.VaultSettings{Address: cfg.Address, Token: cfg.Token, KvV1MountPath: cfg.KvV1MountPath},
			&vault_settings.RelayerSettings{VaultKey: relayerVaultKey},
			mockVaultMetrics,
		),
	}
}

func WhenReadingTheRelayerKeyAfterEach(suite *WhenReadingTheRelayerKeyTestingSuite) {
	container_builders.GetVaultContainerBuilder().DropMount(suite.mountPath)
}

func (suite *WhenReadingTheRelayerKeyTestingSuite) store(t *testing.T, data map[string]any) {
	require.NoError(t, suite.client.KVv1(suite.mountPath).Put(context.Background(), relayerVaultKey, data))
}

func TestWhenReadingTheRelayerKey(t *testing.T) {
	t.Parallel()

	t.Run("Given no relayer secret is stored", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrSecretNotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheRelayerKeyBeforeEach(t)
			defer WhenReadingTheRelayerKeyAfterEach(suite)

			privateKey, err := suite.sut.ReadPrivateKey(context.Background())

			assert.Empty(t, privateKey)
			assert.ErrorIs(t, err, apperrors.ErrSecretNotFound)
		})
	})

	t.Run("Given the relayer secret has no private_key field", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrSecretNotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheRelayerKeyBeforeEach(t)
			defer WhenReadingTheRelayerKeyAfterEach(suite)
			suite.store(t, map[string]any{"address": common_tests.RandomEthAddress().Hex()})

			privateKey, err := suite.sut.ReadPrivateKey(context.Background())

			assert.Empty(t, privateKey)
			assert.ErrorIs(t, err, apperrors.ErrSecretNotFound)
		})
	})

	t.Run("Given the relayer secret is stored", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its private key", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheRelayerKeyBeforeEach(t)
			defer WhenReadingTheRelayerKeyAfterEach(suite)
			suite.store(t, map[string]any{"private_key": "0xabc123"})

			privateKey, err := suite.sut.ReadPrivateKey(context.Background())

			require.NoError(t, err)
			assert.Equal(t, "0xabc123", privateKey)
		})
	})
}
