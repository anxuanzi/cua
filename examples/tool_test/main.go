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

	fmt.Println("\n--- Starting interactive tests in 3 seconds ---")
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
	time.Sleep(1 * time.Second)

	// Test 5: Type in spotlight
	fmt.Println("\n5. Testing keyboard_type in Spotlight (typing 'Calculator')...")
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

	// Test 7: Mouse click
	fmt.Println("\n7. Testing mouse_click at center (500, 500)...")
	click := tools.NewClickTool()
	result, err = click.Execute(ctx, `{"x": 500, "y": 500}`)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Result: %s\n", result)
	}

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("Did you see:")
	fmt.Println("  - 'hello' typed?")
	fmt.Println("  - Spotlight open?")
	fmt.Println("  - 'Calculator' typed in Spotlight?")
	fmt.Println("  - Mouse click in center?")
	fmt.Println("\nIf not, check System Preferences > Privacy & Security > Accessibility")
	fmt.Println("and ensure your terminal/IDE has permission.")
}
