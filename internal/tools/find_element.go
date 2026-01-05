package tools

import (
	"fmt"

	"github.com/anxuanzi/cua/pkg/element"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// FindElementArgs defines the arguments for the find_element tool.
type FindElementArgs struct {
	// Role is the UI element role to search for (e.g., "button", "textfield", "window").
	Role string `json:"role,omitempty" jsonschema:"UI element role to search for (e.g., 'button', 'textfield', 'window', 'checkbox', 'link')"`

	// Name is the exact name/label to match.
	Name string `json:"name,omitempty" jsonschema:"Exact name/label of the element to find"`

	// NameContains is a substring to match in the element name.
	NameContains string `json:"name_contains,omitempty" jsonschema:"Substring to find in element names (case-insensitive)"`

	// Title is the exact title to match (for windows).
	Title string `json:"title,omitempty" jsonschema:"Exact title of the element (typically for windows)"`

	// MaxResults limits the number of results returned.
	MaxResults int `json:"max_results,omitzero" jsonschema:"Maximum number of results to return (default: 10)"`
}

// FoundElement represents a found UI element.
type FoundElement struct {
	// ID is the element's unique identifier (not stable across queries).
	ID string `json:"id"`

	// Role is the semantic type of the element.
	Role string `json:"role"`

	// Name is the accessible name/label.
	Name string `json:"name,omitempty"`

	// Title is the element title.
	Title string `json:"title,omitempty"`

	// Value is the current value (for inputs).
	Value string `json:"value,omitempty"`

	// Bounds contains the screen position and size.
	Bounds ElementBounds `json:"bounds"`

	// CenterX is the X coordinate of the element center (for clicking).
	CenterX int `json:"center_x"`

	// CenterY is the Y coordinate of the element center (for clicking).
	CenterY int `json:"center_y"`

	// Enabled indicates if the element is interactive.
	Enabled bool `json:"enabled"`

	// Focused indicates if the element has keyboard focus.
	Focused bool `json:"focused"`
}

// ElementBounds represents an element's screen rectangle.
type ElementBounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// FindElementResult contains the found elements.
type FindElementResult struct {
	// Success indicates if the search succeeded.
	Success bool `json:"success"`

	// Count is the number of elements found.
	Count int `json:"count"`

	// Elements are the found elements.
	Elements []FoundElement `json:"elements,omitempty"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// findElement handles the find_element tool invocation.
func findElement(ctx tool.Context, args FindElementArgs) (FindElementResult, error) {
	// Create the finder
	finder, err := element.NewFinder()
	if err != nil {
		return FindElementResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create finder: %v", err),
		}, nil
	}
	defer finder.Close()

	// Build selector based on arguments
	var selectors []element.Selector

	if args.Role != "" {
		selectors = append(selectors, element.ByRole(element.Role(args.Role)))
	}
	if args.Name != "" {
		selectors = append(selectors, element.ByName(args.Name))
	}
	if args.NameContains != "" {
		selectors = append(selectors, element.ByNameContains(args.NameContains))
	}
	if args.Title != "" {
		selectors = append(selectors, element.ByTitle(args.Title))
	}

	// Must have at least one selector
	if len(selectors) == 0 {
		return FindElementResult{
			Success: false,
			Error:   "at least one search criteria is required (role, name, name_contains, or title)",
		}, nil
	}

	// Combine selectors with AND
	var selector element.Selector
	if len(selectors) == 1 {
		selector = selectors[0]
	} else {
		selector = element.And(selectors...)
	}

	// Find all matching elements
	elements, err := finder.FindAll(selector)
	if err != nil {
		return FindElementResult{
			Success: false,
			Error:   fmt.Sprintf("search failed: %v", err),
		}, nil
	}

	// Apply max results limit
	maxResults := args.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}
	if len(elements) > maxResults {
		elements = elements[:maxResults]
	}

	// Convert to result format
	foundElements := make([]FoundElement, len(elements))
	for i, el := range elements {
		center := el.Bounds.Center()
		foundElements[i] = FoundElement{
			ID:    el.ID,
			Role:  string(el.Role),
			Name:  el.Name,
			Title: el.Title,
			Value: el.Value,
			Bounds: ElementBounds{
				X:      el.Bounds.X,
				Y:      el.Bounds.Y,
				Width:  el.Bounds.Width,
				Height: el.Bounds.Height,
			},
			CenterX: center.X,
			CenterY: center.Y,
			Enabled: el.Enabled,
			Focused: el.Focused,
		}
	}

	return FindElementResult{
		Success:  true,
		Count:    len(foundElements),
		Elements: foundElements,
	}, nil
}

// NewFindElementTool creates the find_element tool for ADK agents.
func NewFindElementTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "find_element",
			Description: "Finds UI elements on screen using the accessibility tree. Search by role (button, textfield, etc.), name, or title. Returns element positions for clicking.",
		},
		findElement,
	)
}
