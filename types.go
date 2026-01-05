// Package cua provides a Computer Use Agent for desktop automation.
//
// CUA is a library-first desktop automation agent that uses vision and
// accessibility APIs to perform any task a human can do on macOS or Windows.
//
// Basic usage:
//
//	agent := cua.New(cua.WithAPIKey("your-api-key"))
//	result, err := agent.Do("Open Safari and search for golang")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Summary)
package cua

import (
	"image"
	"time"

	"github.com/anxuanzi/cua/pkg/element"
)

// Model represents the AI model to use for the agent.
type Model string

const (
	// Gemini2Flash is the fast, cost-effective model for most tasks.
	// Recommended for perception and action agents.
	Gemini2Flash Model = "gemini-2.5-flash"

	// Gemini2Pro is the advanced model for complex reasoning tasks.
	// Recommended for coordinator agent.
	Gemini2Pro Model = "gemini-2.5-pro"

	// Gemini3Flash is an alias for backward compatibility.
	Gemini3Flash Model = "gemini-3-pro-preview"

	// Gemini3Pro is an alias for backward compatibility.
	Gemini3Pro Model = "gemini-3-flash-preview"
)

// SafetyLevel controls how cautious the agent is.
type SafetyLevel int

const (
	// SafetyNormal applies standard safety checks.
	// The agent will pause for confirmation on sensitive actions.
	SafetyNormal SafetyLevel = iota

	// SafetyStrict applies maximum safety checks.
	// The agent will pause more frequently and avoid risky actions.
	SafetyStrict

	// SafetyMinimal applies minimal safety checks.
	// Only use in controlled environments. Not recommended for production.
	SafetyMinimal
)

// Result contains the outcome of a completed task.
type Result struct {
	// Success indicates whether the task completed successfully.
	Success bool

	// Summary is a human-readable description of what was accomplished.
	Summary string

	// Steps contains the individual actions taken during the task.
	Steps []Step

	// Duration is how long the task took to complete.
	Duration time.Duration

	// FinalScreenshot is the screen state when the task completed.
	// May be nil if screenshot capture failed.
	FinalScreenshot image.Image

	// Error contains any error that occurred. Nil on success.
	Error error
}

// Step represents a single action taken by the agent.
type Step struct {
	// Number is the 1-based step number.
	Number int

	// Action is the type of action taken (e.g., "click", "type", "scroll").
	Action string

	// Description is a human-readable description of the action.
	Description string

	// Target describes what element or location was acted upon.
	Target string

	// Success indicates whether this step succeeded.
	Success bool

	// Duration is how long this step took.
	Duration time.Duration

	// Screenshot is the screen state after this step.
	// May be nil if not captured or capture failed.
	Screenshot image.Image

	// Error contains any error from this step. Nil on success.
	Error error
}

// ProgressFunc is called after each step during DoWithProgress.
type ProgressFunc func(step Step)

// Re-export element package types for convenience.
// Users can use cua.Element, cua.Role, etc. directly.
type (
	// Element represents a UI element on screen.
	Element = element.Element

	// Role represents the semantic type of a UI element.
	Role = element.Role

	// Rect represents a rectangle on screen.
	Rect = element.Rect

	// Point represents a point on screen.
	Point = element.Point

	// Selector is used to find elements by various criteria.
	Selector = element.Selector
)

// Re-export element Role constants.
const (
	RoleWindow      = element.RoleWindow
	RoleButton      = element.RoleButton
	RoleTextField   = element.RoleTextField
	RoleTextArea    = element.RoleTextArea
	RoleStaticText  = element.RoleStaticText
	RoleCheckbox    = element.RoleCheckbox
	RoleRadioButton = element.RoleRadioButton
	RoleList        = element.RoleList
	RoleListItem    = element.RoleListItem
	RoleMenu        = element.RoleMenu
	RoleMenuItem    = element.RoleMenuItem
	RoleMenuBar     = element.RoleMenuBar
	RoleToolbar     = element.RoleToolbar
	RoleScrollArea  = element.RoleScrollArea
	RoleImage       = element.RoleImage
	RoleLink        = element.RoleLink
	RoleGroup       = element.RoleGroup
	RoleTab         = element.RoleTab
	RoleTabGroup    = element.RoleTabGroup
	RoleTable       = element.RoleTable
	RoleRow         = element.RoleRow
	RoleCell        = element.RoleCell
	RoleSlider      = element.RoleSlider
	RoleComboBox    = element.RoleComboBox
	RoleUnknown     = element.RoleUnknown
)

// Re-export element selector constructors.
var (
	// ByRole creates a selector that matches elements by their role.
	ByRole = element.ByRole

	// ByName creates a selector that matches elements by their exact name.
	ByName = element.ByName

	// ByNameContains creates a selector that matches elements whose name contains the substring.
	ByNameContains = element.ByNameContains

	// ByTitle creates a selector that matches elements by their exact title.
	ByTitle = element.ByTitle

	// ByTitleContains creates a selector that matches elements whose title contains the substring.
	ByTitleContains = element.ByTitleContains

	// ByValue creates a selector that matches elements by their exact value.
	ByValue = element.ByValue

	// ByEnabled creates a selector that matches only enabled elements.
	ByEnabled = element.ByEnabled

	// ByFocused creates a selector that matches only focused elements.
	ByFocused = element.ByFocused

	// And creates a selector that matches elements matching ALL provided selectors.
	And = element.And

	// Or creates a selector that matches elements matching ANY provided selector.
	Or = element.Or

	// Not creates a selector that matches elements that do NOT match the provided selector.
	Not = element.Not

	// ByPredicate creates a selector using a custom predicate function.
	ByPredicate = element.ByPredicate
)

// Key represents a keyboard key.
type Key string

const (
	KeyEnter     Key = "enter"
	KeyTab       Key = "tab"
	KeyEscape    Key = "escape"
	KeyBackspace Key = "backspace"
	KeyDelete    Key = "delete"
	KeyUp        Key = "up"
	KeyDown      Key = "down"
	KeyLeft      Key = "left"
	KeyRight     Key = "right"
	KeyHome      Key = "home"
	KeyEnd       Key = "end"
	KeyPageUp    Key = "pageup"
	KeyPageDown  Key = "pagedown"
	KeySpace     Key = "space"
	KeyF1        Key = "f1"
	KeyF2        Key = "f2"
	KeyF3        Key = "f3"
	KeyF4        Key = "f4"
	KeyF5        Key = "f5"
	KeyF6        Key = "f6"
	KeyF7        Key = "f7"
	KeyF8        Key = "f8"
	KeyF9        Key = "f9"
	KeyF10       Key = "f10"
	KeyF11       Key = "f11"
	KeyF12       Key = "f12"
)

// Modifier represents a keyboard modifier key.
type Modifier string

const (
	ModCmd   Modifier = "cmd"   // Command on macOS, Windows key on Windows
	ModCtrl  Modifier = "ctrl"  // Control key
	ModAlt   Modifier = "alt"   // Option on macOS, Alt on Windows
	ModShift Modifier = "shift" // Shift key
	ModFn    Modifier = "fn"    // Function key (macOS)
)
