package tools

import (
	"fmt"
	"strings"

	"github.com/anxuanzi/cua/pkg/logging"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

var keyPressLog = logging.NewToolLogger("key_press")

// KeyPressArgs defines the arguments for the key_press tool.
type KeyPressArgs struct {
	// Key is the key to press (e.g., "enter", "tab", "escape", "a", "f1").
	Key string `json:"key" jsonschema:"The key to press (e.g., 'enter', 'tab', 'escape', 'backspace', 'up', 'down', 'left', 'right', 'space', 'f1'-'f12')"`

	// Modifiers are optional modifier keys to hold while pressing the key.
	// Supported: "cmd", "ctrl", "alt", "shift", "fn"
	Modifiers []string `json:"modifiers,omitempty" jsonschema:"Optional modifier keys: 'cmd', 'ctrl', 'alt', 'shift', 'fn'"`
}

// KeyPressResult contains the result of a key press operation.
type KeyPressResult struct {
	// Success indicates if the key press succeeded.
	Success bool `json:"success"`

	// Key is the key that was pressed.
	Key string `json:"key"`

	// Modifiers are the modifier keys that were held.
	Modifiers []string `json:"modifiers,omitempty"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// pressKey handles the key_press tool invocation.
func pressKey(ctx tool.Context, args KeyPressArgs) (KeyPressResult, error) {
	// Log at INFO level so we always see tool invocations
	keyPressLog.Info(">>> KEY_PRESS TOOL CALLED: key=%s modifiers=%v", args.Key, args.Modifiers)

	if args.Key == "" {
		keyPressLog.Failure("key_press", fmt.Errorf("key cannot be empty"))
		return KeyPressResult{
			Success: false,
			Error:   "key cannot be empty",
		}, nil
	}

	// Format the key combination for logging
	keyCombo := args.Key
	if len(args.Modifiers) > 0 {
		keyCombo = strings.Join(args.Modifiers, "+") + "+" + args.Key
	}
	keyPressLog.Start("key_press", keyCombo)

	err := keyPressNative(args.Key, args.Modifiers)
	if err != nil {
		keyPressLog.Failure("key_press", err)
		return KeyPressResult{
			Success:   false,
			Key:       args.Key,
			Modifiers: args.Modifiers,
			Error:     fmt.Sprintf("failed to press key: %v", err),
		}, nil
	}

	keyPressLog.Success("key_press", keyCombo)
	return KeyPressResult{
		Success:   true,
		Key:       args.Key,
		Modifiers: args.Modifiers,
	}, nil
}

// NewKeyPressTool creates the key_press tool for ADK agents.
func NewKeyPressTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "key_press",
			Description: "Presses a keyboard key with optional modifier keys. Supports special keys like enter, tab, escape, arrow keys, and function keys.",
		},
		pressKey,
	)
}
