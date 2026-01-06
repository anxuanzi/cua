package tools

import (
	"fmt"

	"github.com/anxuanzi/cua/pkg/input"
	"github.com/anxuanzi/cua/pkg/logging"
	"github.com/anxuanzi/cua/pkg/screen"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// ScrollArgs defines the arguments for the scroll tool.
type ScrollArgs struct {
	// X is the X coordinate where scrolling occurs (normalized 0-1000 from Gemini).
	X int `json:"x" jsonschema:"X coordinate (normalized 0-1000 from model output) where scroll occurs"`

	// Y is the Y coordinate where scrolling occurs (normalized 0-1000 from Gemini).
	Y int `json:"y" jsonschema:"Y coordinate (normalized 0-1000 from model output) where scroll occurs"`

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
		logging.Error("[scroll] At least one of delta_x or delta_y must be non-zero")
		return ScrollResult{
			Success: false,
			X:       args.X,
			Y:       args.Y,
			Error:   "at least one of delta_x or delta_y must be non-zero",
		}, nil
	}

	// Denormalize Gemini's 0-1000 coordinates to logical screen coordinates
	logicalX, logicalY := screen.DenormalizeCoord(args.X, args.Y)

	// Get screen size for logging
	screenW, screenH := screen.LogicalScreenSize()

	logging.Info("[scroll] at normalized(%d, %d) → logical(%d, %d), delta=(%d, %d) [screen=%dx%d]",
		args.X, args.Y, logicalX, logicalY, args.DeltaX, args.DeltaY, screenW, screenH)

	err := input.ScrollAt(input.Point{X: logicalX, Y: logicalY}, args.DeltaX, args.DeltaY)
	if err != nil {
		logging.Error("[scroll] Failed: %v", err)
		return ScrollResult{
			Success: false,
			X:       args.X,
			Y:       args.Y,
			DeltaX:  args.DeltaX,
			DeltaY:  args.DeltaY,
			Error:   fmt.Sprintf("scroll failed: %v", err),
		}, nil
	}

	logging.Info("[scroll] Success at logical(%d, %d)", logicalX, logicalY)
	return ScrollResult{
		Success: true,
		X:       args.X,
		Y:       args.Y,
		DeltaX:  args.DeltaX,
		DeltaY:  args.DeltaY,
	}, nil
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
