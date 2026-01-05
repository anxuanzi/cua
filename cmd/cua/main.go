// Command cua is a CLI wrapper around the CUA (Computer Use Agent) library.
//
// This is a thin wrapper that provides command-line access to CUA functionality.
// For programmatic use, import the github.com/anxuanzi/cua package directly.
//
// Usage:
//
//	cua "Open Safari and search for golang"
//	cua --verbose "Fill out the contact form"
//	cua --model pro "Complex multi-step task"
//
// Environment Variables:
//
//	GOOGLE_API_KEY - Required. Your Google API key for Gemini access.
//
// Examples:
//
//	# Simple task
//	cua "Open Calculator and compute 42 * 17"
//
//	# With verbose output
//	cua --verbose "Take a screenshot of the desktop"
//
//	# Using the Pro model for complex tasks
//	cua --model pro "Analyze the spreadsheet and create a summary chart"
//
//	# With longer timeout
//	cua --timeout 5m "Complete the multi-page form"
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anxuanzi/cua"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// Define flags
	var (
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		vFlag       = flag.Bool("v", false, "Enable verbose output (shorthand)")
		model       = flag.String("model", "flash", "Model to use: flash (default) or pro")
		timeout     = flag.Duration("timeout", 2*time.Minute, "Maximum time for task completion")
		maxActions  = flag.Int("max-actions", 50, "Maximum number of actions")
		safetyLevel = flag.String("safety", "normal", "Safety level: minimal, normal, or strict")
		headless    = flag.Bool("headless", false, "Run without human takeover UI")
		showVersion = flag.Bool("version", false, "Show version and exit")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "CUA - Computer Use Agent v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: cua [options] <task>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  cua \"Open Safari and search for golang\"\n")
		fmt.Fprintf(os.Stderr, "  cua --verbose \"Fill out the form\"\n")
		fmt.Fprintf(os.Stderr, "  cua --model pro \"Complex analysis task\"\n")
		fmt.Fprintf(os.Stderr, "\nEnvironment:\n")
		fmt.Fprintf(os.Stderr, "  GOOGLE_API_KEY - Required. Your Google API key.\n")
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("CUA v%s (built: %s)\n", version, buildTime)
		os.Exit(0)
	}

	// Get task from remaining arguments
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No task provided")
		fmt.Fprintln(os.Stderr, "Usage: cua [options] <task>")
		fmt.Fprintln(os.Stderr, "Run 'cua --help' for more information")
		os.Exit(1)
	}
	task := strings.Join(args, " ")

	// Build options
	opts := []cua.Option{
		cua.WithTimeout(*timeout),
		cua.WithMaxActions(*maxActions),
		cua.WithVerbose(*verbose || *vFlag),
		cua.WithHeadless(*headless),
	}

	// Parse model
	switch strings.ToLower(*model) {
	case "flash", "gemini-3-flash":
		opts = append(opts, cua.WithModel(cua.Gemini3Flash))
	case "pro", "gemini-3-pro":
		opts = append(opts, cua.WithModel(cua.Gemini3Pro))
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown model %q. Use 'flash' or 'pro'.\n", *model)
		os.Exit(1)
	}

	// Parse safety level
	switch strings.ToLower(*safetyLevel) {
	case "minimal", "min":
		opts = append(opts, cua.WithSafetyLevel(cua.SafetyMinimal))
	case "normal", "":
		opts = append(opts, cua.WithSafetyLevel(cua.SafetyNormal))
	case "strict", "max":
		opts = append(opts, cua.WithSafetyLevel(cua.SafetyStrict))
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown safety level %q. Use 'minimal', 'normal', or 'strict'.\n", *safetyLevel)
		os.Exit(1)
	}

	// Create agent
	agent := cua.New(opts...)

	// Run with progress if verbose
	if *verbose || *vFlag {
		fmt.Printf("CUA v%s - Starting task: %s\n", version, task)
		fmt.Println(strings.Repeat("-", 60))

		err := agent.DoWithProgress(task, func(step cua.Step) {
			status := "✓"
			if !step.Success {
				status = "✗"
			}
			fmt.Printf("[%s] Step %d: %s - %s\n", status, step.Number, step.Action, step.Description)
			if step.Error != nil {
				fmt.Printf("    Error: %v\n", step.Error)
			}
		})

		fmt.Println(strings.Repeat("-", 60))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Task failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task completed successfully")
	} else {
		// Simple execution
		result, err := agent.Do(task)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if result.Success {
			fmt.Println(result.Summary)
		} else {
			fmt.Fprintf(os.Stderr, "Task failed: %v\n", result.Error)
			os.Exit(1)
		}
	}
}
