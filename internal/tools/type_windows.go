//go:build windows

package tools

import (
	"context"
	"time"

	"github.com/go-vgo/robotgo"
)

// typeText types text on Windows using robotgo.
func typeText(_ context.Context, text string, delayMs int) (string, error) {
	// Delay before typing to ensure UI is ready
	time.Sleep(150 * time.Millisecond)

	// Type character by character with delay for reliability
	for _, char := range text {
		robotgo.TypeStr(string(char))
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	return SuccessResponse(map[string]interface{}{
		"typed_text": text,
		"char_count": len(text),
		"delay_ms":   delayMs,
		"method":     "robotgo",
	}), nil
}
