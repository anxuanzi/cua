package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"

	"github.com/anxuanzi/cua/internal/tools"
)

// ActionInstruction defines the system prompt for the Action Agent.
const ActionInstruction = `You are an action execution specialist for a desktop automation agent.

Your job: Execute desktop actions reliably and verify they succeeded.

## Capabilities
- Click at screen coordinates (left, right, or double click)
- Type text into focused elements
- Press keyboard keys with optional modifiers
- Scroll the mouse wheel
- Wait for specified durations

## Execution Rules
1. Execute exactly ONE action per request
2. Always verify the action succeeded when possible
3. Report any errors or unexpected results
4. Suggest alternatives if an action fails

## Action Verification
After each action, briefly note:
- What was attempted
- Whether it appeared to succeed
- Any visible changes on screen

## Response Format
ACTION_RESULT:
- Executed: [Description of the action taken]
- Parameters: [Key parameters used]
- Success: [yes/no/uncertain]
- Verification: [How you verified success, or why uncertain]
- Error: [Error message if failed, otherwise omit]
- Suggestion: [If failed, what to try next]

## Safety Notes
- Avoid clicking on system-critical areas without explicit instruction
- Be careful with keyboard shortcuts that could have unintended effects
- Wait appropriately after actions that trigger loading or transitions`

// NewActionAgent creates the Action Agent for executing desktop actions.
// It uses Gemini Flash for fast, reliable action execution.
func NewActionAgent(m model.LLM) (agent.Agent, error) {
	// Create the action tools
	clickTool, err := tools.NewClickTool()
	if err != nil {
		return nil, err
	}

	typeTextTool, err := tools.NewTypeTextTool()
	if err != nil {
		return nil, err
	}

	keyPressTool, err := tools.NewKeyPressTool()
	if err != nil {
		return nil, err
	}

	scrollTool, err := tools.NewScrollTool()
	if err != nil {
		return nil, err
	}

	waitTool, err := tools.NewWaitTool()
	if err != nil {
		return nil, err
	}

	return llmagent.New(llmagent.Config{
		Name:        "action_agent",
		Model:       m,
		Description: "Executes desktop actions like clicking, typing, and scrolling.",
		Instruction: ActionInstruction,
		Tools: []tool.Tool{
			clickTool,
			typeTextTool,
			keyPressTool,
			scrollTool,
			waitTool,
		},
	})
}
