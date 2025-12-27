package services

import (
	"encoding/json"
	"log/slog"
	"net/http"

	clientmanager "visory/internal/clientManager"
	"visory/internal/utils"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/labstack/echo/v4"
)

type ClientInfo struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

type DockerService struct {
	Dispatcher    *utils.Dispatcher
	Logger        *slog.Logger
	ClientManager *clientmanager.Docker
}

// NewDockerService creates a new DockerService with dependency injection
func NewDockerService(dispatcher *utils.Dispatcher, logger *slog.Logger) *DockerService {
	service := &DockerService{
		Dispatcher:    dispatcher.WithGroup("docker"),
		Logger:        logger.WithGroup("docker"),
		ClientManager: clientmanager.NewDockerClientManager(dispatcher, logger),
	}
	service.ClientManager.InitializeDockerClients()
	service.ClientManager.InitilizeDefaultDockerClient()
	return service
}

// GetAvailableClients returns all registered Docker clients
func (s *DockerService) GetAvailableClients(c echo.Context) error {
	clients := s.ClientManager.ListClients()
	return c.JSON(http.StatusOK, clients)
}

// ListContainers returns list of containers
func (s *DockerService) ListContainers(c echo.Context) error {
	ctx := c.Request().Context()
	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:    true,
		Latest: true,
	})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to list containers", err)
	}

	return c.JSON(http.StatusOK, containers)
}

// ListImages returns list of images
func (s *DockerService) ListImages(c echo.Context) error {
	ctx := c.Request().Context()
	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	images, err := cli.ImageList(ctx, image.ListOptions{
		All:            true,
		SharedSize:     true,
		ContainerCount: true,
		Manifests:      true,
	})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to list images", err)
	}

	return c.JSON(http.StatusOK, images)
}

// DeleteImage removes an image
func (s *DockerService) DeleteImage(c echo.Context) error {
	ctx := c.Request().Context()
	imageID := c.Param("id")

	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	_, err = cli.ImageRemove(ctx, imageID, image.RemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to delete image", err)
	}

	return c.NoContent(http.StatusNoContent)
}

// InspectContainer returns detailed information about a container
func (s *DockerService) InspectContainer(c echo.Context) error {
	ctx := c.Request().Context()
	containerID := c.Param("id")

	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to inspect container", err)
	}

	return c.JSON(http.StatusOK, containerInfo)
}

// ContainerStats returns container statistics
func (s *DockerService) ContainerStats(c echo.Context) error {
	ctx := c.Request().Context()
	containerID := c.Param("id")

	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	stats, err := cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to get container stats", err)
	}
	defer stats.Body.Close()

	var statsBody container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&statsBody); err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to decode stats", err)
	}

	return c.JSON(http.StatusOK, statsBody)
}

// ContainerStatsStream streams container statistics
func (s *DockerService) ContainerStatsStream(c echo.Context) error {
	ctx := c.Request().Context()
	containerID := c.Param("id")

	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	stats, err := cli.ContainerStats(ctx, containerID, true)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to get container stats stream", err)
	}
	defer stats.Body.Close()

	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	return c.Stream(http.StatusOK, "application/json", stats.Body)
}

// ContainerLogs returns container logs
func (s *DockerService) ContainerLogs(c echo.Context) error {
	ctx := c.Request().Context()
	containerID := c.Param("id")

	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	logs, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     false,
	})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to get container logs", err)
	}
	defer logs.Close()

	c.Response().Header().Set(echo.HeaderContentType, "text/plain")
	return c.Stream(http.StatusOK, "text/plain", logs)
}

// CreateContainer creates a new container
func (s *DockerService) CreateContainer(c echo.Context) error {
	ctx := c.Request().Context()

	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	var config container.Config
	if err := c.Bind(&config); err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to bind request", err)
	}

	containerName := c.QueryParam("name")

	resp, err := cli.ContainerCreate(ctx, &config, nil, nil, nil, containerName)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to create container", err)
	}

	return c.JSON(http.StatusCreated, resp)
}

// StartContainer starts a container
func (s *DockerService) StartContainer(c echo.Context) error {
	ctx := c.Request().Context()
	containerID := c.Param("id")

	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	err = cli.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to start container", err)
	}

	return c.NoContent(http.StatusOK)
}

// StopContainer stops a container
func (s *DockerService) StopContainer(c echo.Context) error {
	ctx := c.Request().Context()
	containerID := c.Param("id")

	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	err = cli.ContainerStop(ctx, containerID, container.StopOptions{})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to stop container", err)
	}

	return c.NoContent(http.StatusOK)
}

// RestartContainer restarts a container
func (s *DockerService) RestartContainer(c echo.Context) error {
	ctx := c.Request().Context()
	containerID := c.Param("id")

	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	err = cli.ContainerRestart(ctx, containerID, container.StopOptions{})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to restart container", err)
	}

	return c.NoContent(http.StatusOK)
}

// DeleteContainer removes a container
func (s *DockerService) DeleteContainer(c echo.Context) error {
	ctx := c.Request().Context()
	containerID := c.Param("id")

	cli, err := s.ClientManager.GetClient(c)
	if err != nil {
		return err
	}

	err = cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to delete container", err)
	}

	return c.NoContent(http.StatusNoContent)
}

// ValidateDockerClientMiddleware validates that the Docker client ID exists
func (s *DockerService) ValidateDockerClientMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := s.ClientManager.GetClient(c)
		if err != nil {
			return err
		}

		return next(c)
	}
}
