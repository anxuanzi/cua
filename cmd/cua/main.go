// Command cua is a CLI wrapper around the CUA (Computer Use Agent) library.
//
// This is a thin wrapper that provides command-line access to CUA functionality.
// For programmatic use, import the github.com/anxuanzi/cua package directly.
//
// Usage:
//
//	cua do "Open Safari and search for golang"
//	cua click 100 200
//	cua type "Hello, world!"
//	cua screenshot output.png
//	cua elements
//
// Environment Variables:
//
//	GOOGLE_API_KEY - Required for agent tasks. Your Google API key for Gemini access.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/anxuanzi/cua"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Handle global flags that can appear before command
	if os.Args[1] == "--version" || os.Args[1] == "-version" {
		fmt.Printf("CUA v%s (built: %s)\n", version, buildTime)
		os.Exit(0)
	}
	if os.Args[1] == "--help" || os.Args[1] == "-help" || os.Args[1] == "-h" {
		printUsage()
		os.Exit(0)
	}

	// Get the command
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "do":
		cmdDo(args)
	case "click":
		cmdClick(args)
	case "type":
		cmdType(args)
	case "screenshot":
		cmdScreenshot(args)
	case "elements":
		cmdElements(args)
	case "screen":
		cmdScreen(args)
	default:
		// If not a known command, treat as a task (backward compatibility)
		// Reconstruct the task from all args
		task := strings.Join(os.Args[1:], " ")
		cmdDoTask(task, false)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `CUA - Computer Use Agent v%s

Usage: cua <command> [arguments]

Commands:
  do <task>           Run an AI agent task
  click <x> <y>       Click at screen coordinates
  type <text>         Type text at current cursor position
  screenshot [file]   Capture screen (default: screenshot.png)
  elements            List visible UI elements
  screen              Show screen dimensions

Options for 'do' command:
  --verbose, -v       Enable verbose output
  --model <model>     Model to use: flash (default) or pro
  --timeout <dur>     Maximum time for task completion (default: 2m)
  --max-actions <n>   Maximum number of actions (default: 50)
  --safety <level>    Safety level: minimal, normal, or strict
  --headless          Run without human takeover UI

Examples:
  cua do "Open Safari and search for golang"
  cua do --verbose "Fill out the form"
  cua click 100 200
  cua type "Hello, world!"
  cua screenshot desktop.png
  cua elements

Environment:
  GOOGLE_API_KEY - Required for agent tasks. Your Google API key.
`, version)
}

// cmdDo handles the 'do' command for running agent tasks
func cmdDo(args []string) {
	fs := flag.NewFlagSet("do", flag.ExitOnError)
	var (
		verbose     = fs.Bool("verbose", false, "Enable verbose output")
		vFlag       = fs.Bool("v", false, "Enable verbose output (shorthand)")
		model       = fs.String("model", "flash", "Model to use: flash (default) or pro")
		timeout     = fs.Duration("timeout", 2*time.Minute, "Maximum time for task completion")
		maxActions  = fs.Int("max-actions", 50, "Maximum number of actions")
		safetyLevel = fs.String("safety", "normal", "Safety level: minimal, normal, or strict")
		headless    = fs.Bool("headless", false, "Run without human takeover UI")
	)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: cua do [options] <task>\n\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No task provided")
		fmt.Fprintln(os.Stderr, "Usage: cua do [options] <task>")
		os.Exit(1)
	}
	task := strings.Join(remaining, " ")

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

	cmdDoTaskWithOpts(task, *verbose || *vFlag, opts)
}

func cmdDoTask(task string, verbose bool) {
	opts := []cua.Option{
		cua.WithTimeout(2 * time.Minute),
		cua.WithMaxActions(50),
		cua.WithVerbose(verbose),
	}
	cmdDoTaskWithOpts(task, verbose, opts)
}

func cmdDoTaskWithOpts(task string, verbose bool, opts []cua.Option) {
	// Create agent
	agent := cua.New(opts...)

	// Run with progress if verbose
	if verbose {
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

// cmdClick handles the 'click' command for direct clicking
func cmdClick(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: click requires x and y coordinates")
		fmt.Fprintln(os.Stderr, "Usage: cua click <x> <y>")
		os.Exit(1)
	}

	x, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid x coordinate: %v\n", err)
		os.Exit(1)
	}

	y, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid y coordinate: %v\n", err)
		os.Exit(1)
	}

	if err := cua.Click(x, y); err != nil {
		fmt.Fprintf(os.Stderr, "Error clicking: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Clicked at (%d, %d)\n", x, y)
}

// cmdType handles the 'type' command for direct typing
func cmdType(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: type requires text")
		fmt.Fprintln(os.Stderr, "Usage: cua type <text>")
		os.Exit(1)
	}

	text := strings.Join(args, " ")

	if err := cua.TypeText(text); err != nil {
		fmt.Fprintf(os.Stderr, "Error typing: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Typed: %s\n", text)
}

// cmdScreenshot handles the 'screenshot' command for screen capture
func cmdScreenshot(args []string) {
	filename := "screenshot.png"
	if len(args) > 0 {
		filename = args[0]
		// Ensure .png extension
		if !strings.HasSuffix(strings.ToLower(filename), ".png") {
			filename += ".png"
		}
	}

	img, err := cua.CaptureScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error capturing screen: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding PNG: %v\n", err)
		os.Exit(1)
	}

	bounds := img.Bounds()
	fmt.Printf("Screenshot saved to %s (%dx%d)\n", filename, bounds.Dx(), bounds.Dy())
}

// cmdElements handles the 'elements' command for listing UI elements
func cmdElements(args []string) {
	fs := flag.NewFlagSet("elements", flag.ExitOnError)
	var (
		maxDepth = fs.Int("depth", 3, "Maximum depth to traverse")
		role     = fs.String("role", "", "Filter by role (button, textfield, etc.)")
	)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: cua elements [options]\n\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nRoles: button, textfield, window, menu, menuitem, checkbox, etc.\n")
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// Get focused application elements
	var elements []*cua.Element
	var err error

	if *role != "" {
		elements, err = cua.FindElements(cua.ByRole(cua.Role(*role)))
	} else {
		// Get the focused application and list its elements
		app, appErr := cua.FocusedApplication()
		if appErr != nil {
			fmt.Fprintf(os.Stderr, "Error getting focused app: %v\n", appErr)
			os.Exit(1)
		}
		if app != nil {
			// Load children of the focused app
			if loadErr := app.LoadChildren(); loadErr != nil {
				fmt.Fprintf(os.Stderr, "Error loading children: %v\n", loadErr)
			}
			elements = app.Children
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding elements: %v\n", err)
		os.Exit(1)
	}

	if len(elements) == 0 {
		fmt.Println("No elements found")
		return
	}

	fmt.Printf("Found %d elements:\n", len(elements))
	for i, elem := range elements {
		if i >= 50 { // Limit output
			fmt.Printf("... and %d more elements\n", len(elements)-50)
			break
		}
		printElement(elem, 0, *maxDepth)
	}
}

func printElement(elem *cua.Element, indent, maxDepth int) {
	if indent > maxDepth || elem == nil {
		return
	}

	prefix := strings.Repeat("  ", indent)

	name := elem.Name
	if name == "" {
		name = elem.Title
	}
	if name == "" {
		name = "(unnamed)"
	}
	if len(name) > 30 {
		name = name[:27] + "..."
	}

	fmt.Printf("%s[%s] %s @ (%d,%d) %dx%d\n",
		prefix,
		elem.Role,
		name,
		elem.Bounds.X, elem.Bounds.Y,
		elem.Bounds.Width, elem.Bounds.Height,
	)

	// Print children
	for _, child := range elem.Children {
		printElement(child, indent+1, maxDepth)
	}
}

// cmdScreen handles the 'screen' command for showing screen dimensions
func cmdScreen(args []string) {
	width, height, err := cua.ScreenSize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting screen size: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Screen: %dx%d\n", width, height)
}
