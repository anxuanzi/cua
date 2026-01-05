// Package input provides cross-platform mouse, keyboard, and scroll operations
// using robotgo for desktop automation.
package input

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
)

// Point represents a screen coordinate.
type Point struct {
	X, Y int
}

// MouseButton represents a mouse button.
type MouseButton string

const (
	ButtonLeft   MouseButton = "left"
	ButtonRight  MouseButton = "right"
	ButtonMiddle MouseButton = "center"
)

// Modifier represents a keyboard modifier key.
type Modifier string

const (
	ModCmd   Modifier = "cmd"
	ModCtrl  Modifier = "ctrl"
	ModAlt   Modifier = "alt"
	ModShift Modifier = "shift"
)

// Click performs a mouse click at the specified coordinates.
func Click(p Point) error {
	robotgo.Move(p.X, p.Y)
	time.Sleep(10 * time.Millisecond) // Small delay for reliability
	robotgo.Click("left", false)
	return nil
}

// DoubleClick performs a double-click at the specified coordinates.
func DoubleClick(p Point) error {
	robotgo.Move(p.X, p.Y)
	time.Sleep(10 * time.Millisecond)
	robotgo.Click("left", true) // true = double-click
	return nil
}

// RightClick performs a right-click at the specified coordinates.
func RightClick(p Point) error {
	robotgo.Move(p.X, p.Y)
	time.Sleep(10 * time.Millisecond)
	robotgo.Click("right", false)
	return nil
}

// ClickButton performs a click with the specified button at the coordinates.
func ClickButton(p Point, button MouseButton, doubleClick bool) error {
	robotgo.Move(p.X, p.Y)
	time.Sleep(10 * time.Millisecond)
	robotgo.Click(string(button), doubleClick)
	return nil
}

// MoveTo moves the mouse cursor to the specified coordinates.
func MoveTo(p Point) error {
	robotgo.Move(p.X, p.Y)
	return nil
}

// MoveSmooth moves the mouse cursor smoothly to the specified coordinates.
func MoveSmooth(p Point) error {
	robotgo.MoveSmooth(p.X, p.Y)
	return nil
}

// Drag performs a mouse drag from start to end coordinates.
func Drag(start, end Point) error {
	robotgo.Move(start.X, start.Y)
	time.Sleep(10 * time.Millisecond)
	robotgo.DragSmooth(end.X, end.Y)
	return nil
}

// MouseLocation returns the current mouse cursor position.
func MouseLocation() Point {
	x, y := robotgo.Location()
	return Point{X: x, Y: y}
}

// Scroll performs a scroll operation at the current mouse position.
// Positive deltaY scrolls down, negative scrolls up.
// Positive deltaX scrolls right, negative scrolls left.
func Scroll(deltaX, deltaY int) error {
	robotgo.Scroll(deltaX, deltaY)
	return nil
}

// ScrollAt performs a scroll operation at the specified coordinates.
func ScrollAt(p Point, deltaX, deltaY int) error {
	robotgo.Move(p.X, p.Y)
	time.Sleep(10 * time.Millisecond)
	robotgo.Scroll(deltaX, deltaY)
	return nil
}

// ScrollSmooth performs a smooth scroll operation.
func ScrollSmooth(deltaY int) error {
	robotgo.ScrollSmooth(deltaY, 6)
	return nil
}

// TypeText types the given text string.
// Supports Unicode characters.
func TypeText(text string) error {
	if text == "" {
		return nil
	}
	robotgo.TypeStr(text)
	return nil
}

// TypeTextWithDelay types text with a delay between characters (in milliseconds).
func TypeTextWithDelay(text string, delayMs int) error {
	if text == "" {
		return nil
	}
	// robotgo.TypeStr accepts delay as second parameter (in milliseconds)
	robotgo.TypeStr(text, delayMs)
	return nil
}

// KeyTap presses and releases a key, optionally with modifiers.
func KeyTap(key string, modifiers ...Modifier) error {
	key = normalizeKeyName(key)
	if len(modifiers) == 0 {
		robotgo.KeyTap(key)
	} else {
		mods := make([]interface{}, len(modifiers))
		for i, m := range modifiers {
			mods[i] = string(m)
		}
		robotgo.KeyTap(key, mods...)
	}
	return nil
}

// KeyDown presses a key down (without releasing).
func KeyDown(key string) error {
	key = normalizeKeyName(key)
	robotgo.KeyToggle(key, "down")
	return nil
}

// KeyUp releases a key.
func KeyUp(key string) error {
	key = normalizeKeyName(key)
	robotgo.KeyToggle(key, "up")
	return nil
}

// KeyTapWithModifiers presses a key with multiple modifiers.
func KeyTapWithModifiers(key string, modifiers []string) error {
	key = normalizeKeyName(key)
	if len(modifiers) == 0 {
		robotgo.KeyTap(key)
	} else {
		// Convert to interface slice for robotgo
		mods := make([]interface{}, len(modifiers))
		for i, m := range modifiers {
			mods[i] = normalizeModifier(m)
		}
		robotgo.KeyTap(key, mods...)
	}
	return nil
}

// normalizeKeyName converts key names to robotgo format.
func normalizeKeyName(key string) string {
	key = strings.ToLower(key)

	// Map common key names to robotgo names
	switch key {
	case "enter", "return":
		return "enter"
	case "backspace":
		return "backspace"
	case "delete":
		return "delete"
	case "escape", "esc":
		return "escape"
	case "tab":
		return "tab"
	case "space":
		return "space"
	case "up":
		return "up"
	case "down":
		return "down"
	case "left":
		return "left"
	case "right":
		return "right"
	case "home":
		return "home"
	case "end":
		return "end"
	case "pageup":
		return "pageup"
	case "pagedown":
		return "pagedown"
	case "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12":
		return key
	default:
		return key
	}
}

// normalizeModifier converts modifier names to robotgo format.
func normalizeModifier(mod string) string {
	mod = strings.ToLower(mod)
	switch mod {
	case "cmd", "command", "meta", "win", "windows":
		return "cmd"
	case "ctrl", "control":
		return "ctrl"
	case "alt", "option":
		return "alt"
	case "shift":
		return "shift"
	case "fn":
		return "fn"
	default:
		return mod
	}
}

// WriteToClipboard writes text to the system clipboard.
func WriteToClipboard(text string) error {
	return robotgo.WriteAll(text)
}

// ReadFromClipboard reads text from the system clipboard.
func ReadFromClipboard() (string, error) {
	return robotgo.ReadAll()
}

// PasteFromClipboard writes text to clipboard and pastes it.
// This is useful for pasting large amounts of text quickly.
func PasteFromClipboard(text string) error {
	if err := WriteToClipboard(text); err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}
	// Cmd+V on macOS, Ctrl+V on Windows/Linux
	robotgo.KeyTap("v", "cmd")
	return nil
}

// Sleep pauses execution for the specified duration.
func Sleep(d time.Duration) {
	time.Sleep(d)
}

// MilliSleep pauses execution for the specified milliseconds.
func MilliSleep(ms int) {
	robotgo.MilliSleep(ms)
}
