// Package cua provides a Computer Use Agent for desktop automation.
//
// CUA is a library-first Go package that enables AI-powered desktop automation.
// It uses vision (screenshots) and native accessibility APIs to perform any
// task a human can do on macOS or Windows.
//
// # Quick Start
//
// The simplest way to use CUA:
//
//	agent := cua.New(cua.WithAPIKey("your-api-key"))
//	result, err := agent.Do("Open Safari and search for golang")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Summary)
//
// # Configuration Options
//
// CUA uses the functional options pattern for configuration:
//
//	agent := cua.New(
//	    cua.WithAPIKey("your-api-key"),
//	    cua.WithModel(cua.Gemini3Pro),       // Use advanced model
//	    cua.WithSafetyLevel(cua.SafetyStrict), // Maximum safety
//	    cua.WithTimeout(5 * time.Minute),      // Longer timeout
//	    cua.WithMaxActions(100),               // More actions allowed
//	    cua.WithVerbose(true),                 // Enable logging
//	)
//
// # Progress Monitoring
//
// For tasks where you want to see progress:
//
//	err := agent.DoWithProgress("Fill out the form", func(step cua.Step) {
//	    fmt.Printf("Step %d: %s\n", step.Number, step.Description)
//	})
//
// # Low-Level Access
//
// CUA also provides direct access to input and element operations:
//
//	// Click at coordinates
//	cua.Click(100, 200)
//
//	// Type text
//	cua.TypeText("Hello, World!")
//
//	// Find and interact with elements
//	elements, _ := cua.FindElements(cua.ByRole(cua.RoleButton))
//	elements[0].Click()
//
//	// Keyboard shortcuts
//	cua.KeyPress(cua.KeyEnter)
//	cua.KeyPress(cua.KeyA, cua.ModCmd) // Cmd+A
//
// # Architecture
//
// CUA uses a multi-agent architecture internally:
//   - Coordinator (Gemini Pro): ReAct reasoning loop, task decomposition
//   - Perception Agent (Gemini Flash): Screen analysis, element identification
//   - Action Agent (Gemini Flash): Tool execution, verification
//
// All internal complexity is hidden from the library user.
//
// # Platform Support
//
// CUA supports:
//   - macOS: Requires Accessibility and Screen Recording permissions
//   - Windows 11: May require Administrator for some applications
package cua

import (
	"image"
	"sync"
	"time"

	"github.com/anxuanzi/cua/pkg/element"
	"github.com/anxuanzi/cua/pkg/input"
	"github.com/anxuanzi/cua/pkg/screen"
)

// Global finder instance (lazily initialized)
var (
	globalFinder     *element.Finder
	globalFinderOnce sync.Once
	globalFinderErr  error
)

// getFinder returns the global element finder, initializing it if needed.
func getFinder() (*element.Finder, error) {
	globalFinderOnce.Do(func() {
		globalFinder, globalFinderErr = element.NewFinder()
	})
	return globalFinder, globalFinderErr
}

// New creates a new CUA Agent with the given options.
//
// At minimum, you should provide an API key:
//
//	agent := cua.New(cua.WithAPIKey("your-key"))
//
// If no API key is provided, CUA will look for the GOOGLE_API_KEY
// environment variable.
//
// See the With* functions for available options.
func New(opts ...Option) *Agent {
	agent, err := newAgent(opts...)
	if err != nil {
		// Return an agent that will return the error on first use
		return &Agent{
			config: defaultConfig(),
		}
	}
	return agent
}

// MustNew creates a new CUA Agent or panics if configuration is invalid.
// Use this in init() or when you know the configuration is valid.
func MustNew(opts ...Option) *Agent {
	agent, err := newAgent(opts...)
	if err != nil {
		panic(err)
	}
	return agent
}

// --- Low-Level Input Functions ---
// These provide direct access to mouse and keyboard control.

// Click performs a left click at the given screen coordinates.
func Click(x, y int) error {
	return input.Click(input.Point{X: x, Y: y})
}

// DoubleClick performs a double left click at the given screen coordinates.
func DoubleClick(x, y int) error {
	return input.DoubleClick(input.Point{X: x, Y: y})
}

// RightClick performs a right click at the given screen coordinates.
func RightClick(x, y int) error {
	return input.RightClick(input.Point{X: x, Y: y})
}

// MiddleClick performs a middle click at the given screen coordinates.
func MiddleClick(x, y int) error {
	return input.ClickButton(input.Point{X: x, Y: y}, input.ButtonMiddle, false)
}

// MoveMouse moves the mouse cursor to the given screen coordinates.
func MoveMouse(x, y int) error {
	return input.MoveTo(input.Point{X: x, Y: y})
}

// DragMouse drags from (x1, y1) to (x2, y2) with the left button held.
func DragMouse(x1, y1, x2, y2 int) error {
	return input.Drag(input.Point{X: x1, Y: y1}, input.Point{X: x2, Y: y2})
}

// Scroll scrolls the mouse wheel at the given coordinates.
// Positive dy scrolls down, negative scrolls up.
// Positive dx scrolls right, negative scrolls left.
func Scroll(x, y int, dx, dy int) error {
	return input.ScrollAt(input.Point{X: x, Y: y}, dx, dy)
}

// TypeText types the given text string using keyboard input.
// This simulates individual key presses for each character.
func TypeText(text string) error {
	return input.TypeText(text)
}

// KeyPress presses a key with optional modifiers.
//
// Examples:
//
//	cua.KeyPress(cua.KeyEnter)              // Press Enter
//	cua.KeyPress(cua.KeyA, cua.ModCmd)      // Cmd+A
//	cua.KeyPress(cua.KeyS, cua.ModCmd, cua.ModShift) // Cmd+Shift+S
func KeyPress(key Key, mods ...Modifier) error {
	inputMods := make([]input.Modifier, len(mods))
	for i, m := range mods {
		inputMods[i] = input.Modifier(m)
	}
	return input.KeyTap(string(key), inputMods...)
}

// HoldKey presses and holds a key. Must be followed by ReleaseKey.
func HoldKey(key Key) error {
	return input.KeyDown(string(key))
}

// ReleaseKey releases a key that was pressed with HoldKey.
func ReleaseKey(key Key) error {
	return input.KeyUp(string(key))
}

// --- Element Functions ---
// These provide access to UI elements via the accessibility tree.

// FindElements finds all elements matching the given selector.
//
// Examples:
//
//	// Find all buttons
//	buttons, _ := cua.FindElements(cua.ByRole(cua.RoleButton))
//
//	// Find elements by name
//	elements, _ := cua.FindElements(cua.ByName("Submit"))
//
//	// Combine selectors
//	elements, _ := cua.FindElements(cua.And(
//	    cua.ByRole(cua.RoleButton),
//	    cua.ByNameContains("Save"),
//	))
func FindElements(selector Selector) ([]*Element, error) {
	finder, err := getFinder()
	if err != nil {
		return nil, err
	}
	return finder.FindAll(selector)
}

// FindElement finds the first element matching the given selector.
// Returns ErrElementNotFound if no element matches.
func FindElement(selector Selector) (*Element, error) {
	finder, err := getFinder()
	if err != nil {
		return nil, err
	}
	return finder.Find(selector)
}

// FocusedElement returns the currently focused element.
func FocusedElement() (*Element, error) {
	finder, err := getFinder()
	if err != nil {
		return nil, err
	}
	return finder.FocusedElement()
}

// FocusedApplication returns the frontmost application element.
func FocusedApplication() (*Element, error) {
	finder, err := getFinder()
	if err != nil {
		return nil, err
	}
	return finder.FocusedApplication()
}

// WaitForElement waits for an element matching the selector to appear.
// Returns when the element is found or when the timeout is reached.
func WaitForElement(selector Selector, timeout time.Duration) (*Element, error) {
	finder, err := getFinder()
	if err != nil {
		return nil, err
	}
	return finder.WaitFor(selector, timeout)
}

// --- Screen Functions ---
// These provide access to screen capture.

// CaptureScreen captures the entire primary screen.
func CaptureScreen() (image.Image, error) {
	return screen.CapturePrimary()
}

// CaptureRect captures a rectangular region of the screen.
func CaptureRect(rect Rect) (image.Image, error) {
	return screen.CaptureRect(screen.Rect{
		X:      rect.X,
		Y:      rect.Y,
		Width:  rect.Width,
		Height: rect.Height,
	})
}

// ScreenSize returns the size of the primary screen.
func ScreenSize() (width, height int, err error) {
	display, err := screen.PrimaryDisplay()
	if err != nil {
		return 0, 0, err
	}
	return display.Bounds.Width, display.Bounds.Height, nil
}

// --- Utility Functions ---

// Version returns the CUA library version.
func Version() string {
	return "0.1.0-dev"
}
