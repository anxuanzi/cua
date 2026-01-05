package cua

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure conditions.
var (
	// ErrNoAPIKey indicates no API key was provided and none found in environment.
	ErrNoAPIKey = errors.New("cua: no API key provided (set GOOGLE_API_KEY or use WithAPIKey)")

	// ErrTimeout indicates the task exceeded the configured timeout.
	ErrTimeout = errors.New("cua: task timed out")

	// ErrMaxActions indicates the task exceeded the maximum allowed actions.
	ErrMaxActions = errors.New("cua: exceeded maximum actions")

	// ErrMaxActionsExceeded is an alias for ErrMaxActions.
	ErrMaxActionsExceeded = ErrMaxActions

	// ErrAgentBusy indicates another task is already running.
	ErrAgentBusy = errors.New("cua: agent is busy with another task")

	// ErrHumanTakeover indicates the user requested to take over control.
	ErrHumanTakeover = errors.New("cua: human takeover requested")

	// ErrAgentStuck indicates the agent couldn't make progress.
	ErrAgentStuck = errors.New("cua: agent stuck, unable to proceed")

	// ErrPermissionDenied indicates missing system permissions (accessibility, screen recording).
	ErrPermissionDenied = errors.New("cua: permission denied (check accessibility/screen recording settings)")

	// ErrElementNotFound indicates no element matched the selector.
	ErrElementNotFound = errors.New("cua: element not found")

	// ErrNotSupported indicates the operation is not supported on this platform.
	ErrNotSupported = errors.New("cua: operation not supported on this platform")

	// ErrCanceled indicates the operation was canceled via context.
	ErrCanceled = errors.New("cua: operation canceled")

	// ErrRateLimited indicates too many actions in a short period.
	ErrRateLimited = errors.New("cua: rate limited, too many actions")

	// ErrSafetyBlock indicates a safety check prevented the action.
	ErrSafetyBlock = errors.New("cua: action blocked by safety check")
)

// ActionError wraps an error with context about which action failed.
type ActionError struct {
	Action      string // The action that failed (e.g., "click", "type")
	Description string // Human-readable description of what was attempted
	Step        int    // The step number when this occurred
	Err         error  // The underlying error
}

func (e *ActionError) Error() string {
	return fmt.Sprintf("cua: step %d failed: %s (%s): %v", e.Step, e.Action, e.Description, e.Err)
}

func (e *ActionError) Unwrap() error {
	return e.Err
}

// TaskError wraps an error with context about the overall task.
type TaskError struct {
	Task        string // The original task description
	StepsTotal  int    // Total steps attempted
	StepsFailed int    // Number of failed steps
	LastAction  string // The last action attempted
	Err         error  // The underlying error
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("cua: task failed after %d steps (%d failed): %v", e.StepsTotal, e.StepsFailed, e.Err)
}

func (e *TaskError) Unwrap() error {
	return e.Err
}

// ElementError wraps an error with context about element operations.
type ElementError struct {
	Selector string // String representation of the selector used
	Err      error  // The underlying error
}

func (e *ElementError) Error() string {
	return fmt.Sprintf("cua: element error (%s): %v", e.Selector, e.Err)
}

func (e *ElementError) Unwrap() error {
	return e.Err
}

// PermissionError provides details about missing permissions.
type PermissionError struct {
	Permission string // The permission that's missing
	Platform   string // The platform (darwin, windows)
	Hint       string // A hint for how to fix
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("cua: missing %s permission on %s. %s", e.Permission, e.Platform, e.Hint)
}

func (e *PermissionError) Is(target error) bool {
	return target == ErrPermissionDenied
}

// SafetyError provides details about why a safety check blocked an action.
type SafetyError struct {
	Action   string // The action that was blocked
	Reason   string // Why it was blocked
	Severity string // "warning", "blocked", "critical"
}

func (e *SafetyError) Error() string {
	return fmt.Sprintf("cua: safety %s: %s blocked: %s", e.Severity, e.Action, e.Reason)
}

func (e *SafetyError) Is(target error) bool {
	return target == ErrSafetyBlock
}

// IsRetryable returns true if the error might succeed on retry.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// These errors are retryable
	if errors.Is(err, ErrRateLimited) {
		return true
	}

	// Check for specific retryable underlying errors
	var actionErr *ActionError
	if errors.As(err, &actionErr) {
		// Element not found might succeed on retry (element may appear)
		if errors.Is(actionErr.Err, ErrElementNotFound) {
			return true
		}
	}

	return false
}

// IsFatal returns true if the error cannot be recovered from.
func IsFatal(err error) bool {
	if err == nil {
		return false
	}

	// These errors are fatal
	return errors.Is(err, ErrNoAPIKey) ||
		errors.Is(err, ErrPermissionDenied) ||
		errors.Is(err, ErrNotSupported) ||
		errors.Is(err, ErrHumanTakeover) ||
		errors.Is(err, ErrCanceled)
}
