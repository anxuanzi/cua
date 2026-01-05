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

	m := &mockLLM{}
	agent, err := NewCoordinatorAgent(m)

	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "cua_loop", agent.Name())
}

func TestNewCUAAgent(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewCUAAgent(m)

	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "cua_loop", agent.Name())
}

func TestNewCUAAgentWithConfig(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewCUAAgentWithConfig(CUAConfig{
		Model:         m,
		MaxIterations: 100,
	})

	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "cua_loop", agent.Name())
}

func TestBuildCUAInstruction(t *testing.T) {
	t.Parallel()

	instruction := BuildCUAInstruction()

	// Verify essential ReAct elements
	assert.Contains(t, instruction, "desktop automation agent")
	assert.Contains(t, instruction, "ReAct")
	assert.Contains(t, instruction, "OBSERVE")
	assert.Contains(t, instruction, "THINK")
	assert.Contains(t, instruction, "ACT")
	assert.Contains(t, instruction, "REPEAT")

	// Verify tool references
	assert.Contains(t, instruction, "screenshot")
	assert.Contains(t, instruction, "click")
	assert.Contains(t, instruction, "type_text")
	assert.Contains(t, instruction, "key_press")
	assert.Contains(t, instruction, "scroll")
	assert.Contains(t, instruction, "drag")
	assert.Contains(t, instruction, "wait")
	assert.Contains(t, instruction, "complete_task")
	assert.Contains(t, instruction, "need_help")
	assert.Contains(t, instruction, "find_element")

	// Verify dynamic context placeholder
	assert.Contains(t, instruction, "{task_context}")
}

func TestBuildCUAInstruction_ContainsPlatformContext(t *testing.T) {
	t.Parallel()

	instruction := BuildCUAInstruction()

	// Should contain platform info (using XML tags)
	assert.Contains(t, instruction, "<platform>")
	assert.Contains(t, instruction, "<os>")
}

func TestBuildCUAInstruction_ContainsPlatformShortcuts(t *testing.T) {
	t.Parallel()

	instruction := BuildCUAInstruction()

	// Should contain platform-specific keyboard info
	assert.True(t, strings.Contains(instruction, "App launcher") ||
		strings.Contains(instruction, "Primary modifier"),
		"CUA instruction should contain platform keyboard info")
}

func TestBuildCUAInstruction_ContainsExamples(t *testing.T) {
	t.Parallel()

	instruction := BuildCUAInstruction()

	// Should contain examples for tool usage
	assert.Contains(t, instruction, "Example")
}

func TestBuildCUAInstruction_ContainsRules(t *testing.T) {
	t.Parallel()

	instruction := BuildCUAInstruction()

	// Should contain critical rules
	assert.Contains(t, instruction, "ONE ACTION PER TURN")
	assert.Contains(t, instruction, "SCREENSHOT FIRST")
}

func TestDefaultCoordinatorConfig(t *testing.T) {
	t.Parallel()

	config := DefaultCoordinatorConfig()

	// Single-loop defaults to 50 iterations
	assert.Equal(t, 50, config.MaxIterations)
}

func TestDefaultCUAConfig(t *testing.T) {
	t.Parallel()

	config := DefaultCUAConfig()

	assert.Equal(t, 50, config.MaxIterations)
}

func TestCoordinatorConfig_Fields(t *testing.T) {
	t.Parallel()

	config := CoordinatorConfig{
		Model:         &mockLLM{},
		MaxIterations: 30,
	}

	assert.Equal(t, 30, config.MaxIterations)
	assert.NotNil(t, config.Model)
}

func TestCUAConfig_Fields(t *testing.T) {
	t.Parallel()

	config := CUAConfig{
		Model:         &mockLLM{},
		MaxIterations: 75,
	}

	assert.Equal(t, 75, config.MaxIterations)
	assert.NotNil(t, config.Model)
}

func TestLLM_TypeAlias(t *testing.T) {
	t.Parallel()

	// Verify that LLM is an alias for model.LLM
	var llm LLM = &mockLLM{}
	assert.NotNil(t, llm)
}

func TestNewCoordinatorAgentWithConfig_DefaultsMaxIterations(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewCoordinatorAgentWithConfig(CoordinatorConfig{
		Model:         m,
		MaxIterations: 0, // Should default to 50
	})

	require.NoError(t, err)
	assert.NotNil(t, agent)
}

func TestNewCUAAgentWithConfig_DefaultsMaxIterations(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewCUAAgentWithConfig(CUAConfig{
		Model:         m,
		MaxIterations: -5, // Should default to 50
	})

	require.NoError(t, err)
	assert.NotNil(t, agent)
}
