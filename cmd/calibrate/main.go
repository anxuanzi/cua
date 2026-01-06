// Command calibrate tests the coordinate mapping system visually.
package main

import (
	"fmt"
	"time"

	"github.com/anxuanzi/cua/pkg/screen"
	"github.com/go-vgo/robotgo"
)

func main() {
	fmt.Println("=== CUA Visual Calibration Test ===")
	fmt.Println("This tests the coordinate conversion system")
	fmt.Println("")

	// Get screen dimensions
	robotgoW, robotgoH := robotgo.GetScreenSize()

	primary, _ := screen.PrimaryDisplay()
	logicalW, logicalH := primary.Bounds.Width, primary.Bounds.Height

	img, _ := screen.CaptureDisplay(0)
	physW, physH := img.Bounds().Dx(), img.Bounds().Dy()

	scaleFactor := screen.ScaleFactor()

	fmt.Println("## Screen Dimensions")
	fmt.Printf("  Robotgo:  %dx%d\n", robotgoW, robotgoH)
	fmt.Printf("  Logical:  %dx%d\n", logicalW, logicalH)
	fmt.Printf("  Physical: %dx%d\n", physW, physH)
	fmt.Printf("  Scale:    %.2f\n", scaleFactor)

	// Calculate resized image dimensions (same as screenshot tool)
	maxDim := 1280
	var imgW, imgH int
	if physW > physH {
		imgW = maxDim
		imgH = int(float64(physH) * float64(maxDim) / float64(physW))
	} else {
		imgH = maxDim
		imgW = int(float64(physW) * float64(maxDim) / float64(physH))
	}

	// Set up dimension tracking (same as screenshot tool does)
	screen.SetLogicalScreenSize(logicalW, logicalH)
	screen.SetImageSize(imgW, imgH)

	fmt.Printf("  Image:    %dx%d (resized for Gemini)\n", imgW, imgH)

	fmt.Println("\n## Auto-Detection Logic")
	fmt.Printf("  Image > 1000px: %v (will use normalized for coords < 1000)\n", imgW > 1000 || imgH > 1000)

	fmt.Println("\n## Starting clicks in 3 seconds...")
	fmt.Println("Watch where the mouse moves and clicks!")
	time.Sleep(3 * time.Second)

	// Test 1: Direct robotgo clicks (baseline)
	fmt.Println("\n### Test 1: Direct robotgo clicks (baseline)")
	fmt.Println("These should land at exact positions:")
	directTests := []struct {
		name string
		x, y int
	}{
		{"Center", logicalW / 2, logicalH / 2},
		{"Top-left (100,100)", 100, 100},
		{"Top-right", logicalW - 100, 100},
	}

	for _, t := range directTests {
		fmt.Printf("  Clicking %s at logical(%d, %d)...\n", t.name, t.x, t.y)
		robotgo.Move(t.x, t.y)
		time.Sleep(300 * time.Millisecond)
		robotgo.Click()
		time.Sleep(700 * time.Millisecond)
	}

	// Test 2: ConvertModelCoord with normalized coordinates (primary use case)
	fmt.Println("\n### Test 2: ConvertModelCoord (simulating Gemini output)")
	fmt.Println("Testing normalized 0-1000 coordinates (what Gemini typically outputs):")
	convertTests := []struct {
		name     string
		inputX   int
		inputY   int
		expected string
	}{
		{"Center (500, 500)", 500, 500, "screen center"},
		{"Top-left (50, 50)", 50, 50, "near top-left corner"},
		{"Top-right (950, 50)", 950, 50, "near top-right corner"},
		{"Bottom-right (950, 950)", 950, 950, "near bottom-right corner"},
		{"Quarter (250, 250)", 250, 250, "25% from top-left"},
	}

	for _, t := range convertTests {
		screenX, screenY, mode := screen.ConvertModelCoord(t.inputX, t.inputY)
		fmt.Printf("  %s: input(%d,%d) → logical(%d,%d) [mode=%s] (%s)\n",
			t.name, t.inputX, t.inputY, screenX, screenY, mode, t.expected)
		robotgo.Move(screenX, screenY)
		time.Sleep(300 * time.Millisecond)
		robotgo.Click()
		time.Sleep(700 * time.Millisecond)
	}

	// Test 3: ConvertModelCoord with image pixel coordinates (coords > 1000)
	fmt.Println("\n### Test 3: ConvertModelCoord with coords > 1000 (image pixels)")
	fmt.Println("Testing image pixel coordinates (when coords exceed 1000):")
	largeCoordTests := []struct {
		name   string
		inputX int
		inputY int
	}{
		{"Large X (1200, 500)", 1200, 500},
	}

	for _, t := range largeCoordTests {
		screenX, screenY, mode := screen.ConvertModelCoord(t.inputX, t.inputY)
		fmt.Printf("  %s: input(%d,%d) → logical(%d,%d) [mode=%s]\n",
			t.name, t.inputX, t.inputY, screenX, screenY, mode)
		robotgo.Move(screenX, screenY)
		time.Sleep(300 * time.Millisecond)
		robotgo.Click()
		time.Sleep(700 * time.Millisecond)
	}

	// Summary
	fmt.Println("\n=== Calibration Summary ===")
	fmt.Println("Expected behavior:")
	fmt.Println("  - Coords 0-999 with image > 1000px → treated as NORMALIZED")
	fmt.Println("  - Coords > 1000 → treated as IMAGE PIXELS")
	fmt.Println("  - Center (500, 500) should click screen center")
	fmt.Println("  - (250, 250) should click 25% from top-left")
	fmt.Println("")
	fmt.Println("If clicks are landing at wrong positions:")
	fmt.Println("  1. Check if mode detection matches expected")
	fmt.Println("  2. Verify logical vs physical dimensions")
	fmt.Println("  3. Check scale factor calculation")
}
