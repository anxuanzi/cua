package screen

import (
	"bytes"
	"image/png"
	"runtime"
	"testing"
)

func TestNumDisplays(t *testing.T) {
	n := NumDisplays()
	if n < 1 {
		t.Errorf("NumDisplays() = %d, want at least 1", n)
	}
	t.Logf("Found %d displays", n)
}

func TestDisplays(t *testing.T) {
	displays := Displays()
	if len(displays) < 1 {
		t.Fatal("Displays() returned empty list, want at least 1")
	}

	for _, d := range displays {
		t.Logf("Display %d: %dx%d at (%d,%d) primary=%v",
			d.Index, d.Bounds.Width, d.Bounds.Height, d.Bounds.X, d.Bounds.Y, d.Primary)

		if d.Bounds.Width <= 0 || d.Bounds.Height <= 0 {
			t.Errorf("Display %d has invalid dimensions: %dx%d",
				d.Index, d.Bounds.Width, d.Bounds.Height)
		}
	}

	// First display should be primary
	if !displays[0].Primary {
		t.Error("First display should be marked as primary")
	}
}

func TestPrimaryDisplay(t *testing.T) {
	primary, err := PrimaryDisplay()
	if err != nil {
		t.Fatalf("PrimaryDisplay() error: %v", err)
	}

	if primary.Index != 0 {
		t.Errorf("PrimaryDisplay().Index = %d, want 0", primary.Index)
	}

	if !primary.Primary {
		t.Error("PrimaryDisplay().Primary should be true")
	}

	if primary.Bounds.Width <= 0 || primary.Bounds.Height <= 0 {
		t.Errorf("PrimaryDisplay() has invalid dimensions: %dx%d",
			primary.Bounds.Width, primary.Bounds.Height)
	}

	t.Logf("Primary display: %dx%d", primary.Bounds.Width, primary.Bounds.Height)
}

func TestGetDisplayBounds(t *testing.T) {
	// Valid display index
	bounds, err := GetDisplayBounds(0)
	if err != nil {
		t.Fatalf("GetDisplayBounds(0) error: %v", err)
	}

	if bounds.Width <= 0 || bounds.Height <= 0 {
		t.Errorf("GetDisplayBounds(0) invalid dimensions: %dx%d", bounds.Width, bounds.Height)
	}

	// Invalid display index
	_, err = GetDisplayBounds(-1)
	if err != ErrInvalidDisplay {
		t.Errorf("GetDisplayBounds(-1) error = %v, want ErrInvalidDisplay", err)
	}

	_, err = GetDisplayBounds(999)
	if err != ErrInvalidDisplay {
		t.Errorf("GetDisplayBounds(999) error = %v, want ErrInvalidDisplay", err)
	}
}

func TestCaptureDisplay(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping screenshot test in short mode")
	}

	// Skip on CI where screen capture may not work
	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux CI - may not have display")
	}

	img, err := CaptureDisplay(0)
	if err != nil {
		t.Fatalf("CaptureDisplay(0) error: %v", err)
	}

	if img == nil {
		t.Fatal("CaptureDisplay(0) returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Errorf("Captured image has invalid dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}

	t.Logf("Captured display: %dx%d", bounds.Dx(), bounds.Dy())

	// Verify it's a valid image by encoding to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Errorf("Failed to encode captured image: %v", err)
	}

	t.Logf("PNG size: %d bytes", buf.Len())
}

func TestCapturePrimary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping screenshot test in short mode")
	}

	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux CI - may not have display")
	}

	img, err := CapturePrimary()
	if err != nil {
		t.Fatalf("CapturePrimary() error: %v", err)
	}

	if img == nil {
		t.Fatal("CapturePrimary() returned nil image")
	}

	bounds := img.Bounds()
	t.Logf("Primary display capture: %dx%d", bounds.Dx(), bounds.Dy())
}

func TestCaptureRect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping screenshot test in short mode")
	}

	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux CI - may not have display")
	}

	// Capture a small region
	img, err := CaptureRect(Rect{X: 0, Y: 0, Width: 100, Height: 100})
	if err != nil {
		t.Fatalf("CaptureRect() error: %v", err)
	}

	if img == nil {
		t.Fatal("CaptureRect() returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("CaptureRect() dimensions = %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}
}

func TestCaptureRectInvalid(t *testing.T) {
	tests := []struct {
		name string
		rect Rect
	}{
		{"zero width", Rect{X: 0, Y: 0, Width: 0, Height: 100}},
		{"zero height", Rect{X: 0, Y: 0, Width: 100, Height: 0}},
		{"negative width", Rect{X: 0, Y: 0, Width: -100, Height: 100}},
		{"negative height", Rect{X: 0, Y: 0, Width: 100, Height: -100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CaptureRect(tt.rect)
			if err != ErrInvalidRect {
				t.Errorf("CaptureRect(%v) error = %v, want ErrInvalidRect", tt.rect, err)
			}
		})
	}
}

func TestCapture(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping screenshot test in short mode")
	}

	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux CI - may not have display")
	}

	img, err := Capture(10, 10, 50, 50)
	if err != nil {
		t.Fatalf("Capture() error: %v", err)
	}

	if img == nil {
		t.Fatal("Capture() returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Capture() dimensions = %dx%d, want 50x50", bounds.Dx(), bounds.Dy())
	}
}

func TestCaptureDisplayInvalid(t *testing.T) {
	_, err := CaptureDisplay(-1)
	if err != ErrInvalidDisplay {
		t.Errorf("CaptureDisplay(-1) error = %v, want ErrInvalidDisplay", err)
	}

	_, err = CaptureDisplay(999)
	if err != ErrInvalidDisplay {
		t.Errorf("CaptureDisplay(999) error = %v, want ErrInvalidDisplay", err)
	}
}

func TestSavePNG(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping screenshot test in short mode")
	}

	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux CI - may not have display")
	}

	img, err := Capture(0, 0, 100, 100)
	if err != nil {
		t.Fatalf("Capture() error: %v", err)
	}

	var buf bytes.Buffer
	if err := SavePNG(&buf, img); err != nil {
		t.Fatalf("SavePNG() error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("SavePNG() wrote 0 bytes")
	}

	// Verify it's a valid PNG
	decoded, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to decode saved PNG: %v", err)
	}

	decodedBounds := decoded.Bounds()
	if decodedBounds.Dx() != 100 || decodedBounds.Dy() != 100 {
		t.Errorf("Decoded image dimensions = %dx%d, want 100x100",
			decodedBounds.Dx(), decodedBounds.Dy())
	}
}

func TestCaptureAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping screenshot test in short mode")
	}

	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux CI - may not have display")
	}

	img, err := CaptureAll()
	if err != nil {
		t.Fatalf("CaptureAll() error: %v", err)
	}

	if img == nil {
		t.Fatal("CaptureAll() returned nil image")
	}

	bounds := img.Bounds()
	t.Logf("CaptureAll() dimensions: %dx%d", bounds.Dx(), bounds.Dy())

	// Should be at least as big as primary display
	primary, _ := PrimaryDisplay()
	if bounds.Dx() < primary.Bounds.Width || bounds.Dy() < primary.Bounds.Height {
		t.Errorf("CaptureAll() smaller than primary display: got %dx%d, primary is %dx%d",
			bounds.Dx(), bounds.Dy(), primary.Bounds.Width, primary.Bounds.Height)
	}
}
