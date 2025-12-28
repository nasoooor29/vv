package notifications

import (
	"log/slog"
	"sync"
)

// NotificationLevel represents the severity level of a notification
type NotificationLevel string

const (
	LevelError   NotificationLevel = "ERROR"
	LevelWarning NotificationLevel = "WARNING"
	LevelInfo    NotificationLevel = "INFO"
	LevelSuccess NotificationLevel = "SUCCESS"
)

// Notification represents a notification message
type Notification struct {
	Level   NotificationLevel
	Title   string
	Message string
	Fields  map[string]string
	Group   string
	Version string
}

// Sender is the interface that all notification senders must implement
type Sender interface {
	// Name returns the sender's name for logging purposes
	Name() string
	// IsEnabled returns whether this sender is enabled for the given level
	IsEnabled(level NotificationLevel) bool
	// Send sends the notification
	Send(notification Notification) error
}

// Manager manages multiple notification senders
type Manager struct {
	senders []Sender
	mu      sync.RWMutex
}

// NewManager creates a new notification manager
func NewManager() *Manager {
	return &Manager{
		senders: []Sender{},
	}
}

// RegisterSender adds a sender to the manager
func (m *Manager) RegisterSender(sender Sender) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.senders = append(m.senders, sender)
	slog.Info("registered notification sender", "sender", sender.Name())
}

// ClearSenders removes all registered senders
func (m *Manager) ClearSenders() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.senders = []Sender{}
	slog.Info("cleared all notification senders")
}

// Send sends a notification to all enabled senders
func (m *Manager) Send(notification Notification) {
	m.mu.RLock()
	senders := make([]Sender, len(m.senders))
	copy(senders, m.senders)
	m.mu.RUnlock()

	for _, sender := range senders {
		if !sender.IsEnabled(notification.Level) {
			continue
		}

		go func(s Sender) {
			if err := s.Send(notification); err != nil {
				slog.Error("failed to send notification",
					"sender", s.Name(),
					"level", notification.Level,
					"err", err,
				)
			}
		}(sender)
	}
}

// SendError sends an error notification
func (m *Manager) SendError(title, message string, fields map[string]string, group, version string) {
	m.Send(Notification{
		Level:   LevelError,
		Title:   title,
		Message: message,
		Fields:  fields,
		Group:   group,
		Version: version,
	})
}

// SendWarning sends a warning notification
func (m *Manager) SendWarning(title, message string, fields map[string]string, group, version string) {
	m.Send(Notification{
		Level:   LevelWarning,
		Title:   title,
		Message: message,
		Fields:  fields,
		Group:   group,
		Version: version,
	})
}

// SendInfo sends an info notification
func (m *Manager) SendInfo(title, message string, fields map[string]string, group, version string) {
	m.Send(Notification{
		Level:   LevelInfo,
		Title:   title,
		Message: message,
		Fields:  fields,
		Group:   group,
		Version: version,
	})
}

// SendSuccess sends a success notification
func (m *Manager) SendSuccess(title, message string, fields map[string]string, group, version string) {
	m.Send(Notification{
		Level:   LevelSuccess,
		Title:   title,
		Message: message,
		Fields:  fields,
		Group:   group,
		Version: version,
	})
}
