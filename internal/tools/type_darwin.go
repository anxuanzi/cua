//go:build darwin

package tools

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// typeText types text on macOS using AppleScript for reliability with secure input fields.
// AppleScript's "keystroke" command works with Spotlight, password fields, and other secure inputs
// where robotgo's TypeStr() fails.
func typeText(_ context.Context, text string, delayMs int) (string, error) {
	// Delay before typing to ensure UI is ready
	time.Sleep(150 * time.Millisecond)

	// Escape special characters for AppleScript
	// AppleScript needs backslashes and quotes escaped
	escaped := strings.ReplaceAll(text, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")

	// Use AppleScript to type the text
	// "keystroke" works with secure input fields where robotgo fails
	script := `tell application "System Events" to keystroke "` + escaped + `"`

	cmd := exec.Command("osascript", "-e", script)
	err := cmd.Run()
	if err != nil {
		return ErrorResponse(
			"failed to type text: "+err.Error(),
			"Make sure the application is focused and accepts keyboard input",
		), nil
	}

	// Small delay after typing for reliability
	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	return SuccessResponse(map[string]interface{}{
		"typed_text": text,
		"char_count": len(text),
		"delay_ms":   delayMs,
		"method":     "applescript",
	}), nil
}
