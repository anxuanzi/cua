package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"

	"github.com/anxuanzi/cua/internal/tools"
)

// NewPerceptionAgent creates the Perception Agent for screen analysis.
// It uses Gemini Flash for fast, specialized screen understanding.
// The instruction is built dynamically to include platform-specific context.
func NewPerceptionAgent(m model.LLM) (agent.Agent, error) {
	// Create the tools for perception
	screenshotTool, err := tools.NewScreenshotTool()
	if err != nil {
		return nil, err
	}

	findElementTool, err := tools.NewFindElementTool()
	if err != nil {
		return nil, err
	}

	// Build instruction with platform context
	instruction := BuildPerceptionInstruction()

	return llmagent.New(llmagent.Config{
		Name:        "perception_agent",
		Model:       m,
		Description: "Analyzes screenshots and UI elements to understand the current screen state.",
		Instruction: instruction,
		Tools: []tool.Tool{
			screenshotTool,
			findElementTool,
		},
	})
}
