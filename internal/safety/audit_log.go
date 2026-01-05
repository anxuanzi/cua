package safety

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// AuditLevel represents the importance level of an audit entry.
type AuditLevel string

const (
	AuditLevelInfo    AuditLevel = "INFO"
	AuditLevelWarning AuditLevel = "WARNING"
	AuditLevelError   AuditLevel = "ERROR"
	AuditLevelAction  AuditLevel = "ACTION"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       AuditLevel             `json:"level"`
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	Target      string                 `json:"target,omitempty"`
	Result      string                 `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AuditLogger records all agent actions for security and debugging.
type AuditLogger struct {
	mu      sync.Mutex
	writer  io.Writer
	entries []AuditEntry
	maxSize int // Maximum entries to keep in memory
}

// NewAuditLogger creates a new audit logger.
// If writer is nil, logs are only kept in memory.
func NewAuditLogger(writer io.Writer, maxSize int) *AuditLogger {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &AuditLogger{
		writer:  writer,
		entries: make([]AuditEntry, 0, 100),
		maxSize: maxSize,
	}
}

// NewFileAuditLogger creates an audit logger that writes to a file.
func NewFileAuditLogger(path string) (*AuditLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	return NewAuditLogger(file, 1000), nil
}

// Log records an audit entry.
func (a *AuditLogger) Log(entry AuditEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Add to in-memory buffer
	if len(a.entries) >= a.maxSize {
		// Remove oldest entries (keep last 75%)
		keepFrom := a.maxSize / 4
		a.entries = a.entries[keepFrom:]
	}
	a.entries = append(a.entries, entry)

	// Write to output if configured
	if a.writer != nil {
		data, err := json.Marshal(entry)
		if err == nil {
			a.writer.Write(data)
			a.writer.Write([]byte("\n"))
		}
	}
}

// LogAction is a convenience method for logging an action.
func (a *AuditLogger) LogAction(action, description, target string) {
	a.Log(AuditEntry{
		Level:       AuditLevelAction,
		Action:      action,
		Description: description,
		Target:      target,
	})
}

// LogActionResult logs an action with its result.
func (a *AuditLogger) LogActionResult(action, description, target, result string, err error) {
	entry := AuditEntry{
		Level:       AuditLevelAction,
		Action:      action,
		Description: description,
		Target:      target,
		Result:      result,
	}
	if err != nil {
		entry.Error = err.Error()
		entry.Level = AuditLevelError
	}
	a.Log(entry)
}

// LogWarning logs a warning.
func (a *AuditLogger) LogWarning(description string, metadata map[string]interface{}) {
	a.Log(AuditEntry{
		Level:       AuditLevelWarning,
		Description: description,
		Metadata:    metadata,
	})
}

// LogError logs an error.
func (a *AuditLogger) LogError(description string, err error) {
	a.Log(AuditEntry{
		Level:       AuditLevelError,
		Description: description,
		Error:       err.Error(),
	})
}

// GetEntries returns a copy of all entries.
func (a *AuditLogger) GetEntries() []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := make([]AuditEntry, len(a.entries))
	copy(result, a.entries)
	return result
}

// GetEntriesSince returns entries since the given time.
func (a *AuditLogger) GetEntriesSince(since time.Time) []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	var result []AuditEntry
	for _, entry := range a.entries {
		if entry.Timestamp.After(since) || entry.Timestamp.Equal(since) {
			result = append(result, entry)
		}
	}
	return result
}

// Clear removes all entries from memory.
func (a *AuditLogger) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = a.entries[:0]
}

// Count returns the number of entries in memory.
func (a *AuditLogger) Count() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.entries)
}
