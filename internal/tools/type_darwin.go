//go:build darwin

package tools

import (
	"context"
	"math/rand"
	"os/exec"
	"time"
)

// typeText types text on macOS using AppleScript for reliability with secure input fields.
// Types CHARACTER BY CHARACTER with human-like delays to appear natural and work reliably.
// AppleScript's "keystroke" command works with Spotlight, password fields, and other secure inputs
// where robotgo's TypeStr() fails.
func typeText(_ context.Context, text string, delayMs int) (string, error) {
	// Delay before typing to ensure UI is ready
	time.Sleep(200 * time.Millisecond)

	// Type each character individually with human-like delays
	for _, char := range text {
		// Escape special characters for AppleScript
		charStr := string(char)
		escaped := charStr
		if charStr == "\\" {
			escaped = "\\\\"
		} else if charStr == "\"" {
			escaped = "\\\""
		}

		// Use AppleScript to type single character
		script := `tell application "System Events" to keystroke "` + escaped + `"`

		cmd := exec.Command("osascript", "-e", script)
		err := cmd.Run()
		if err != nil {
			return ErrorResponse(
				"failed to type character '"+charStr+"': "+err.Error(),
				"Make sure the application is focused and accepts keyboard input",
			), nil
		}

		// Human-like delay between characters
		// Add some randomness to make it more natural (±30% variation)
		baseDelay := delayMs
		if baseDelay < 30 {
			baseDelay = 30 // Minimum 30ms for reliability
		}
		variation := rand.Intn(baseDelay*60/100) - baseDelay*30/100 // ±30%
		actualDelay := baseDelay + variation
		if actualDelay < 20 {
			actualDelay = 20
		}
		time.Sleep(time.Duration(actualDelay) * time.Millisecond)
	}

	// Small delay after typing to let UI catch up
	time.Sleep(100 * time.Millisecond)

	return SuccessResponse(map[string]interface{}{
		"typed_text":   text,
		"char_count":   len(text),
		"delay_ms":     delayMs,
		"method":       "applescript",
		"human_typing": true,
	}), nil
}
