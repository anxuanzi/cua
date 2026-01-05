package tools

import (
	"fmt"

	"github.com/anxuanzi/cua/pkg/input"
	"github.com/anxuanzi/cua/pkg/screen"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// ScrollArgs defines the arguments for the scroll tool.
type ScrollArgs struct {
	// X is the X coordinate where scrolling occurs (in physical pixels from screenshot).
	X int `json:"x" jsonschema:"X coordinate in physical pixels (from screenshot) where scroll occurs"`

	// Y is the Y coordinate where scrolling occurs (in physical pixels from screenshot).
	Y int `json:"y" jsonschema:"Y coordinate in physical pixels (from screenshot) where scroll occurs"`

	// DeltaX is horizontal scroll amount. Positive = right, negative = left.
	DeltaX int `json:"delta_x,omitzero" jsonschema:"Horizontal scroll amount (positive = right, negative = left)"`

	// DeltaY is vertical scroll amount. Positive = down, negative = up.
	DeltaY int `json:"delta_y,omitzero" jsonschema:"Vertical scroll amount (positive = down, negative = up)"`
}

// ScrollResult contains the result of a scroll operation.
type ScrollResult struct {
	// Success indicates if scrolling succeeded.
	Success bool `json:"success"`

	// X is the X coordinate where scroll occurred.
	X int `json:"x"`

	// Y is the Y coordinate where scroll occurred.
	Y int `json:"y"`

	// DeltaX is the horizontal scroll amount.
	DeltaX int `json:"delta_x"`

	// DeltaY is the vertical scroll amount.
	DeltaY int `json:"delta_y"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// performScroll handles the scroll tool invocation.
func performScroll(ctx tool.Context, args ScrollArgs) (ScrollResult, error) {
	if args.DeltaX == 0 && args.DeltaY == 0 {
		return ScrollResult{
			Success: false,
			X:       args.X,
			Y:       args.Y,
			Error:   "at least one of delta_x or delta_y must be non-zero",
		}, nil
	}

	err := scrollNative(args.X, args.Y, args.DeltaX, args.DeltaY)
	if err != nil {
		return ScrollResult{
			Success: false,
			X:       args.X,
			Y:       args.Y,
			DeltaX:  args.DeltaX,
			DeltaY:  args.DeltaY,
			Error:   fmt.Sprintf("scroll failed: %v", err),
		}, nil
	}

	return ScrollResult{
		Success: true,
		X:       args.X,
		Y:       args.Y,
		DeltaX:  args.DeltaX,
		DeltaY:  args.DeltaY,
	}, nil
}

// scrollNative performs a scroll operation using robotgo.
func scrollNative(x, y, deltaX, deltaY int) error {
	// Convert physical pixels (from screenshot) to logical coordinates (for mouse)
	logicalX, logicalY := screen.PhysicalToLogical(x, y)
	return input.ScrollAt(input.Point{X: logicalX, Y: logicalY}, deltaX, deltaY)
}

// NewScrollTool creates the scroll tool for ADK agents.
func NewScrollTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "scroll",
			Description: "Scrolls the mouse wheel at the specified screen coordinates. Positive delta_y scrolls down, negative scrolls up. Positive delta_x scrolls right, negative scrolls left.",
		},
		performScroll,
	)
}
