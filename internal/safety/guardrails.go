// Package safety provides safety guardrails for the CUA agent.
// It includes rate limiting, sensitive action detection, and audit logging.
package safety

import (
	"errors"
	"fmt"
	"sync"
)

// Common errors returned by guardrails.
var (
	ErrRateLimited            = errors.New("action rate limited")
	ErrSensitiveActionBlocked = errors.New("sensitive action blocked")
	ErrTakeoverRequested      = errors.New("human takeover requested")
	ErrConsecutiveFailures    = errors.New("too many consecutive failures")
)

// Level represents the overall safety level.
type Level int

const (
	// LevelMinimal applies minimal safety checks.
	LevelMinimal Level = iota

	// LevelNormal applies standard safety checks.
	LevelNormal

	// LevelStrict applies maximum safety checks.
	LevelStrict
)

// GuardrailsConfig configures the safety guardrails.
type GuardrailsConfig struct {
	// Level is the overall safety level.
	Level Level

	// MaxActionsPerMinute limits action rate.
	MaxActionsPerMinute int

	// MaxConsecutiveFailures before requiring help.
	MaxConsecutiveFailures int

	// AuditLogPath is the path for the audit log file.
	// Empty means no file logging.
	AuditLogPath string
}

// DefaultGuardrailsConfig returns the default configuration.
func DefaultGuardrailsConfig() GuardrailsConfig {
	return GuardrailsConfig{
		Level:                  LevelNormal,
		MaxActionsPerMinute:    60,
		MaxConsecutiveFailures: 5,
		AuditLogPath:           "",
	}
}

// Guardrails provides safety checks and controls for the agent.
type Guardrails struct {
	mu sync.Mutex

	config              GuardrailsConfig
	rateLimiter         *RateLimiter
	sensitiveDetector   *SensitiveDetector
	auditLogger         *AuditLogger
	takeoverChan        chan struct{}
	consecutiveFailures int
	paused              bool
}

// NewGuardrails creates a new guardrails instance.
func NewGuardrails(config GuardrailsConfig) (*Guardrails, error) {
	var auditLogger *AuditLogger
	if config.AuditLogPath != "" {
		var err error
		auditLogger, err = NewFileAuditLogger(config.AuditLogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create audit logger: %w", err)
		}
	} else {
		auditLogger = NewAuditLogger(nil, 1000)
	}

	return &Guardrails{
		config:            config,
		rateLimiter:       NewRateLimiter(config.MaxActionsPerMinute),
		sensitiveDetector: NewSensitiveDetector(),
		auditLogger:       auditLogger,
		takeoverChan:      make(chan struct{}, 1),
	}, nil
}

// NewDefaultGuardrails creates guardrails with default configuration.
func NewDefaultGuardrails() *Guardrails {
	g, _ := NewGuardrails(DefaultGuardrailsConfig())
	return g
}

// ValidateAction checks if an action should be allowed.
// Returns nil if allowed, or an error explaining why not.
func (g *Guardrails) ValidateAction(action, target, description string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check for takeover request
	select {
	case <-g.takeoverChan:
		g.paused = true
		return ErrTakeoverRequested
	default:
	}

	// Check if paused
	if g.paused {
		return ErrTakeoverRequested
	}

	// Check consecutive failures
	if g.consecutiveFailures >= g.config.MaxConsecutiveFailures {
		return ErrConsecutiveFailures
	}

	// Check rate limit
	if !g.rateLimiter.Allow() {
		g.auditLogger.LogWarning("Rate limited", map[string]interface{}{
			"action": action,
			"target": target,
		})
		return ErrRateLimited
	}

	// Check sensitive patterns (only in Normal or Strict mode)
	if g.config.Level >= LevelNormal {
		matches := g.sensitiveDetector.Check(action, target, description)
		if len(matches) > 0 {
			level := g.sensitiveDetector.GetHighestLevel(matches)

			// Log the sensitive action attempt
			g.auditLogger.LogWarning("Sensitive action detected", map[string]interface{}{
				"action":  action,
				"target":  target,
				"matches": len(matches),
				"level":   level,
			})

			// In strict mode, block even confirmation-level actions
			if g.config.Level == LevelStrict && level >= SensitiveLevelConfirm {
				return ErrSensitiveActionBlocked
			}

			// Block actions at block level
			if level >= SensitiveLevelBlock {
				return ErrSensitiveActionBlocked
			}
		}
	}

	// Log the action
	g.auditLogger.LogAction(action, description, target)

	return nil
}

// RecordSuccess records a successful action.
func (g *Guardrails) RecordSuccess(action, target, result string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.consecutiveFailures = 0
	g.auditLogger.LogActionResult(action, "Action succeeded", target, result, nil)
}

// RecordFailure records a failed action.
func (g *Guardrails) RecordFailure(action, target string, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.consecutiveFailures++
	g.auditLogger.LogActionResult(action, "Action failed", target, "", err)
}

// GetConsecutiveFailures returns the current consecutive failure count.
func (g *Guardrails) GetConsecutiveFailures() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.consecutiveFailures
}

// ResetFailures resets the consecutive failure counter.
func (g *Guardrails) ResetFailures() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.consecutiveFailures = 0
}

// RequestTakeover signals that human takeover is requested.
func (g *Guardrails) RequestTakeover() {
	select {
	case g.takeoverChan <- struct{}{}:
	default:
		// Channel already has a pending request
	}
}

// TakeoverRequested returns true if takeover was requested.
func (g *Guardrails) TakeoverRequested() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.paused
}

// Resume resumes operation after a takeover.
func (g *Guardrails) Resume() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.paused = false
	// Clear any pending takeover request
	select {
	case <-g.takeoverChan:
	default:
	}
}

// Pause pauses the agent.
func (g *Guardrails) Pause() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.paused = true
}

// IsPaused returns true if the agent is paused.
func (g *Guardrails) IsPaused() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.paused
}

// GetAuditEntries returns recent audit log entries.
func (g *Guardrails) GetAuditEntries() []AuditEntry {
	return g.auditLogger.GetEntries()
}

// GetRateLimiter returns the rate limiter for direct access.
func (g *Guardrails) GetRateLimiter() *RateLimiter {
	return g.rateLimiter
}

// GetSensitiveDetector returns the sensitive detector for customization.
func (g *Guardrails) GetSensitiveDetector() *SensitiveDetector {
	return g.sensitiveDetector
}

// SetLevel changes the safety level.
func (g *Guardrails) SetLevel(level Level) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.config.Level = level
}
