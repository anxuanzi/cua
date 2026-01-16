package tools

import (
	"context"
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

	// Platform-specific typing implementation
	return typeText(ctx, args.Text, charDelay)
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *TypeTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
