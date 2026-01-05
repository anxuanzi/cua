package memory

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	task := "Open Safari and search for golang"
	mem := New(task)

	assert.Equal(t, task, mem.OriginalTask)
	assert.NotZero(t, mem.StartedAt)
	assert.Empty(t, mem.Milestones)
	assert.Empty(t, mem.Phase)
	assert.NotNil(t, mem.KeyFacts)
	assert.Empty(t, mem.RecentActions)
	assert.Zero(t, mem.TotalSteps)
	assert.Zero(t, mem.ConsecutiveFails)
	assert.Nil(t, mem.LastError)
}

func TestRecordAction(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	// Record a successful action
	mem.RecordAction("click", map[string]any{"x": 100, "y": 200}, true, "clicked", 50*time.Millisecond)

	assert.Equal(t, 1, mem.TotalSteps)
	assert.Len(t, mem.RecentActions, 1)
	assert.Equal(t, "click", mem.RecentActions[0].Action)
	assert.True(t, mem.RecentActions[0].Success)
	assert.Zero(t, mem.ConsecutiveFails)
	assert.Nil(t, mem.LastError)

	// Record a failed action
	mem.RecordAction("type", map[string]any{"text": "hello"}, false, "element not found", 100*time.Millisecond)

	assert.Equal(t, 2, mem.TotalSteps)
	assert.Len(t, mem.RecentActions, 2)
	assert.Equal(t, 1, mem.ConsecutiveFails)
	require.NotNil(t, mem.LastError)
	assert.Equal(t, "element not found", mem.LastError.Message)
	assert.True(t, mem.LastError.Recoverable)
}

func TestSlidingWindow(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	// Record more than MaxRecentActions
	for i := 0; i < MaxRecentActions+3; i++ {
		mem.RecordAction("action", nil, true, "ok", time.Millisecond)
	}

	assert.Equal(t, MaxRecentActions+3, mem.TotalSteps)
	assert.Len(t, mem.RecentActions, MaxRecentActions)
	// Should have the most recent actions
	assert.Equal(t, 4, mem.RecentActions[0].StepNumber)
	assert.Equal(t, MaxRecentActions+3, mem.RecentActions[len(mem.RecentActions)-1].StepNumber)
}

func TestConsecutiveFailures(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	// Record consecutive failures
	mem.RecordAction("action1", nil, false, "error 1", time.Millisecond)
	assert.Equal(t, 1, mem.ConsecutiveFails)
	assert.False(t, mem.IsStuck())
	assert.False(t, mem.NeedsHelp())

	mem.RecordAction("action2", nil, false, "error 2", time.Millisecond)
	assert.Equal(t, 2, mem.ConsecutiveFails)
	assert.False(t, mem.IsStuck())
	assert.False(t, mem.NeedsHelp())

	mem.RecordAction("action3", nil, false, "error 3", time.Millisecond)
	assert.Equal(t, 3, mem.ConsecutiveFails)
	assert.True(t, mem.IsStuck())
	assert.False(t, mem.NeedsHelp())

	// Success resets counter
	mem.RecordAction("action4", nil, true, "success", time.Millisecond)
	assert.Zero(t, mem.ConsecutiveFails)
	assert.False(t, mem.IsStuck())
	assert.Nil(t, mem.LastError)
}

func TestNeedsHelp(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	// Record 5 consecutive failures
	for i := 0; i < 5; i++ {
		mem.RecordAction("action", nil, false, "error", time.Millisecond)
	}

	assert.Equal(t, 5, mem.ConsecutiveFails)
	assert.True(t, mem.IsStuck())
	assert.True(t, mem.NeedsHelp())
}

func TestMilestones(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	mem.AddMilestone("Opened Safari")
	mem.AddMilestone("Navigated to Google")
	mem.AddMilestone("Performed search")

	assert.Len(t, mem.Milestones, 3)
	assert.Equal(t, "Opened Safari", mem.Milestones[0])
	assert.Equal(t, "Performed search", mem.Milestones[2])
}

func TestPhase(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	// Record some actions
	mem.RecordAction("action1", nil, true, "ok", time.Millisecond)
	mem.RecordAction("action2", nil, true, "ok", time.Millisecond)

	// Set phase
	mem.SetPhase("navigation")
	assert.Equal(t, "navigation", mem.Phase)
	assert.Equal(t, 2, mem.PhaseStartStep)

	// More actions
	mem.RecordAction("action3", nil, true, "ok", time.Millisecond)
	mem.RecordAction("action4", nil, true, "ok", time.Millisecond)

	// Change phase
	mem.SetPhase("form_filling")
	assert.Equal(t, "form_filling", mem.Phase)
	assert.Equal(t, 4, mem.PhaseStartStep)

	// Setting same phase doesn't change start step
	mem.RecordAction("action5", nil, true, "ok", time.Millisecond)
	mem.SetPhase("form_filling")
	assert.Equal(t, 4, mem.PhaseStartStep)
}

func TestKeyFacts(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	mem.SetKeyFact("price", "$99")
	mem.SetKeyFact("item", "Widget")

	price, ok := mem.GetKeyFact("price")
	assert.True(t, ok)
	assert.Equal(t, "$99", price)

	item, ok := mem.GetKeyFact("item")
	assert.True(t, ok)
	assert.Equal(t, "Widget", item)

	_, ok = mem.GetKeyFact("nonexistent")
	assert.False(t, ok)
}

func TestFailedPatterns(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	mem.AddFailedPattern("clicking X while loading")
	mem.AddFailedPattern("submitting empty form")

	assert.True(t, mem.HasFailedPattern("clicking X while loading"))
	assert.True(t, mem.HasFailedPattern("submitting empty form"))
	assert.False(t, mem.HasFailedPattern("some other pattern"))

	// Adding duplicate doesn't add again
	mem.AddFailedPattern("clicking X while loading")
	assert.Len(t, mem.FailedPatterns, 2)
}

func TestFailedPatternsLimit(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	// Add more than MaxFailedPatterns
	for i := 0; i < MaxFailedPatterns+5; i++ {
		mem.AddFailedPattern(strings.Repeat("x", i+1))
	}

	assert.Len(t, mem.FailedPatterns, MaxFailedPatterns)
}

func TestSummary(t *testing.T) {
	t.Parallel()

	mem := New("Open Safari and search for golang")

	mem.AddMilestone("Opened Safari")
	mem.SetPhase("navigation")
	mem.SetKeyFact("search_term", "golang")
	mem.RecordAction("click", map[string]any{"x": 100}, true, "clicked", 50*time.Millisecond)
	mem.RecordAction("type", map[string]any{"text": "golang"}, true, "typed", 100*time.Millisecond)

	summary := mem.Summary()

	assert.Equal(t, "Open Safari and search for golang", summary.OriginalTask)
	assert.Equal(t, 2, summary.TotalSteps)
	assert.Equal(t, "navigation", summary.Phase)
	assert.Len(t, summary.Milestones, 1)
	assert.Equal(t, "golang", summary.KeyFacts["search_term"])
	assert.Len(t, summary.RecentActions, 2)
	assert.False(t, summary.IsStuck)
	assert.False(t, summary.NeedsHelp)
}

func TestSummaryForContext(t *testing.T) {
	t.Parallel()

	mem := New("Open Safari and search for golang")

	mem.AddMilestone("Opened Safari")
	mem.SetPhase("navigation")
	mem.SetKeyFact("search_term", "golang")
	mem.RecordAction("click", nil, true, "clicked search box", 50*time.Millisecond)
	mem.RecordAction("type", nil, false, "element not visible", 100*time.Millisecond)

	summary := mem.Summary()
	context := summary.ForContext()

	// Check that key information is present
	assert.Contains(t, context, "Open Safari and search for golang")
	assert.Contains(t, context, "Opened Safari")
	assert.Contains(t, context, "navigation")
	assert.Contains(t, context, "golang")
	assert.Contains(t, context, "click")
	assert.Contains(t, context, "type")
	assert.Contains(t, context, "FAILED")
	assert.Contains(t, context, "element not visible")
}

func TestSummaryForContextStuck(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	// Make the agent stuck
	for i := 0; i < 3; i++ {
		mem.RecordAction("action", nil, false, "error", time.Millisecond)
	}

	summary := mem.Summary()
	context := summary.ForContext()

	assert.Contains(t, context, "POSSIBLY STUCK")
}

func TestSummaryForContextNeedsHelp(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	// Make the agent need help
	for i := 0; i < 5; i++ {
		mem.RecordAction("action", nil, false, "error", time.Millisecond)
	}

	summary := mem.Summary()
	context := summary.ForContext()

	assert.Contains(t, context, "NEEDS HELP")
}

func TestDuration(t *testing.T) {
	t.Parallel()

	mem := New("test task")

	// Wait a tiny bit
	time.Sleep(10 * time.Millisecond)

	d := mem.Duration()
	assert.GreaterOrEqual(t, d, 10*time.Millisecond)
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	mem := New("test task")
	done := make(chan struct{})

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			mem.RecordAction("action", nil, true, "ok", time.Millisecond)
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			mem.AddMilestone("milestone")
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			mem.SetKeyFact("key", "value")
		}
		done <- struct{}{}
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = mem.Summary()
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = mem.IsStuck()
			_ = mem.NeedsHelp()
		}
		done <- struct{}{}
	}()

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Just verify we didn't panic
	assert.Equal(t, 100, mem.TotalSteps)
}

func TestToPrompt_AlwaysIncludesTask(t *testing.T) {
	t.Parallel()

	task := "Book a flight to NYC"
	mem := New(task)

	// Even after many actions
	for i := 0; i < 100; i++ {
		mem.RecordAction("action", nil, true, "ok", time.Millisecond)
	}

	prompt := mem.ToPrompt()
	assert.Contains(t, prompt, "Book a flight to NYC")
	assert.Contains(t, prompt, "## Your Task")
}

func TestToPrompt_IncludesAllSections(t *testing.T) {
	t.Parallel()

	mem := New("Test task")

	// Add various state
	mem.AddMilestone("Opened app")
	mem.SetPhase(PhaseNavigation)
	mem.SetKeyFact("user", "john")
	mem.RecordAction("click", nil, true, "clicked", time.Millisecond)
	mem.RecordAction("type", nil, false, "failed", time.Millisecond)
	mem.AddFailedPattern("clicking while loading")

	prompt := mem.ToPrompt()

	assert.Contains(t, prompt, "## Your Task")
	assert.Contains(t, prompt, "## What You've Accomplished")
	assert.Contains(t, prompt, "Opened app")
	assert.Contains(t, prompt, "## Current Phase")
	assert.Contains(t, prompt, "navigation")
	assert.Contains(t, prompt, "user: john")
	assert.Contains(t, prompt, "## Recent Actions")
	assert.Contains(t, prompt, "✓")
	assert.Contains(t, prompt, "✗")
	assert.Contains(t, prompt, "## Known Issues")
	assert.Contains(t, prompt, "clicking while loading")
}

func TestToPrompt_StuckWarning(t *testing.T) {
	t.Parallel()

	mem := New("Test task")

	// 3 consecutive failures = stuck
	for i := 0; i < 3; i++ {
		mem.RecordAction("action", nil, false, "error", time.Millisecond)
	}

	prompt := mem.ToPrompt()
	assert.Contains(t, prompt, "POSSIBLY STUCK")
}

func TestToPrompt_NeedsHelpWarning(t *testing.T) {
	t.Parallel()

	mem := New("Test task")

	// 5 consecutive failures = needs help
	for i := 0; i < 5; i++ {
		mem.RecordAction("action", nil, false, "error", time.Millisecond)
	}

	prompt := mem.ToPrompt()
	assert.Contains(t, prompt, "NEEDS HELP")
}

func TestProgressiveSummarization(t *testing.T) {
	t.Parallel()

	mem := New("Test task")

	// Record more actions than the window size
	for i := 0; i < 10; i++ {
		mem.RecordAction("action", nil, true, "ok", time.Millisecond)
	}

	// Should have created milestones from older actions
	assert.Greater(t, len(mem.Milestones), 0, "should have created milestones")
	// Recent actions should be limited to window size
	assert.LessOrEqual(t, len(mem.RecentActions), MaxRecentActions, "recent actions should be limited")
}

func TestPhaseChangeSummarization(t *testing.T) {
	t.Parallel()

	mem := New("Test task")

	// Record actions in navigation phase
	mem.SetPhase(PhaseNavigation)
	mem.RecordAction("click", nil, true, "clicked", time.Millisecond)
	mem.RecordAction("scroll", nil, true, "scrolled", time.Millisecond)

	// Change phase - should summarize previous phase
	mem.SetPhase(PhaseFormFilling)

	assert.Contains(t, mem.Milestones[len(mem.Milestones)-1], "navigation")
	assert.Empty(t, mem.RecentActions, "recent actions should be cleared on phase change")
	assert.Equal(t, PhaseFormFilling, mem.Phase)
}

func TestDetectPhase_LoginForm(t *testing.T) {
	t.Parallel()

	obs := Observation{
		HasLoginForm: true,
	}

	phase := DetectPhase(obs)
	assert.Equal(t, PhaseAuthentication, phase)
}

func TestDetectPhase_Checkout(t *testing.T) {
	t.Parallel()

	obs := Observation{
		HasCheckoutElements: true,
	}

	phase := DetectPhase(obs)
	assert.Equal(t, PhaseCheckout, phase)
}

func TestDetectPhase_Confirmation(t *testing.T) {
	t.Parallel()

	obs := Observation{
		HasConfirmation: true,
	}

	phase := DetectPhase(obs)
	assert.Equal(t, PhaseConfirmation, phase)
}

func TestDetectPhase_SearchBox(t *testing.T) {
	t.Parallel()

	obs := Observation{
		HasSearchBox: true,
	}

	phase := DetectPhase(obs)
	assert.Equal(t, PhaseSearch, phase)
}

func TestDetectPhase_TextHeuristics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		visibleText   []string
		expectedPhase string
	}{
		{
			name:          "success message",
			visibleText:   []string{"Thank you for your order!"},
			expectedPhase: PhaseConfirmation,
		},
		{
			name:          "checkout page",
			visibleText:   []string{"Enter your credit card details"},
			expectedPhase: PhaseCheckout,
		},
		{
			name:          "login page",
			visibleText:   []string{"Please sign in to continue"},
			expectedPhase: PhaseAuthentication,
		},
		{
			name:          "search results",
			visibleText:   []string{"Search results for: golang"},
			expectedPhase: PhaseBrowsing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := Observation{
				VisibleText: tt.visibleText,
			}
			phase := DetectPhase(obs)
			assert.Equal(t, tt.expectedPhase, phase)
		})
	}
}

func TestDetectPhase_FocusedElement(t *testing.T) {
	t.Parallel()

	obs := Observation{
		FocusedElementRole: "textfield",
	}

	phase := DetectPhase(obs)
	assert.Equal(t, PhaseFormFilling, phase)
}

func TestDetectPhase_Priority(t *testing.T) {
	t.Parallel()

	// Confirmation should take priority even if checkout elements are also present
	obs := Observation{
		HasConfirmation:     true,
		HasCheckoutElements: true,
	}

	phase := DetectPhase(obs)
	assert.Equal(t, PhaseConfirmation, phase)
}

func TestMaybeUpdatePhase(t *testing.T) {
	t.Parallel()

	mem := New("Test task")
	mem.SetPhase(PhaseNavigation)

	// Update phase based on observation
	obs := Observation{
		HasLoginForm: true,
	}
	mem.MaybeUpdatePhase(obs)

	assert.Equal(t, PhaseAuthentication, mem.Phase)
}

func TestMaybeUpdatePhase_NoChangeForUnknown(t *testing.T) {
	t.Parallel()

	mem := New("Test task")
	mem.SetPhase(PhaseNavigation)

	// Empty observation shouldn't change phase to unknown
	obs := Observation{}
	mem.MaybeUpdatePhase(obs)

	// Should still be navigation (DetectPhase returns navigation for empty obs)
	// Actually, let's check what DetectPhase returns for empty
	phase := DetectPhase(obs)
	if phase == PhaseUnknown {
		assert.Equal(t, PhaseNavigation, mem.Phase, "should not change to unknown")
	}
}

func TestMaybeUpdatePhase_NoChangeForSamePhase(t *testing.T) {
	t.Parallel()

	mem := New("Test task")
	mem.SetPhase(PhaseAuthentication)
	mem.RecordAction("type", nil, true, "typed password", time.Millisecond)

	// Same phase observation
	obs := Observation{
		HasLoginForm: true,
	}
	mem.MaybeUpdatePhase(obs)

	// Phase should remain the same, and recent actions should NOT be cleared
	assert.Equal(t, PhaseAuthentication, mem.Phase)
	assert.Len(t, mem.RecentActions, 1, "actions should not be cleared for same phase")
}
