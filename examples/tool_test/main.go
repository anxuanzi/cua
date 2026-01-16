// Tool test - tests individual tools directly without LLM
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/anxuanzi/cua/internal/tools"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== CUA Tool Test ===")
	fmt.Println("This will test each tool directly.")
	fmt.Println("Watch your screen to verify actions are happening.")
	fmt.Println()

	// Test: App List (no permissions needed)
	fmt.Println("=== APP LIST TEST ===")
	fmt.Println("Listing installed applications...")
	appList := tools.NewAppListTool()
	result, err := appList.Execute(ctx, `{"limit": 10}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   First 10 apps: %s\n", result)
	}

	// Test: Search for Calculator
	fmt.Println("\nSearching for 'Calculator'...")
	result, err = appList.Execute(ctx, `{"search": "calculator"}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Search result: %s\n", result)
	}

	// Test: App Launch
	fmt.Println("\n=== APP LAUNCH TEST ===")
	fmt.Println("Launching Calculator using app_launch tool...")
	appLaunch := tools.NewAppLaunchTool()
	result, err = appLaunch.Execute(ctx, `{"app_name": "Calculator"}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", result)
	}
	fmt.Println("   Did Calculator open? (This bypasses Spotlight entirely!)")
	time.Sleep(2 * time.Second)

	// Test: Screen info
	fmt.Println("\n=== SCREEN INFO TEST ===")
	fmt.Println("Getting screen dimensions...")
	screenInfo := tools.NewScreenInfoTool()
	result, err = screenInfo.Execute(ctx, `{"screen_index": 0}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Screen info: %s\n", result)
	}

	// Test 1: Screen capture (should work if permissions are OK)
	fmt.Println("\n=== SCREEN CAPTURE TEST ===")
	fmt.Println("Testing screen_capture...")
	screenshot := tools.NewScreenshotTool()
	result, err = screenshot.Execute(ctx, `{}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		// Parse and display key info (not full base64)
		fmt.Printf("   OK - Screenshot captured\n")
		// The result contains JSON with screen_width, screen_height, image_width, image_height, scale_factor
		fmt.Printf("   Metadata: %s...\n", truncateJSON(result, 300))
	}
	fmt.Println("\n   NOTE: screen_width/screen_height = LOGICAL dimensions (for mouse coords)")
	fmt.Println("         image_width/image_height = actual image dimensions sent to LLM")
	fmt.Println("         scale_factor = detected Retina scale (capture vs logical)")

	// Test: Mouse movement demo - WATCH YOUR CURSOR!
	fmt.Println("\n=== MOUSE MOVEMENT TEST ===")
	fmt.Println("Watch your cursor move to different positions!")
	fmt.Println("Starting in 2 seconds...")
	time.Sleep(2 * time.Second)

	moveTool := tools.NewMoveTool()

	// Move to corners and center
	positions := []struct {
		name string
		x, y int
	}{
		{"TOP-LEFT (0, 0)", 50, 50},
		{"TOP-RIGHT (1000, 0)", 950, 50},
		{"BOTTOM-RIGHT (1000, 1000)", 950, 950},
		{"BOTTOM-LEFT (0, 1000)", 50, 950},
		{"CENTER (500, 500)", 500, 500},
	}

	for _, pos := range positions {
		fmt.Printf("   Moving to %s...\n", pos.name)
		result, err = moveTool.Execute(ctx, fmt.Sprintf(`{"x": %d, "y": %d}`, pos.x, pos.y))
		if err != nil {
			fmt.Printf("   ERROR: %v\n", err)
		} else {
			fmt.Printf("   OK: %s\n", result)
		}
		time.Sleep(800 * time.Millisecond) // Pause so you can see each position
	}

	fmt.Println("\n   Did you see the cursor move to all 5 positions?")
	fmt.Println("   If not, check Accessibility permissions.")

	fmt.Println("\n--- Starting keyboard tests in 3 seconds ---")
	fmt.Println("    Please focus on a text editor or terminal")
	time.Sleep(3 * time.Second)

	// Test 2: Type some text
	fmt.Println("\n2. Testing keyboard_type (typing 'hello')...")
	typeTool := tools.NewTypeTool()
	result, err = typeTool.Execute(ctx, `{"text": "hello"}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", result)
	}
	time.Sleep(500 * time.Millisecond)

	// Test 3: Press enter
	fmt.Println("\n3. Testing keyboard_press (pressing 'enter')...")
	keypress := tools.NewKeyPressTool()
	result, err = keypress.Execute(ctx, `{"key": "enter"}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", result)
	}
	time.Sleep(500 * time.Millisecond)

	// Test 4: Cmd+Space (Spotlight)
	fmt.Println("\n4. Testing keyboard_press (Cmd+Space for Spotlight)...")
	result, err = keypress.Execute(ctx, `{"key": "cmd+space"}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", result)
	}
	// Note: keypress now has 300ms built-in delay for modifier combos
	fmt.Println("   (Waiting for Spotlight to open...)")
	time.Sleep(500 * time.Millisecond) // Additional wait for Spotlight animation

	// Test 5: Type in spotlight
	fmt.Println("\n5. Testing keyboard_type in Spotlight (typing 'Calculator')...")
	fmt.Println("   (Now typing character by character with 10ms delay...)")
	result, err = typeTool.Execute(ctx, `{"text": "Calculator"}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", result)
	}
	time.Sleep(500 * time.Millisecond)

	// Test 6: Press escape to close spotlight
	fmt.Println("\n6. Testing keyboard_press (Escape to close)...")
	result, err = keypress.Execute(ctx, `{"key": "escape"}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", result)
	}

	// Test 7: Mouse click - move first, then click
	fmt.Println("\n7. Testing mouse_click...")
	fmt.Println("   Moving cursor to center then clicking...")
	click := tools.NewClickTool()
	result, err = click.Execute(ctx, `{"x": 500, "y": 500}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", result)
	}

	fmt.Println("\n=== Test Complete ===")
	fmt.Println()
	fmt.Println("CHECKLIST - Did you see:")
	fmt.Println("  [ ] Cursor move to all 4 corners and center?")
	fmt.Println("  [ ] 'hello' typed in your text editor?")
	fmt.Println("  [ ] Spotlight open (Cmd+Space)?")
	fmt.Println("  [ ] 'Calculator' typed in Spotlight?")
	fmt.Println("  [ ] Mouse click in center?")
	fmt.Println()
	fmt.Println("If mouse didn't move: Check System Preferences > Privacy & Security > Accessibility")
	fmt.Println("If typing didn't work: Also check Input Monitoring permissions")
	fmt.Println()
	fmt.Println("Your terminal/IDE needs BOTH Accessibility AND Input Monitoring permissions!")
}

// truncateJSON truncates a JSON string, removing large base64 data
func truncateJSON(s string, maxLen int) string {
	// Remove the image_base64 field content to show just the metadata
	// Find "image_base64":"..." and replace with "image_base64":"[truncated]"
	start := 0
	for {
		idx := indexOfString(s[start:], `"image_base64":"`)
		if idx == -1 {
			break
		}
		idx += start + len(`"image_base64":"`)
		// Find the closing quote
		end := idx
		for end < len(s) && s[end] != '"' {
			if s[end] == '\\' && end+1 < len(s) {
				end++ // Skip escaped char
			}
			end++
		}
		s = s[:idx] + "[base64 data truncated]" + s[end:]
		start = idx + len("[base64 data truncated]")
	}

	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

func indexOfString(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
