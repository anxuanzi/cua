package tools

import (
	"context"
	"time"

	"github.com/go-vgo/robotgo"
)

// TypeTool types text at the current cursor position.
type TypeTool struct {
	BaseTool
}

// NewTypeTool creates a new type tool.
func NewTypeTool() *TypeTool {
	return &TypeTool{}
}

func (t *TypeTool) Name() string {
	return "keyboard_type"
}

func (t *TypeTool) Description() string {
	return `Type text at the current cursor position. The text is typed character by character to simulate natural typing. Use this to fill in forms, enter commands, or input any text. Make sure the target input field is focused before typing.`
}

func (t *TypeTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"text": {
			Type:        "string",
			Description: "The text to type",
			Required:    true,
		},
		"delay_ms": {
			Type:        "integer",
			Description: "Delay between characters in milliseconds (0 for fastest)",
			Required:    false,
			Default:     0,
		},
	}
}

func (t *TypeTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Text    string `json:"text"`
		DelayMs int    `json:"delay_ms"`
	}

	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide the text to type"), nil
	}

	if args.Text == "" {
		return ErrorResponse("text cannot be empty", "Provide the text to type"), nil
	}

	// Small delay before typing to ensure focus
	time.Sleep(50 * time.Millisecond)

	// Type the text
	if args.DelayMs > 0 {
		// Type with delay between characters
		for _, char := range args.Text {
			robotgo.TypeStr(string(char))
			time.Sleep(time.Duration(args.DelayMs) * time.Millisecond)
		}
	} else {
		// Fast typing
		robotgo.TypeStr(args.Text)
	}

	return SuccessResponse(map[string]interface{}{
		"typed_text": args.Text,
		"char_count": len(args.Text),
		"delay_ms":   args.DelayMs,
	}), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *TypeTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
