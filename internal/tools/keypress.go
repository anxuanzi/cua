package tools

import (
	"context"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
)

// KeyPressTool presses keyboard keys or key combinations.
type KeyPressTool struct {
	BaseTool
}

// NewKeyPressTool creates a new keypress tool.
func NewKeyPressTool() *KeyPressTool {
	return &KeyPressTool{}
}

func (t *KeyPressTool) Name() string {
	return "keyboard_press"
}

func (t *KeyPressTool) Description() string {
	return `Press a key or key combination. Use this for keyboard shortcuts, special keys, or navigation.

Common keys: enter, tab, escape, backspace, delete, space, up, down, left, right, home, end, pageup, pagedown
Modifier keys: cmd (or command), ctrl (or control), alt (or option), shift
Function keys: f1-f12

For combinations, separate keys with '+'. Examples:
- "enter" - Press Enter
- "cmd+c" - Copy (macOS)
- "ctrl+c" - Copy (Windows/Linux)
- "cmd+shift+s" - Save As (macOS)
- "alt+tab" - Switch windows`
}

func (t *KeyPressTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"key": {
			Type:        "string",
			Description: "Key or key combination to press (e.g., 'enter', 'cmd+c', 'ctrl+shift+s')",
			Required:    true,
		},
		"hold_ms": {
			Type:        "integer",
			Description: "How long to hold the key in milliseconds (default: 0)",
			Required:    false,
			Default:     0,
		},
	}
}

func (t *KeyPressTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Key    string `json:"key"`
		HoldMs int    `json:"hold_ms"`
	}

	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide the key to press"), nil
	}

	if args.Key == "" {
		return ErrorResponse("key cannot be empty", "Provide the key to press"), nil
	}

	// Parse key combination
	parts := strings.Split(strings.ToLower(args.Key), "+")
	key := normalizeKeyName(parts[len(parts)-1])
	modifiers := make([]string, 0)

	for i := 0; i < len(parts)-1; i++ {
		mod := normalizeModifier(parts[i])
		if mod != "" {
			modifiers = append(modifiers, mod)
		}
	}

	// Small delay before key press
	time.Sleep(100 * time.Millisecond)

	// Press the key
	if args.HoldMs > 0 {
		// Hold the key - press modifiers first, then main key
		for _, mod := range modifiers {
			robotgo.KeyToggle(mod, "down")
		}
		robotgo.KeyToggle(key, "down")
		time.Sleep(time.Duration(args.HoldMs) * time.Millisecond)
		robotgo.KeyToggle(key, "up")
		// Release modifiers in reverse order
		for i := len(modifiers) - 1; i >= 0; i-- {
			robotgo.KeyToggle(modifiers[i], "up")
		}
	} else {
		// Quick tap with modifiers
		if len(modifiers) > 0 {
			robotgo.KeyTap(key, modifiers)
		} else {
			robotgo.KeyTap(key)
		}
	}

	// Extra delay after modifier combos (Spotlight, app launchers need time)
	if len(modifiers) > 0 {
		time.Sleep(300 * time.Millisecond)
	}

	return SuccessResponse(map[string]interface{}{
		"pressed_key": args.Key,
		"key":         key,
		"modifiers":   modifiers,
		"hold_ms":     args.HoldMs,
	}), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *KeyPressTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}

// normalizeKeyName converts common key name variations to robotgo format.
func normalizeKeyName(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))

	// Map common aliases
	aliases := map[string]string{
		"return":     "enter",
		"esc":        "escape",
		"del":        "delete",
		"bs":         "backspace",
		"pgup":       "pageup",
		"pgdn":       "pagedown",
		"pgdown":     "pagedown",
		"arrowup":    "up",
		"arrowdown":  "down",
		"arrowleft":  "left",
		"arrowright": "right",
	}

	if mapped, ok := aliases[key]; ok {
		return mapped
	}
	return key
}

// normalizeModifier converts modifier key names to robotgo format.
func normalizeModifier(mod string) string {
	mod = strings.TrimSpace(strings.ToLower(mod))

	modMap := map[string]string{
		"cmd":     "cmd",
		"command": "cmd",
		"ctrl":    "ctrl",
		"control": "ctrl",
		"alt":     "alt",
		"option":  "alt",
		"shift":   "shift",
		"meta":    "cmd", // For cross-platform compatibility
		"win":     "cmd", // Windows key maps to cmd
		"windows": "cmd",
	}

	if mapped, ok := modMap[mod]; ok {
		return mapped
	}
	return ""
}
