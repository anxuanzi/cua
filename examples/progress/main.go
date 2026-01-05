// Package main demonstrates using CUA with progress callbacks.
//
// This example shows how to monitor task execution in real-time
// using the DoWithProgress method. This is useful for:
//   - Displaying progress to users
//   - Logging each action for debugging
//   - Taking screenshots at each step
//   - Building interactive UIs
//
// Prerequisites:
//   - Set GOOGLE_API_KEY or GEMINI_API_KEY environment variable
//   - macOS: Grant Accessibility and Screen Recording permissions
//   - Windows: May need to run as Administrator for some applications
//
// Run with:
//
//	go run ./examples/progress
package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/anxuanzi/cua"
)

func main() {
	fmt.Println("=== CUA Progress Example ===")
	fmt.Println()

	// Create agent with custom options
	agent := cua.New(
		// Use verbose mode to see internal reasoning
		cua.WithVerbose(true),
		// Allow more time for complex tasks
		cua.WithTimeout(5*time.Minute),
		// Allow more actions
		cua.WithMaxActions(100),
	)

	// Define the task
	task := "Open Safari, go to google.com, and search for 'Go programming language'"

	// Allow overriding the task from command line
	if len(os.Args) > 1 {
		task = os.Args[1]
	}

	fmt.Printf("Task: %s\n", task)
	fmt.Println()
	fmt.Println("Executing with progress tracking...")
	fmt.Println()

	// Create output directory for screenshots
	screenshotDir := "screenshots"
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		log.Printf("Warning: Could not create screenshot directory: %v", err)
	}

	// Track timing
	start := time.Now()

	// Execute with progress callback
	err := agent.DoWithProgress(task, func(step cua.Step) {
		elapsed := time.Since(start).Round(time.Millisecond)

		// Print step info
		status := "✓"
		if !step.Success {
			status = "✗"
		}

		fmt.Printf("[%s] %s Step %d: %s\n", elapsed, status, step.Number, step.Description)

		if step.Target != "" {
			fmt.Printf("         Target: %s\n", step.Target)
		}

		if step.Duration > 0 {
			fmt.Printf("         Duration: %v\n", step.Duration)
		}

		if step.Error != nil {
			fmt.Printf("         Error: %v\n", step.Error)
		}

		// Save screenshot if available
		if step.Screenshot != nil {
			filename := filepath.Join(screenshotDir, fmt.Sprintf("step_%03d.png", step.Number))
			if err := saveScreenshot(step.Screenshot, filename); err != nil {
				log.Printf("Warning: Could not save screenshot: %v", err)
			} else {
				fmt.Printf("         Screenshot: %s\n", filename)
			}
		}

		fmt.Println()
	})

	if err != nil {
		log.Fatalf("Task failed: %v", err)
	}

	fmt.Println("=== Task Complete ===")
	fmt.Printf("Total time: %v\n", time.Since(start).Round(time.Millisecond))
	fmt.Printf("Screenshots saved to: %s/\n", screenshotDir)
}

// saveScreenshot saves an image.Image to a PNG file.
func saveScreenshot(img image.Image, filename string) error {
	if img == nil {
		return fmt.Errorf("nil image")
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode PNG: %w", err)
	}

	return nil
}
