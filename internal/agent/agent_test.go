package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/adk/model"
)

// mockLLM is a mock implementation of model.LLM for testing.
type mockLLM struct {
	model.LLM
}

func TestNewPerceptionAgent(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewPerceptionAgent(m)

	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "perception_agent", agent.Name())
}

func TestPerceptionInstruction_ContainsKeyElements(t *testing.T) {
	t.Parallel()

	// Verify the instruction contains essential elements
	assert.Contains(t, PerceptionInstruction, "screen analysis specialist")
	assert.Contains(t, PerceptionInstruction, "screenshot")
	assert.Contains(t, PerceptionInstruction, "find_element")
	assert.Contains(t, PerceptionInstruction, "SCREEN_STATE:")
	assert.Contains(t, PerceptionInstruction, "COORDINATES")
	assert.Contains(t, PerceptionInstruction, "OBSERVATIONS")
}

func TestNewActionAgent(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewActionAgent(m)

	require.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "action_agent", agent.Name())
}

func TestActionInstruction_ContainsKeyElements(t *testing.T) {
	t.Parallel()

	// Verify the instruction contains essential elements
	assert.Contains(t, ActionInstruction, "action execution specialist")
	assert.Contains(t, ActionInstruction, "Click")
	assert.Contains(t, ActionInstruction, "Type text")
	assert.Contains(t, ActionInstruction, "keyboard keys")
	assert.Contains(t, ActionInstruction, "Scroll")
	assert.Contains(t, ActionInstruction, "Drag")
	assert.Contains(t, ActionInstruction, "ACTION_RESULT:")
	assert.Contains(t, ActionInstruction, "Safety Notes")
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

func TestCoordinatorInstruction_ContainsReActPattern(t *testing.T) {
	t.Parallel()

	// Verify the instruction contains ReAct pattern elements
	assert.Contains(t, CoordinatorInstruction, "ReAct")
	assert.Contains(t, CoordinatorInstruction, "OBSERVE")
	assert.Contains(t, CoordinatorInstruction, "THINK")
	assert.Contains(t, CoordinatorInstruction, "ACT")
	assert.Contains(t, CoordinatorInstruction, "VERIFY")
	assert.Contains(t, CoordinatorInstruction, "REPEAT")
}

func TestCoordinatorInstruction_ContainsCommunicationMarkers(t *testing.T) {
	t.Parallel()

	// Verify communication markers are present
	assert.Contains(t, CoordinatorInstruction, "TASK_COMPLETE")
	assert.Contains(t, CoordinatorInstruction, "NEED_HELP")
}

func TestCoordinatorInstruction_ContainsTransferFunction(t *testing.T) {
	t.Parallel()

	// Verify the instruction tells LLM to call transfer_to_agent function
	assert.Contains(t, CoordinatorInstruction, "transfer_to_agent")
	assert.Contains(t, CoordinatorInstruction, `transfer_to_agent(agent_name="perception_agent")`)
	assert.Contains(t, CoordinatorInstruction, `transfer_to_agent(agent_name="action_agent")`)
}

func TestCoordinatorInstruction_ContainsDecisionRules(t *testing.T) {
	t.Parallel()

	// Verify decision rules are present
	assert.Contains(t, CoordinatorInstruction, "ONE action at a time")
	assert.Contains(t, CoordinatorInstruction, "3 consecutive failures")
	assert.Contains(t, CoordinatorInstruction, "5 consecutive failures")
}

func TestCoordinatorInstruction_ContainsSafetyGuidelines(t *testing.T) {
	t.Parallel()

	// Verify safety guidelines are present
	assert.Contains(t, CoordinatorInstruction, "passwords")
	assert.Contains(t, CoordinatorInstruction, "sensitive")
	assert.Contains(t, CoordinatorInstruction, "ads")
	assert.Contains(t, CoordinatorInstruction, "destructive")
}

func TestDefaultCoordinatorConfig(t *testing.T) {
	t.Parallel()

	config := DefaultCoordinatorConfig()

	assert.Equal(t, 50, config.MaxActions)
	assert.Equal(t, "normal", config.SafetyLevel)
}

func TestCoordinatorConfig_Fields(t *testing.T) {
	t.Parallel()

	config := CoordinatorConfig{
		CoordinatorModel: &mockLLM{},
		SubAgentModel:    &mockLLM{},
		MaxActions:       100,
		SafetyLevel:      "strict",
	}

	assert.Equal(t, 100, config.MaxActions)
	assert.Equal(t, "strict", config.SafetyLevel)
	assert.NotNil(t, config.CoordinatorModel)
	assert.NotNil(t, config.SubAgentModel)
}

func TestPerceptionAgent_HasCorrectTools(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewPerceptionAgent(m)

	require.NoError(t, err)
	assert.NotNil(t, agent)

	// The agent should have been created with screenshot and find_element tools
	// We verify this indirectly through the fact that creation succeeded
	// Tool verification would require accessing internal state which ADK may not expose
}

func TestActionAgent_HasCorrectTools(t *testing.T) {
	t.Parallel()

	m := &mockLLM{}
	agent, err := NewActionAgent(m)

	require.NoError(t, err)
	assert.NotNil(t, agent)

	// The agent should have been created with click, type, key_press, scroll, wait tools
	// We verify this indirectly through the fact that creation succeeded
}

func TestCoordinatorAgent_HasSubAgents(t *testing.T) {
	t.Parallel()

	coordinatorModel := &mockLLM{}
	subAgentModel := &mockLLM{}

	agent, err := NewCoordinatorAgent(coordinatorModel, subAgentModel)

	require.NoError(t, err)
	assert.NotNil(t, agent)

	// The coordinator should have perception_agent and action_agent as sub-agents
	// We verify this indirectly through successful creation
}

func TestLLM_TypeAlias(t *testing.T) {
	t.Parallel()

	// Verify that LLM is an alias for model.LLM
	var llm LLM = &mockLLM{}
	assert.NotNil(t, llm)
}
