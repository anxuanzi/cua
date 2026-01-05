package element

// WalkFunc is called for each element during tree traversal.
// Return true to continue walking, false to stop.
type WalkFunc func(e *Element) bool

// errStopWalk is a sentinel error used to stop tree traversal early.
var errStopWalk = &stopWalkError{}

type stopWalkError struct{}

func (e *stopWalkError) Error() string { return "walk stopped" }

// walkElements performs a depth-first traversal of the element tree.
// Loads children as needed. Returns errStopWalk if the callback returns false.
func walkElements(root *Element, fn WalkFunc) error {
	if root == nil {
		return nil
	}

	// Visit this element
	if !fn(root) {
		return errStopWalk // stopped by callback
	}

	// Load children if not already loaded
	if root.Children == nil {
		if err := root.LoadChildren(); err != nil {
			// Some elements don't support children - that's ok
			return nil
		}
	}

	// Visit children recursively
	for _, child := range root.Children {
		if err := walkElements(child, fn); err != nil {
			return err // propagate stop signal
		}
	}

	return nil
}

// Walk performs a depth-first traversal of the element tree.
// The callback is called for each element. Return true to continue, false to stop.
// Returns nil even if the walk was stopped early by the callback.
func Walk(root *Element, fn WalkFunc) error {
	err := walkElements(root, fn)
	if err == errStopWalk {
		return nil // stopping early is not an error
	}
	return err
}

// Find finds the first element in the tree matching the predicate.
func Find(root *Element, predicate func(*Element) bool) *Element {
	var result *Element
	_ = walkElements(root, func(e *Element) bool {
		if predicate(e) {
			result = e
			return false // stop walking
		}
		return true
	})
	return result
}

// FindAll finds all elements in the tree matching the predicate.
func FindAll(root *Element, predicate func(*Element) bool) []*Element {
	var results []*Element
	_ = walkElements(root, func(e *Element) bool {
		if predicate(e) {
			results = append(results, e)
		}
		return true // continue walking
	})
	return results
}

// Count counts elements in the tree matching the predicate.
func Count(root *Element, predicate func(*Element) bool) int {
	count := 0
	_ = walkElements(root, func(e *Element) bool {
		if predicate(e) {
			count++
		}
		return true
	})
	return count
}

// Depth returns the maximum depth of the element tree.
func Depth(root *Element) int {
	if root == nil {
		return 0
	}

	maxChildDepth := 0
	if root.Children == nil {
		_ = root.LoadChildren()
	}

	for _, child := range root.Children {
		d := Depth(child)
		if d > maxChildDepth {
			maxChildDepth = d
		}
	}

	return 1 + maxChildDepth
}

// Ancestors returns all ancestor elements from the element up to the root.
// The first element is the immediate parent, the last is the root.
func Ancestors(e *Element) []*Element {
	var ancestors []*Element
	current := e.Parent
	for current != nil {
		ancestors = append(ancestors, current)
		current = current.Parent
	}
	return ancestors
}

// Siblings returns all sibling elements (elements with the same parent).
// The element itself is not included in the result.
func Siblings(e *Element) []*Element {
	if e.Parent == nil {
		return nil
	}

	if e.Parent.Children == nil {
		_ = e.Parent.LoadChildren()
	}

	var siblings []*Element
	for _, child := range e.Parent.Children {
		if child.ID != e.ID {
			siblings = append(siblings, child)
		}
	}
	return siblings
}

// Path returns the path from the root to this element as a slice.
// The first element is the root, the last is the element itself.
func Path(e *Element) []*Element {
	ancestors := Ancestors(e)
	if len(ancestors) == 0 {
		return []*Element{e}
	}

	// Reverse ancestors and append self
	path := make([]*Element, len(ancestors)+1)
	for i, a := range ancestors {
		path[len(ancestors)-1-i] = a
	}
	path[len(ancestors)] = e

	return path
}

// PrintTree prints a text representation of the element tree for debugging.
// Returns a slice of lines.
func PrintTree(root *Element, maxDepth int) []string {
	var lines []string
	printTreeRecursive(root, 0, maxDepth, "", true, &lines)
	return lines
}

func printTreeRecursive(e *Element, depth, maxDepth int, prefix string, isLast bool, lines *[]string) {
	if e == nil || (maxDepth > 0 && depth >= maxDepth) {
		return
	}

	// Build the connector
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if depth == 0 {
		connector = ""
	}

	// Build the line
	name := e.Name
	if name == "" {
		name = e.Title
	}
	if name == "" && e.Value != "" {
		name = "value:" + truncate(e.Value, 20)
	}
	if name == "" {
		name = "(unnamed)"
	}

	line := prefix + connector + string(e.Role) + ": " + truncate(name, 40)
	if e.Bounds.Width > 0 && e.Bounds.Height > 0 {
		line += " [" + rectString(e.Bounds) + "]"
	}
	*lines = append(*lines, line)

	// Load children if needed
	if e.Children == nil {
		_ = e.LoadChildren()
	}

	// Print children
	childPrefix := prefix
	if depth > 0 {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i, child := range e.Children {
		isLastChild := i == len(e.Children)-1
		printTreeRecursive(child, depth+1, maxDepth, childPrefix, isLastChild, lines)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func rectString(r Rect) string {
	return string(rune('0'+r.X/100)) + string(rune('0'+(r.X%100)/10)) + string(rune('0'+r.X%10)) +
		"," + string(rune('0'+r.Y/100)) + string(rune('0'+(r.Y%100)/10)) + string(rune('0'+r.Y%10)) +
		" " + string(rune('0'+r.Width/100)) + string(rune('0'+(r.Width%100)/10)) + string(rune('0'+r.Width%10)) +
		"x" + string(rune('0'+r.Height/100)) + string(rune('0'+(r.Height%100)/10)) + string(rune('0'+r.Height%10))
}
