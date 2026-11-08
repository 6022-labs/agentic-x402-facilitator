package interfaces

import "context"

// RelayerKeyStorageInterface reads the relayer's private key from the secret store.
type RelayerKeyStorageInterface interface {
	// ReadPrivateKey returns the hex private key, or common/errors.ErrSecretNotFound when none is stored.
	ReadPrivateKey(ctx context.Context) (string, error)
}
