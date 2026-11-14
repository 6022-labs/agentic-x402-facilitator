package container_builders

import (
	"context"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	redisContainer   *RedisContainerBuilder
	redisContainerMu sync.Mutex
)

type RedisContainerBuilder struct {
	container     testcontainers.Container
	address       string
	containerMu   sync.Mutex
	connectionsMu sync.Mutex
	connections   []*redis.Client
}

func GetRedisContainerBuilder() *RedisContainerBuilder {
	redisContainerMu.Lock()
	defer redisContainerMu.Unlock()

	if redisContainer == nil {
		redisContainer = &RedisContainerBuilder{}
	}
	return redisContainer
}

func (b *RedisContainerBuilder) start() string {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container == nil {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "redis:8-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			log.Fatalf("failed to start redis container: %v", err)
		}
		b.container = container

		host, _ := container.Host(ctx)
		port, _ := container.MappedPort(ctx, "6379")
		b.address = host + ":" + port.Port()
		log.Println("Redis container ready")
	}
	return b.address
}

func (b *RedisContainerBuilder) stop() {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container != nil {
		_ = b.container.Terminate(context.Background())
		b.container = nil
		log.Println("Redis container stopped")
	}

	redisContainerMu.Lock()
	redisContainer = nil
	redisContainerMu.Unlock()
}

// Create returns a client on the shared container; tests isolate themselves through their own keys.
func (b *RedisContainerBuilder) Create() *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: b.start()})

	b.connectionsMu.Lock()
	b.connections = append(b.connections, client)
	b.connectionsMu.Unlock()

	return client
}

func (b *RedisContainerBuilder) DropConnection(client *redis.Client) {
	b.connectionsMu.Lock()
	defer b.connectionsMu.Unlock()
	for i, connection := range b.connections {
		if connection == client {
			b.connections = append(b.connections[:i], b.connections[i+1:]...)
			_ = client.Close()
			break
		}
	}
}
