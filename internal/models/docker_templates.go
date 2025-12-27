package models

// PortainerTemplate represents a single template from the Portainer templates JSON
type PortainerTemplate struct {
	Type          int              `json:"type"`
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	Categories    []string         `json:"categories"`
	Platform      string           `json:"platform"`
	Logo          string           `json:"logo"`
	Image         string           `json:"image"`
	Name          string           `json:"name"`
	Registry      string           `json:"registry"`
	Command       string           `json:"command"`
	Network       string           `json:"network"`
	Privileged    bool             `json:"privileged"`
	Interactive   bool             `json:"interactive"`
	Hostname      string           `json:"hostname"`
	Note          string           `json:"note"`
	Ports         []string         `json:"ports"`
	Volumes       []TemplateVolume `json:"volumes"`
	Env           []TemplateEnv    `json:"env"`
	Labels        []TemplateLabel  `json:"labels"`
	RestartPolicy string           `json:"restart_policy"`
}

// TemplateVolume represents a volume mapping in a template
type TemplateVolume struct {
	Container string `json:"container"`
	Bind      string `json:"bind"`
	ReadOnly  bool   `json:"readonly"`
}

// TemplateEnv represents an environment variable in a template
type TemplateEnv struct {
	Name        string      `json:"name"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Default     string      `json:"default"`
	Preset      bool        `json:"preset"`
	Select      []EnvSelect `json:"select"`
}

// EnvSelect represents a select option for environment variables
type EnvSelect struct {
	Text    string `json:"text"`
	Value   string `json:"value"`
	Default bool   `json:"default"`
}

// TemplateLabel represents a label in a template
type TemplateLabel struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// TemplatesResponse represents the response from the Portainer templates API
type TemplatesResponse struct {
	Version   string              `json:"version"`
	Templates []PortainerTemplate `json:"templates"`
}

// DeployTemplateRequest represents a request to deploy a template
type DeployTemplateRequest struct {
	Name          string            `json:"name"`           // Container name (optional, uses template name if empty)
	Env           map[string]string `json:"env"`            // Environment variable overrides
	Ports         []string          `json:"ports"`          // Port mapping overrides
	Volumes       []TemplateVolume  `json:"volumes"`        // Volume mapping overrides
	Network       string            `json:"network"`        // Network to attach to
	RestartPolicy string            `json:"restart_policy"` // Restart policy override
}

// TemplateListItem represents a simplified template for listing
type TemplateListItem struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
	Platform    string   `json:"platform"`
	Logo        string   `json:"logo"`
	Image       string   `json:"image"`
}

// DeployResponse represents the response after deploying a template
type DeployResponse struct {
	ContainerID string `json:"container_id"`
	Name        string `json:"name"`
	Message     string `json:"message"`
}
