package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
)

// LLM is an alias for model.LLM for use by callers.
type LLM = model.LLM

// NewCoordinatorAgent creates the Coordinator Agent that orchestrates the ReAct loop.
// It uses Gemini Pro for complex reasoning and delegates perception/action to sub-agents.
// The instruction is built dynamically to include platform-specific context.
func NewCoordinatorAgent(coordinatorModel, subAgentModel model.LLM) (agent.Agent, error) {
	// Create the perception agent (uses Flash for speed)
	perceptionAgent, err := NewPerceptionAgent(subAgentModel)
	if err != nil {
		return nil, err
	}

	// Create the action agent (uses Flash for reliability)
	actionAgent, err := NewActionAgent(subAgentModel)
	if err != nil {
		return nil, err
	}

	// Build instruction with platform context
	instruction := BuildCoordinatorInstruction()

	// Create the coordinator with sub-agents
	return llmagent.New(llmagent.Config{
		Name:        "coordinator",
		Model:       coordinatorModel,
		Description: "Orchestrates desktop automation tasks using the ReAct pattern.",
		Instruction: instruction,
		SubAgents: []agent.Agent{
			perceptionAgent,
			actionAgent,
		},
	})
}

// CoordinatorConfig holds configuration for creating a Coordinator.
type CoordinatorConfig struct {
	// CoordinatorModel is the model for the coordinator (typically Gemini Pro).
	CoordinatorModel model.LLM

	// SubAgentModel is the model for sub-agents (typically Gemini Flash).
	SubAgentModel model.LLM

	// MaxActions is the maximum number of actions to take (default: 50).
	MaxActions int

	// SafetyLevel controls how cautious the agent is.
	SafetyLevel string
}

// DefaultCoordinatorConfig returns the default configuration.
func DefaultCoordinatorConfig() CoordinatorConfig {
	return CoordinatorConfig{
		MaxActions:  50,
		SafetyLevel: "normal",
	}
}
