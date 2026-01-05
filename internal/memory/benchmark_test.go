package memory

import (
	"fmt"
	"testing"
	"time"
)

// Benchmarks for TaskMemory operations.
// Run with: go test -bench=. -benchmem ./internal/memory/...

// BenchmarkTaskMemoryNew benchmarks creating a new TaskMemory.
func BenchmarkTaskMemoryNew(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = New("Open Safari and navigate to google.com")
	}
}

// BenchmarkRecordAction benchmarks recording an action.
func BenchmarkRecordAction(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")

	args := map[string]any{
		"x": 100,
		"y": 200,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.RecordAction("click", args, true, "Clicked button", 50*time.Millisecond)
	}
}

// BenchmarkRecordActionConcurrent benchmarks concurrent action recording.
func BenchmarkRecordActionConcurrent(b *testing.B) {
	m := New("Test task")

	args := map[string]any{
		"x": 100,
		"y": 200,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.RecordAction("click", args, true, "Clicked button", 50*time.Millisecond)
		}
	})
}

// BenchmarkToPrompt benchmarks generating the prompt string.
func BenchmarkToPrompt(b *testing.B) {
	b.ReportAllocs()

	// Create a TaskMemory with realistic state
	m := New("Open Safari and navigate to google.com")
	m.SetPhase(PhaseNavigation)
	m.AddMilestone("Opened Safari browser")
	m.AddMilestone("Navigated to google.com")
	m.SetKeyFact("current_url", "https://google.com")
	m.SetKeyFact("search_query", "golang tutorial")

	// Record some actions
	for i := 0; i < 5; i++ {
		m.RecordAction("click", map[string]any{"x": 100, "y": 200}, true, "Clicked element", 50*time.Millisecond)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.ToPrompt()
	}
}

// BenchmarkToPromptLarge benchmarks prompt generation with large state.
func BenchmarkToPromptLarge(b *testing.B) {
	b.ReportAllocs()

	// Create a TaskMemory with large realistic state
	m := New("Complete complex workflow with multiple phases")

	// Add many milestones
	for i := 0; i < 20; i++ {
		m.AddMilestone(fmt.Sprintf("Milestone %d completed", i+1))
	}

	// Add key facts
	for i := 0; i < 10; i++ {
		m.SetKeyFact(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
	}

	// Add failed patterns
	for i := 0; i < 10; i++ {
		m.AddFailedPattern(fmt.Sprintf("pattern_%d", i))
	}

	// Record many actions (only last 5 kept)
	for i := 0; i < 100; i++ {
		m.RecordAction("click", map[string]any{"step": i}, true, "Action completed", 50*time.Millisecond)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.ToPrompt()
	}
}

// BenchmarkSummary benchmarks generating a summary.
func BenchmarkSummary(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")
	m.AddMilestone("Step 1 done")
	m.AddMilestone("Step 2 done")
	m.SetKeyFact("key", "value")

	for i := 0; i < 5; i++ {
		m.RecordAction("click", nil, true, "Clicked", 50*time.Millisecond)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.Summary()
	}
}

// BenchmarkSetPhase benchmarks phase changes.
func BenchmarkSetPhase(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")

	phases := []string{
		PhaseNavigation,
		PhaseFormFilling,
		PhaseAuthentication,
		PhaseSearch,
		PhaseBrowsing,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.SetPhase(phases[i%len(phases)])
	}
}

// BenchmarkMaybeUpdatePhase benchmarks automatic phase detection.
func BenchmarkMaybeUpdatePhase(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")

	obs := Observation{
		VisibleText:  []string{"Login", "Username", "Password"},
		ActiveApp:    "Safari",
		HasLoginForm: true,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.MaybeUpdatePhase(obs)
	}
}

// BenchmarkSetKeyFact benchmarks setting key facts.
func BenchmarkSetKeyFact(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.SetKeyFact(fmt.Sprintf("key_%d", i%100), fmt.Sprintf("value_%d", i))
	}
}

// BenchmarkGetKeyFact benchmarks retrieving key facts.
func BenchmarkGetKeyFact(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")
	// Pre-populate some facts
	for i := 0; i < 10; i++ {
		m.SetKeyFact(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = m.GetKeyFact(fmt.Sprintf("key_%d", i%10))
	}
}

// BenchmarkAddFailedPattern benchmarks adding failed patterns.
func BenchmarkAddFailedPattern(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.AddFailedPattern(fmt.Sprintf("pattern_%d", i%20))
	}
}

// BenchmarkHasFailedPattern benchmarks checking failed patterns.
func BenchmarkHasFailedPattern(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")
	// Pre-populate some patterns
	for i := 0; i < 10; i++ {
		m.AddFailedPattern(fmt.Sprintf("pattern_%d", i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.HasFailedPattern(fmt.Sprintf("pattern_%d", i%15))
	}
}

// BenchmarkIsStuck benchmarks the stuck detection.
func BenchmarkIsStuck(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")
	// Add some failures
	for i := 0; i < 3; i++ {
		m.RecordAction("click", nil, false, "Failed", 50*time.Millisecond)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.IsStuck()
	}
}

// BenchmarkNeedsHelp benchmarks the help detection.
func BenchmarkNeedsHelp(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")
	// Add some failures
	for i := 0; i < 5; i++ {
		m.RecordAction("click", nil, false, "Failed", 50*time.Millisecond)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.NeedsHelp()
	}
}

// BenchmarkDuration benchmarks getting the task duration.
func BenchmarkDuration(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.Duration()
	}
}

// BenchmarkAddMilestone benchmarks adding milestones.
func BenchmarkAddMilestone(b *testing.B) {
	b.ReportAllocs()

	m := New("Test task")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.AddMilestone(fmt.Sprintf("Milestone %d", i%30))
	}
}
