// Package agent contains the single-loop CUA agent implementation.
package agent

import (
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/loopagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"

	"github.com/anxuanzi/cua/internal/tools"
)

// CUAConfig holds configuration for the CUA agent.
type CUAConfig struct {
	// Model is the LLM to use (default: should be set by caller, typically Gemini Pro).
	Model model.LLM

	// MaxIterations is the maximum ReAct loop iterations (default: 50).
	MaxIterations int
}

// DefaultCUAConfig returns the default CUA configuration.
func DefaultCUAConfig() CUAConfig {
	return CUAConfig{
		MaxIterations: 50,
	}
}

// NewCUAAgent creates the single-loop CUA agent with all tools.
//
// Architecture:
//
//	LoopAgent (cua_loop)
//	└── LlmAgent (cua)
//	    └── All 10 tools directly attached
//	    └── State injection via {task_context}
//	    └── Exit via complete_task/need_help tools
//
// The loop continues until:
//   - complete_task or need_help tool is called (sets Escalate = true)
//   - MaxIterations is reached
func NewCUAAgent(m model.LLM) (agent.Agent, error) {
	return NewCUAAgentWithConfig(CUAConfig{
		Model:         m,
		MaxIterations: 50,
	})
}

// NewCUAAgentWithConfig creates the CUA agent with custom configuration.
func NewCUAAgentWithConfig(cfg CUAConfig) (agent.Agent, error) {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 50
	}

	// Create all tools
	allTools, err := createAllTools()
	if err != nil {
		return nil, fmt.Errorf("failed to create tools: %w", err)
	}

	// Build the world-class ReAct instruction
	instruction := BuildCUAInstruction()

	// Create the core CUA LLM agent with all tools
	cuaAgent, err := llmagent.New(llmagent.Config{
		Name:        "cua",
		Model:       cfg.Model,
		Description: "Desktop automation agent using ReAct pattern. Observes screen, thinks, and acts.",
		Instruction: instruction,
		Tools:       allTools,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create CUA agent: %w", err)
	}

	// Wrap in LoopAgent for continuous operation until task completion
	cuaLoop, err := loopagent.New(loopagent.Config{
		MaxIterations: uint(cfg.MaxIterations),
		AgentConfig: agent.Config{
			Name:        "cua_loop",
			Description: "ReAct loop that continues until task is complete or help is needed.",
			SubAgents:   []agent.Agent{cuaAgent},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create CUA loop: %w", err)
	}

	return cuaLoop, nil
}

// createAllTools creates all 10 tools for the CUA agent.
func createAllTools() ([]tool.Tool, error) {
	var allTools []tool.Tool

	// Observation tools
	screenshotTool, err := tools.NewScreenshotTool()
	if err != nil {
		return nil, fmt.Errorf("screenshot tool: %w", err)
	}
	allTools = append(allTools, screenshotTool)

	findElementTool, err := tools.NewFindElementTool()
	if err != nil {
		return nil, fmt.Errorf("find_element tool: %w", err)
	}
	allTools = append(allTools, findElementTool)

	// Action tools
	clickTool, err := tools.NewClickTool()
	if err != nil {
		return nil, fmt.Errorf("click tool: %w", err)
	}
	allTools = append(allTools, clickTool)

	typeTextTool, err := tools.NewTypeTextTool()
	if err != nil {
		return nil, fmt.Errorf("type_text tool: %w", err)
	}
	allTools = append(allTools, typeTextTool)

	keyPressTool, err := tools.NewKeyPressTool()
	if err != nil {
		return nil, fmt.Errorf("key_press tool: %w", err)
	}
	allTools = append(allTools, keyPressTool)

	scrollTool, err := tools.NewScrollTool()
	if err != nil {
		return nil, fmt.Errorf("scroll tool: %w", err)
	}
	allTools = append(allTools, scrollTool)

	dragTool, err := tools.NewDragTool()
	if err != nil {
		return nil, fmt.Errorf("drag tool: %w", err)
	}
	allTools = append(allTools, dragTool)

	waitTool, err := tools.NewWaitTool()
	if err != nil {
		return nil, fmt.Errorf("wait tool: %w", err)
	}
	allTools = append(allTools, waitTool)

	// Exit tools
	completeTaskTool, err := tools.NewCompleteTaskTool()
	if err != nil {
		return nil, fmt.Errorf("complete_task tool: %w", err)
	}
	allTools = append(allTools, completeTaskTool)

	needHelpTool, err := tools.NewNeedHelpTool()
	if err != nil {
		return nil, fmt.Errorf("need_help tool: %w", err)
	}
	allTools = append(allTools, needHelpTool)

	return allTools, nil
}
