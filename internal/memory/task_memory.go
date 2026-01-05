package memory

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// MaxRecentActions is the maximum number of recent actions to keep in detail.
const MaxRecentActions = 5

// MaxFailedPatterns is the maximum number of failed patterns to track.
const MaxFailedPatterns = 10

// MaxMilestones is the maximum number of milestones to keep.
const MaxMilestones = 20

// Phase constants for common workflow phases.
const (
	PhaseNavigation     = "navigation"
	PhaseFormFilling    = "form_filling"
	PhaseAuthentication = "authentication"
	PhaseSearch         = "search"
	PhaseBrowsing       = "browsing"
	PhaseConfirmation   = "confirmation"
	PhaseCheckout       = "checkout"
	PhaseUnknown        = ""
)

// Observation represents the current screen state for phase detection.
type Observation struct {
	// VisibleText contains key text visible on screen.
	VisibleText []string

	// ActiveApp is the name of the active application.
	ActiveApp string

	// HasLoginForm indicates if a login form is visible.
	HasLoginForm bool

	// HasSearchBox indicates if a search box is visible.
	HasSearchBox bool

	// HasCheckoutElements indicates if checkout elements are visible.
	HasCheckoutElements bool

	// HasConfirmation indicates if a confirmation message is visible.
	HasConfirmation bool

	// FocusedElementRole is the role of the currently focused element.
	FocusedElementRole string

	// Custom allows arbitrary key-value pairs for extensibility.
	Custom map[string]any
}

// ActionResult represents the outcome of an executed action.
type ActionResult struct {
	// StepNumber is the sequential step number.
	StepNumber int `json:"step_number"`

	// Action is the action that was executed.
	Action string `json:"action"`

	// Args contains the action arguments.
	Args map[string]any `json:"args,omitempty"`

	// Success indicates if the action succeeded.
	Success bool `json:"success"`

	// Result contains the action result or error message.
	Result string `json:"result,omitempty"`

	// Duration is how long the action took.
	Duration time.Duration `json:"duration"`

	// Timestamp is when the action was executed.
	Timestamp time.Time `json:"timestamp"`
}

// Error represents an error that occurred during task execution.
type Error struct {
	// Message is the error message.
	Message string `json:"message"`

	// Action is the action that caused the error.
	Action string `json:"action,omitempty"`

	// StepNumber is when the error occurred.
	StepNumber int `json:"step_number"`

	// Timestamp is when the error occurred.
	Timestamp time.Time `json:"timestamp"`

	// Recoverable indicates if the agent can recover from this error.
	Recoverable bool `json:"recoverable"`
}

// TaskMemory provides context engineering for long-running tasks.
// It maintains a structured view of task progress that fits within
// the context window while preserving critical information.
type TaskMemory struct {
	mu sync.RWMutex

	// === IMMUTABLE (always in context) ===

	// OriginalTask is the user's original request (never truncated).
	OriginalTask string `json:"original_task"`

	// StartedAt is when the task began.
	StartedAt time.Time `json:"started_at"`

	// === PROGRESSIVE SUMMARY (grows slowly) ===

	// Milestones are completed major steps (e.g., "Opened Safari", "Logged in").
	Milestones []string `json:"milestones,omitempty"`

	// === CURRENT STATE ===

	// Phase describes the current phase (e.g., "navigation", "form_filling").
	Phase string `json:"phase,omitempty"`

	// PhaseStartStep is reset when phase changes.
	PhaseStartStep int `json:"phase_start_step,omitempty"`

	// KeyFacts are extracted values (e.g., {"price": "$99", "item": "Widget"}).
	KeyFacts map[string]string `json:"key_facts,omitempty"`

	// === SLIDING WINDOW (recent detail) ===

	// RecentActions contains the last MaxRecentActions actions with full detail.
	RecentActions []ActionResult `json:"recent_actions,omitempty"`

	// TotalSteps is the total number of steps executed.
	TotalSteps int `json:"total_steps"`

	// === ERROR TRACKING ===

	// ConsecutiveFails counts consecutive failed actions.
	ConsecutiveFails int `json:"consecutive_fails,omitempty"`

	// LastError is the most recent error, if any.
	LastError *Error `json:"last_error,omitempty"`

	// === LEARNING (avoid repeating mistakes) ===

	// FailedPatterns are patterns that have been tried and failed.
	FailedPatterns []string `json:"failed_patterns,omitempty"`

	// === INTERNAL ===

	// summarizer handles progressive summarization of actions.
	summarizer *Summarizer
}

// New creates a new TaskMemory for the given task.
func New(task string) *TaskMemory {
	return &TaskMemory{
		OriginalTask: task,
		StartedAt:    time.Now(),
		KeyFacts:     make(map[string]string),
		summarizer:   NewSummarizer(),
	}
}

// RecordAction records the result of an action.
func (m *TaskMemory) RecordAction(action string, args map[string]any, success bool, result string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalSteps++

	actionResult := ActionResult{
		StepNumber: m.TotalSteps,
		Action:     action,
		Args:       args,
		Success:    success,
		Result:     result,
		Duration:   duration,
		Timestamp:  time.Now(),
	}

	// Add to recent actions
	m.RecentActions = append(m.RecentActions, actionResult)

	// Progressive summarization: when we exceed the window, summarize older actions into milestones
	if m.summarizer != nil && m.summarizer.ShouldSummarize(m.RecentActions) {
		milestones, remaining := m.summarizer.Summarize(m.RecentActions)
		for _, milestone := range milestones {
			m.addMilestoneLocked(milestone)
		}
		m.RecentActions = remaining
	} else if len(m.RecentActions) > MaxRecentActions {
		// Fallback if summarizer not initialized
		m.RecentActions = m.RecentActions[len(m.RecentActions)-MaxRecentActions:]
	}

	// Update error tracking
	if success {
		m.ConsecutiveFails = 0
		m.LastError = nil
	} else {
		m.ConsecutiveFails++
		m.LastError = &Error{
			Message:     result,
			Action:      action,
			StepNumber:  m.TotalSteps,
			Timestamp:   time.Now(),
			Recoverable: m.ConsecutiveFails < 3,
		}
	}
}

// addMilestoneLocked adds a milestone without locking (caller must hold lock).
func (m *TaskMemory) addMilestoneLocked(milestone string) {
	m.Milestones = append(m.Milestones, milestone)
	// Trim milestones if we have too many
	if len(m.Milestones) > MaxMilestones {
		// Compress oldest milestones
		m.Milestones = m.Milestones[len(m.Milestones)-MaxMilestones:]
	}
}

// AddMilestone adds a completed milestone to the summary.
func (m *TaskMemory) AddMilestone(milestone string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addMilestoneLocked(milestone)
}

// SetPhase updates the current phase.
// When changing phases, the previous phase is summarized into a milestone.
func (m *TaskMemory) SetPhase(phase string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Phase == phase {
		return // No change
	}

	// Summarize the previous phase if it existed
	if m.Phase != "" && m.summarizer != nil {
		milestone := m.summarizer.SummarizeForPhaseChange(m.Phase, m.RecentActions)
		if milestone != "" {
			m.addMilestoneLocked(milestone)
		}
		// Clear recent actions on phase change
		m.RecentActions = nil
	}

	m.Phase = phase
	m.PhaseStartStep = m.TotalSteps
}

// MaybeUpdatePhase detects the current phase from an observation and updates if changed.
// This is useful for automatic phase detection based on screen content.
func (m *TaskMemory) MaybeUpdatePhase(obs Observation) {
	newPhase := DetectPhase(obs)
	if newPhase != PhaseUnknown && newPhase != m.Phase {
		m.SetPhase(newPhase)
	}
}

// DetectPhase analyzes an observation to determine the current workflow phase.
// Returns PhaseUnknown if no specific phase can be determined.
func DetectPhase(obs Observation) string {
	// Priority-ordered phase detection based on observation signals

	// Confirmation phase takes priority (end state)
	if obs.HasConfirmation {
		return PhaseConfirmation
	}

	// Checkout phase
	if obs.HasCheckoutElements {
		return PhaseCheckout
	}

	// Authentication phase
	if obs.HasLoginForm {
		return PhaseAuthentication
	}

	// Search phase
	if obs.HasSearchBox {
		return PhaseSearch
	}

	// Text-based heuristics
	for _, text := range obs.VisibleText {
		lower := strings.ToLower(text)

		// Confirmation indicators
		if strings.Contains(lower, "success") ||
			strings.Contains(lower, "thank you") ||
			strings.Contains(lower, "order confirmed") ||
			strings.Contains(lower, "completed") {
			return PhaseConfirmation
		}

		// Checkout indicators
		if strings.Contains(lower, "checkout") ||
			strings.Contains(lower, "payment") ||
			strings.Contains(lower, "credit card") ||
			strings.Contains(lower, "place order") {
			return PhaseCheckout
		}

		// Authentication indicators
		if strings.Contains(lower, "sign in") ||
			strings.Contains(lower, "log in") ||
			strings.Contains(lower, "password") ||
			strings.Contains(lower, "username") {
			return PhaseAuthentication
		}

		// Search results indicators
		if strings.Contains(lower, "results for") ||
			strings.Contains(lower, "search results") {
			return PhaseBrowsing
		}
	}

	// Focused element heuristics
	switch obs.FocusedElementRole {
	case "textfield", "searchbox":
		return PhaseFormFilling
	}

	// Default based on what's visible
	if len(obs.VisibleText) > 0 {
		return PhaseBrowsing
	}

	return PhaseNavigation
}

// SetKeyFact stores an extracted key fact.
func (m *TaskMemory) SetKeyFact(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.KeyFacts[key] = value
}

// GetKeyFact retrieves a stored key fact.
func (m *TaskMemory) GetKeyFact(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.KeyFacts[key]
	return v, ok
}

// AddFailedPattern records a pattern that failed.
func (m *TaskMemory) AddFailedPattern(pattern string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already recorded
	for _, p := range m.FailedPatterns {
		if p == pattern {
			return
		}
	}

	m.FailedPatterns = append(m.FailedPatterns, pattern)
	if len(m.FailedPatterns) > MaxFailedPatterns {
		m.FailedPatterns = m.FailedPatterns[1:]
	}
}

// HasFailedPattern checks if a pattern has been tried and failed.
func (m *TaskMemory) HasFailedPattern(pattern string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.FailedPatterns {
		if p == pattern {
			return true
		}
	}
	return false
}

// IsStuck returns true if the agent appears to be stuck.
func (m *TaskMemory) IsStuck() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ConsecutiveFails >= 3
}

// NeedsHelp returns true if the agent should ask for human help.
func (m *TaskMemory) NeedsHelp() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ConsecutiveFails >= 5
}

// Duration returns how long the task has been running.
func (m *TaskMemory) Duration() time.Duration {
	return time.Since(m.StartedAt)
}

// ToPrompt formats the memory for direct inclusion in the agent's context window.
// This is the critical method that builds the context for each iteration.
func (m *TaskMemory) ToPrompt() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder

	// 1. Task anchor (ALWAYS present - never truncated)
	sb.WriteString("## Your Task\n")
	sb.WriteString(m.OriginalTask)
	sb.WriteString("\n\n")

	// 2. Progress summary (compressed milestones)
	if len(m.Milestones) > 0 {
		sb.WriteString("## What You've Accomplished\n")
		for _, milestone := range m.Milestones {
			sb.WriteString("- ")
			sb.WriteString(milestone)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// 3. Current phase and key facts
	if m.Phase != "" {
		fmt.Fprintf(&sb, "## Current Phase: %s\n", m.Phase)
		if len(m.KeyFacts) > 0 {
			sb.WriteString("Key information:\n")
			for k, v := range m.KeyFacts {
				fmt.Fprintf(&sb, "- %s: %s\n", k, v)
			}
		}
		sb.WriteString("\n")
	} else if len(m.KeyFacts) > 0 {
		sb.WriteString("## Key Information\n")
		for k, v := range m.KeyFacts {
			fmt.Fprintf(&sb, "- %s: %s\n", k, v)
		}
		sb.WriteString("\n")
	}

	// 4. Recent actions (sliding window)
	if len(m.RecentActions) > 0 {
		sb.WriteString("## Recent Actions\n")
		for _, action := range m.RecentActions {
			status := "✓"
			if !action.Success {
				status = "✗"
			}
			if action.Result != "" {
				fmt.Fprintf(&sb, "%s %s → %s\n", status, action.Action, action.Result)
			} else {
				fmt.Fprintf(&sb, "%s %s\n", status, action.Action)
			}
		}
		sb.WriteString("\n")
	}

	// 5. Known issues (avoid repeating mistakes)
	if len(m.FailedPatterns) > 0 {
		sb.WriteString("## Known Issues (avoid these)\n")
		for _, pattern := range m.FailedPatterns {
			sb.WriteString("- ")
			sb.WriteString(pattern)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// 6. Status warnings
	if m.ConsecutiveFails >= 5 {
		sb.WriteString("## ⚠️ NEEDS HELP\nMultiple consecutive failures. Consider asking the user for guidance.\n\n")
	} else if m.ConsecutiveFails >= 3 {
		sb.WriteString("## ⚠️ POSSIBLY STUCK\nTry a different approach.\n\n")
	}

	return sb.String()
}

// Summary returns the current state summary for the context window.
func (m *TaskMemory) Summary() TaskSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return TaskSummary{
		OriginalTask:     m.OriginalTask,
		Duration:         time.Since(m.StartedAt),
		TotalSteps:       m.TotalSteps,
		Phase:            m.Phase,
		Milestones:       append([]string{}, m.Milestones...),
		KeyFacts:         copyMap(m.KeyFacts),
		RecentActions:    append([]ActionResult{}, m.RecentActions...),
		ConsecutiveFails: m.ConsecutiveFails,
		LastError:        m.LastError,
		FailedPatterns:   append([]string{}, m.FailedPatterns...),
		IsStuck:          m.ConsecutiveFails >= 3,
		NeedsHelp:        m.ConsecutiveFails >= 5,
	}
}

// TaskSummary is an immutable snapshot of the task state.
type TaskSummary struct {
	OriginalTask     string
	Duration         time.Duration
	TotalSteps       int
	Phase            string
	Milestones       []string
	KeyFacts         map[string]string
	RecentActions    []ActionResult
	ConsecutiveFails int
	LastError        *Error
	FailedPatterns   []string
	IsStuck          bool
	NeedsHelp        bool
}

// ForContext formats the summary for inclusion in the agent's context window.
func (s TaskSummary) ForContext() string {
	var b strings.Builder

	// Task anchor (never truncated)
	b.WriteString("## TASK\n")
	b.WriteString(s.OriginalTask)
	b.WriteString("\n\n")

	// Progress summary
	if len(s.Milestones) > 0 {
		b.WriteString("## COMPLETED MILESTONES\n")
		for i, m := range s.Milestones {
			fmt.Fprintf(&b, "%d. %s\n", i+1, m)
		}
		b.WriteString("\n")
	}

	// Current state
	if s.Phase != "" {
		fmt.Fprintf(&b, "## CURRENT PHASE: %s\n\n", s.Phase)
	}

	// Key facts
	if len(s.KeyFacts) > 0 {
		b.WriteString("## KEY FACTS\n")
		for k, v := range s.KeyFacts {
			fmt.Fprintf(&b, "- %s: %s\n", k, v)
		}
		b.WriteString("\n")
	}

	// Recent actions
	if len(s.RecentActions) > 0 {
		b.WriteString("## RECENT ACTIONS\n")
		for _, a := range s.RecentActions {
			status := "SUCCESS"
			if !a.Success {
				status = "FAILED"
			}
			fmt.Fprintf(&b, "Step %d: %s [%s] - %s\n", a.StepNumber, a.Action, status, a.Result)
		}
		b.WriteString("\n")
	}

	// Error state
	if s.LastError != nil {
		fmt.Fprintf(&b, "## LAST ERROR\n%s\n\n", s.LastError.Message)
	}

	// Status warnings
	if s.NeedsHelp {
		b.WriteString("## STATUS: NEEDS HELP\nMultiple consecutive failures. Consider asking the user for guidance.\n\n")
	} else if s.IsStuck {
		b.WriteString("## STATUS: POSSIBLY STUCK\nTry a different approach.\n\n")
	}

	// Failed patterns to avoid
	if len(s.FailedPatterns) > 0 {
		b.WriteString("## FAILED PATTERNS (avoid these)\n")
		for _, p := range s.FailedPatterns {
			fmt.Fprintf(&b, "- %s\n", p)
		}
		b.WriteString("\n")
	}

	// Stats
	fmt.Fprintf(&b, "## STATS\n- Total steps: %d\n- Duration: %s\n", s.TotalSteps, s.Duration.Round(time.Second))

	return b.String()
}

func copyMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
