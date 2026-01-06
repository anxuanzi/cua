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

func TestScaleFactor(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux CI - may not have display")
	}

	scale := ScaleFactor()
	t.Logf("Display scale factor: %.2f", scale)

	// Scale should be between 1.0 and 3.0 for reasonable displays
	if scale < 1.0 || scale > 3.0 {
		t.Errorf("ScaleFactor() = %.2f, want between 1.0 and 3.0", scale)
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

	// Capture a small region (in logical coordinates)
	logicalWidth, logicalHeight := 100, 100
	img, err := CaptureRect(Rect{X: 0, Y: 0, Width: logicalWidth, Height: logicalHeight})
	if err != nil {
		t.Fatalf("CaptureRect() error: %v", err)
	}

	if img == nil {
		t.Fatal("CaptureRect() returned nil image")
	}

	bounds := img.Bounds()
	scale := ScaleFactor()

	// Captured image is at physical resolution (logical * scale)
	expectedWidth := int(float64(logicalWidth) * scale)
	expectedHeight := int(float64(logicalHeight) * scale)

	t.Logf("CaptureRect(%dx%d logical) = %dx%d physical (scale=%.2f)",
		logicalWidth, logicalHeight, bounds.Dx(), bounds.Dy(), scale)

	if bounds.Dx() != expectedWidth || bounds.Dy() != expectedHeight {
		t.Errorf("CaptureRect() dimensions = %dx%d, want %dx%d (at scale %.2f)",
			bounds.Dx(), bounds.Dy(), expectedWidth, expectedHeight, scale)
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

	logicalWidth, logicalHeight := 50, 50
	img, err := Capture(10, 10, logicalWidth, logicalHeight)
	if err != nil {
		t.Fatalf("Capture() error: %v", err)
	}

	if img == nil {
		t.Fatal("Capture() returned nil image")
	}

	bounds := img.Bounds()
	scale := ScaleFactor()

	// Captured image is at physical resolution
	expectedWidth := int(float64(logicalWidth) * scale)
	expectedHeight := int(float64(logicalHeight) * scale)

	t.Logf("Capture(%dx%d logical) = %dx%d physical (scale=%.2f)",
		logicalWidth, logicalHeight, bounds.Dx(), bounds.Dy(), scale)

	if bounds.Dx() != expectedWidth || bounds.Dy() != expectedHeight {
		t.Errorf("Capture() dimensions = %dx%d, want %dx%d (at scale %.2f)",
			bounds.Dx(), bounds.Dy(), expectedWidth, expectedHeight, scale)
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

	logicalWidth, logicalHeight := 100, 100
	img, err := Capture(0, 0, logicalWidth, logicalHeight)
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
	scale := ScaleFactor()
	expectedWidth := int(float64(logicalWidth) * scale)
	expectedHeight := int(float64(logicalHeight) * scale)

	if decodedBounds.Dx() != expectedWidth || decodedBounds.Dy() != expectedHeight {
		t.Errorf("Decoded image dimensions = %dx%d, want %dx%d (at scale %.2f)",
			decodedBounds.Dx(), decodedBounds.Dy(), expectedWidth, expectedHeight, scale)
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

	// Should be at least as big as primary display (accounting for scale)
	primary, _ := PrimaryDisplay()
	scale := ScaleFactor()
	expectedMinWidth := int(float64(primary.Bounds.Width) * scale)
	expectedMinHeight := int(float64(primary.Bounds.Height) * scale)

	if bounds.Dx() < expectedMinWidth || bounds.Dy() < expectedMinHeight {
		t.Errorf("CaptureAll() smaller than primary display: got %dx%d, expected at least %dx%d (scale %.2f)",
			bounds.Dx(), bounds.Dy(), expectedMinWidth, expectedMinHeight, scale)
	}
}

func TestPhysicalToLogical(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux CI - may not have display")
	}

	scale := ScaleFactor()
	t.Logf("Scale factor: %.2f", scale)

	// Test coordinate conversion
	physicalX, physicalY := 200, 400
	logicalX, logicalY := PhysicalToLogical(physicalX, physicalY)

	expectedLogicalX := int(float64(physicalX) / scale)
	expectedLogicalY := int(float64(physicalY) / scale)

	if logicalX != expectedLogicalX || logicalY != expectedLogicalY {
		t.Errorf("PhysicalToLogical(%d, %d) = (%d, %d), want (%d, %d)",
			physicalX, physicalY, logicalX, logicalY, expectedLogicalX, expectedLogicalY)
	}

	t.Logf("PhysicalToLogical(%d, %d) = (%d, %d)", physicalX, physicalY, logicalX, logicalY)
}

func TestLogicalToPhysical(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux CI - may not have display")
	}

	scale := ScaleFactor()
	t.Logf("Scale factor: %.2f", scale)

	// Test coordinate conversion
	logicalX, logicalY := 100, 200
	physicalX, physicalY := LogicalToPhysical(logicalX, logicalY)

	expectedPhysicalX := int(float64(logicalX) * scale)
	expectedPhysicalY := int(float64(logicalY) * scale)

	if physicalX != expectedPhysicalX || physicalY != expectedPhysicalY {
		t.Errorf("LogicalToPhysical(%d, %d) = (%d, %d), want (%d, %d)",
			logicalX, logicalY, physicalX, physicalY, expectedPhysicalX, expectedPhysicalY)
	}

	t.Logf("LogicalToPhysical(%d, %d) = (%d, %d)", logicalX, logicalY, physicalX, physicalY)
}

func TestCoordinateRoundTrip(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping on Linux CI - may not have display")
	}

	// Test that converting logical -> physical -> logical gives same result
	// (for coordinates that are exact multiples of scale)
	scale := ScaleFactor()
	if scale != float64(int(scale)) {
		t.Skip("Skipping round-trip test for non-integer scale factors")
	}

	logicalX, logicalY := 100, 200
	physicalX, physicalY := LogicalToPhysical(logicalX, logicalY)
	backLogicalX, backLogicalY := PhysicalToLogical(physicalX, physicalY)

	if backLogicalX != logicalX || backLogicalY != logicalY {
		t.Errorf("Round-trip failed: (%d,%d) -> (%d,%d) -> (%d,%d)",
			logicalX, logicalY, physicalX, physicalY, backLogicalX, backLogicalY)
	}
}

func TestDenormalizeCoord(t *testing.T) {
	// Test Gemini's normalized 0-1000 coordinate system
	// Gemini outputs coords in 0-999 range regardless of screen size

	tests := []struct {
		name           string
		screenW        int
		screenH        int
		modelX, modelY int
		wantX, wantY   int
	}{
		{
			name:    "center of 1440x900 screen",
			screenW: 1440, screenH: 900,
			modelX: 500, modelY: 500,
			wantX: 720, wantY: 450,
		},
		{
			name:    "top-left corner",
			screenW: 1920, screenH: 1080,
			modelX: 0, modelY: 0,
			wantX: 0, wantY: 0,
		},
		{
			name:    "bottom-right corner (999)",
			screenW: 1920, screenH: 1080,
			modelX: 999, modelY: 999,
			wantX: 1918, wantY: 1079, // Clamped to max-1
		},
		{
			name:    "quarter positions",
			screenW: 1000, screenH: 1000,
			modelX: 250, modelY: 750,
			wantX: 250, wantY: 750,
		},
		{
			name:    "MacBook-like resolution",
			screenW: 1512, screenH: 982,
			modelX: 500, modelY: 300,
			wantX: 756, wantY: 295, // 500/1000*1512=756, 300/1000*982=294.6≈295
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set the screen size
			SetLogicalScreenSize(tc.screenW, tc.screenH)

			// Verify it was stored
			gotW, gotH := LogicalScreenSize()
			if gotW != tc.screenW || gotH != tc.screenH {
				t.Errorf("LogicalScreenSize() = (%d, %d), want (%d, %d)",
					gotW, gotH, tc.screenW, tc.screenH)
			}

			// Test denormalization
			gotX, gotY := DenormalizeCoord(tc.modelX, tc.modelY)

			if gotX != tc.wantX || gotY != tc.wantY {
				t.Errorf("DenormalizeCoord(%d, %d) with screen %dx%d = (%d, %d), want (%d, %d)",
					tc.modelX, tc.modelY, tc.screenW, tc.screenH, gotX, gotY, tc.wantX, tc.wantY)
			}

			t.Logf("DenormalizeCoord(%d, %d) with screen %dx%d = (%d, %d)",
				tc.modelX, tc.modelY, tc.screenW, tc.screenH, gotX, gotY)
		})
	}
}

func TestDenormalizeCoordEdgeCases(t *testing.T) {
	SetLogicalScreenSize(1920, 1080)

	// Test clamping at boundaries
	tests := []struct {
		name        string
		modelX      int
		modelY      int
		wantClamped bool
	}{
		{"negative X should clamp to 0", -100, 500, true},
		{"negative Y should clamp to 0", 500, -100, true},
		{"X > 1000 should clamp to max", 1500, 500, true},
		{"Y > 1000 should clamp to max", 500, 1500, true},
		{"normal values", 500, 500, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x, y := DenormalizeCoord(tc.modelX, tc.modelY)

			// Check that values are within bounds
			screenW, screenH := LogicalScreenSize()
			if x < 0 || x >= screenW {
				if tc.wantClamped {
					// Verify clamping worked
					if x < 0 || x >= screenW {
						t.Errorf("Expected clamped X, got %d (screen width: %d)", x, screenW)
					}
				} else {
					t.Errorf("Unexpected out-of-bounds X: %d", x)
				}
			}
			if y < 0 || y >= screenH {
				if tc.wantClamped {
					// Verify clamping worked
					if y < 0 || y >= screenH {
						t.Errorf("Expected clamped Y, got %d (screen height: %d)", y, screenH)
					}
				} else {
					t.Errorf("Unexpected out-of-bounds Y: %d", y)
				}
			}

			t.Logf("DenormalizeCoord(%d, %d) = (%d, %d) clamped=%v", tc.modelX, tc.modelY, x, y, tc.wantClamped)
		})
	}
}
