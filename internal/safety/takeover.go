package safety

import (
	"context"
	"sync"
	"time"
)

// TakeoverReason describes why takeover was triggered.
type TakeoverReason string

const (
	TakeoverReasonHotkey           TakeoverReason = "hotkey"
	TakeoverReasonConsecutiveFails TakeoverReason = "consecutive_failures"
	TakeoverReasonSensitiveAction  TakeoverReason = "sensitive_action"
	TakeoverReasonProgrammatic     TakeoverReason = "programmatic"
	TakeoverReasonTimeout          TakeoverReason = "timeout"
)

// TakeoverEvent represents a takeover occurrence.
type TakeoverEvent struct {
	Timestamp time.Time
	Reason    TakeoverReason
	Message   string
}

// TakeoverHandler is called when takeover is triggered.
type TakeoverHandler func(event TakeoverEvent) TakeoverResponse

// TakeoverResponse indicates what to do after takeover.
type TakeoverResponse int

const (
	// TakeoverResponseAbort stops the current task entirely.
	TakeoverResponseAbort TakeoverResponse = iota

	// TakeoverResponseResume continues with the current task.
	TakeoverResponseResume

	// TakeoverResponseRetry retries the last failed action.
	TakeoverResponseRetry
)

// TakeoverController manages human takeover functionality.
type TakeoverController struct {
	mu sync.Mutex

	// handler is called when takeover is triggered.
	handler TakeoverHandler

	// requestChan receives takeover requests.
	requestChan chan TakeoverEvent

	// responseChan sends responses back.
	responseChan chan TakeoverResponse

	// active indicates if takeover is currently active.
	active bool

	// lastEvent is the most recent takeover event.
	lastEvent *TakeoverEvent

	// history of takeover events.
	history []TakeoverEvent
}

// NewTakeoverController creates a new takeover controller.
func NewTakeoverController(handler TakeoverHandler) *TakeoverController {
	if handler == nil {
		handler = func(event TakeoverEvent) TakeoverResponse {
			// Default handler just aborts
			return TakeoverResponseAbort
		}
	}

	return &TakeoverController{
		handler:      handler,
		requestChan:  make(chan TakeoverEvent, 1),
		responseChan: make(chan TakeoverResponse, 1),
		history:      make([]TakeoverEvent, 0, 10),
	}
}

// Request triggers a takeover with the given reason.
// Returns the response from the handler.
func (t *TakeoverController) Request(reason TakeoverReason, message string) TakeoverResponse {
	event := TakeoverEvent{
		Timestamp: time.Now(),
		Reason:    reason,
		Message:   message,
	}

	t.mu.Lock()
	t.active = true
	t.lastEvent = &event
	t.history = append(t.history, event)
	handler := t.handler
	t.mu.Unlock()

	// Call handler
	response := handler(event)

	t.mu.Lock()
	t.active = false
	t.mu.Unlock()

	return response
}

// RequestAsync triggers a takeover asynchronously.
// Use WaitForResponse to get the response.
func (t *TakeoverController) RequestAsync(reason TakeoverReason, message string) {
	event := TakeoverEvent{
		Timestamp: time.Now(),
		Reason:    reason,
		Message:   message,
	}

	t.mu.Lock()
	t.active = true
	t.lastEvent = &event
	t.history = append(t.history, event)
	t.mu.Unlock()

	select {
	case t.requestChan <- event:
	default:
		// Channel full, event already pending
	}
}

// WaitForResponse waits for a response to an async takeover request.
func (t *TakeoverController) WaitForResponse(ctx context.Context) (TakeoverResponse, error) {
	select {
	case response := <-t.responseChan:
		t.mu.Lock()
		t.active = false
		t.mu.Unlock()
		return response, nil
	case <-ctx.Done():
		return TakeoverResponseAbort, ctx.Err()
	}
}

// Respond provides a response to an async takeover request.
func (t *TakeoverController) Respond(response TakeoverResponse) {
	select {
	case t.responseChan <- response:
	default:
		// No pending request
	}
}

// IsActive returns true if a takeover is currently active.
func (t *TakeoverController) IsActive() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.active
}

// LastEvent returns the most recent takeover event.
func (t *TakeoverController) LastEvent() *TakeoverEvent {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.lastEvent == nil {
		return nil
	}
	event := *t.lastEvent
	return &event
}

// History returns a copy of all takeover events.
func (t *TakeoverController) History() []TakeoverEvent {
	t.mu.Lock()
	defer t.mu.Unlock()

	result := make([]TakeoverEvent, len(t.history))
	copy(result, t.history)
	return result
}

// ClearHistory removes all takeover history.
func (t *TakeoverController) ClearHistory() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.history = t.history[:0]
	t.lastEvent = nil
}

// SetHandler updates the takeover handler.
func (t *TakeoverController) SetHandler(handler TakeoverHandler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handler = handler
}

// DefaultCLIHandler provides a simple CLI-based takeover handler.
// It prints the event and waits for user input.
func DefaultCLIHandler(event TakeoverEvent) TakeoverResponse {
	// This is a placeholder - in a real implementation,
	// this would read from stdin or show a dialog
	return TakeoverResponseAbort
}
