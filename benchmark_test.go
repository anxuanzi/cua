package cua

import (
	"testing"
	"time"
)

// Benchmarks for the CUA library.
// Run with: go test -bench=. -benchmem ./...
//
// Performance targets:
//   - Screenshot: <100ms
//   - Element find: <50ms per query
//   - Input operations: <10ms

// BenchmarkCaptureScreen benchmarks full screen capture.
// Target: <100ms
func BenchmarkCaptureScreen(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		img, err := CaptureScreen()
		if err != nil {
			b.Fatalf("CaptureScreen failed: %v", err)
		}
		if img == nil {
			b.Fatal("CaptureScreen returned nil image")
		}
	}
}

// BenchmarkCaptureRect benchmarks region capture.
func BenchmarkCaptureRect(b *testing.B) {
	b.ReportAllocs()

	rect := Rect{X: 0, Y: 0, Width: 800, Height: 600}

	for i := 0; i < b.N; i++ {
		img, err := CaptureRect(rect)
		if err != nil {
			b.Fatalf("CaptureRect failed: %v", err)
		}
		if img == nil {
			b.Fatal("CaptureRect returned nil image")
		}
	}
}

// BenchmarkScreenSize benchmarks getting screen dimensions.
func BenchmarkScreenSize(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w, h, err := ScreenSize()
		if err != nil {
			b.Fatalf("ScreenSize failed: %v", err)
		}
		if w == 0 || h == 0 {
			b.Fatal("ScreenSize returned zero dimensions")
		}
	}
}

// BenchmarkFindElements benchmarks finding UI elements.
// Target: <50ms per query
func BenchmarkFindElements(b *testing.B) {
	b.ReportAllocs()

	// Skip if accessibility is not available
	if _, err := getFinder(); err != nil {
		b.Skipf("Finder not available: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := FindElements(ByRole(RoleButton))
		if err != nil {
			// Don't fail - just continue
			continue
		}
	}
}

// BenchmarkFindElementsCombined benchmarks finding with combined selectors.
func BenchmarkFindElementsCombined(b *testing.B) {
	b.ReportAllocs()

	// Skip if accessibility is not available
	if _, err := getFinder(); err != nil {
		b.Skipf("Finder not available: %v", err)
	}

	selector := And(
		ByRole(RoleButton),
		ByEnabled(),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := FindElements(selector)
		if err != nil {
			continue
		}
	}
}

// BenchmarkFocusedApplication benchmarks getting focused app.
func BenchmarkFocusedApplication(b *testing.B) {
	b.ReportAllocs()

	// Skip if accessibility is not available
	if _, err := getFinder(); err != nil {
		b.Skipf("Finder not available: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app, err := FocusedApplication()
		if err != nil {
			continue
		}
		if app == nil {
			b.Fatal("FocusedApplication returned nil")
		}
	}
}

// BenchmarkFocusedElement benchmarks getting focused element.
func BenchmarkFocusedElement(b *testing.B) {
	b.ReportAllocs()

	// Skip if accessibility is not available
	if _, err := getFinder(); err != nil {
		b.Skipf("Finder not available: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := FocusedElement()
		if err != nil {
			// No element focused is OK
			continue
		}
	}
}

// BenchmarkAgentCreation benchmarks creating an agent.
func BenchmarkAgentCreation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create with mock API key (won't actually work but tests creation speed)
		agent := New(WithAPIKey("test-key"))
		if agent == nil {
			b.Fatal("New returned nil agent")
		}
	}
}

// BenchmarkAgentCreationWithOptions benchmarks creating an agent with all options.
func BenchmarkAgentCreationWithOptions(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		agent := New(
			WithAPIKey("test-key"),
			WithModel(Gemini2Flash),
			WithSafetyLevel(SafetyNormal),
			WithTimeout(5*time.Minute),
			WithMaxActions(100),
			WithVerbose(false),
			WithHeadless(true),
			WithRateLimit(60),
		)
		if agent == nil {
			b.Fatal("New returned nil agent")
		}
	}
}

// BenchmarkSelectorMatching benchmarks selector matching performance.
func BenchmarkSelectorMatching(b *testing.B) {
	// Create a sample element
	elem := &Element{
		Role:    RoleButton,
		Name:    "Submit Button",
		Title:   "Click to Submit",
		Enabled: true,
		Focused: false,
	}

	selectors := map[string]Selector{
		"ByRole":         ByRole(RoleButton),
		"ByName":         ByName("Submit Button"),
		"ByNameContains": ByNameContains("Submit"),
		"ByEnabled":      ByEnabled(),
		"And2":           And(ByRole(RoleButton), ByEnabled()),
		"And3":           And(ByRole(RoleButton), ByEnabled(), ByNameContains("Submit")),
		"Or2":            Or(ByRole(RoleButton), ByRole(RoleTextField)),
		"Not":            Not(ByFocused()),
	}

	for name, sel := range selectors {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = sel.Matches(elem)
			}
		})
	}
}

// BenchmarkRectOperations benchmarks Rect method performance.
func BenchmarkRectOperations(b *testing.B) {
	rect := Rect{X: 100, Y: 100, Width: 800, Height: 600}
	point := Point{X: 500, Y: 400}

	b.Run("Center", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = rect.Center()
		}
	})

	b.Run("Contains", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = rect.Contains(point)
		}
	})

	b.Run("IsEmpty", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = rect.IsEmpty()
		}
	})
}
