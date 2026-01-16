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

	// Test 1: Screen capture (should work if permissions are OK)
	fmt.Println("1. Testing screen_capture...")
	screenshot := tools.NewScreenshotTool()
	result, err := screenshot.Execute(ctx, `{}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   OK - got %d chars of base64\n", len(result))
	}

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
