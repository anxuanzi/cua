package agent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/adk/model"
)

// mockLLM is a mock implementation of model.LLM for testing.
type mockLLM struct {
	model.LLM
}

func TestNewCoordinatorAgent(t *testing.T) {
	t.Parallel()

	coordinatorModel := &mockLLM{}
	subAgentModel := &mockLLM{}

	agent, err := NewCoordinatorAgent(coordinatorModel, subAgentModel)

	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "coordinator", agent.Name())
}

func TestBuildPerceptionInstruction(t *testing.T) {
	t.Parallel()

	instruction := BuildPerceptionInstruction()

	// Verify essential elements
	assert.Contains(t, instruction, "screen analysis specialist")
	assert.Contains(t, instruction, "screenshot")
	assert.Contains(t, instruction, "find_element")
	assert.Contains(t, instruction, "Current State:")
	assert.Contains(t, instruction, "Key Elements")
	assert.Contains(t, instruction, "Observations")
}

func TestBuildDecisionInstruction(t *testing.T) {
	t.Parallel()

	instruction := BuildDecisionInstruction()

	// Verify essential elements
	assert.Contains(t, instruction, "decision-making component")
	assert.Contains(t, instruction, "{screen_state}") // Template variable for ADK state
	assert.Contains(t, instruction, "exit_loop")
	assert.Contains(t, instruction, "Analysis:")
	assert.Contains(t, instruction, "Decision:")
	assert.Contains(t, instruction, "ONE action per decision")
}

func TestBuildActionInstruction(t *testing.T) {
	t.Parallel()

	instruction := BuildActionInstruction()

	// Verify essential elements
	assert.Contains(t, instruction, "action executor")
	assert.Contains(t, instruction, "{next_action}") // Template variable for ADK state
	assert.Contains(t, instruction, "click")
	assert.Contains(t, instruction, "type_text")
	assert.Contains(t, instruction, "key_press")
	assert.Contains(t, instruction, "scroll")
	assert.Contains(t, instruction, "wait")
	assert.Contains(t, instruction, "drag")
	assert.Contains(t, instruction, "Executed:")
	assert.Contains(t, instruction, "Result:")
}

func TestBuildPerceptionInstruction_ContainsPlatformContext(t *testing.T) {
	t.Parallel()

	instruction := BuildPerceptionInstruction()

	// Should contain platform info (using XML tags)
	assert.Contains(t, instruction, "<platform>")
	assert.Contains(t, instruction, "<os>")
}

func TestBuildDecisionInstruction_ContainsPlatformShortcuts(t *testing.T) {
	t.Parallel()

	instruction := BuildDecisionInstruction()

	// Should contain platform-specific keyboard info
	assert.True(t, strings.Contains(instruction, "App launcher") ||
		strings.Contains(instruction, "Primary modifier"),
		"Decision instruction should contain platform keyboard info")
}

func TestBuildActionInstruction_ContainsToolUsageWarnings(t *testing.T) {
	t.Parallel()

	instruction := BuildActionInstruction()

	// Verify tool usage warnings are present (to prevent agent prefixing issues)
	assert.Contains(t, instruction, "EXECUTE EXACTLY")
	assert.Contains(t, instruction, "without any prefix")
}

func TestDefaultCoordinatorConfig(t *testing.T) {
	t.Parallel()

	config := DefaultCoordinatorConfig()

	assert.Equal(t, 20, config.MaxIterations)
}

func TestCoordinatorConfig_Fields(t *testing.T) {
	t.Parallel()

	config := CoordinatorConfig{
		CoordinatorModel: &mockLLM{},
		SubAgentModel:    &mockLLM{},
		MaxIterations:    30,
	}

	assert.Equal(t, 30, config.MaxIterations)
	assert.NotNil(t, config.CoordinatorModel)
	assert.NotNil(t, config.SubAgentModel)
}

func TestNewPerceptionAgent(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewPerceptionAgent(m)

	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "perception", agent.Name())
}

func TestNewDecisionAgent(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewDecisionAgent(m)

	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "decision", agent.Name())
}

func TestNewActionAgent(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewActionAgent(m)

	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "action", agent.Name())
}

func TestLLM_TypeAlias(t *testing.T) {
	t.Parallel()

	// Verify that LLM is an alias for model.LLM
	var llm LLM = &mockLLM{}
	assert.NotNil(t, llm)
}

func TestExitLoopArgs(t *testing.T) {
	t.Parallel()

	args := ExitLoopArgs{
		Summary: "Task completed successfully",
	}

	assert.Equal(t, "Task completed successfully", args.Summary)
}

func TestExitLoopResult(t *testing.T) {
	t.Parallel()

	result := ExitLoopResult{
		Success: true,
		Summary: "Opened calculator and computed 42+8=50",
	}

	assert.True(t, result.Success)
	assert.Equal(t, "Opened calculator and computed 42+8=50", result.Summary)
}
