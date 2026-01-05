// Package main demonstrates low-level CUA APIs for direct control.
//
// This example shows how to use CUA's direct input and element APIs
// without going through the AI agent. This is useful for:
//   - Building custom automation scripts
//   - Testing specific UI interactions
//   - Integrating with existing automation workflows
//   - Cases where AI-based task execution isn't needed
//
// Prerequisites:
//   - macOS: Grant Accessibility and Screen Recording permissions
//   - Windows: May need to run as Administrator for some applications
//
// Run with:
//
//	go run ./examples/low_level
package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/anxuanzi/cua"
)

func main() {
	fmt.Println("=== CUA Low-Level API Example ===")
	fmt.Println()

	// Demonstrate each low-level API
	demoScreenInfo()
	demoScreenCapture()
	demoElementFinding()
	demoMouseControl()
	demoKeyboardControl()
}

// demoScreenInfo shows how to get screen information.
func demoScreenInfo() {
	fmt.Println("--- Screen Information ---")

	width, height, err := cua.ScreenSize()
	if err != nil {
		log.Printf("Error getting screen size: %v", err)
		return
	}

	fmt.Printf("Primary screen size: %dx%d pixels\n", width, height)
	fmt.Println()
}

// demoScreenCapture shows how to capture screenshots.
func demoScreenCapture() {
	fmt.Println("--- Screen Capture ---")

	// Capture full screen
	img, err := cua.CaptureScreen()
	if err != nil {
		log.Printf("Error capturing screen: %v", err)
		return
	}

	bounds := img.Bounds()
	fmt.Printf("Captured full screen: %dx%d\n", bounds.Dx(), bounds.Dy())

	// Save to file
	filename := "screenshot_full.png"
	if err := saveImage(img, filename); err != nil {
		log.Printf("Error saving screenshot: %v", err)
	} else {
		fmt.Printf("Saved to: %s\n", filename)
	}

	// Capture a region
	rect := cua.Rect{X: 0, Y: 0, Width: 200, Height: 200}
	regionImg, err := cua.CaptureRect(rect)
	if err != nil {
		log.Printf("Error capturing region: %v", err)
	} else {
		regionFilename := "screenshot_region.png"
		if err := saveImage(regionImg, regionFilename); err != nil {
			log.Printf("Error saving region screenshot: %v", err)
		} else {
			fmt.Printf("Captured region (0,0,200,200): saved to %s\n", regionFilename)
		}
	}

	fmt.Println()
}

// demoElementFinding shows how to find UI elements.
func demoElementFinding() {
	fmt.Println("--- Element Finding ---")

	// Find the focused application
	app, err := cua.FocusedApplication()
	if err != nil {
		log.Printf("Error finding focused app: %v", err)
	} else {
		fmt.Printf("Focused application: %s (Role: %s)\n", app.Name, app.Role)
	}

	// Find the currently focused element
	focused, err := cua.FocusedElement()
	if err != nil {
		log.Printf("Error finding focused element: %v", err)
	} else {
		fmt.Printf("Focused element: %s (Role: %s)\n", focused.Name, focused.Role)
	}

	// Find all buttons
	buttons, err := cua.FindElements(cua.ByRole(cua.RoleButton))
	if err != nil {
		log.Printf("Error finding buttons: %v", err)
	} else {
		fmt.Printf("Found %d buttons\n", len(buttons))
		// Print first 5 buttons
		for i, btn := range buttons {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(buttons)-5)
				break
			}
			fmt.Printf("  - %s at (%d, %d)\n", btn.Name, btn.Bounds.X, btn.Bounds.Y)
		}
	}

	// Find elements with combined selectors
	// Example: Find enabled buttons with "OK" in the name
	okButtons, err := cua.FindElements(cua.And(
		cua.ByRole(cua.RoleButton),
		cua.ByEnabled(),
		cua.ByNameContains("OK"),
	))
	if err != nil {
		log.Printf("Error finding OK buttons: %v", err)
	} else {
		fmt.Printf("Found %d enabled buttons containing 'OK'\n", len(okButtons))
	}

	// Find text fields
	textFields, err := cua.FindElements(cua.ByRole(cua.RoleTextField))
	if err != nil {
		log.Printf("Error finding text fields: %v", err)
	} else {
		fmt.Printf("Found %d text fields\n", len(textFields))
	}

	fmt.Println()
}

// demoMouseControl shows how to control the mouse.
func demoMouseControl() {
	fmt.Println("--- Mouse Control ---")
	fmt.Println("(Demonstrating mouse movement - will move cursor)")
	fmt.Println()

	// Get current screen size for safe movement
	width, height, err := cua.ScreenSize()
	if err != nil {
		log.Printf("Error getting screen size: %v", err)
		return
	}

	// Calculate safe center position
	centerX := width / 2
	centerY := height / 2

	// Move mouse to center
	fmt.Printf("Moving mouse to center (%d, %d)\n", centerX, centerY)
	if err := cua.MoveMouse(centerX, centerY); err != nil {
		log.Printf("Error moving mouse: %v", err)
		return
	}
	time.Sleep(500 * time.Millisecond)

	// Demonstrate different click types (commented out to avoid unintended clicks)
	fmt.Println()
	fmt.Println("Available click methods (not executed to avoid unintended actions):")
	fmt.Println("  cua.Click(x, y)       - Left click")
	fmt.Println("  cua.DoubleClick(x, y) - Double click")
	fmt.Println("  cua.RightClick(x, y)  - Right click (context menu)")
	fmt.Println("  cua.MiddleClick(x, y) - Middle click")
	fmt.Println("  cua.DragMouse(x1, y1, x2, y2) - Drag from point to point")
	fmt.Println("  cua.Scroll(x, y, dx, dy)     - Scroll at position")

	// Example of how clicks would be done:
	// if err := cua.Click(centerX, centerY); err != nil {
	//     log.Printf("Error clicking: %v", err)
	// }

	fmt.Println()
}

// demoKeyboardControl shows how to control the keyboard.
func demoKeyboardControl() {
	fmt.Println("--- Keyboard Control ---")
	fmt.Println("(Keyboard methods shown but not executed)")
	fmt.Println()

	fmt.Println("Available keyboard methods:")
	fmt.Println("  cua.TypeText(\"hello\")           - Type text string")
	fmt.Println("  cua.KeyPress(cua.KeyEnter)       - Press Enter")
	fmt.Println("  cua.KeyPress(cua.KeyTab)         - Press Tab")
	fmt.Println("  cua.KeyPress(cua.KeyEscape)      - Press Escape")
	fmt.Println()
	fmt.Println("Keyboard shortcuts with modifiers:")
	fmt.Println("  cua.KeyPress(cua.KeyA, cua.ModCmd)              - Cmd+A (Select All)")
	fmt.Println("  cua.KeyPress(cua.KeyC, cua.ModCmd)              - Cmd+C (Copy)")
	fmt.Println("  cua.KeyPress(cua.KeyV, cua.ModCmd)              - Cmd+V (Paste)")
	fmt.Println("  cua.KeyPress(cua.KeyS, cua.ModCmd, cua.ModShift) - Cmd+Shift+S (Save As)")
	fmt.Println()
	fmt.Println("Arrow keys and navigation:")
	fmt.Println("  cua.KeyPress(cua.KeyUp)          - Up arrow")
	fmt.Println("  cua.KeyPress(cua.KeyDown)        - Down arrow")
	fmt.Println("  cua.KeyPress(cua.KeyLeft)        - Left arrow")
	fmt.Println("  cua.KeyPress(cua.KeyRight)       - Right arrow")
	fmt.Println("  cua.KeyPress(cua.KeyHome)        - Home")
	fmt.Println("  cua.KeyPress(cua.KeyEnd)         - End")
	fmt.Println("  cua.KeyPress(cua.KeyPageUp)      - Page Up")
	fmt.Println("  cua.KeyPress(cua.KeyPageDown)    - Page Down")
	fmt.Println()
	fmt.Println("Hold and release keys:")
	fmt.Println("  cua.HoldKey(cua.KeyShift)        - Hold Shift")
	fmt.Println("  cua.ReleaseKey(cua.KeyShift)     - Release Shift")

	// Example of how keyboard input would be done:
	// if err := cua.TypeText("Hello, World!"); err != nil {
	//     log.Printf("Error typing: %v", err)
	// }
	// if err := cua.KeyPress(cua.KeyEnter); err != nil {
	//     log.Printf("Error pressing Enter: %v", err)
	// }

	fmt.Println()
}

// saveImage saves an image.Image to a PNG file.
func saveImage(img image.Image, filename string) error {
	if img == nil {
		return fmt.Errorf("nil image")
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	return png.Encode(f, img)
}
