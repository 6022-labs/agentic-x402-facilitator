package helpers

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// RecoverSigner recovers a 65-byte signature's signer, with a v of 27/28 or 0/1 as wallets emit either.
func RecoverSigner(digest []byte, signature []byte) (common.Address, error) {
	normalized := make([]byte, len(signature))
	copy(normalized, signature)
	if normalized[64] >= 27 {
		normalized[64] -= 27
	}
	publicKey, err := crypto.SigToPub(digest, normalized)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*publicKey), nil
}
