package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"

	"github.com/anxuanzi/cua/internal/tools"
)

// NewActionAgent creates the Action Agent for executing desktop actions.
// It uses Gemini Flash for fast, reliable action execution.
// Output is saved to "action_result" in session state.
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

	dragTool, err := tools.NewDragTool()
	if err != nil {
		return nil, err
	}

	instruction := BuildActionInstruction()

	return llmagent.New(llmagent.Config{
		Name:        "action",
		Model:       m,
		Description: "Executes desktop actions like clicking, typing, and scrolling.",
		Instruction: instruction,
		Tools: []tool.Tool{
			clickTool,
			typeTextTool,
			keyPressTool,
			scrollTool,
			waitTool,
			dragTool,
		},
		OutputKey: "action_result",
	})
}
