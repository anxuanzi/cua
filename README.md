# CUA - Computer Use Agent

A pure desktop Computer Use Agent built in Go using Google ADK and Gemini 3.

## Overview

CUA is an AI-powered desktop automation agent that can perform any task a human can on macOS or Windows 11. It uses a
vision-first approach combined with native accessibility APIs to understand and interact with the desktop environment.

**Design Philosophy:**

- **Library-First**: Import and use in 5 lines of code
- **CLI as Wrapper**: Command-line app built on the library
- **Own Implementations**: We build our own element/accessibility layer (no unmaintained dependencies)
- **ADK-Powered**: All agent logic uses Google ADK

## Quick Start

### As a Library (Recommended)

```go
package main

import "github.com/anxuanzi/cua"

func main() {
	// Create agent with API key
	agent := cua.New(cua.WithAPIKey("your-gemini-api-key"))

	// Do any task
	result, err := agent.Do("Open Safari and search for 'golang tutorials'")
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
```

### With Progress Callback

```go
agent := cua.New(cua.WithAPIKey("your-key"))

err := agent.DoWithProgress("Fill out the registration form", func (step cua.Step) {
fmt.Printf("Step %d: %s\n", step.Number, step.Description)
})
```

### Low-Level API

```go
// Direct element interaction
elements, _ := cua.FindElements(cua.ByRole("button"))
elements[0].Click()

// Direct input control
cua.Click(100, 200)
cua.TypeText("Hello, World!")
cua.KeyPress(cua.KeyEnter, cua.ModCmd)

// Screen capture
screenshot, _ := cua.CaptureScreen()
```

### As a CLI

```bash
# Run a task
cua do "Open Terminal and run 'ls -la'"

# Direct commands
cua click 100 200
cua type "Hello"
cua screenshot output.png
cua elements  # List visible elements
```

## Features

- **Vision-First**: Uses Gemini 3 vision to understand screen content
- **Element Enhancement**: Native accessibility APIs for precise element identification
- **Cross-Platform**: macOS (primary) and Windows 11
- **Safe by Design**: Rate limiting, sensitive area detection, human takeover (Cmd+Shift+Esc)
- **Simple API**: Dead-simple library interface for developers

## Requirements

- Go 1.25+
- macOS 14+ or Windows 11
- Google API Key (for Gemini)
- Accessibility permissions:
    - macOS: System Settings → Privacy & Security → Accessibility
    - Windows: May need Administrator for some apps

## Installation

```bash
# As a library
go get github.com/anxuanzi/cua

# Build CLI from source
git clone https://github.com/anxuanzi/cua.git
cd cua
go build -o cua ./cmd/cua
```

## Configuration Options

```go
agent := cua.New(
cua.WithAPIKey("your-key"), // Required
cua.WithModel(cua.Gemini3Pro), // Default: Gemini3Flash
cua.WithSafetyLevel(cua.SafetyStrict),// Default: SafetyNormal
cua.WithTimeout(5 * time.Minute), // Default: 2 minutes
cua.WithMaxActions(100),          // Default: 50
cua.WithVerbose(true), // Default: false
)
```

## Architecture

```
cua.New() → Agent.Do("task")
     │
     ├── pkg/element (OUR OWN CODE - cross-platform element finding)
     ├── pkg/input   (mouse/keyboard via RobotGo)
     └── pkg/screen  (screen capture)
     │
     └── internal/agent (ADK agents - hidden from users)
         ├── Coordinator (Gemini 3 Pro - planning)
         ├── Perception (Gemini 3 Flash - vision)
         └── Action (Gemini 3 Flash - execution)
```

## Safety

CUA includes multiple safety mechanisms:

1. **Rate Limiting**: Maximum 60 actions per minute
2. **Sensitive Detection**: Warns before passwords, payments, system settings
3. **Human Takeover**: Press `Cmd+Shift+Esc` (macOS) or `Ctrl+Shift+Esc` (Windows)
4. **Audit Logging**: All actions logged

## Examples

The `examples/` directory contains runnable examples:

- **[examples/simple/](examples/simple/)** - Basic 5-line quick start
- **[examples/progress/](examples/progress/)** - Real-time progress monitoring with callbacks
- **[examples/low_level/](examples/low_level/)** - Direct element/input control without AI

Run any example:

```bash
go run ./examples/simple
go run ./examples/progress
go run ./examples/low_level
```

## Benchmarks

Run performance benchmarks:

```bash
# All benchmarks
go test -bench=. -benchmem ./...

# Screen capture (target: <100ms)
go test -bench=BenchmarkCapture -benchmem ./pkg/screen/...

# TaskMemory operations
go test -bench=. -benchmem ./internal/memory/...
```

## Troubleshooting

Having issues? Check the [Troubleshooting Guide](docs/TROUBLESHOOTING.md) for common problems and solutions:

- Permission issues on macOS/Windows
- API key configuration
- Agent execution problems
- Performance optimization

## License

MIT