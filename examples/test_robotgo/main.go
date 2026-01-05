// Package main tests that robotgo actually works on macOS.
// This is a manual verification script.
//
// Run with:
//
//	go run ./examples/test_robotgo
package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	fmt.Println("=== RobotGo Test ===")
	fmt.Printf("OS: %s\n", runtime.GOOS)
	fmt.Printf("Arch: %s\n", runtime.GOARCH)
	fmt.Println()

	// Test 1: Get mouse location
	fmt.Println("[Test 1] Getting mouse location...")
	x, y := robotgo.Location()
	fmt.Printf("✓ Mouse at: (%d, %d)\n", x, y)
	fmt.Println()

	// Test 2: Move mouse (small movement)
	fmt.Println("[Test 2] Moving mouse to (100, 100)...")
	robotgo.Move(100, 100)
	time.Sleep(100 * time.Millisecond)
	newX, newY := robotgo.Location()
	if newX == 100 && newY == 100 {
		fmt.Println("✓ Mouse moved successfully!")
	} else {
		fmt.Printf("✗ Mouse at (%d, %d) - expected (100, 100)\n", newX, newY)
		fmt.Println("  NOTE: This may fail if Accessibility permission is not granted.")
	}
	fmt.Println()

	// Test 3: Get screen size
	fmt.Println("[Test 3] Getting screen size...")
	width, height := robotgo.GetScreenSize()
	fmt.Printf("✓ Screen size: %dx%d\n", width, height)
	fmt.Println()

	// Test 4: Type text (optional - requires user confirmation)
	if len(os.Args) > 1 && os.Args[1] == "--type" {
		fmt.Println("[Test 4] Typing test...")
		fmt.Println("Open a text editor and focus it. You have 3 seconds...")
		time.Sleep(3 * time.Second)
		robotgo.TypeStr("Hello from CUA!")
		fmt.Println("✓ Typed text (check your text editor)")
	} else {
		fmt.Println("[Test 4] SKIPPED - pass --type to test typing")
	}
	fmt.Println()

	// Test 5: Key tap (optional - requires user confirmation)
	if len(os.Args) > 1 && os.Args[1] == "--key" {
		fmt.Println("[Test 5] Key tap test...")
		fmt.Println("This will press Cmd+Space (Spotlight). You have 3 seconds...")
		time.Sleep(3 * time.Second)

		// On macOS, Cmd+Space opens Spotlight
		if runtime.GOOS == "darwin" {
			robotgo.KeyTap("space", "cmd")
			fmt.Println("✓ Pressed Cmd+Space (check if Spotlight opened)")
		} else if runtime.GOOS == "windows" {
			robotgo.KeyTap("lwin") // Windows key
			fmt.Println("✓ Pressed Windows key (check if Start Menu opened)")
		}
	} else {
		fmt.Println("[Test 5] SKIPPED - pass --key to test key tap")
	}
	fmt.Println()

	// Test 6: Screenshot (capture region)
	fmt.Println("[Test 6] Taking screenshot...")
	bitmap := robotgo.CaptureScreen(0, 0, 100, 100)
	if bitmap != nil {
		fmt.Println("✓ Captured 100x100 region successfully")
		// Convert to image to verify it worked
		img := robotgo.ToImage(bitmap)
		if img != nil {
			bounds := img.Bounds()
			fmt.Printf("  Image dimensions: %dx%d\n", bounds.Dx(), bounds.Dy())
		}
		robotgo.FreeBitmap(bitmap)
	} else {
		fmt.Println("✗ Failed to capture screenshot")
		fmt.Println("  NOTE: This may fail if Screen Recording permission is not granted.")
	}
	fmt.Println()

	fmt.Println("=== Test Complete ===")
	fmt.Println()
	fmt.Println("If tests failed, ensure you have granted:")
	fmt.Println("1. Accessibility permission (System Settings > Privacy & Security > Accessibility)")
	fmt.Println("2. Screen Recording permission (System Settings > Privacy & Security > Screen Recording)")
}
