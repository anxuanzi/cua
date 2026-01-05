package tools

import (
	"time"

	"github.com/anxuanzi/cua/pkg/logging"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// WaitArgs defines the arguments for the wait tool.
type WaitArgs struct {
	// Duration is the time to wait in milliseconds.
	Duration int `json:"duration" jsonschema:"Time to wait in milliseconds (1-30000)"`
}

// WaitResult contains the result of a wait operation.
type WaitResult struct {
	// Success indicates if the wait completed.
	Success bool `json:"success"`

	// WaitedMs is the actual time waited in milliseconds.
	WaitedMs int `json:"waited_ms"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// performWait handles the wait tool invocation.
func performWait(ctx tool.Context, args WaitArgs) (WaitResult, error) {
	logging.Info("[wait] Waiting %dms", args.Duration)

	// Validate duration
	if args.Duration < 1 {
		logging.Error("[wait] Duration must be at least 1 millisecond")
		return WaitResult{
			Success: false,
			Error:   "duration must be at least 1 millisecond",
		}, nil
	}

	if args.Duration > 30000 {
		logging.Error("[wait] Duration cannot exceed 30000 milliseconds")
		return WaitResult{
			Success: false,
			Error:   "duration cannot exceed 30000 milliseconds (30 seconds)",
		}, nil
	}

	// Perform the wait
	start := time.Now()
	time.Sleep(time.Duration(args.Duration) * time.Millisecond)
	elapsed := time.Since(start)

	logging.Info("[wait] Completed after %dms", elapsed.Milliseconds())
	return WaitResult{
		Success:  true,
		WaitedMs: int(elapsed.Milliseconds()),
	}, nil
}

// NewWaitTool creates the wait tool for ADK agents.
func NewWaitTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "wait",
			Description: "Pauses execution for a specified duration. Use this to wait for UI transitions, loading, or animations to complete.",
		},
		performWait,
	)
}
