package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Discord embed colors
const (
	discordColorError   = 0xFF0000 // Red
	discordColorWarning = 0xFFA500 // Orange
	discordColorInfo    = 0x00BFFF // Light Blue
	discordColorSuccess = 0x00FF00 // Green
)

// DiscordConfig holds configuration for the Discord sender
type DiscordConfig struct {
	WebhookURL    string
	NotifyOnError bool
	NotifyOnWarn  bool
	NotifyOnInfo  bool
	Username      string
}

// DiscordSender implements the Sender interface for Discord webhooks
type DiscordSender struct {
	config DiscordConfig
}

// Discord webhook message types
type discordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []discordEmbedField `json:"fields,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	Footer      *discordFooter      `json:"footer,omitempty"`
}

type discordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type discordFooter struct {
	Text string `json:"text"`
}

type discordWebhookMessage struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []discordEmbed `json:"embeds,omitempty"`
}

// NewDiscordSender creates a new Discord notification sender
func NewDiscordSender(config DiscordConfig) *DiscordSender {
	if config.Username == "" {
		config.Username = "Visory Notifications"
	}
	return &DiscordSender{config: config}
}

// Name returns the sender's name
func (d *DiscordSender) Name() string {
	return "discord"
}

// IsEnabled returns whether Discord notifications are enabled for the given level
func (d *DiscordSender) IsEnabled(level NotificationLevel) bool {
	if d.config.WebhookURL == "" {
		return false
	}

	switch level {
	case LevelError:
		return d.config.NotifyOnError
	case LevelWarning:
		return d.config.NotifyOnWarn
	case LevelInfo:
		return d.config.NotifyOnInfo
	case LevelSuccess:
		return true // Success is always enabled if webhook is configured
	default:
		return false
	}
}

// Send sends a notification to Discord
func (d *DiscordSender) Send(notification Notification) error {
	if d.config.WebhookURL == "" {
		return nil
	}

	color := d.getColor(notification.Level)

	embedFields := make([]discordEmbedField, 0, len(notification.Fields))
	for name, value := range notification.Fields {
		embedFields = append(embedFields, discordEmbedField{
			Name:   name,
			Value:  value,
			Inline: true,
		})
	}

	group := notification.Group
	if group == "" {
		group = "Visory"
	}

	embed := discordEmbed{
		Title:       fmt.Sprintf("[%s] %s", notification.Level, notification.Title),
		Description: notification.Message,
		Color:       color,
		Fields:      embedFields,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &discordFooter{
			Text: fmt.Sprintf("Visory - %s | %s", group, notification.Version),
		},
	}

	msg := discordWebhookMessage{
		Username: d.config.Username,
		Embeds:   []discordEmbed{embed},
	}

	return d.sendWebhook(msg)
}

func (d *DiscordSender) getColor(level NotificationLevel) int {
	switch level {
	case LevelError:
		return discordColorError
	case LevelWarning:
		return discordColorWarning
	case LevelInfo:
		return discordColorInfo
	case LevelSuccess:
		return discordColorSuccess
	default:
		return discordColorInfo
	}
}

func (d *DiscordSender) sendWebhook(msg discordWebhookMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal discord message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create discord webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send discord webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}
