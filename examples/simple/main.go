// Package main demonstrates the simplest way to use CUA.
//
// This example shows how to get started with CUA in just 5 lines of code.
// It creates an agent and executes a simple task.
//
// Prerequisites:
//   - Set GOOGLE_API_KEY or GEMINI_API_KEY environment variable
//   - macOS: Grant Accessibility and Screen Recording permissions
//   - Windows: May need to run as Administrator for some applications
//
// Run with:
//
//	go run ./examples/simple
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/anxuanzi/cua"
)

func main() {
	// Create agent - API key is read from GOOGLE_API_KEY or GEMINI_API_KEY env var
	agent := cua.New()

	// Example 1: Simple task
	fmt.Println("=== CUA Simple Example ===")
	fmt.Println()

	// Define the task
	task := "Open the Calculator app and calculate 42 + 8"

	// Allow overriding the task from command line
	if len(os.Args) > 1 {
		task = os.Args[1]
	}

	fmt.Printf("Task: %s\n", task)
	fmt.Println("Executing...")
	fmt.Println()

	// Execute the task
	result, err := agent.Do(task)
	if err != nil {
		log.Fatalf("Task failed: %v", err)
	}

	// Print the result
	fmt.Println("=== Result ===")
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Summary: %s\n", result.Summary)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Steps taken: %d\n", len(result.Steps))
	fmt.Println()

	// Print each step
	if len(result.Steps) > 0 {
		fmt.Println("=== Steps ===")
		for _, step := range result.Steps {
			status := "✓"
			if !step.Success {
				status = "✗"
			}
			fmt.Printf("%s Step %d: %s\n", status, step.Number, step.Description)
			if step.Target != "" {
				fmt.Printf("    Target: %s\n", step.Target)
			}
			if step.Error != nil {
				fmt.Printf("    Error: %v\n", step.Error)
			}
		}
	}
}
