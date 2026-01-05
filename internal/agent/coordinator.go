// Package agent contains the ADK agent implementations for CUA.
//
// The agent system uses a single-loop architecture with the ReAct pattern:
//
//	LoopAgent (cua_loop)
//	└── LlmAgent (cua)
//	    └── All 10 tools: screenshot, find_element, click, type_text,
//	        key_press, scroll, drag, wait, complete_task, need_help
//
// The agent observes the screen, thinks about the next action, executes it,
// and repeats until calling complete_task or need_help.
package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
)

// LLM is an alias for model.LLM for use by callers.
type LLM = model.LLM

// CoordinatorConfig holds configuration for creating a Coordinator.
type CoordinatorConfig struct {
	// Model is the model for the CUA agent (configurable, default should be Pro for reasoning).
	Model model.LLM

	// MaxIterations is the maximum ReAct loop iterations (default: 50).
	MaxIterations int
}

// DefaultCoordinatorConfig returns the default configuration.
func DefaultCoordinatorConfig() CoordinatorConfig {
	return CoordinatorConfig{
		MaxIterations: 50,
	}
}

// NewCoordinatorAgent creates the single-loop ReAct agent for desktop automation.
//
// Architecture:
//
//	LoopAgent (cua_loop)
//	└── LlmAgent (cua)
//	    └── 10 tools attached: observation, action, and exit tools
//	    └── State injection via {task_context} placeholder
//
// The loop continues until:
//   - complete_task tool is called (task succeeded)
//   - need_help tool is called (needs human intervention)
//   - MaxIterations is reached
func NewCoordinatorAgent(m model.LLM) (agent.Agent, error) {
	return NewCoordinatorAgentWithConfig(CoordinatorConfig{
		Model:         m,
		MaxIterations: 50,
	})
}

// NewCoordinatorAgentWithConfig creates the coordinator with custom configuration.
func NewCoordinatorAgentWithConfig(cfg CoordinatorConfig) (agent.Agent, error) {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 50
	}

	// Create the single-loop CUA agent
	return NewCUAAgentWithConfig(CUAConfig{
		Model:         cfg.Model,
		MaxIterations: cfg.MaxIterations,
	})
}
