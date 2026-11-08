package services

import (
	"context"
	"errors"
	"net/http"
	"time"

	apperrors "github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/metrics/interfaces"
	vault_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_vault/settings"
	vault "github.com/hashicorp/vault/api"
)

type KvV1RelayerKeyStorage struct {
	client          *vault.Client
	vaultSettings   *vault_settings.VaultSettings
	relayerSettings *vault_settings.RelayerSettings
	vaultMetrics    metrics_interfaces.VaultMetricsInterface
}

func NewKvV1RelayerKeyStorage(
	client *vault.Client,
	vaultSettings *vault_settings.VaultSettings,
	relayerSettings *vault_settings.RelayerSettings,
	vaultMetrics metrics_interfaces.VaultMetricsInterface,
) *KvV1RelayerKeyStorage {
	return &KvV1RelayerKeyStorage{
		client:          client,
		vaultSettings:   vaultSettings,
		relayerSettings: relayerSettings,
		vaultMetrics:    vaultMetrics,
	}
}

func (s *KvV1RelayerKeyStorage) ReadPrivateKey(ctx context.Context) (privateKey string, err error) {
	defer s.recordOperation(ctx, "read", time.Now(), &err)

	secret, err := s.client.KVv1(s.vaultSettings.KvV1MountPath).Get(ctx, s.relayerSettings.VaultKey)
	if err != nil {
		var responseErr *vault.ResponseError
		if errors.Is(err, vault.ErrSecretNotFound) ||
			(errors.As(err, &responseErr) && responseErr.StatusCode == http.StatusNotFound) {
			return "", apperrors.ErrSecretNotFound
		}
		return "", err
	}
	if secret == nil || secret.Data == nil {
		return "", apperrors.ErrSecretNotFound
	}

	value, ok := secret.Data[vault_settings.RelayerPrivateKeyField].(string)
	if !ok || value == "" {
		return "", apperrors.ErrSecretNotFound
	}
	return value, nil
}

// recordOperation counts a missing secret as a failure: no settlement runs without it.
func (s *KvV1RelayerKeyStorage) recordOperation(ctx context.Context, operation string, start time.Time, err *error) {
	s.vaultMetrics.RecordOperation(ctx, operation, s.vaultSettings.KvV1MountPath, time.Since(start), *err)
}
