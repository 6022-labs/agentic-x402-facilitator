package container_builders

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	anvilChainId     = "31337"
	anvilChainIdUint = uint64(31337)
)

// AnvilAccountSettings is what the builder hands a test: the shared anvil chain plus its own pre-funded key.
type AnvilAccountSettings struct {
	ChainId       string
	ChainIdUint   uint64
	HttpUrl       string
	PrivateKeyHex string
	Address       common.Address
}

var (
	anvilContainer   *AnvilContainerBuilder
	anvilContainerMu sync.Mutex
)

type AnvilContainerBuilder struct {
	container   testcontainers.Container
	httpUrl     string
	rpcClient   *rpc.Client
	containerMu sync.Mutex
}

// GetAnvilContainerBuilder is called from parallel tests, so the singleton itself needs guarding.
func GetAnvilContainerBuilder() *AnvilContainerBuilder {
	anvilContainerMu.Lock()
	defer anvilContainerMu.Unlock()

	if anvilContainer == nil {
		anvilContainer = &AnvilContainerBuilder{}
	}
	return anvilContainer
}

func (b *AnvilContainerBuilder) start() *rpc.Client {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container == nil {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "ghcr.io/foundry-rs/foundry:v1.7.1",
			ExposedPorts: []string{"8545/tcp"},
			// The image entrypoint is a shell, so anvil has to be the explicit entrypoint.
			Entrypoint: []string{"anvil", "--host", "0.0.0.0"},
			WaitingFor: wait.ForListeningPort("8545/tcp"),
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			log.Fatalf("failed to start anvil container: %v", err)
		}
		b.container = container

		host, _ := container.Host(ctx)
		port, _ := container.MappedPort(ctx, "8545")
		b.httpUrl = fmt.Sprintf("http://%s:%s", host, port.Port())

		rpcClient, err := rpc.DialContext(ctx, b.httpUrl)
		if err != nil {
			log.Fatalf("failed to dial anvil at %s: %v", b.httpUrl, err)
		}
		var chainId string
		if err := rpcClient.CallContext(ctx, &chainId, "eth_chainId"); err != nil {
			log.Fatalf("anvil did not answer eth_chainId: %v", err)
		}
		b.rpcClient = rpcClient

		log.Println("Anvil container ready")
	}
	return b.rpcClient
}

func (b *AnvilContainerBuilder) stop() {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container != nil {
		if b.rpcClient != nil {
			b.rpcClient.Close()
			b.rpcClient = nil
		}
		_ = b.container.Terminate(context.Background())
		b.container = nil
		log.Println("Anvil container stopped")
	}

	anvilContainerMu.Lock()
	anvilContainer = nil
	anvilContainerMu.Unlock()
}

// Create generates a funded key so parallel tests never share a nonce sequence.
func (b *AnvilContainerBuilder) Create() AnvilAccountSettings {
	rpcClient := b.start()

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("failed to generate anvil test key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	balance := new(big.Int).Exp(big.NewInt(10), big.NewInt(22), nil)
	if err := rpcClient.CallContext(
		context.Background(),
		nil,
		"anvil_setBalance",
		address,
		hexutil.EncodeBig(balance),
	); err != nil {
		log.Fatalf("failed to fund %s on anvil: %v", address.Hex(), err)
	}

	return AnvilAccountSettings{
		ChainId:       anvilChainId,
		ChainIdUint:   anvilChainIdUint,
		HttpUrl:       b.httpUrl,
		PrivateKeyHex: common.Bytes2Hex(crypto.FromECDSA(privateKey)),
		Address:       address,
	}
}
