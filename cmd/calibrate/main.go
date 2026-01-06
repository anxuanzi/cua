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
	fmt.Println("This will click using robotgo on the calibration page")
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

	// Set logical screen size for denormalization
	screen.SetLogicalScreenSize(logicalW, logicalH)

	fmt.Println("\n## Starting clicks in 3 seconds...")
	fmt.Println("Make sure the calibration page is visible!")
	time.Sleep(3 * time.Second)

	// Test 1: Direct robotgo clicks at absolute positions
	fmt.Println("\n### Test 1: Direct robotgo clicks (logical coordinates)")
	directTests := []struct {
		name string
		x, y int
	}{
		{"Center", logicalW / 2, logicalH / 2},
		{"Top-left (100,100)", 100, 100},
		{"Top-right", logicalW - 100, 100},
		{"Bottom-right", logicalW - 100, logicalH - 100},
	}

	for _, t := range directTests {
		fmt.Printf("  Clicking %s at (%d, %d)...\n", t.name, t.x, t.y)
		robotgo.Move(t.x, t.y)
		time.Sleep(300 * time.Millisecond)
		robotgo.Click()
		time.Sleep(700 * time.Millisecond)
	}

	fmt.Println("\n### Test 2: Denormalized clicks (0-1000 → logical)")
	// These simulate what Gemini would output
	normalizedTests := []struct {
		name         string
		normX, normY int
	}{
		{"Norm center (500,500)", 500, 500},
		{"Norm TL (50,50)", 50, 50},
		{"Norm TR (950,50)", 950, 50},
		{"Norm BR (950,950)", 950, 950},
		{"Norm BL (50,950)", 50, 950},
	}

	for _, t := range normalizedTests {
		logX, logY := screen.DenormalizeCoord(t.normX, t.normY)
		fmt.Printf("  %s: norm(%d,%d) → logical(%d,%d)\n", t.name, t.normX, t.normY, logX, logY)
		robotgo.Move(logX, logY)
		time.Sleep(300 * time.Millisecond)
		robotgo.Click()
		time.Sleep(700 * time.Millisecond)
	}

	fmt.Println("\n### Test 3: Image coordinate conversion")
	// Simulate Gemini outputting coordinates relative to a resized image
	maxDim := 1280
	var imgW, imgH int
	if physW > physH {
		imgW = maxDim
		imgH = int(float64(physH) * float64(maxDim) / float64(physW))
	} else {
		imgH = maxDim
		imgW = int(float64(physW) * float64(maxDim) / float64(physH))
	}

	fmt.Printf("  Gemini sees image: %dx%d\n", imgW, imgH)
	convX := float64(logicalW) / float64(imgW)
	convY := float64(logicalH) / float64(imgH)
	fmt.Printf("  Conversion factors: X=%.4f, Y=%.4f\n", convX, convY)

	imageTests := []struct {
		name       string
		imgX, imgY int
	}{
		{"Img center", imgW / 2, imgH / 2},
		{"Img TL (64, 42)", 64, 42},
		{"Img TR", imgW - 64, 42},
		{"Img BR", imgW - 64, imgH - 42},
	}

	for _, t := range imageTests {
		screenX := int(float64(t.imgX) * convX)
		screenY := int(float64(t.imgY) * convY)
		fmt.Printf("  %s: img(%d,%d) → screen(%d,%d)\n", t.name, t.imgX, t.imgY, screenX, screenY)
		robotgo.Move(screenX, screenY)
		time.Sleep(300 * time.Millisecond)
		robotgo.Click()
		time.Sleep(700 * time.Millisecond)
	}

	fmt.Println("\n=== Calibration complete ===")
	fmt.Println("Check the calibration page for red dots!")
	fmt.Println("Compare where dots landed vs expected positions.")
}
