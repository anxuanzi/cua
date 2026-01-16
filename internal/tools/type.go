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
			Description: "Delay between characters in milliseconds (default: 10 for reliability)",
			Required:    false,
			Default:     10,
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

	// Default delay if not specified (10ms for reliability)
	charDelay := args.DelayMs
	if charDelay == 0 {
		charDelay = 10
	}

	// Delay before typing to ensure UI is ready (Spotlight, dialogs need time)
	time.Sleep(150 * time.Millisecond)

	// Type character by character with delay for reliability
	// This works better with macOS secure text fields like Spotlight
	for _, char := range args.Text {
		robotgo.TypeStr(string(char))
		time.Sleep(time.Duration(charDelay) * time.Millisecond)
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
