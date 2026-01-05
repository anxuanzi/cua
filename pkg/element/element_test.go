package element

import (
	"runtime"
	"testing"
)

func TestRectCenter(t *testing.T) {
	tests := []struct {
		name string
		rect Rect
		want Point
	}{
		{
			name: "simple rectangle",
			rect: Rect{X: 0, Y: 0, Width: 100, Height: 100},
			want: Point{X: 50, Y: 50},
		},
		{
			name: "offset rectangle",
			rect: Rect{X: 100, Y: 200, Width: 50, Height: 60},
			want: Point{X: 125, Y: 230},
		},
		{
			name: "zero size",
			rect: Rect{X: 10, Y: 20, Width: 0, Height: 0},
			want: Point{X: 10, Y: 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rect.Center()
			if got != tt.want {
				t.Errorf("Center() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRectContains(t *testing.T) {
	rect := Rect{X: 10, Y: 20, Width: 100, Height: 50}

	tests := []struct {
		name  string
		point Point
		want  bool
	}{
		{"inside", Point{50, 40}, true},
		{"top-left corner", Point{10, 20}, true},
		{"bottom-right edge", Point{109, 69}, true},
		{"outside left", Point{5, 40}, false},
		{"outside right", Point{111, 40}, false},
		{"outside top", Point{50, 15}, false},
		{"outside bottom", Point{50, 71}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rect.Contains(tt.point); got != tt.want {
				t.Errorf("Contains(%v) = %v, want %v", tt.point, got, tt.want)
			}
		})
	}
}

func TestRectIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		rect Rect
		want bool
	}{
		{"normal", Rect{X: 0, Y: 0, Width: 100, Height: 50}, false},
		{"zero width", Rect{X: 0, Y: 0, Width: 0, Height: 50}, true},
		{"zero height", Rect{X: 0, Y: 0, Width: 100, Height: 0}, true},
		{"negative width", Rect{X: 0, Y: 0, Width: -10, Height: 50}, true},
		{"negative height", Rect{X: 0, Y: 0, Width: 100, Height: -10}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rect.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElementString(t *testing.T) {
	elem := &Element{
		Role:   RoleButton,
		Name:   "Submit",
		Bounds: Rect{X: 100, Y: 200, Width: 80, Height: 30},
	}

	got := elem.String()
	if got == "" {
		t.Error("String() returned empty string")
	}

	// Should contain role and name
	if !containsSubstring(got, "button") {
		t.Errorf("String() should contain role, got %s", got)
	}
	if !containsSubstring(got, "Submit") {
		t.Errorf("String() should contain name, got %s", got)
	}
}

func TestSelectorByRole(t *testing.T) {
	selector := ByRole(RoleButton)

	tests := []struct {
		name    string
		element *Element
		want    bool
	}{
		{"matches button", &Element{Role: RoleButton}, true},
		{"no match textfield", &Element{Role: RoleTextField}, false},
		{"no match unknown", &Element{Role: RoleUnknown}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selector.Matches(tt.element); got != tt.want {
				t.Errorf("ByRole(Button).Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectorByName(t *testing.T) {
	selector := ByName("Submit")

	tests := []struct {
		name    string
		element *Element
		want    bool
	}{
		{"exact match", &Element{Name: "Submit"}, true},
		{"case mismatch", &Element{Name: "submit"}, false},
		{"partial match", &Element{Name: "Submit Button"}, false},
		{"no match", &Element{Name: "Cancel"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selector.Matches(tt.element); got != tt.want {
				t.Errorf("ByName(Submit).Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectorByNameContains(t *testing.T) {
	selector := ByNameContains("save")

	tests := []struct {
		name    string
		element *Element
		want    bool
	}{
		{"exact match", &Element{Name: "save"}, true},
		{"contains lowercase", &Element{Name: "save file"}, true},
		{"contains uppercase", &Element{Name: "Save File"}, true},
		{"contains mixed", &Element{Name: "AutoSave"}, true},
		{"no match", &Element{Name: "Cancel"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selector.Matches(tt.element); got != tt.want {
				t.Errorf("ByNameContains(save).Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectorAnd(t *testing.T) {
	selector := And(ByRole(RoleButton), ByNameContains("submit"))

	tests := []struct {
		name    string
		element *Element
		want    bool
	}{
		{"both match", &Element{Role: RoleButton, Name: "Submit"}, true},
		{"role only", &Element{Role: RoleButton, Name: "Cancel"}, false},
		{"name only", &Element{Role: RoleTextField, Name: "Submit"}, false},
		{"neither", &Element{Role: RoleTextField, Name: "Cancel"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selector.Matches(tt.element); got != tt.want {
				t.Errorf("And().Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectorOr(t *testing.T) {
	selector := Or(ByRole(RoleButton), ByRole(RoleLink))

	tests := []struct {
		name    string
		element *Element
		want    bool
	}{
		{"first matches", &Element{Role: RoleButton}, true},
		{"second matches", &Element{Role: RoleLink}, true},
		{"neither matches", &Element{Role: RoleTextField}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selector.Matches(tt.element); got != tt.want {
				t.Errorf("Or().Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectorNot(t *testing.T) {
	selector := Not(ByRole(RoleButton))

	tests := []struct {
		name    string
		element *Element
		want    bool
	}{
		{"not button", &Element{Role: RoleTextField}, true},
		{"is button", &Element{Role: RoleButton}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selector.Matches(tt.element); got != tt.want {
				t.Errorf("Not().Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectorByEnabled(t *testing.T) {
	selector := ByEnabled()

	tests := []struct {
		name    string
		element *Element
		want    bool
	}{
		{"enabled", &Element{Enabled: true}, true},
		{"disabled", &Element{Enabled: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selector.Matches(tt.element); got != tt.want {
				t.Errorf("ByEnabled().Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectorByPredicate(t *testing.T) {
	// Custom predicate: element width > 100
	selector := ByPredicate(func(e *Element) bool {
		return e.Bounds.Width > 100
	})

	tests := []struct {
		name    string
		element *Element
		want    bool
	}{
		{"wide element", &Element{Bounds: Rect{Width: 200}}, true},
		{"narrow element", &Element{Bounds: Rect{Width: 50}}, false},
		{"exactly 100", &Element{Bounds: Rect{Width: 100}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selector.Matches(tt.element); got != tt.want {
				t.Errorf("ByPredicate().Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTreeWalk(t *testing.T) {
	// Create a simple tree
	root := &Element{Name: "root", Children: []*Element{
		{Name: "child1", Children: []*Element{
			{Name: "grandchild1"},
			{Name: "grandchild2"},
		}},
		{Name: "child2"},
	}}

	var visited []string
	Walk(root, func(e *Element) bool {
		visited = append(visited, e.Name)
		return true
	})

	// Should visit all elements in depth-first order
	expected := []string{"root", "child1", "grandchild1", "grandchild2", "child2"}
	if len(visited) != len(expected) {
		t.Errorf("Walk visited %d elements, want %d", len(visited), len(expected))
	}

	for i, name := range expected {
		if i < len(visited) && visited[i] != name {
			t.Errorf("Walk order[%d] = %s, want %s", i, visited[i], name)
		}
	}
}

func TestTreeFind(t *testing.T) {
	root := &Element{Name: "root", Role: RoleWindow, Children: []*Element{
		{Name: "button1", Role: RoleButton},
		{Name: "text1", Role: RoleTextField},
		{Name: "button2", Role: RoleButton},
	}}

	// Find first button
	found := Find(root, func(e *Element) bool {
		return e.Role == RoleButton
	})

	if found == nil {
		t.Fatal("Find() returned nil")
	}
	if found.Name != "button1" {
		t.Errorf("Find() found %s, want button1", found.Name)
	}
}

func TestTreeFindAll(t *testing.T) {
	root := &Element{Name: "root", Role: RoleWindow, Children: []*Element{
		{Name: "button1", Role: RoleButton},
		{Name: "text1", Role: RoleTextField},
		{Name: "button2", Role: RoleButton},
	}}

	// Find all buttons
	found := FindAll(root, func(e *Element) bool {
		return e.Role == RoleButton
	})

	if len(found) != 2 {
		t.Errorf("FindAll() found %d elements, want 2", len(found))
	}
}

func TestTreeCount(t *testing.T) {
	root := &Element{Name: "root", Children: []*Element{
		{Name: "child1", Children: []*Element{
			{Name: "grandchild1"},
		}},
		{Name: "child2"},
	}}

	count := Count(root, func(e *Element) bool {
		return true // count all
	})

	if count != 4 {
		t.Errorf("Count() = %d, want 4", count)
	}
}

func TestMapRole(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("mapRole is only available on macOS")
	}

	tests := []struct {
		axRole string
		want   Role
	}{
		{"AXWindow", RoleWindow},
		{"AXButton", RoleButton},
		{"AXTextField", RoleTextField},
		{"AXStaticText", RoleStaticText},
		{"AXUnknownRole", RoleUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.axRole, func(t *testing.T) {
			got := mapRole(tt.axRole)
			if got != tt.want {
				t.Errorf("mapRole(%s) = %s, want %s", tt.axRole, got, tt.want)
			}
		})
	}
}

// Integration test - requires accessibility permissions
func TestFinderIntegration(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Finder integration test only available on macOS")
	}

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	finder, err := NewFinder()
	if err != nil {
		if err == ErrPermissionDenied {
			t.Skip("Accessibility permissions not granted")
		}
		t.Fatalf("NewFinder() error: %v", err)
	}
	defer finder.Close()

	// Get frontmost application
	app, err := finder.FocusedApplication()
	if err != nil {
		t.Fatalf("FocusedApplication() error: %v", err)
	}

	t.Logf("Focused app: %s (PID: %d)", app.Name, app.PID)

	// Find all windows in the app
	windows, err := finder.FindAll(ByRole(RoleWindow))
	if err != nil {
		t.Fatalf("FindAll(Window) error: %v", err)
	}

	t.Logf("Found %d windows", len(windows))
	for _, w := range windows {
		t.Logf("  Window: %s at %v", w.Title, w.Bounds)
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
