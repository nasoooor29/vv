package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	clientmanager "visory/internal/clientManager"
	"visory/internal/models"
	"visory/internal/utils"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/labstack/echo/v4"
)

const (
	templatesURL  = "https://raw.githubusercontent.com/Lissy93/portainer-templates/main/templates_v3.json"
	cacheDuration = 1 * time.Hour
)

type TemplatesService struct {
	Dispatcher    *utils.Dispatcher
	Logger        *slog.Logger
	clientManager *clientmanager.Docker

	// Cache
	cacheMutex      sync.RWMutex
	cachedTemplates []models.PortainerTemplate
	cacheExpiry     time.Time
}

// NewTemplatesService creates a new TemplatesService with dependency injection
func NewTemplatesService(dispatcher *utils.Dispatcher, logger *slog.Logger, clientManager *clientmanager.Docker) *TemplatesService {
	return &TemplatesService{
		Dispatcher:    dispatcher.WithGroup("templates"),
		Logger:        logger.WithGroup("templates"),
		clientManager: clientManager,
	}
}

// fetchTemplates fetches templates from the remote URL
func (s *TemplatesService) fetchTemplates() ([]models.PortainerTemplate, error) {
	s.cacheMutex.RLock()
	if time.Now().Before(s.cacheExpiry) && len(s.cachedTemplates) > 0 {
		templates := s.cachedTemplates
		s.cacheMutex.RUnlock()
		return templates, nil
	}
	s.cacheMutex.RUnlock()

	// Fetch from remote
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", templatesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch templates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var templatesResponse models.TemplatesResponse
	if err := json.Unmarshal(body, &templatesResponse); err != nil {
		return nil, fmt.Errorf("failed to parse templates JSON: %w", err)
	}

	// Update cache
	s.cacheMutex.Lock()
	s.cachedTemplates = templatesResponse.Templates
	s.cacheExpiry = time.Now().Add(cacheDuration)
	s.cacheMutex.Unlock()

	s.Logger.Info("Templates fetched and cached", "count", len(templatesResponse.Templates))
	return templatesResponse.Templates, nil
}

// ListTemplates returns all available templates
func (s *TemplatesService) ListTemplates(c echo.Context) error {
	templates, err := s.fetchTemplates()
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to fetch templates", err)
	}

	// Convert to list items with IDs
	listItems := make([]models.TemplateListItem, 0, len(templates))
	for i, t := range templates {
		// Only include container templates (type 1)
		if t.Type != 1 {
			continue
		}
		listItems = append(listItems, models.TemplateListItem{
			ID:          i,
			Title:       t.Title,
			Description: t.Description,
			Categories:  t.Categories,
			Platform:    t.Platform,
			Logo:        t.Logo,
			Image:       t.Image,
		})
	}

	return c.JSON(http.StatusOK, listItems)
}

// GetTemplate returns a single template by ID
func (s *TemplatesService) GetTemplate(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid template ID")
	}

	templates, err := s.fetchTemplates()
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to fetch templates", err)
	}

	if id < 0 || id >= len(templates) {
		return echo.NewHTTPError(http.StatusNotFound, "Template not found")
	}

	return c.JSON(http.StatusOK, templates[id])
}

// DeployTemplate deploys a template to a Docker client
func (s *TemplatesService) DeployTemplate(c echo.Context) error {
	ctx := c.Request().Context()

	// Get template ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid template ID")
	}

	// Get Docker client
	cli, err := s.clientManager.GetClient(c)
	if err != nil {
		return err
	}

	// Get templates
	templates, err := s.fetchTemplates()
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to fetch templates", err)
	}

	if id < 0 || id >= len(templates) {
		return echo.NewHTTPError(http.StatusNotFound, "Template not found")
	}

	template := templates[id]

	// Parse request body for overrides
	var deployReq models.DeployTemplateRequest
	if err := c.Bind(&deployReq); err != nil {
		// It's okay if body is empty, use defaults
		deployReq = models.DeployTemplateRequest{}
	}

	// Determine container name
	containerName := deployReq.Name
	if containerName == "" {
		containerName = template.Name
		if containerName == "" {
			containerName = template.Title
		}
	}

	// Pull image first
	s.Logger.Info("Pulling image", "image", template.Image)
	reader, err := cli.ImagePull(ctx, template.Image, image.PullOptions{})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to pull image", err)
	}
	defer reader.Close()
	// Consume the reader to complete the pull
	_, _ = io.Copy(io.Discard, reader)

	// Build environment variables
	env := make([]string, 0)
	for _, e := range template.Env {
		value := e.Default
		// Check for overrides
		if override, exists := deployReq.Env[e.Name]; exists {
			value = override
		}
		if value != "" {
			env = append(env, fmt.Sprintf("%s=%s", e.Name, value))
		}
	}

	// Build port bindings
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}

	ports := template.Ports
	if len(deployReq.Ports) > 0 {
		ports = deployReq.Ports
	}

	for _, p := range ports {
		// Parse port mapping (format: "hostPort:containerPort" or "hostPort:containerPort/protocol")
		portMapping, err := nat.ParsePortSpec(p)
		if err != nil {
			s.Logger.Warn("Failed to parse port mapping", "port", p, "error", err)
			continue
		}
		for _, pm := range portMapping {
			exposedPorts[pm.Port] = struct{}{}
			portBindings[pm.Port] = append(portBindings[pm.Port], pm.Binding)
		}
	}

	// Build volume bindings
	binds := make([]string, 0)
	volumes := template.Volumes
	if len(deployReq.Volumes) > 0 {
		volumes = deployReq.Volumes
	}

	for _, v := range volumes {
		bind := v.Bind
		if bind == "" {
			// Use a named volume if no host path specified
			bind = fmt.Sprintf("%s_%s", containerName, v.Container)
		}
		bindStr := fmt.Sprintf("%s:%s", bind, v.Container)
		if v.ReadOnly {
			bindStr += ":ro"
		}
		binds = append(binds, bindStr)
	}

	// Determine restart policy
	restartPolicy := template.RestartPolicy
	if deployReq.RestartPolicy != "" {
		restartPolicy = deployReq.RestartPolicy
	}
	if restartPolicy == "" {
		restartPolicy = "unless-stopped"
	}

	// Build container config
	config := &container.Config{
		Image:        template.Image,
		Env:          env,
		ExposedPorts: exposedPorts,
		Labels:       make(map[string]string),
	}

	// Add labels
	for _, l := range template.Labels {
		config.Labels[l.Name] = l.Value
	}
	config.Labels["visory.template.title"] = template.Title
	config.Labels["visory.template.id"] = strconv.Itoa(id)

	// Add command if specified
	if template.Command != "" {
		config.Cmd = []string{template.Command}
	}

	// Build host config
	hostConfig := &container.HostConfig{
		PortBindings:  portBindings,
		Binds:         binds,
		Privileged:    template.Privileged,
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyMode(restartPolicy)},
	}

	// Build network config
	networkConfig := &network.NetworkingConfig{}
	networkName := deployReq.Network
	if networkName == "" {
		networkName = template.Network
	}
	if networkName != "" {
		networkConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			networkName: {},
		}
	}

	// Create container
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, containerName)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to create container", err)
	}

	// Start container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to start container", err)
	}

	s.Logger.Info("Template deployed successfully", "template", template.Title, "container_id", resp.ID)

	return c.JSON(http.StatusCreated, models.DeployResponse{
		ContainerID: resp.ID,
		Name:        containerName,
		Message:     fmt.Sprintf("Successfully deployed %s", template.Title),
	})
}

// RefreshCache forces a refresh of the templates cache
func (s *TemplatesService) RefreshCache(c echo.Context) error {
	// Clear cache
	s.cacheMutex.Lock()
	s.cacheExpiry = time.Time{}
	s.cacheMutex.Unlock()

	// Fetch fresh templates
	templates, err := s.fetchTemplates()
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to refresh templates cache", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Cache refreshed successfully",
		"count":   len(templates),
	})
}

// GetCategories returns all unique categories from templates
func (s *TemplatesService) GetCategories(c echo.Context) error {
	templates, err := s.fetchTemplates()
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to fetch templates", err)
	}

	categorySet := make(map[string]bool)
	for _, t := range templates {
		for _, cat := range t.Categories {
			categorySet[cat] = true
		}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}

	return c.JSON(http.StatusOK, categories)
}
