package clientmanager

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"visory/internal/utils"

	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

type ClientInfo struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

type Docker struct {
	Dispatcher   *utils.Dispatcher
	Logger       *slog.Logger
	clients      map[int]*client.Client
	clientsMutex sync.RWMutex
	nextClientID int
}

// NewDockerService creates a new DockerService with dependency injection
func NewDockerClientManager(dispatcher *utils.Dispatcher, logger *slog.Logger) *Docker {
	return &Docker{
		Dispatcher:   dispatcher.WithGroup("dockerClientManager"),
		Logger:       logger.WithGroup("dockerClientManager"),
		clients:      make(map[int]*client.Client),
		nextClientID: 0,
	}
}

// RegisterClient registers a new Docker client and returns the assigned ID
func (s *Docker) RegisterClient(endpoint string, cli *client.Client) int {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	id := s.nextClientID
	s.clients[id] = cli
	s.nextClientID++

	s.Logger.Info("Docker client registered", "id", id, "endpoint", endpoint)
	return id
}

// GetClient retrieves a Docker client by ID from context
func (s *Docker) GetClient(c echo.Context) (*client.Client, error) {
	clientIDStr := c.Param("clientid")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		return nil, s.Dispatcher.NewInternalServerError("Invalid client ID", err)
	}

	s.clientsMutex.RLock()
	cli, exists := s.clients[clientID]
	s.clientsMutex.RUnlock()

	if !exists {
		return nil, s.Dispatcher.NewInternalServerError("Client not found", nil)
	}
	return cli, nil
}

// ListClients returns all available Docker clients
func (s *Docker) ListClients() []ClientInfo {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	clients := make([]ClientInfo, 0, len(s.clients))
	for id, cli := range s.clients {
		status := "connected"
		if cli == nil {
			status = "disconnected"
		}
		clients = append(clients, ClientInfo{
			ID:     id,
			Status: status,
		})
	}
	return clients
}

// WatchDogClients monitors the health of registered Docker clients forever
func (s *Docker) WatchDogClients() {
	for {
		var wg sync.WaitGroup
		for _, c := range s.clients {
			wg.Add(1)
			go func(c *client.Client) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_, err := c.Ping(ctx)
				if err != nil {
					s.Dispatcher.NewInternalServerError("Client health check failed", err)
				}
			}(c)
		}
		wg.Wait()
		// Sleep before next health check iteration
		time.Sleep(5 * time.Second)
	}
}

// GetAvailableClients returns all registered Docker clients
func (s *Docker) GetAvailableClients(c echo.Context) error {
	clients := s.ListClients()
	return c.JSON(http.StatusOK, clients)
}

// func (s *Docker) IsValidClient(clientId string) (*client.Client, error) {
// 	s.clientsMutex.RLock()
// 	defer s.clientsMutex.RUnlock()
// 	id, err := strconv.Atoi(clientId)
// 	if err != nil {
// 		return nil, s.Dispatcher.NewInternalServerError("Invalid client ID", err)
// 	}
// 	client, exists := s.clients[id]
// 	if !exists {
// 		return nil, s.Dispatcher.NewInternalServerError("Client not found", nil)
// 	}
// 	return client, nil
// }

// initializeDockerClients reads Docker client configurations from environment variables
// and registers them with the Docker service
func (s *Docker) InitializeDockerClients() {
	// Look for VISORY_DOCKER_CLIENT_<ID> environment variables
	for i := range 100 {
		i++ // we start from 1 because 0 is reserved for default client
		envKey := fmt.Sprintf("VISORY_DOCKER_CLIENT_%d", i)
		endpoint := os.Getenv(envKey)
		if endpoint == "" {
			continue
		}

		endpoint = strings.TrimSpace(endpoint)

		// Create Docker client
		cli, err := client.NewClientWithOpts(
			client.WithHost(endpoint),
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			s.Logger.Error("failed to create docker client", "id", i, "endpoint", endpoint, "error", err)
			continue
		}

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err = cli.Ping(ctx)
		cancel()
		if err != nil {
			s.Logger.Error("failed to ping docker client", "id", i, "endpoint", endpoint, "error", err)
			cli.Close()
			continue
		}

		s.RegisterClient(endpoint, cli)
	}
}

func (s *Docker) InitilizeDefaultDockerClient() {
	// Create default Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		s.Logger.Error("failed to create default docker client", "error", err)
		return
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = cli.Ping(ctx)
	cancel()
	if err != nil {
		s.Logger.Error("failed to ping default docker client", "error", err)
		cli.Close()
		return
	}
	s.RegisterClient("0", cli)
}
