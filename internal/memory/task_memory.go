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
}

// New creates a new TaskMemory for the given task.
func New(task string) *TaskMemory {
	return &TaskMemory{
		OriginalTask: task,
		StartedAt:    time.Now(),
		KeyFacts:     make(map[string]string),
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

	// Add to recent actions (sliding window)
	m.RecentActions = append(m.RecentActions, actionResult)
	if len(m.RecentActions) > MaxRecentActions {
		m.RecentActions = m.RecentActions[1:]
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

// AddMilestone adds a completed milestone to the summary.
func (m *TaskMemory) AddMilestone(milestone string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Milestones = append(m.Milestones, milestone)
}

// SetPhase updates the current phase.
func (m *TaskMemory) SetPhase(phase string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Phase != phase {
		m.Phase = phase
		m.PhaseStartStep = m.TotalSteps
	}
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
