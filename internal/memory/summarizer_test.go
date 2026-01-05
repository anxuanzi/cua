package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSummarizer(t *testing.T) {
	t.Parallel()

	s := NewSummarizer()

	assert.Equal(t, 3, s.MinActionsToSummarize)
	assert.Equal(t, MaxRecentActions, s.ActionWindowSize)
}

func TestShouldSummarize(t *testing.T) {
	t.Parallel()

	s := NewSummarizer()

	tests := []struct {
		name           string
		numActions     int
		expectedResult bool
	}{
		{"empty actions", 0, false},
		{"few actions", 3, false},
		{"at window size", MaxRecentActions, false},
		{"over window size", MaxRecentActions + 1, true},
		{"way over window size", MaxRecentActions + 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := makeActions(tt.numActions, true)
			result := s.ShouldSummarize(actions)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestSummarize_NoSummarizationNeeded(t *testing.T) {
	t.Parallel()

	s := NewSummarizer()
	actions := makeActions(3, true)

	milestones, remaining := s.Summarize(actions)

	assert.Empty(t, milestones)
	assert.Equal(t, actions, remaining)
}

func TestSummarize_BasicSummarization(t *testing.T) {
	t.Parallel()

	s := NewSummarizer()
	// Create 8 actions (3 will be summarized, 5 will remain)
	actions := makeActions(8, true)

	milestones, remaining := s.Summarize(actions)

	assert.NotEmpty(t, milestones)
	assert.Len(t, remaining, MaxRecentActions)
	// The remaining actions should be the most recent ones
	assert.Equal(t, 4, remaining[0].StepNumber)
	assert.Equal(t, 8, remaining[len(remaining)-1].StepNumber)
}

func TestSummarize_WithFailures(t *testing.T) {
	t.Parallel()

	s := NewSummarizer()
	actions := []ActionResult{
		{StepNumber: 1, Action: "click", Success: true, Result: "ok"},
		{StepNumber: 2, Action: "type", Success: true, Result: "typed"},
		{StepNumber: 3, Action: "click", Success: false, Result: "element not found"},
		{StepNumber: 4, Action: "click", Success: true, Result: "clicked"},
		{StepNumber: 5, Action: "type", Success: true, Result: "typed"},
		{StepNumber: 6, Action: "scroll", Success: true, Result: "scrolled"},
		{StepNumber: 7, Action: "click", Success: true, Result: "clicked"},
		{StepNumber: 8, Action: "type", Success: true, Result: "typed"},
	}

	milestones, remaining := s.Summarize(actions)

	assert.NotEmpty(t, milestones)
	// Should have at least one milestone noting the failure
	foundFailure := false
	for _, m := range milestones {
		if containsString(m, "failed") {
			foundFailure = true
			break
		}
	}
	assert.True(t, foundFailure, "should note the failure in milestones")
	assert.Len(t, remaining, MaxRecentActions)
}

func TestGroupIntoMilestones_SingleAction(t *testing.T) {
	t.Parallel()

	s := NewSummarizer()
	actions := []ActionResult{
		{Action: "click", Success: true, Result: "clicked button"},
	}

	milestones := s.groupIntoMilestones(actions)

	assert.Len(t, milestones, 1)
	assert.Contains(t, milestones[0], "click")
}

func TestGroupIntoMilestones_MultipleSuccessfulActions(t *testing.T) {
	t.Parallel()

	s := NewSummarizer()
	actions := []ActionResult{
		{Action: "click", Success: true, Result: "clicked"},
		{Action: "type", Success: true, Result: "typed"},
		{Action: "click", Success: true, Result: "final click"},
	}

	milestones := s.groupIntoMilestones(actions)

	// Should combine into one or more milestones
	assert.NotEmpty(t, milestones)
}

func TestGroupIntoMilestones_MixedSuccess(t *testing.T) {
	t.Parallel()

	s := NewSummarizer()
	actions := []ActionResult{
		{Action: "click", Success: true, Result: "clicked"},
		{Action: "type", Success: false, Result: "failed"},
		{Action: "click", Success: true, Result: "clicked"},
	}

	milestones := s.groupIntoMilestones(actions)

	assert.GreaterOrEqual(t, len(milestones), 2, "should have separate milestones around the failure")
}

func TestSummarizeForPhaseChange(t *testing.T) {
	t.Parallel()

	s := NewSummarizer()

	tests := []struct {
		name            string
		oldPhase        string
		actions         []ActionResult
		expectMilestone bool
		expectContains  string
	}{
		{
			name:            "empty phase",
			oldPhase:        "",
			actions:         makeActions(3, true),
			expectMilestone: false,
		},
		{
			name:            "navigation phase with actions",
			oldPhase:        PhaseNavigation,
			actions:         makeActions(3, true),
			expectMilestone: true,
			expectContains:  "navigation",
		},
		{
			name:            "form filling phase",
			oldPhase:        PhaseFormFilling,
			actions:         makeActions(5, true),
			expectMilestone: true,
			expectContains:  "5 actions",
		},
		{
			name:            "phase with no successful actions",
			oldPhase:        PhaseAuthentication,
			actions:         makeActions(3, false),
			expectMilestone: true,
			expectContains:  "Attempted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SummarizeForPhaseChange(tt.oldPhase, tt.actions)

			if tt.expectMilestone {
				assert.NotEmpty(t, result)
				if tt.expectContains != "" {
					assert.Contains(t, result, tt.expectContains)
				}
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

// Helper functions

func makeActions(n int, success bool) []ActionResult {
	actions := make([]ActionResult, n)
	for i := 0; i < n; i++ {
		actions[i] = ActionResult{
			StepNumber: i + 1,
			Action:     "action",
			Success:    success,
			Result:     "result",
			Duration:   time.Millisecond,
			Timestamp:  time.Now(),
		}
	}
	return actions
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsSubstr(s, substr)))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
