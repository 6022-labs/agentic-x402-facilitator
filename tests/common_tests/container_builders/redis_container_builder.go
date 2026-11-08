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
	redisContainer     testcontainers.Container
	redisAddr          string
	redisContainerMu   sync.Mutex
	redisConnectionsMu sync.Mutex
	redisConnections   []*redis.Client
)

func startRedis() string {
	redisContainerMu.Lock()
	defer redisContainerMu.Unlock()

	if redisContainer == nil {
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
			log.Fatalf("Failed to start Redis container: %v", err)
		}
		redisContainer = container

		host, _ := container.Host(ctx)
		port, _ := container.MappedPort(ctx, "6379")
		redisAddr = host + ":" + port.Port()

		log.Println("Redis container started and ready")
	}

	return redisAddr
}

func stopRedis() {
	redisContainerMu.Lock()
	defer redisContainerMu.Unlock()

	if redisContainer != nil {
		redisContainer.Terminate(context.Background())
		redisContainer = nil

		log.Println("Redis container stopped")
	}
}

func GetRedisConnection() *redis.Client {
	addr := startRedis()

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	redisConnectionsMu.Lock()
	redisConnections = append(redisConnections, client)
	redisConnectionsMu.Unlock()

	return client
}

func CloseRedisConnection(client *redis.Client) {
	// Just remove from the list, the container will be destroyed after all tests
	for i, connection := range redisConnections {
		if connection == client {
			redisConnectionsMu.Lock()
			redisConnections = append(redisConnections[:i], redisConnections[i+1:]...)
			redisConnectionsMu.Unlock()
			break
		}
	}

	log.Printf("Redis client removed from tracking")
}
