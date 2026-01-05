// Package element provides cross-platform UI element access via accessibility APIs.
//
// This package implements our own element/accessibility layer, inspired by
// patterns from existing implementations but written from scratch in our style.
//
// # Platform Support
//
//   - macOS: Uses AXUIElement API via CGo bindings
//   - Windows: Uses UI Automation API via COM
//
// # Basic Usage
//
//	finder, err := element.NewFinder()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer finder.Close()
//
//	// Find all buttons
//	buttons, err := finder.FindAll(element.ByRole(element.RoleButton))
//
//	// Find by name
//	el, err := finder.Find(element.ByName("Submit"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	el.Click()
//
// # Permissions
//
// On macOS, accessibility permissions are required. The app must be granted
// access in System Settings > Privacy & Security > Accessibility.
//
// On Windows, some applications may require running as Administrator.
package element

import (
	"errors"
	"fmt"
)

// Role represents the semantic type of a UI element.
// These map to accessibility roles across platforms.
type Role string

const (
	RoleWindow      Role = "window"
	RoleButton      Role = "button"
	RoleTextField   Role = "textfield"
	RoleTextArea    Role = "textarea"
	RoleStaticText  Role = "statictext"
	RoleCheckbox    Role = "checkbox"
	RoleRadioButton Role = "radiobutton"
	RoleList        Role = "list"
	RoleListItem    Role = "listitem"
	RoleMenu        Role = "menu"
	RoleMenuItem    Role = "menuitem"
	RoleMenuBar     Role = "menubar"
	RoleToolbar     Role = "toolbar"
	RoleScrollArea  Role = "scrollarea"
	RoleScrollBar   Role = "scrollbar"
	RoleImage       Role = "image"
	RoleLink        Role = "link"
	RoleGroup       Role = "group"
	RoleTab         Role = "tab"
	RoleTabGroup    Role = "tabgroup"
	RoleTable       Role = "table"
	RoleRow         Role = "row"
	RoleCell        Role = "cell"
	RoleColumn      Role = "column"
	RoleSlider      Role = "slider"
	RoleComboBox    Role = "combobox"
	RolePopUpButton Role = "popupbutton"
	RoleProgressBar Role = "progressbar"
	RoleSplitter    Role = "splitter"
	RoleSheet       Role = "sheet"
	RoleDrawer      Role = "drawer"
	RoleDialog      Role = "dialog"
	RoleApplication Role = "application"
	RoleUnknown     Role = "unknown"
)

// Rect represents a rectangle on screen in pixel coordinates.
type Rect struct {
	X      int // Left edge
	Y      int // Top edge
	Width  int
	Height int
}

// Center returns the center point of the rectangle.
func (r Rect) Center() Point {
	return Point{
		X: r.X + r.Width/2,
		Y: r.Y + r.Height/2,
	}
}

// Contains returns true if the point is within the rectangle.
func (r Rect) Contains(p Point) bool {
	return p.X >= r.X && p.X < r.X+r.Width &&
		p.Y >= r.Y && p.Y < r.Y+r.Height
}

// IsEmpty returns true if the rectangle has zero area.
func (r Rect) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

// Point represents a point on screen in pixel coordinates.
type Point struct {
	X int
	Y int
}

// Element represents a UI element on screen.
// Elements are obtained via Finder and provide information about
// the element's properties and hierarchy.
type Element struct {
	// ID is a unique identifier for this element within the current tree.
	// This is NOT stable across queries - don't cache it.
	ID string

	// Role is the semantic type of the element.
	Role Role

	// Name is the accessible name/label of the element.
	// This is what screen readers announce.
	Name string

	// Title is the window/element title (may differ from Name).
	Title string

	// Value is the current value for inputs, sliders, etc.
	Value string

	// Description is additional accessible description text.
	Description string

	// Bounds is the screen rectangle containing this element.
	Bounds Rect

	// Enabled indicates if the element can be interacted with.
	Enabled bool

	// Focused indicates if the element currently has keyboard focus.
	Focused bool

	// Selected indicates if the element is selected (for selectable items).
	Selected bool

	// Children contains child elements in the accessibility tree.
	// May be nil if not yet loaded (use LoadChildren to populate).
	Children []*Element

	// Parent is the parent element. May be nil for root elements.
	Parent *Element

	// PID is the process ID of the owning application.
	PID int

	// Attributes contains additional platform-specific attributes.
	Attributes map[string]interface{}

	// handle is the platform-specific element reference (unexported).
	// On macOS: AXUIElementRef
	// On Windows: IUIAutomationElement pointer
	handle interface{}
}

// Focus sets keyboard focus to this element.
func (e *Element) Focus() error {
	return focusElement(e)
}

// PerformAction performs a named action on this element.
// Common actions: "AXPress", "AXConfirm", "AXCancel", "AXRaise"
func (e *Element) PerformAction(action string) error {
	return performAction(e, action)
}

// SetValue sets the value of this element (for text fields, sliders, etc).
func (e *Element) SetValue(value string) error {
	return setValue(e, value)
}

// LoadChildren populates the Children slice with immediate child elements.
// Call this if you need to traverse the element tree.
func (e *Element) LoadChildren() error {
	return loadChildren(e)
}

// String returns a human-readable representation of the element.
func (e *Element) String() string {
	name := e.Name
	if name == "" {
		name = e.Title
	}
	if name == "" {
		name = "(no name)"
	}
	return fmt.Sprintf("%s[%s] at (%d,%d) %dx%d",
		e.Role, name, e.Bounds.X, e.Bounds.Y, e.Bounds.Width, e.Bounds.Height)
}

// Common errors
var (
	// ErrNotSupported indicates the operation is not supported on this platform.
	ErrNotSupported = errors.New("element: operation not supported on this platform")

	// ErrPermissionDenied indicates missing accessibility permissions.
	ErrPermissionDenied = errors.New("element: accessibility permission denied")

	// ErrNotFound indicates no element matched the query.
	ErrNotFound = errors.New("element: element not found")

	// ErrNoBounds indicates the element has no valid bounds.
	ErrNoBounds = errors.New("element: element has no bounds")

	// ErrInvalidElement indicates the element reference is no longer valid.
	ErrInvalidElement = errors.New("element: element reference is invalid")

	// ErrTimeout indicates a timeout waiting for an element.
	ErrTimeout = errors.New("element: timeout waiting for element")

	// ErrNoFocus indicates no element currently has focus.
	ErrNoFocus = errors.New("element: no focused element")
)

// Platform-specific implementations (defined in darwin.go / windows.go)
var (
	focusElement  func(e *Element) error                = notSupported1[*Element]
	performAction func(e *Element, action string) error = notSupported2[*Element, string]
	setValue      func(e *Element, value string) error  = notSupported2[*Element, string]
	loadChildren  func(e *Element) error                = notSupported1[*Element]
)

// Helper functions for default implementations
func notSupported1[T any](_ T) error {
	return ErrNotSupported
}

func notSupported2[T, U any](_ T, _ U) error {
	return ErrNotSupported
}
