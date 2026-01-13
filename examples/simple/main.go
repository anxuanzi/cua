// Simple example demonstrating basic CUA usage.
//
// This example shows how to:
// 1. Create a CUA instance
// 2. Execute individual tools directly
// 3. Run a full task with the LLM-powered Run method
//
// Run with: ANTHROPIC_API_KEY=your-key go run main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/anxuanzi/cua"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("Note: GEMINI_API_KEY not set. Some features will be limited.")
		fmt.Println("Set it to use the full LLM-powered automation.")
		fmt.Println()
	}

	// Create CUA instance
	agent, err := cua.New(
		cua.WithAPIKey(apiKey),
		cua.WithProvider(cua.ProviderGemini),
		cua.WithScreenIndex(0), // Primary screen
	)
	if err != nil {
		log.Fatalf("Failed to create CUA: %v", err)
	}

	ctx := context.Background()

	// === Direct Tool Usage ===
	fmt.Println("=== Direct Tool Usage ===")
	fmt.Println("(Tools can be used directly without LLM)")
	fmt.Println()

	// Example 1: Get screen info
	fmt.Println("1. Getting Screen Info:")
	result, err := agent.ExecuteTool(ctx, "get_screen_info", `{"screen_index": 0}`)
	if err != nil {
		log.Fatalf("Failed to get screen info: %v", err)
	}
	prettyPrint(result)

	// Example 2: Take a screenshot
	fmt.Println("\n2. Taking Screenshot:")
	result, err = agent.ExecuteTool(ctx, "screenshot", `{}`)
	if err != nil {
		log.Fatalf("Failed to take screenshot: %v", err)
	}
	// Parse and show metadata (not the full base64 image)
	var screenshotResult map[string]interface{}
	json.Unmarshal([]byte(result), &screenshotResult)
	fmt.Printf("   Original: %v x %v\n", screenshotResult["original_width"], screenshotResult["original_height"])
	fmt.Printf("   Scaled:   %v x %v\n", screenshotResult["scaled_width"], screenshotResult["scaled_height"])
	if b64, ok := screenshotResult["image_base64"].(string); ok {
		fmt.Printf("   Base64:   %d characters\n", len(b64))
	}

	// Example 3: List available tools
	fmt.Println("\n3. Available Tools:")
	for _, t := range agent.Tools() {
		fmt.Printf("   - %s: %s\n", t.Name(), truncate(t.Description(), 60))
	}

	// === LLM-Powered Task Execution ===
	fmt.Println("\n=== LLM-Powered Task Execution ===")

	if apiKey == "" {
		fmt.Println("Skipping LLM task execution (no API key)")
		fmt.Println("To see full automation, set ANTHROPIC_API_KEY and run again.")
	} else {
		// Example: Run a simple task using LLM
		fmt.Println("Running task: 'Take a screenshot and describe what you see'")
		fmt.Println("(This will use Claude to analyze the screen)")
		fmt.Println()

		result, err := agent.Run(ctx, "Take a screenshot and briefly describe what you see on the screen.")
		if err != nil {
			log.Printf("Task failed: %v", err)
		} else {
			fmt.Println("Result:")
			fmt.Println(result)
		}
	}

	fmt.Println("\n=== Done! ===")
}

func prettyPrint(jsonStr string) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		fmt.Printf("   %s\n", jsonStr)
		return
	}
	pretty, _ := json.MarshalIndent(data, "   ", "  ")
	fmt.Printf("%s\n", string(pretty))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
