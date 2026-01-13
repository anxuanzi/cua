// Streaming example demonstrating real-time ReAct events.
//
// This example shows how to:
// 1. Create a CUA instance
// 2. Use RunStream to see real-time events
// 3. Handle thinking, tool calls, and observations
//
// Run with: GEMINI_API_KEY=your-key go run main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anxuanzi/cua"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: GEMINI_API_KEY environment variable is required for streaming.")
		fmt.Println("Usage: GEMINI_API_KEY=your-key go run main.go")
		os.Exit(1)
	}

	// Create CUA instance with Gemini
	agent, err := cua.New(
		cua.WithAPIKey(apiKey),
		cua.WithProvider(cua.ProviderGemini),
	)
	if err != nil {
		log.Fatalf("Failed to create CUA: %v", err)
	}

	ctx := context.Background()

	fmt.Println("=== CUA Streaming Example ===")
	fmt.Println("This demonstrates real-time ReAct events from Gemini.")
	fmt.Println()

	// Task to execute - Calculator example
	task := "Open the Calculator app (use Spotlight with cmd+space, type 'Calculator', press enter). Then calculate 987 + 654 and tell me the result."
	fmt.Printf("Task: %s\n", task)
	fmt.Println()
	fmt.Println("--- Starting ReAct Loop ---")
	fmt.Println()

	// Start streaming
	events, err := agent.RunStream(ctx, task)
	if err != nil {
		log.Fatalf("Failed to start streaming: %v", err)
	}

	// Process events as they arrive
	for event := range events {
		switch event.Type {
		case cua.EventThinking:
			// Extended thinking/reasoning from the model
			fmt.Printf("[THINKING] %s\n", truncate(event.Thinking, 200))

		case cua.EventContent:
			// Text content being generated
			fmt.Printf("%s", event.Content)

		case cua.EventToolCall:
			// Tool is being called
			if event.ToolCall != nil {
				fmt.Printf("\n[ACTION] %s", event.ToolCall.Name)
				if event.ToolCall.Arguments != "" && event.ToolCall.Arguments != "{}" {
					fmt.Printf(" - %s", formatArgs(event.ToolCall.Arguments))
				}
				fmt.Println()
			}

		case cua.EventToolResult:
			// Tool returned a result
			fmt.Printf("[OBSERVATION] %s\n", formatResult(event.ToolCall, event.ToolResult))

		case cua.EventComplete:
			// Task completed
			fmt.Println()
			fmt.Println("--- ReAct Loop Complete ---")

		case cua.EventError:
			// Error occurred
			fmt.Printf("\n[ERROR] %v\n", event.Error)
		}
	}

	fmt.Println()
	fmt.Println("=== Done! ===")
}

func formatArgs(args string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(args), &data); err != nil {
		return args
	}
	// Remove empty values
	for k, v := range data {
		if v == "" || v == 0 || v == nil {
			delete(data, k)
		}
	}
	if len(data) == 0 {
		return ""
	}
	pretty, _ := json.Marshal(data)
	return string(pretty)
}

func formatResult(toolCall *cua.ToolCallEvent, result string) string {
	if toolCall == nil {
		return truncate(result, 100)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return truncate(result, 100)
	}

	switch toolCall.Name {
	case "screen_capture":
		// Don't print the full base64
		return fmt.Sprintf("Screenshot captured (%vx%v -> %vx%v)",
			data["original_width"], data["original_height"],
			data["scaled_width"], data["scaled_height"])
	case "screen_info":
		if screens, ok := data["screens"].([]interface{}); ok {
			return fmt.Sprintf("Found %d screen(s)", len(screens))
		}
		return "Screen info retrieved"
	case "mouse_click", "mouse_move", "mouse_drag", "mouse_scroll":
		if success, ok := data["success"].(bool); ok && success {
			return "Success"
		}
		if errMsg, ok := data["error"].(string); ok {
			return fmt.Sprintf("Error: %s", errMsg)
		}
		return "Completed"
	case "keyboard_type", "keyboard_press":
		if success, ok := data["success"].(bool); ok && success {
			return "Text input completed"
		}
		return "Completed"
	default:
		pretty, _ := json.Marshal(data)
		return truncate(string(pretty), 100)
	}
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
