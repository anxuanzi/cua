// Package agent contains the ADK agent implementations for CUA.
//
// The agent system uses a multi-agent architecture with the ReAct pattern:
//   - Coordinator: LoopAgent that repeats until task completion
//   - ReAct Step: SequentialAgent running perception → decision → action
//   - Perception Agent: Analyzes screen state
//   - Decision Agent: Decides next action (can call exit_loop to stop)
//   - Action Agent: Executes desktop actions
package agent

import (
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/workflowagents/loopagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/model"
)

// LLM is an alias for model.LLM for use by callers.
type LLM = model.LLM

// CoordinatorConfig holds configuration for creating a Coordinator.
type CoordinatorConfig struct {
	// CoordinatorModel is the model for the decision agent (typically Gemini Pro).
	CoordinatorModel model.LLM

	// SubAgentModel is the model for perception/action agents (typically Gemini Flash).
	SubAgentModel model.LLM

	// MaxIterations is the maximum ReAct loop iterations (default: 20).
	MaxIterations int
}

// DefaultCoordinatorConfig returns the default configuration.
func DefaultCoordinatorConfig() CoordinatorConfig {
	return CoordinatorConfig{
		MaxIterations: 20,
	}
}

// NewCoordinatorAgent creates the ReAct loop agent for desktop automation.
//
// Architecture:
//
//	LoopAgent (coordinator)
//	└── SequentialAgent (react_step)
//	    ├── LlmAgent (perception) → OutputKey: "screen_state"
//	    ├── LlmAgent (decision)   → OutputKey: "next_action"
//	    └── LlmAgent (action)     → OutputKey: "action_result"
//
// The loop continues until:
//   - Decision agent calls exit_loop tool (sets Escalate = true)
//   - MaxIterations is reached
func NewCoordinatorAgent(coordinatorModel, subAgentModel model.LLM) (agent.Agent, error) {
	return NewCoordinatorAgentWithConfig(CoordinatorConfig{
		CoordinatorModel: coordinatorModel,
		SubAgentModel:    subAgentModel,
		MaxIterations:    20,
	})
}

// NewCoordinatorAgentWithConfig creates the coordinator with custom configuration.
func NewCoordinatorAgentWithConfig(cfg CoordinatorConfig) (agent.Agent, error) {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 20
	}

	// Create sub-agents
	perceptionAgent, err := NewPerceptionAgent(cfg.SubAgentModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create perception agent: %w", err)
	}

	decisionAgent, err := NewDecisionAgent(cfg.CoordinatorModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create decision agent: %w", err)
	}

	actionAgent, err := NewActionAgent(cfg.SubAgentModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create action agent: %w", err)
	}

	// Create ReAct step: sequential execution of perception → decision → action
	reactStep, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        "react_step",
			Description: "One iteration of the ReAct loop: observe, decide, act.",
			SubAgents:   []agent.Agent{perceptionAgent, decisionAgent, actionAgent},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create react step: %w", err)
	}

	// Create ReAct loop: repeats until exit_loop is called or max iterations
	reactLoop, err := loopagent.New(loopagent.Config{
		MaxIterations: uint(cfg.MaxIterations),
		AgentConfig: agent.Config{
			Name:        "coordinator",
			Description: "Desktop automation agent using the ReAct pattern.",
			SubAgents:   []agent.Agent{reactStep},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create react loop: %w", err)
	}

	return reactLoop, nil
}
