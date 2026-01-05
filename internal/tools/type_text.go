package tools

import (
	"fmt"

	"github.com/anxuanzi/cua/pkg/logging"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

var typeTextLog = logging.NewToolLogger("type_text")

// TypeTextArgs defines the arguments for the type_text tool.
type TypeTextArgs struct {
	// Text is the text to type.
	Text string `json:"text" jsonschema:"The text to type using keyboard input"`
}

// TypeTextResult contains the result of a type operation.
type TypeTextResult struct {
	// Success indicates if typing succeeded.
	Success bool `json:"success"`

	// Text is the text that was typed.
	Text string `json:"text"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// typeText handles the type_text tool invocation.
func typeText(ctx tool.Context, args TypeTextArgs) (TypeTextResult, error) {
	if args.Text == "" {
		typeTextLog.Failure("type_text", fmt.Errorf("text cannot be empty"))
		return TypeTextResult{
			Success: false,
			Error:   "text cannot be empty",
		}, nil
	}

	// Truncate text for logging if too long
	logText := args.Text
	if len(logText) > 50 {
		logText = logText[:47] + "..."
	}
	typeTextLog.Start("type_text", logText)

	// Use robotgo keyboard input
	err := typeTextNative(args.Text)
	if err != nil {
		typeTextLog.Failure("type_text", err)
		return TypeTextResult{
			Success: false,
			Text:    args.Text,
			Error:   fmt.Sprintf("failed to type text: %v", err),
		}, nil
	}

	typeTextLog.Success("type_text", logText)
	return TypeTextResult{
		Success: true,
		Text:    args.Text,
	}, nil
}

// NewTypeTextTool creates the type_text tool for ADK agents.
func NewTypeTextTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "type_text",
			Description: "Types the specified text using keyboard input. Simulates pressing each character key.",
		},
		typeText,
	)
}
