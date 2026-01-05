package tools

import (
	"errors"
	"fmt"

	"github.com/anxuanzi/cua/pkg/input"
	"github.com/anxuanzi/cua/pkg/screen"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// DragArgs defines the arguments for the drag tool.
type DragArgs struct {
	// StartX is the starting X coordinate (in physical pixels from screenshot).
	StartX int `json:"start_x" jsonschema:"Starting X coordinate in physical pixels (from screenshot)"`

	// StartY is the starting Y coordinate (in physical pixels from screenshot).
	StartY int `json:"start_y" jsonschema:"Starting Y coordinate in physical pixels (from screenshot)"`

	// EndX is the ending X coordinate (in physical pixels from screenshot).
	EndX int `json:"end_x" jsonschema:"Ending X coordinate in physical pixels (from screenshot)"`

	// EndY is the ending Y coordinate (in physical pixels from screenshot).
	EndY int `json:"end_y" jsonschema:"Ending Y coordinate in physical pixels (from screenshot)"`
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
		return DragResult{
			Success: false,
			StartX:  args.StartX,
			StartY:  args.StartY,
			EndX:    args.EndX,
			EndY:    args.EndY,
			Error:   err.Error(),
		}, nil
	}

	// Perform the drag operation
	err := dragNative(args.StartX, args.StartY, args.EndX, args.EndY)
	if err != nil {
		return DragResult{
			Success: false,
			StartX:  args.StartX,
			StartY:  args.StartY,
			EndX:    args.EndX,
			EndY:    args.EndY,
			Error:   fmt.Sprintf("drag failed: %v", err),
		}, nil
	}

	return DragResult{
		Success: true,
		StartX:  args.StartX,
		StartY:  args.StartY,
		EndX:    args.EndX,
		EndY:    args.EndY,
	}, nil
}

// dragNative performs a drag operation using the input package.
func dragNative(startX, startY, endX, endY int) error {
	// Convert physical pixels (from screenshot) to logical coordinates (for mouse)
	logicalStartX, logicalStartY := screen.PhysicalToLogical(startX, startY)
	logicalEndX, logicalEndY := screen.PhysicalToLogical(endX, endY)

	start := input.Point{X: logicalStartX, Y: logicalStartY}
	end := input.Point{X: logicalEndX, Y: logicalEndY}
	return input.Drag(start, end)
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
