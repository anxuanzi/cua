package memory

import (
	"fmt"
	"strings"
)

// Summarizer provides progressive summarization for long-running tasks.
// When recent actions pile up, it compresses older ones into milestones.
type Summarizer struct {
	// MinActionsToSummarize is the minimum number of actions before summarization triggers.
	// Default: 3
	MinActionsToSummarize int

	// ActionWindowSize is the number of recent actions to keep in detail.
	// Default: MaxRecentActions (5)
	ActionWindowSize int
}

// NewSummarizer creates a new Summarizer with default settings.
func NewSummarizer() *Summarizer {
	return &Summarizer{
		MinActionsToSummarize: 3,
		ActionWindowSize:      MaxRecentActions,
	}
}

// ShouldSummarize returns true if the action list should be summarized.
func (s *Summarizer) ShouldSummarize(actions []ActionResult) bool {
	return len(actions) > s.ActionWindowSize
}

// Summarize takes actions beyond the window and compresses them into milestones.
// It returns the new milestone descriptions and the remaining (recent) actions.
func (s *Summarizer) Summarize(actions []ActionResult) (milestones []string, remaining []ActionResult) {
	if len(actions) <= s.ActionWindowSize {
		return nil, actions
	}

	// Calculate how many actions to summarize
	toSummarize := actions[:len(actions)-s.ActionWindowSize]
	remaining = actions[len(actions)-s.ActionWindowSize:]

	// Group consecutive successful actions into milestones
	milestones = s.groupIntoMilestones(toSummarize)

	return milestones, remaining
}

// groupIntoMilestones groups actions into logical milestones.
// Successful action sequences are combined; failures are noted separately.
func (s *Summarizer) groupIntoMilestones(actions []ActionResult) []string {
	if len(actions) == 0 {
		return nil
	}

	var milestones []string
	var currentGroup []ActionResult

	for _, action := range actions {
		if !action.Success {
			// Flush current group if any
			if len(currentGroup) > 0 {
				milestones = append(milestones, s.summarizeGroup(currentGroup))
				currentGroup = nil
			}
			// Note the failure briefly
			milestones = append(milestones, fmt.Sprintf("Attempted %s (failed)", action.Action))
		} else {
			currentGroup = append(currentGroup, action)
		}
	}

	// Flush remaining group
	if len(currentGroup) > 0 {
		milestones = append(milestones, s.summarizeGroup(currentGroup))
	}

	return milestones
}

// summarizeGroup creates a single milestone from a group of successful actions.
func (s *Summarizer) summarizeGroup(actions []ActionResult) string {
	if len(actions) == 0 {
		return ""
	}

	if len(actions) == 1 {
		return s.describeAction(actions[0])
	}

	// For multiple actions, describe the outcome
	// Use the last action's result as the summary focus
	lastAction := actions[len(actions)-1]

	// Count action types
	actionCounts := make(map[string]int)
	for _, a := range actions {
		actionCounts[a.Action]++
	}

	// Build a summary
	var parts []string
	for action, count := range actionCounts {
		if count > 1 {
			parts = append(parts, fmt.Sprintf("%d %s actions", count, action))
		} else {
			parts = append(parts, action)
		}
	}

	if lastAction.Result != "" {
		return fmt.Sprintf("Completed %s (%s)", strings.Join(parts, ", "), lastAction.Result)
	}

	return fmt.Sprintf("Completed %s", strings.Join(parts, ", "))
}

// describeAction creates a human-readable description of a single action.
func (s *Summarizer) describeAction(action ActionResult) string {
	if action.Result != "" {
		return fmt.Sprintf("%s: %s", action.Action, action.Result)
	}
	return action.Action
}

// SummarizeForPhaseChange creates a milestone for a completed phase.
func (s *Summarizer) SummarizeForPhaseChange(oldPhase string, actions []ActionResult) string {
	if oldPhase == "" {
		return ""
	}

	successCount := 0
	for _, a := range actions {
		if a.Success {
			successCount++
		}
	}

	if successCount == 0 {
		return fmt.Sprintf("Attempted %s phase", oldPhase)
	}

	return fmt.Sprintf("Completed %s phase (%d actions)", oldPhase, successCount)
}
