package screen

import (
	"image/png"
	"io"
	"testing"
)

// Benchmarks for screen capture operations.
// Run with: go test -bench=. -benchmem ./pkg/screen/...
//
// Performance targets:
//   - Full screen capture: <100ms
//   - Region capture: <50ms
//   - Display info: <1ms

// BenchmarkCapturePrimary benchmarks full primary screen capture.
// Target: <100ms
func BenchmarkCapturePrimary(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		img, err := CapturePrimary()
		if err != nil {
			b.Fatalf("CapturePrimary failed: %v", err)
		}
		if img == nil {
			b.Fatal("CapturePrimary returned nil image")
		}
	}
}

// BenchmarkCaptureRect benchmarks region capture.
// Target: <50ms
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

// BenchmarkCaptureRectSmall benchmarks small region capture.
func BenchmarkCaptureRectSmall(b *testing.B) {
	b.ReportAllocs()

	rect := Rect{X: 0, Y: 0, Width: 100, Height: 100}

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

// BenchmarkCaptureRectLarge benchmarks large region capture.
func BenchmarkCaptureRectLarge(b *testing.B) {
	b.ReportAllocs()

	// Get display size for full-screen equivalent
	display, err := PrimaryDisplay()
	if err != nil {
		b.Fatalf("PrimaryDisplay failed: %v", err)
	}

	rect := Rect{X: 0, Y: 0, Width: display.Bounds.Width, Height: display.Bounds.Height}

	b.ResetTimer()

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

// BenchmarkPrimaryDisplay benchmarks getting primary display info.
// Target: <1ms
func BenchmarkPrimaryDisplay(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		display, err := PrimaryDisplay()
		if err != nil {
			b.Fatalf("PrimaryDisplay failed: %v", err)
		}
		if display.Bounds.Width == 0 {
			b.Fatal("PrimaryDisplay returned zero width")
		}
	}
}

// BenchmarkDisplays benchmarks getting all display info.
func BenchmarkDisplays(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		displays := Displays()
		if len(displays) == 0 {
			b.Fatal("Displays returned no displays")
		}
	}
}

// BenchmarkCaptureAndEncode benchmarks capture + PNG encoding.
// This is the full pipeline for sending to the AI model.
func BenchmarkCaptureAndEncode(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		img, err := CapturePrimary()
		if err != nil {
			b.Fatalf("CapturePrimary failed: %v", err)
		}

		// Encode to PNG (discarding output)
		err = png.Encode(io.Discard, img)
		if err != nil {
			b.Fatalf("PNG encode failed: %v", err)
		}
	}
}

// BenchmarkCaptureRectAndEncode benchmarks region capture + PNG encoding.
func BenchmarkCaptureRectAndEncode(b *testing.B) {
	b.ReportAllocs()

	rect := Rect{X: 0, Y: 0, Width: 800, Height: 600}

	for i := 0; i < b.N; i++ {
		img, err := CaptureRect(rect)
		if err != nil {
			b.Fatalf("CaptureRect failed: %v", err)
		}

		err = png.Encode(io.Discard, img)
		if err != nil {
			b.Fatalf("PNG encode failed: %v", err)
		}
	}
}

// BenchmarkCaptureDisplayByIndex benchmarks capturing a specific display.
func BenchmarkCaptureDisplayByIndex(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		img, err := CaptureDisplay(0)
		if err != nil {
			b.Fatalf("CaptureDisplay failed: %v", err)
		}
		if img == nil {
			b.Fatal("CaptureDisplay returned nil image")
		}
	}
}
