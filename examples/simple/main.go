// Simple example demonstrating basic CUA usage.
//
// This example shows how to:
// 1. Create a CUA instance with token limit monitoring
// 2. Execute individual tools directly
// 3. Run a full task with the LLM-powered Run method
// 4. Track token usage across multiple runs (including failed runs)
//
// Run with: GEMINI_API_KEY=your-key go run main.go
//
// Optional:
//   - Set GOOGLE_GEMINI_BASE_URL to use a custom API endpoint
//   - Set USE_STREAMING=1 to use RunStreamWithTracking for better failure tracking
//
// NOTE: Token counts require agent-sdk-go to return usage data. On some failures
// (like context limit exceeded), the SDK may not return usage info. In this case,
// use RunStreamWithTracking which tracks tool calls and execution time from events.
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

	// Optional: Custom API endpoint
	baseURL := os.Getenv("GOOGLE_GEMINI_BASE_URL")

	// Create CUA instance with token limit monitoring
	// Gemini Tier 1 has a limit of 1,000,000 input tokens per minute
	opts := []cua.Option{
		cua.WithAPIKey(apiKey),
		cua.WithProvider(cua.ProviderGemini),
		cua.WithScreenIndex(0), // Primary screen
		// Set token limit to 900K to leave some buffer (Gemini tier 1 is 1M/min)
		cua.WithTokenLimit(900000),
		// Warn at 80% usage (720K tokens)
		cua.WithTokenLimitWarning(80, func(current, limit int, percentUsed float64) {
			fmt.Printf("\n⚠️  Token Usage Warning: %d/%d tokens (%.1f%%)\n", current, limit, percentUsed)
			fmt.Println("Consider waiting before next request to avoid rate limits.")
		}),
	}

	// Add custom base URL if provided
	if baseURL != "" {
		fmt.Printf("Using custom API endpoint: %s\n\n", baseURL)
		opts = append(opts, cua.WithBaseURL(baseURL))
	}

	agent, err := cua.New(opts...)
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
	result, err := agent.ExecuteTool(ctx, "screen_info", `{"screen_index": 0}`)
	if err != nil {
		log.Fatalf("Failed to get screen info: %v", err)
	}
	prettyPrint(result)

	// Example 2: Take a screenshot
	fmt.Println("\n2. Taking Screenshot:")
	result, err = agent.ExecuteTool(ctx, "screen_capture", `{}`)
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
		fmt.Println("To see full automation, set GEMINI_API_KEY and run again.")
	} else {
		// Example: Run a calculator task using LLM
		fmt.Println("Running task: 'Open Calculator and compute 123 * 456'")
		fmt.Println("(This will use Gemini to automate the Calculator app)")
		fmt.Println()

		task := "Open the Calculator app (use Spotlight with cmd+space, type 'Calculator', press enter). Then calculate 123 * 456 and tell me the result."

		var result string
		var err error

		// Use streaming mode if requested - this provides better tracking on failures
		// because we can count tool calls and LLM iterations from events
		if os.Getenv("USE_STREAMING") != "" {
			fmt.Println("Using RunStreamWithTracking for better failure tracking...")
			result, err = agent.RunStreamWithTracking(ctx, task)
		} else {
			result, err = agent.Run(ctx, task)
		}

		if err != nil {
			fmt.Printf("Task failed: %v\n", err)
			fmt.Println("(Tool calls and execution time are still tracked)")
		} else {
			fmt.Println("Result:")
			fmt.Println(result)
		}
	}

	// === Token Usage Statistics ===
	// NOTE: Usage is tracked even when tasks fail, so you can monitor
	// token consumption that led to failures (e.g., context limit exceeded)
	fmt.Println("\n=== Token Usage Statistics ===")
	usage := agent.Usage()
	fmt.Printf("Total Runs:         %d\n", usage.TotalRuns)
	fmt.Printf("Total LLM Calls:    %d\n", usage.TotalLLMCalls)
	fmt.Printf("Total Tool Calls:   %d\n", usage.TotalToolCalls)
	fmt.Printf("Total Input Tokens: %d\n", usage.TotalInputTokens)
	fmt.Printf("Total Output Tokens:%d\n", usage.TotalOutputTokens)
	fmt.Printf("Total Tokens:       %d\n", usage.TotalTokens)
	fmt.Printf("Execution Time:     %dms\n", usage.TotalTimeMs)

	// Show percentage of limit used
	if usage.TotalInputTokens > 0 {
		percentUsed := float64(usage.TotalInputTokens) / 900000 * 100
		fmt.Printf("Rate Limit Usage:   %.2f%% of 900K limit\n", percentUsed)
	} else if usage.TotalRuns > 0 {
		// Token count may be 0 if agent-sdk-go doesn't return usage data
		fmt.Println("Note: Token count unavailable (agent-sdk-go may not provide this data)")
		fmt.Println("      Execution time is still tracked for rate limiting purposes.")
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
