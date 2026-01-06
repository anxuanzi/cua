package tools

import (
	"errors"
	"fmt"

	"github.com/anxuanzi/cua/pkg/input"
	"github.com/anxuanzi/cua/pkg/logging"
	"github.com/anxuanzi/cua/pkg/screen"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// DragArgs defines the arguments for the drag tool.
type DragArgs struct {
	// StartX is the starting X coordinate (normalized 0-1000 from Gemini).
	StartX int `json:"start_x" jsonschema:"Starting X coordinate (normalized 0-1000 from model output)"`

	// StartY is the starting Y coordinate (normalized 0-1000 from Gemini).
	StartY int `json:"start_y" jsonschema:"Starting Y coordinate (normalized 0-1000 from model output)"`

	// EndX is the ending X coordinate (normalized 0-1000 from Gemini).
	EndX int `json:"end_x" jsonschema:"Ending X coordinate (normalized 0-1000 from model output)"`

	// EndY is the ending Y coordinate (normalized 0-1000 from Gemini).
	EndY int `json:"end_y" jsonschema:"Ending Y coordinate (normalized 0-1000 from model output)"`
}

// DragResult contains the result of a drag operation.
type DragResult struct {
	// Success indicates if the drag operation succeeded.
	Success bool `json:"success"`

	// StartX is the starting X coordinate.
	StartX int `json:"start_x"`

	// StartY is the starting Y coordinate.
	StartY int `json:"start_y"`

	// EndX is the ending X coordinate.
	EndX int `json:"end_x"`

	// EndY is the ending Y coordinate.
	EndY int `json:"end_y"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// validateDragArgs validates the drag arguments.
func validateDragArgs(args DragArgs) error {
	if args.StartX < 0 {
		return errors.New("start_x cannot be negative")
	}
	if args.StartY < 0 {
		return errors.New("start_y cannot be negative")
	}
	if args.EndX < 0 {
		return errors.New("end_x cannot be negative")
	}
	if args.EndY < 0 {
		return errors.New("end_y cannot be negative")
	}
	if args.StartX == args.EndX && args.StartY == args.EndY {
		return errors.New("start and end coordinates cannot be the same")
	}
	return nil
}

// performDrag handles the drag tool invocation.
func performDrag(ctx tool.Context, args DragArgs) (DragResult, error) {
	// Validate arguments
	if err := validateDragArgs(args); err != nil {
		logging.Error("[drag] Validation failed: %v", err)
		return DragResult{
			Success: false,
			StartX:  args.StartX,
			StartY:  args.StartY,
			EndX:    args.EndX,
			EndY:    args.EndY,
			Error:   err.Error(),
		}, nil
	}

	// Denormalize Gemini's 0-1000 coordinates to logical screen coordinates
	logicalStartX, logicalStartY := screen.DenormalizeCoord(args.StartX, args.StartY)
	logicalEndX, logicalEndY := screen.DenormalizeCoord(args.EndX, args.EndY)

	// Get screen size for logging
	screenW, screenH := screen.LogicalScreenSize()

	logging.Info("[drag] from normalized(%d, %d) → logical(%d, %d) to normalized(%d, %d) → logical(%d, %d) [screen=%dx%d]",
		args.StartX, args.StartY, logicalStartX, logicalStartY,
		args.EndX, args.EndY, logicalEndX, logicalEndY, screenW, screenH)

	start := input.Point{X: logicalStartX, Y: logicalStartY}
	end := input.Point{X: logicalEndX, Y: logicalEndY}
	err := input.Drag(start, end)
	if err != nil {
		logging.Error("[drag] Failed: %v", err)
		return DragResult{
			Success: false,
			StartX:  args.StartX,
			StartY:  args.StartY,
			EndX:    args.EndX,
			EndY:    args.EndY,
			Error:   fmt.Sprintf("drag failed: %v", err),
		}, nil
	}

	logging.Info("[drag] Success from logical(%d, %d) to logical(%d, %d)",
		logicalStartX, logicalStartY, logicalEndX, logicalEndY)
	return DragResult{
		Success: true,
		StartX:  args.StartX,
		StartY:  args.StartY,
		EndX:    args.EndX,
		EndY:    args.EndY,
	}, nil
}

// NewDragTool creates the drag tool for ADK agents.
func NewDragTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "drag",
			Description: "Performs a mouse drag operation from start coordinates to end coordinates. Useful for moving elements, resizing windows, selecting text, or slider adjustments.",
		},
		performDrag,
	)
}
