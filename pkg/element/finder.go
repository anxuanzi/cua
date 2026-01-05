package element

import (
	"strings"
	"time"
)

// Finder locates UI elements on screen using the accessibility API.
// Create a Finder with NewFinder() and remember to call Close() when done.
type Finder struct {
	// impl holds the platform-specific implementation
	impl finderImpl
}

// finderImpl is the platform-specific finder implementation.
// Defined in darwin.go / windows.go
type finderImpl interface {
	// Root returns the root element (typically the system-wide element).
	Root() (*Element, error)

	// FocusedApplication returns the frontmost application element.
	FocusedApplication() (*Element, error)

	// FocusedElement returns the element that currently has keyboard focus.
	FocusedElement() (*Element, error)

	// ApplicationByPID returns the application element for a process ID.
	ApplicationByPID(pid int) (*Element, error)

	// ApplicationByName returns the application element by name.
	ApplicationByName(name string) (*Element, error)

	// AllApplications returns all running application elements.
	AllApplications() ([]*Element, error)

	// Close releases any resources held by the finder.
	Close() error
}

// NewFinder creates a new Finder for locating UI elements.
// On macOS, this requires accessibility permissions.
// Call Close() when done to release resources.
func NewFinder() (*Finder, error) {
	impl, err := newFinderImpl()
	if err != nil {
		return nil, err
	}
	return &Finder{impl: impl}, nil
}

// Close releases resources held by the Finder.
func (f *Finder) Close() error {
	if f.impl != nil {
		return f.impl.Close()
	}
	return nil
}

// Root returns the system-wide root element.
// All applications are children of this element.
func (f *Finder) Root() (*Element, error) {
	return f.impl.Root()
}

// FocusedApplication returns the frontmost application.
func (f *Finder) FocusedApplication() (*Element, error) {
	return f.impl.FocusedApplication()
}

// FocusedElement returns the element that currently has keyboard focus.
func (f *Finder) FocusedElement() (*Element, error) {
	return f.impl.FocusedElement()
}

// ApplicationByPID returns the application element for a process ID.
func (f *Finder) ApplicationByPID(pid int) (*Element, error) {
	return f.impl.ApplicationByPID(pid)
}

// ApplicationByName returns the application element by name.
// The name is matched case-insensitively.
func (f *Finder) ApplicationByName(name string) (*Element, error) {
	return f.impl.ApplicationByName(name)
}

// AllApplications returns all running application elements.
func (f *Finder) AllApplications() ([]*Element, error) {
	return f.impl.AllApplications()
}

// Find returns the first element matching the selector.
// Returns ErrNotFound if no element matches.
func (f *Finder) Find(selector Selector) (*Element, error) {
	elements, err := f.FindAll(selector)
	if err != nil {
		return nil, err
	}
	if len(elements) == 0 {
		return nil, ErrNotFound
	}
	return elements[0], nil
}

// FindAll returns all elements matching the selector.
// Searches within the focused application by default.
func (f *Finder) FindAll(selector Selector) ([]*Element, error) {
	return f.FindAllIn(nil, selector)
}

// FindIn returns the first element matching the selector within the given root.
// If root is nil, searches within the focused application.
func (f *Finder) FindIn(root *Element, selector Selector) (*Element, error) {
	elements, err := f.FindAllIn(root, selector)
	if err != nil {
		return nil, err
	}
	if len(elements) == 0 {
		return nil, ErrNotFound
	}
	return elements[0], nil
}

// FindAllIn returns all elements matching the selector within the given root.
// If root is nil, searches within the focused application.
func (f *Finder) FindAllIn(root *Element, selector Selector) ([]*Element, error) {
	if root == nil {
		var err error
		root, err = f.FocusedApplication()
		if err != nil {
			return nil, err
		}
	}

	var results []*Element
	err := walkElements(root, func(e *Element) bool {
		if selector.Matches(e) {
			results = append(results, e)
		}
		return true // continue walking
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// WaitFor waits for an element matching the selector to appear.
// Returns ErrTimeout if the element doesn't appear within the timeout.
func (f *Finder) WaitFor(selector Selector, timeout time.Duration) (*Element, error) {
	return f.WaitForIn(nil, selector, timeout)
}

// WaitForIn waits for an element matching the selector within root.
// Returns ErrTimeout if the element doesn't appear within the timeout.
func (f *Finder) WaitForIn(root *Element, selector Selector, timeout time.Duration) (*Element, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := 100 * time.Millisecond

	for {
		el, err := f.FindIn(root, selector)
		if err == nil {
			return el, nil
		}
		if err != ErrNotFound {
			return nil, err
		}

		if time.Now().After(deadline) {
			return nil, ErrTimeout
		}

		time.Sleep(pollInterval)
	}
}

// Selector is used to find elements by various criteria.
type Selector interface {
	// Matches returns true if the element matches this selector.
	Matches(e *Element) bool
}

// roleSelector matches elements by their role.
type roleSelector struct {
	role Role
}

func (s roleSelector) Matches(e *Element) bool {
	return e.Role == s.role
}

// ByRole creates a selector that matches elements by their role.
func ByRole(role Role) Selector {
	return roleSelector{role: role}
}

// nameSelector matches elements by their exact name.
type nameSelector struct {
	name string
}

func (s nameSelector) Matches(e *Element) bool {
	return e.Name == s.name
}

// ByName creates a selector that matches elements by their exact name.
func ByName(name string) Selector {
	return nameSelector{name: name}
}

// nameContainsSelector matches elements whose name contains a substring.
type nameContainsSelector struct {
	substring string
}

func (s nameContainsSelector) Matches(e *Element) bool {
	return strings.Contains(strings.ToLower(e.Name), strings.ToLower(s.substring))
}

// ByNameContains creates a selector that matches elements whose name contains the substring.
// The match is case-insensitive.
func ByNameContains(substring string) Selector {
	return nameContainsSelector{substring: substring}
}

// titleSelector matches elements by their exact title.
type titleSelector struct {
	title string
}

func (s titleSelector) Matches(e *Element) bool {
	return e.Title == s.title
}

// ByTitle creates a selector that matches elements by their exact title.
func ByTitle(title string) Selector {
	return titleSelector{title: title}
}

// titleContainsSelector matches elements whose title contains a substring.
type titleContainsSelector struct {
	substring string
}

func (s titleContainsSelector) Matches(e *Element) bool {
	return strings.Contains(strings.ToLower(e.Title), strings.ToLower(s.substring))
}

// ByTitleContains creates a selector that matches elements whose title contains the substring.
// The match is case-insensitive.
func ByTitleContains(substring string) Selector {
	return titleContainsSelector{substring: substring}
}

// valueSelector matches elements by their exact value.
type valueSelector struct {
	value string
}

func (s valueSelector) Matches(e *Element) bool {
	return e.Value == s.value
}

// ByValue creates a selector that matches elements by their exact value.
func ByValue(value string) Selector {
	return valueSelector{value: value}
}

// enabledSelector matches elements that are enabled.
type enabledSelector struct{}

func (s enabledSelector) Matches(e *Element) bool {
	return e.Enabled
}

// ByEnabled creates a selector that matches only enabled elements.
func ByEnabled() Selector {
	return enabledSelector{}
}

// focusedSelector matches elements that are focused.
type focusedSelector struct{}

func (s focusedSelector) Matches(e *Element) bool {
	return e.Focused
}

// ByFocused creates a selector that matches only focused elements.
func ByFocused() Selector {
	return focusedSelector{}
}

// andSelector matches elements that match ALL provided selectors.
type andSelector struct {
	selectors []Selector
}

func (s andSelector) Matches(e *Element) bool {
	for _, sel := range s.selectors {
		if !sel.Matches(e) {
			return false
		}
	}
	return true
}

// And creates a selector that matches elements matching ALL provided selectors.
func And(selectors ...Selector) Selector {
	return andSelector{selectors: selectors}
}

// orSelector matches elements that match ANY provided selector.
type orSelector struct {
	selectors []Selector
}

func (s orSelector) Matches(e *Element) bool {
	for _, sel := range s.selectors {
		if sel.Matches(e) {
			return true
		}
	}
	return false
}

// Or creates a selector that matches elements matching ANY provided selector.
func Or(selectors ...Selector) Selector {
	return orSelector{selectors: selectors}
}

// notSelector matches elements that do NOT match the provided selector.
type notSelector struct {
	selector Selector
}

func (s notSelector) Matches(e *Element) bool {
	return !s.selector.Matches(e)
}

// Not creates a selector that matches elements that do NOT match the provided selector.
func Not(selector Selector) Selector {
	return notSelector{selector: selector}
}

// predicateSelector matches elements using a custom predicate function.
type predicateSelector struct {
	fn func(*Element) bool
}

func (s predicateSelector) Matches(e *Element) bool {
	return s.fn(e)
}

// ByPredicate creates a selector using a custom predicate function.
func ByPredicate(fn func(*Element) bool) Selector {
	return predicateSelector{fn: fn}
}

// Platform-specific implementation constructor (defined in darwin.go / windows.go)
var newFinderImpl func() (finderImpl, error) = func() (finderImpl, error) {
	return nil, ErrNotSupported
}
