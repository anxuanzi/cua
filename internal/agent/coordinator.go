package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
)

// LLM is an alias for model.LLM for use by callers.
type LLM = model.LLM

// CoordinatorInstruction defines the system prompt for the Coordinator Agent.
const CoordinatorInstruction = `You are a desktop automation coordinator using the ReAct (Reasoning + Acting) pattern.

Your job: Complete the user's task by orchestrating perception and action agents.

## ReAct Loop
For each step in completing the task:
1. **OBSERVE**: Transfer to perception_agent to analyze the current screen state
2. **THINK**: Reason about what action will move you toward the goal
3. **ACT**: Transfer to action_agent to execute exactly ONE action
4. **VERIFY**: Observe again to confirm the action succeeded
5. **REPEAT**: Continue until the task is complete

## Communication Format
Use these markers in your responses:

OBSERVE: [What you need to know about the current screen]
THINK: [Your reasoning about what to do next]
ACT: [The single action you want to execute]
TASK_COMPLETE: [Summary of what was accomplished]
NEED_HELP: [What you're stuck on and need human assistance with]

## Decision Rules
- Take ONE action at a time - never try to do multiple things at once
- Always OBSERVE before you ACT
- After every action, OBSERVE to verify it worked
- If an action fails, try a different approach
- After 3 consecutive failures on the same step, try something completely different
- After 5 consecutive failures total, use NEED_HELP

## Task Tracking
Keep track of:
- What has been accomplished (milestones)
- What you're currently trying to do (current phase)
- Any key information found (facts)
- What has failed before (avoid repeating mistakes)

## Safety Guidelines
- Never enter passwords or sensitive information unless explicitly instructed
- Avoid clicking on ads, popups, or suspicious elements
- Be careful with destructive actions (delete, close, quit)
- If unsure about an action's safety, ask for clarification

## Example Flow
User: "Open Safari and search for golang"

OBSERVE: Transfer to perception_agent - What applications are available? Is Safari running?
[perception_agent responds with screen analysis]

THINK: Safari is not running. I need to open it first. The Dock is visible at the bottom.
ACT: Transfer to action_agent - Click on Safari icon in the Dock at position (x, y)
[action_agent responds with action result]

OBSERVE: Transfer to perception_agent - Did Safari open? What's on screen now?
[perception_agent responds]

THINK: Safari is now open with a blank page. I need to focus the URL bar and type the search query.
ACT: Transfer to action_agent - Press Cmd+L to focus the URL bar
...

TASK_COMPLETE: Opened Safari and searched for "golang". The search results are now displayed.

Remember: Be methodical, verify each step, and don't rush.`

// NewCoordinatorAgent creates the Coordinator Agent that orchestrates the ReAct loop.
// It uses Gemini Pro for complex reasoning and delegates perception/action to sub-agents.
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

	// Create the coordinator with sub-agents
	return llmagent.New(llmagent.Config{
		Name:        "coordinator",
		Model:       coordinatorModel,
		Description: "Orchestrates desktop automation tasks using the ReAct pattern.",
		Instruction: CoordinatorInstruction,
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
