package container_builders

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	vault "github.com/hashicorp/vault/api"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const vaultDevToken = "root"

// VaultMountSettings is what the builder hands a test: a fresh KV v1 mount in the shared Vault plus address and token.
type VaultMountSettings struct {
	Address       string
	Token         string
	KvV1MountPath string
}

var (
	vaultContainer   *VaultContainerBuilder
	vaultContainerMu sync.Mutex
)

type VaultContainerBuilder struct {
	container   testcontainers.Container
	address     string
	client      *vault.Client
	containerMu sync.Mutex
	mountsMu    sync.Mutex
	mounts      []string
}

func GetVaultContainerBuilder() *VaultContainerBuilder {
	vaultContainerMu.Lock()
	defer vaultContainerMu.Unlock()

	if vaultContainer == nil {
		vaultContainer = &VaultContainerBuilder{}
	}
	return vaultContainer
}

func (b *VaultContainerBuilder) start() *vault.Client {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container == nil {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "hashicorp/vault:1.21",
			ExposedPorts: []string{"8200/tcp"},
			Env: map[string]string{
				"VAULT_DEV_ROOT_TOKEN_ID":  vaultDevToken,
				"VAULT_DEV_LISTEN_ADDRESS": "0.0.0.0:8200",
			},
			Cmd:        []string{"server", "-dev", "-dev-listen-address=0.0.0.0:8200"},
			WaitingFor: wait.ForListeningPort("8200/tcp"),
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			log.Fatalf("failed to start vault container: %v", err)
		}
		b.container = container

		host, _ := container.Host(ctx)
		port, _ := container.MappedPort(ctx, "8200")
		b.address = fmt.Sprintf("http://%s:%s", host, port.Port())

		cfg := vault.DefaultConfig()
		cfg.Address = b.address
		client, err := vault.NewClient(cfg)
		if err != nil {
			log.Fatalf("failed to build vault client: %v", err)
		}
		client.SetToken(vaultDevToken)
		b.client = client

		log.Println("Vault container ready")
	}
	return b.client
}

func (b *VaultContainerBuilder) stop() {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container != nil {
		_ = b.container.Terminate(context.Background())
		b.container = nil
		log.Println("Vault container stopped")
	}
	vaultContainerMu.Lock()
	vaultContainer = nil
	vaultContainerMu.Unlock()
}

// Create mounts a fresh KV v1 engine at a unique path so parallel tests don't trample each other.
func (b *VaultContainerBuilder) Create() VaultMountSettings {
	client := b.start()

	mountPath := "kvv1_" + uuid.New().String()[:8]
	if err := client.Sys().Mount(mountPath, &vault.MountInput{
		Type:    "kv",
		Options: map[string]string{"version": "1"},
	}); err != nil {
		log.Fatalf("failed to mount kv v1 at %s: %v", mountPath, err)
	}

	b.mountsMu.Lock()
	b.mounts = append(b.mounts, mountPath)
	b.mountsMu.Unlock()

	return VaultMountSettings{
		Address:       b.address,
		Token:         vaultDevToken,
		KvV1MountPath: mountPath,
	}
}

func (b *VaultContainerBuilder) DropMount(mountPath string) {
	if b.client != nil {
		_ = b.client.Sys().Unmount(mountPath)
	}
	b.mountsMu.Lock()
	defer b.mountsMu.Unlock()
	for i, m := range b.mounts {
		if m == mountPath {
			b.mounts = append(b.mounts[:i], b.mounts[i+1:]...)
			break
		}
	}
}
