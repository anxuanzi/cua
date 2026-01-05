package safety

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(10)
	assert.NotNil(t, limiter)
	assert.Equal(t, 10, limiter.Available())
}

func TestRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(3)

	// First 3 should be allowed
	assert.True(t, limiter.Allow())
	assert.True(t, limiter.Allow())
	assert.True(t, limiter.Allow())

	// 4th should be rate limited
	assert.False(t, limiter.Allow())
	assert.Equal(t, 0, limiter.Available())
}

func TestRateLimiter_Reset(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(3)
	limiter.Allow()
	limiter.Allow()
	limiter.Allow()

	assert.Equal(t, 0, limiter.Available())

	limiter.Reset()
	assert.Equal(t, 3, limiter.Available())
}

func TestNewAuditLogger(t *testing.T) {
	t.Parallel()

	logger := NewAuditLogger(nil, 100)
	assert.NotNil(t, logger)
	assert.Equal(t, 0, logger.Count())
}

func TestAuditLogger_Log(t *testing.T) {
	t.Parallel()

	logger := NewAuditLogger(nil, 100)

	logger.LogAction("click", "Clicked button", "(100, 200)")
	assert.Equal(t, 1, logger.Count())

	entries := logger.GetEntries()
	assert.Len(t, entries, 1)
	assert.Equal(t, "click", entries[0].Action)
	assert.Equal(t, AuditLevelAction, entries[0].Level)
}

func TestAuditLogger_LogError(t *testing.T) {
	t.Parallel()

	logger := NewAuditLogger(nil, 100)

	logger.LogError("Something failed", assert.AnError)
	entries := logger.GetEntries()

	require.Len(t, entries, 1)
	assert.Equal(t, AuditLevelError, entries[0].Level)
	assert.Contains(t, entries[0].Error, "assert.AnError")
}

func TestAuditLogger_MaxSize(t *testing.T) {
	t.Parallel()

	logger := NewAuditLogger(nil, 10)

	// Add 15 entries
	for i := 0; i < 15; i++ {
		logger.LogAction("action", "test", "target")
	}

	// Should have pruned some entries
	assert.Less(t, logger.Count(), 15)
}

func TestAuditLogger_GetEntriesSince(t *testing.T) {
	t.Parallel()

	logger := NewAuditLogger(nil, 100)

	start := time.Now()
	time.Sleep(10 * time.Millisecond)

	logger.LogAction("action1", "test1", "target1")
	time.Sleep(10 * time.Millisecond)
	logger.LogAction("action2", "test2", "target2")

	entries := logger.GetEntriesSince(start)
	assert.Len(t, entries, 2)
}

func TestNewSensitiveDetector(t *testing.T) {
	t.Parallel()

	detector := NewSensitiveDetector()
	assert.NotNil(t, detector)
}

func TestSensitiveDetector_Check_Password(t *testing.T) {
	t.Parallel()

	detector := NewSensitiveDetector()

	matches := detector.Check("type", "password_field", "Entering password")
	assert.NotEmpty(t, matches)
	assert.True(t, detector.IsSensitive("type", "password_field", ""))
}

func TestSensitiveDetector_Check_APIKey(t *testing.T) {
	t.Parallel()

	detector := NewSensitiveDetector()

	matches := detector.Check("type", "api_key input", "")
	assert.NotEmpty(t, matches)

	level := detector.GetHighestLevel(matches)
	assert.Equal(t, SensitiveLevelBlock, level)
}

func TestSensitiveDetector_Check_Safe(t *testing.T) {
	t.Parallel()

	detector := NewSensitiveDetector()

	matches := detector.Check("click", "Submit button", "Submitting form")
	assert.Empty(t, matches)
	assert.False(t, detector.IsSensitive("click", "Submit button", ""))
}

func TestSensitiveDetector_Payment(t *testing.T) {
	t.Parallel()

	detector := NewSensitiveDetector()

	matches := detector.Check("click", "Buy Now", "Making purchase")
	assert.NotEmpty(t, matches)
}

func TestSensitiveDetector_Delete(t *testing.T) {
	t.Parallel()

	detector := NewSensitiveDetector()

	matches := detector.Check("click", "Delete Account", "")
	assert.NotEmpty(t, matches)

	level := detector.GetHighestLevel(matches)
	assert.Equal(t, SensitiveLevelConfirm, level)
}

func TestNewGuardrails(t *testing.T) {
	t.Parallel()

	g, err := NewGuardrails(DefaultGuardrailsConfig())
	require.NoError(t, err)
	assert.NotNil(t, g)
}

func TestGuardrails_ValidateAction_Allowed(t *testing.T) {
	t.Parallel()

	g := NewDefaultGuardrails()

	err := g.ValidateAction("click", "(100, 200)", "Clicking button")
	assert.NoError(t, err)
}

func TestGuardrails_ValidateAction_RateLimited(t *testing.T) {
	t.Parallel()

	config := DefaultGuardrailsConfig()
	config.MaxActionsPerMinute = 2
	g, _ := NewGuardrails(config)

	assert.NoError(t, g.ValidateAction("click", "1", ""))
	assert.NoError(t, g.ValidateAction("click", "2", ""))
	assert.ErrorIs(t, g.ValidateAction("click", "3", ""), ErrRateLimited)
}

func TestGuardrails_ValidateAction_SensitiveBlocked(t *testing.T) {
	t.Parallel()

	config := DefaultGuardrailsConfig()
	config.Level = LevelNormal
	g, _ := NewGuardrails(config)

	err := g.ValidateAction("type", "api_key field", "")
	assert.ErrorIs(t, err, ErrSensitiveActionBlocked)
}

func TestGuardrails_ConsecutiveFailures(t *testing.T) {
	t.Parallel()

	config := DefaultGuardrailsConfig()
	config.MaxConsecutiveFailures = 3
	g, _ := NewGuardrails(config)

	// Record 3 failures
	g.RecordFailure("click", "target", assert.AnError)
	g.RecordFailure("click", "target", assert.AnError)
	g.RecordFailure("click", "target", assert.AnError)

	assert.Equal(t, 3, g.GetConsecutiveFailures())

	// Next action should be blocked
	err := g.ValidateAction("click", "target", "")
	assert.ErrorIs(t, err, ErrConsecutiveFailures)

	// Reset and try again
	g.ResetFailures()
	err = g.ValidateAction("click", "target", "")
	assert.NoError(t, err)
}

func TestGuardrails_RecordSuccess_ResetsFailures(t *testing.T) {
	t.Parallel()

	g := NewDefaultGuardrails()

	g.RecordFailure("click", "target", assert.AnError)
	g.RecordFailure("click", "target", assert.AnError)
	assert.Equal(t, 2, g.GetConsecutiveFailures())

	g.RecordSuccess("click", "target", "success")
	assert.Equal(t, 0, g.GetConsecutiveFailures())
}

func TestGuardrails_Takeover(t *testing.T) {
	t.Parallel()

	g := NewDefaultGuardrails()

	// Request takeover
	g.RequestTakeover()

	// Next action should return takeover error
	err := g.ValidateAction("click", "target", "")
	assert.ErrorIs(t, err, ErrTakeoverRequested)

	assert.True(t, g.TakeoverRequested())
	assert.True(t, g.IsPaused())

	// Resume
	g.Resume()
	assert.False(t, g.IsPaused())

	// Actions should work again
	err = g.ValidateAction("click", "target", "")
	assert.NoError(t, err)
}

func TestGuardrails_PauseResume(t *testing.T) {
	t.Parallel()

	g := NewDefaultGuardrails()

	g.Pause()
	assert.True(t, g.IsPaused())

	err := g.ValidateAction("click", "target", "")
	assert.Error(t, err)

	g.Resume()
	assert.False(t, g.IsPaused())
}

func TestGuardrails_SetLevel(t *testing.T) {
	t.Parallel()

	g := NewDefaultGuardrails()

	// In minimal mode, sensitive actions should pass
	g.SetLevel(LevelMinimal)
	err := g.ValidateAction("type", "password_field", "")
	assert.NoError(t, err)

	// Reset rate limiter
	g.GetRateLimiter().Reset()

	// In strict mode, confirmation-level should be blocked
	g.SetLevel(LevelStrict)
	err = g.ValidateAction("type", "password_field", "")
	assert.ErrorIs(t, err, ErrSensitiveActionBlocked)
}

func TestGuardrails_GetAuditEntries(t *testing.T) {
	t.Parallel()

	g := NewDefaultGuardrails()

	g.ValidateAction("click", "button1", "test1")
	g.ValidateAction("click", "button2", "test2")

	entries := g.GetAuditEntries()
	assert.GreaterOrEqual(t, len(entries), 2)
}
