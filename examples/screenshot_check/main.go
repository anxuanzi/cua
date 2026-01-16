// Screenshot and click verification test for 0-1000 normalized coordinates
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/anxuanzi/cua/internal/tools"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Screenshot & Click Verification (0-1000 Normalized) ===")
	fmt.Println("This test verifies the normalized coordinate system.")
	fmt.Println()

	// Step 1: Take a screenshot and save it
	fmt.Println("1. Taking screenshot...")
	screenshot := tools.NewScreenshotTool()
	result, err := screenshot.Execute(ctx, `{}`)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	// Parse the result to get dimensions
	var resp struct {
		ImageBase64  string `json:"image_base64"`
		ScreenWidth  int    `json:"screen_width"`
		ScreenHeight int    `json:"screen_height"`
		ImageWidth   int    `json:"image_width"`
		ImageHeight  int    `json:"image_height"`
	}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		fmt.Printf("ERROR parsing response: %v\n", err)
		return
	}

	fmt.Printf("   Screen (logical): %dx%d\n", resp.ScreenWidth, resp.ScreenHeight)
	fmt.Printf("   Image dimensions: %dx%d\n", resp.ImageWidth, resp.ImageHeight)

	// Save screenshot
	imgData, _ := base64.StdEncoding.DecodeString(resp.ImageBase64)
	screenshotPath := "/tmp/cua_coordinate_test.jpg"
	os.WriteFile(screenshotPath, imgData, 0644)
	fmt.Printf("   Screenshot saved to: %s\n", screenshotPath)

	// Step 2: Visual click test with 0-1000 normalized coordinates
	fmt.Println()
	fmt.Println("2. VISUAL CLICK TEST (0-1000 Normalized Coordinates)")
	fmt.Println("   Watch where the cursor moves!")
	fmt.Println()
	fmt.Println("   Coordinate System:")
	fmt.Println("   - (0, 0) = TOP-LEFT corner")
	fmt.Println("   - (1000, 1000) = BOTTOM-RIGHT corner")
	fmt.Println("   - (500, 500) = CENTER")
	fmt.Println()
	fmt.Println("   Starting in 3 seconds...")
	time.Sleep(3 * time.Second)

	moveTool := tools.NewMoveTool()

	// Test positions using 0-1000 normalized coordinates
	tests := []struct {
		name   string
		x, y   int
		expect string
	}{
		{"FAR LEFT", 50, 500, "Cursor should be on LEFT side of screen, vertically centered"},
		{"FAR RIGHT", 950, 500, "Cursor should be on RIGHT side of screen, vertically centered"},
		{"CENTER", 500, 500, "Cursor should be at CENTER of screen"},
		{"TOP-LEFT corner", 20, 20, "Cursor should be near APPLE MENU (top-left)"},
		{"TOP-RIGHT corner", 980, 20, "Cursor should be near CLOCK/DATE (top-right)"},
	}

	for i, test := range tests {
		fmt.Printf("\n   Test %d: Moving to %s (normalized x=%d, y=%d)\n", i+1, test.name, test.x, test.y)
		fmt.Printf("   Expected: %s\n", test.expect)

		result, err := moveTool.Execute(ctx, fmt.Sprintf(`{"x": %d, "y": %d}`, test.x, test.y))
		if err != nil {
			fmt.Printf("   ERROR: %v\n", err)
		} else {
			// Parse to show the screen coordinates it calculated
			var moveResp map[string]interface{}
			json.Unmarshal([]byte(result), &moveResp)
			if screenCoords, ok := moveResp["moved_to_screen"].(map[string]interface{}); ok {
				fmt.Printf("   Converted to screen: x=%.0f, y=%.0f\n", screenCoords["x"], screenCoords["y"])
			}
		}

		fmt.Println("   >>> Look at your screen - is the cursor where expected? <<<")
		time.Sleep(2 * time.Second)
	}

	// Step 3: Click test at center
	fmt.Println()
	fmt.Println("3. CLICK TEST at center (will click once)")
	fmt.Println("   Clicking at normalized (500, 500)...")
	click := tools.NewClickTool()
	click.Execute(ctx, `{"x": 500, "y": 500}`)

	fmt.Println()
	fmt.Println("=== Test Complete ===")
	fmt.Println()
	fmt.Println("VERIFICATION CHECKLIST:")
	fmt.Println("  [ ] When moving to FAR LEFT (x=50), was cursor on LEFT side?")
	fmt.Println("  [ ] When moving to FAR RIGHT (x=950), was cursor on RIGHT side?")
	fmt.Println("  [ ] When moving to TOP-LEFT (20,20), was cursor near Apple menu?")
	fmt.Println("  [ ] When moving to TOP-RIGHT (980,20), was cursor near the clock?")
	fmt.Println()
	fmt.Println("If LEFT/RIGHT are swapped, coordinates are mirrored!")
	fmt.Println("If TOP/BOTTOM are swapped, Y-axis is inverted!")
	fmt.Println()

	// Open the screenshot for reference
	fmt.Println("Opening screenshot for reference...")
	exec.Command("open", screenshotPath).Run()
}
