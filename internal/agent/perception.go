package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"

	"github.com/anxuanzi/cua/internal/tools"
)

// PerceptionInstruction defines the system prompt for the Perception Agent.
const PerceptionInstruction = `You are a screen analysis specialist for a desktop automation agent.

Your job: Analyze screenshots and UI elements to help the Coordinator understand the current screen state.

## Capabilities
- Capture screenshots of the current display
- Find UI elements using the accessibility tree
- Identify interactive elements (buttons, text fields, links, etc.)
- Locate specific elements by role, name, or title

## When analyzing:
1. Take a screenshot first to see the current state
2. Use find_element to locate specific UI elements when needed
3. Report coordinates for clicking if relevant elements are found

## Response Format
Always respond with a structured analysis:

SCREEN_STATE:
- Current app/window: [Name of the focused application or window]
- Visible elements: [Key interactive elements with their roles and positions]
- Focused element: [What element currently has focus, if any]
- Notable text: [Important text visible on screen]
- Suggested targets: [Elements likely relevant for the current task]

COORDINATES (if applicable):
- [Element name]: (x, y) center point

OBSERVATIONS:
- [Any relevant observations about the screen state]
- [Potential issues or blockers visible]

Be concise but thorough. Focus on actionable information that helps decide what to do next.`

// NewPerceptionAgent creates the Perception Agent for screen analysis.
// It uses Gemini Flash for fast, specialized screen understanding.
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

	return llmagent.New(llmagent.Config{
		Name:        "perception_agent",
		Model:       m,
		Description: "Analyzes screenshots and UI elements to understand the current screen state.",
		Instruction: PerceptionInstruction,
		Tools: []tool.Tool{
			screenshotTool,
			findElementTool,
		},
	})
}
