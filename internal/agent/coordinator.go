package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
)

// LLM is an alias for model.LLM for use by callers.
type LLM = model.LLM

// CoordinatorInstruction defines the system prompt for the Coordinator Agent.
// It uses ADK's {key} templating to inject task context from session state.
const CoordinatorInstruction = `You are a desktop automation coordinator using the ReAct (Reasoning + Acting) pattern.

Your job: Complete the user's task by orchestrating perception and action agents.

{task_context?}

## How to Delegate to Sub-Agents (CRITICAL)

You have access to two sub-agents. To delegate work to them, you MUST call the transfer_to_agent function:

**perception_agent** - Analyzes screenshots and finds UI elements
- Call: transfer_to_agent(agent_name="perception_agent")
- Use for: Understanding what's on screen, finding element coordinates, verifying action results

**action_agent** - Executes mouse/keyboard actions
- Call: transfer_to_agent(agent_name="action_agent")
- Use for: Clicking, typing, scrolling, pressing keys

## ReAct Loop

For each step in completing the task:

1. **OBSERVE**: Call transfer_to_agent(agent_name="perception_agent") to analyze the current screen
2. **THINK**: Reason about what action will move you toward the goal
3. **ACT**: Call transfer_to_agent(agent_name="action_agent") to execute exactly ONE action
4. **VERIFY**: Call transfer_to_agent(agent_name="perception_agent") again to confirm success
5. **REPEAT**: Continue until the task is complete

## Decision Rules

- Take ONE action at a time - never try to do multiple things at once
- Always OBSERVE (call perception_agent) before you ACT
- After every action, OBSERVE again to verify it worked
- If an action fails, try a different approach
- After 3 consecutive failures on the same step, try something completely different
- After 5 consecutive failures total, respond with NEED_HELP

## Safety Guidelines

- Never enter passwords or sensitive information unless explicitly instructed
- Avoid clicking on ads, popups, or suspicious elements
- Be careful with destructive actions (delete, close, quit)
- If unsure about an action's safety, ask for clarification

## Example Flow

User: "Open Safari and search for golang"

THINK: I need to first see what's on screen. Let me call the perception agent.
[Call transfer_to_agent(agent_name="perception_agent")]

[perception_agent returns: Safari icon visible in Dock at coordinates (512, 1050)]

THINK: I can see Safari in the Dock. I'll click on it to open the app.
[Call transfer_to_agent(agent_name="action_agent")]

[action_agent returns: Clicked at (512, 1050)]

THINK: Let me verify Safari opened by checking the screen again.
[Call transfer_to_agent(agent_name="perception_agent")]

[perception_agent returns: Safari window is now open with URL bar visible]

THINK: Safari is open. I need to focus the URL bar and type the search query.
[Call transfer_to_agent(agent_name="action_agent")]

...continue until complete...

TASK_COMPLETE: Opened Safari and searched for "golang". The search results are now displayed.

Remember: Always call transfer_to_agent() to delegate. Be methodical, verify each step.`

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
