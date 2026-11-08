package common_tests

import (
	"crypto/ecdsa"
	"crypto/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func RandomEthAddress() common.Address {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("error casting public key to ECDSA")
	}

	return crypto.PubkeyToAddress(*publicKeyECDSA)
}
