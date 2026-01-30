// Package session provides session management for the TUI.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cooperations/internal/tui/stream"
)

// Session represents a saved TUI session.
type Session struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Task        string         `json:"task"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Status      string         `json:"status"` // "running", "paused", "complete", "error"
	Events      []SessionEvent `json:"events"`
	Checkpoints []Checkpoint   `json:"checkpoints"`
	Metrics     SessionMetrics `json:"metrics"`
}

// SessionEvent represents a recorded event in the session.
type SessionEvent struct {
	Timestamp time.Time   `json:"timestamp"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
}

// Checkpoint represents a point in the session that can be resumed from.
type Checkpoint struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Timestamp   time.Time `json:"timestamp"`
	EventIndex  int       `json:"event_index"`
	Description string    `json:"description"`
}

// SessionMetrics contains aggregate metrics for the session.
type SessionMetrics struct {
	TotalTokens      int           `json:"total_tokens"`
	EstimatedCostUSD float64       `json:"estimated_cost_usd"`
	Duration         time.Duration `json:"duration"`
	AgentCycles      int           `json:"agent_cycles"`
	HandoffCount     int           `json:"handoff_count"`
}

// Manager handles session persistence and replay.
type Manager struct {
	SessionDir  string
	Current     *Session
	EventBuffer []SessionEvent
}

// NewManager creates a new session manager.
func NewManager(sessionDir string) (*Manager, error) {
	// Create session directory if it doesn't exist
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}

	return &Manager{
		SessionDir: sessionDir,
	}, nil
}

// NewSession creates a new session.
func (m *Manager) NewSession(task string) *Session {
	session := &Session{
		ID:        generateSessionID(),
		Task:      task,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    "running",
	}
	m.Current = session
	m.EventBuffer = nil
	return session
}

// generateSessionID creates a unique session ID.
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// RecordEvent adds an event to the current session.
func (m *Manager) RecordEvent(eventType string, data interface{}) {
	if m.Current == nil {
		return
	}

	event := SessionEvent{
		Timestamp: time.Now(),
		Type:      eventType,
		Data:      data,
	}

	m.Current.Events = append(m.Current.Events, event)
	m.EventBuffer = append(m.EventBuffer, event)
	m.Current.UpdatedAt = time.Now()
}

// RecordStreamEvent records a stream event.
func (m *Manager) RecordStreamEvent(event interface{}) {
	switch e := event.(type) {
	case stream.TokenChunk:
		m.RecordEvent("token", e)
	case stream.ProgressUpdate:
		m.RecordEvent("progress", e)
	case stream.HandoffEvent:
		m.RecordEvent("handoff", e)
		if m.Current != nil {
			m.Current.Metrics.HandoffCount++
		}
	case stream.CodeUpdate:
		m.RecordEvent("code", e)
	case stream.FileDiff:
		m.RecordEvent("diff", e)
	case stream.AgentLogEntry:
		m.RecordEvent("log", e)
	case stream.MetricsSnapshot:
		m.RecordEvent("metrics", e)
		if m.Current != nil {
			m.Current.Metrics.TotalTokens = e.TotalTokens
			m.Current.Metrics.EstimatedCostUSD = e.EstimatedCostUSD
			m.Current.Metrics.AgentCycles = e.AgentCycles
		}
	case stream.ToastNotification:
		m.RecordEvent("toast", e)
	case stream.DecisionRequest:
		m.RecordEvent("decision", e)
	}
}

// CreateCheckpoint creates a checkpoint at the current position.
func (m *Manager) CreateCheckpoint(name, description string) *Checkpoint {
	if m.Current == nil {
		return nil
	}

	checkpoint := Checkpoint{
		ID:          fmt.Sprintf("cp_%d", time.Now().UnixNano()),
		Name:        name,
		Timestamp:   time.Now(),
		EventIndex:  len(m.Current.Events),
		Description: description,
	}

	m.Current.Checkpoints = append(m.Current.Checkpoints, checkpoint)
	return &checkpoint
}

// Save persists the current session to disk.
func (m *Manager) Save() error {
	if m.Current == nil {
		return fmt.Errorf("no current session")
	}

	m.Current.UpdatedAt = time.Now()
	m.Current.Metrics.Duration = time.Since(m.Current.CreatedAt)

	filename := filepath.Join(m.SessionDir, m.Current.ID+".json")
	data, err := json.MarshalIndent(m.Current, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("write session: %w", err)
	}

	return nil
}

// Load loads a session from disk.
func (m *Manager) Load(sessionID string) (*Session, error) {
	filename := filepath.Join(m.SessionDir, sessionID+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	m.Current = &session
	return &session, nil
}

// List returns all saved sessions.
func (m *Manager) List() ([]Session, error) {
	entries, err := os.ReadDir(m.SessionDir)
	if err != nil {
		return nil, fmt.Errorf("read session dir: %w", err)
	}

	var sessions []Session
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(m.SessionDir, entry.Name()))
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Delete removes a session from disk.
func (m *Manager) Delete(sessionID string) error {
	filename := filepath.Join(m.SessionDir, sessionID+".json")
	return os.Remove(filename)
}

// SetStatus updates the current session status.
func (m *Manager) SetStatus(status string) {
	if m.Current != nil {
		m.Current.Status = status
		m.Current.UpdatedAt = time.Now()
	}
}

// Replay replays events from a session to a stream.
func (m *Manager) Replay(session *Session, s *stream.WorkflowStream, speed float64) error {
	if len(session.Events) == 0 {
		return nil
	}

	startTime := session.Events[0].Timestamp
	replayStart := time.Now()

	for i, event := range session.Events {
		// Calculate delay based on original timing and replay speed
		if i > 0 {
			originalDelay := event.Timestamp.Sub(session.Events[i-1].Timestamp)
			replayDelay := time.Duration(float64(originalDelay) / speed)

			// Wait until the appropriate time
			targetTime := replayStart.Add(time.Duration(float64(event.Timestamp.Sub(startTime)) / speed))
			waitTime := time.Until(targetTime)
			if waitTime > 0 {
				time.Sleep(waitTime)
			}

			_ = replayDelay // unused but calculated for documentation
		}

		// Send the event to the stream
		m.replayEvent(event, s)
	}

	return nil
}

// replayEvent sends a recorded event to the stream.
func (m *Manager) replayEvent(event SessionEvent, s *stream.WorkflowStream) {
	switch event.Type {
	case "token":
		if data, ok := event.Data.(map[string]interface{}); ok {
			s.SendToken(stream.TokenChunk{
				AgentRole: getString(data, "agent_role"),
				Token:     getString(data, "token"),
				IsFinal:   getBool(data, "is_final"),
			})
		}

	case "progress":
		if data, ok := event.Data.(map[string]interface{}); ok {
			s.SendProgress(stream.ProgressUpdate{
				Percent: getFloat(data, "percent"),
				Stage:   getString(data, "stage"),
				Message: getString(data, "message"),
			})
		}

	case "handoff":
		if data, ok := event.Data.(map[string]interface{}); ok {
			s.SendHandoff(stream.HandoffEvent{
				From:   getString(data, "from"),
				To:     getString(data, "to"),
				Reason: getString(data, "reason"),
			})
		}

	case "code":
		if data, ok := event.Data.(map[string]interface{}); ok {
			s.SendCode(stream.CodeUpdate{
				Path:     getString(data, "path"),
				Language: getString(data, "language"),
				Content:  getString(data, "content"),
			})
		}

	case "log":
		if data, ok := event.Data.(map[string]interface{}); ok {
			s.SendLog(stream.AgentLogEntry{
				AgentRole: getString(data, "agent_role"),
				Level:     getString(data, "level"),
				Message:   getString(data, "message"),
			})
		}

	case "metrics":
		if data, ok := event.Data.(map[string]interface{}); ok {
			s.SendMetrics(stream.MetricsSnapshot{
				TotalTokens:      getInt(data, "total_tokens"),
				PromptTokens:     getInt(data, "prompt_tokens"),
				CompletionTokens: getInt(data, "completion_tokens"),
				EstimatedCostUSD: getFloat(data, "estimated_cost_usd"),
			})
		}

	case "toast":
		if data, ok := event.Data.(map[string]interface{}); ok {
			s.SendToast(stream.ToastNotification{
				Level:   getString(data, "level"),
				Message: getString(data, "message"),
			})
		}
	}
}

// Helper functions for type conversion
func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(data map[string]interface{}, key string) float64 {
	if v, ok := data[key].(float64); ok {
		return v
	}
	return 0
}

func getInt(data map[string]interface{}, key string) int {
	if v, ok := data[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getBool(data map[string]interface{}, key string) bool {
	if v, ok := data[key].(bool); ok {
		return v
	}
	return false
}
