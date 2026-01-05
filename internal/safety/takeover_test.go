package safety

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTakeoverController(t *testing.T) {
	t.Parallel()

	tc := NewTakeoverController(nil)
	assert.NotNil(t, tc)
	assert.False(t, tc.IsActive())
}

func TestTakeoverController_Request(t *testing.T) {
	t.Parallel()

	handlerCalled := false
	tc := NewTakeoverController(func(event TakeoverEvent) TakeoverResponse {
		handlerCalled = true
		assert.Equal(t, TakeoverReasonHotkey, event.Reason)
		assert.Equal(t, "User pressed hotkey", event.Message)
		return TakeoverResponseResume
	})

	response := tc.Request(TakeoverReasonHotkey, "User pressed hotkey")

	assert.True(t, handlerCalled)
	assert.Equal(t, TakeoverResponseResume, response)
	assert.False(t, tc.IsActive()) // Should be inactive after handler returns
}

func TestTakeoverController_RequestWithAbort(t *testing.T) {
	t.Parallel()

	tc := NewTakeoverController(func(event TakeoverEvent) TakeoverResponse {
		return TakeoverResponseAbort
	})

	response := tc.Request(TakeoverReasonConsecutiveFails, "Too many failures")
	assert.Equal(t, TakeoverResponseAbort, response)
}

func TestTakeoverController_RequestWithRetry(t *testing.T) {
	t.Parallel()

	tc := NewTakeoverController(func(event TakeoverEvent) TakeoverResponse {
		return TakeoverResponseRetry
	})

	response := tc.Request(TakeoverReasonSensitiveAction, "Sensitive action")
	assert.Equal(t, TakeoverResponseRetry, response)
}

func TestTakeoverController_DefaultHandler(t *testing.T) {
	t.Parallel()

	tc := NewTakeoverController(nil)

	// Default handler should abort
	response := tc.Request(TakeoverReasonProgrammatic, "Test")
	assert.Equal(t, TakeoverResponseAbort, response)
}

func TestTakeoverController_LastEvent(t *testing.T) {
	t.Parallel()

	tc := NewTakeoverController(func(event TakeoverEvent) TakeoverResponse {
		return TakeoverResponseAbort
	})

	assert.Nil(t, tc.LastEvent())

	tc.Request(TakeoverReasonHotkey, "Test message")

	event := tc.LastEvent()
	assert.NotNil(t, event)
	assert.Equal(t, TakeoverReasonHotkey, event.Reason)
	assert.Equal(t, "Test message", event.Message)
	assert.False(t, event.Timestamp.IsZero())
}

func TestTakeoverController_History(t *testing.T) {
	t.Parallel()

	tc := NewTakeoverController(func(event TakeoverEvent) TakeoverResponse {
		return TakeoverResponseAbort
	})

	assert.Empty(t, tc.History())

	tc.Request(TakeoverReasonHotkey, "First")
	tc.Request(TakeoverReasonConsecutiveFails, "Second")
	tc.Request(TakeoverReasonProgrammatic, "Third")

	history := tc.History()
	assert.Len(t, history, 3)
	assert.Equal(t, TakeoverReasonHotkey, history[0].Reason)
	assert.Equal(t, TakeoverReasonConsecutiveFails, history[1].Reason)
	assert.Equal(t, TakeoverReasonProgrammatic, history[2].Reason)
}

func TestTakeoverController_ClearHistory(t *testing.T) {
	t.Parallel()

	tc := NewTakeoverController(func(event TakeoverEvent) TakeoverResponse {
		return TakeoverResponseAbort
	})

	tc.Request(TakeoverReasonHotkey, "Test")
	assert.Len(t, tc.History(), 1)
	assert.NotNil(t, tc.LastEvent())

	tc.ClearHistory()
	assert.Empty(t, tc.History())
	assert.Nil(t, tc.LastEvent())
}

func TestTakeoverController_SetHandler(t *testing.T) {
	t.Parallel()

	tc := NewTakeoverController(func(event TakeoverEvent) TakeoverResponse {
		return TakeoverResponseAbort
	})

	response := tc.Request(TakeoverReasonHotkey, "Test")
	assert.Equal(t, TakeoverResponseAbort, response)

	tc.SetHandler(func(event TakeoverEvent) TakeoverResponse {
		return TakeoverResponseResume
	})

	response = tc.Request(TakeoverReasonHotkey, "Test")
	assert.Equal(t, TakeoverResponseResume, response)
}

func TestTakeoverController_AsyncRequest(t *testing.T) {
	t.Parallel()

	tc := NewTakeoverController(nil)

	tc.RequestAsync(TakeoverReasonHotkey, "Async test")
	assert.True(t, tc.IsActive())

	// Respond to the request
	tc.Respond(TakeoverResponseResume)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	response, err := tc.WaitForResponse(ctx)
	assert.NoError(t, err)
	assert.Equal(t, TakeoverResponseResume, response)
	assert.False(t, tc.IsActive())
}

func TestTakeoverController_AsyncRequestTimeout(t *testing.T) {
	t.Parallel()

	tc := NewTakeoverController(nil)

	tc.RequestAsync(TakeoverReasonHotkey, "Timeout test")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	response, err := tc.WaitForResponse(ctx)
	assert.Error(t, err)
	assert.Equal(t, TakeoverResponseAbort, response)
}

func TestTakeoverReasons(t *testing.T) {
	t.Parallel()

	reasons := []TakeoverReason{
		TakeoverReasonHotkey,
		TakeoverReasonConsecutiveFails,
		TakeoverReasonSensitiveAction,
		TakeoverReasonProgrammatic,
		TakeoverReasonTimeout,
	}

	for _, reason := range reasons {
		assert.NotEmpty(t, string(reason))
	}
}

func TestTakeoverResponses(t *testing.T) {
	t.Parallel()

	assert.Equal(t, TakeoverResponse(0), TakeoverResponseAbort)
	assert.Equal(t, TakeoverResponse(1), TakeoverResponseResume)
	assert.Equal(t, TakeoverResponse(2), TakeoverResponseRetry)
}
