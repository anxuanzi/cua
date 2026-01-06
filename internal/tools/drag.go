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
	// StartX is the starting X coordinate (in image pixels).
	// Use the pixel position from the screenshot image.
	StartX int `json:"start_x" jsonschema:"Starting X coordinate in image pixels (from the screenshot)"`

	// StartY is the starting Y coordinate (in image pixels).
	// Use the pixel position from the screenshot image.
	StartY int `json:"start_y" jsonschema:"Starting Y coordinate in image pixels (from the screenshot)"`

	// EndX is the ending X coordinate (in image pixels).
	// Use the pixel position from the screenshot image.
	EndX int `json:"end_x" jsonschema:"Ending X coordinate in image pixels (from the screenshot)"`

	// EndY is the ending Y coordinate (in image pixels).
	// Use the pixel position from the screenshot image.
	EndY int `json:"end_y" jsonschema:"Ending Y coordinate in image pixels (from the screenshot)"`
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

	// Convert image pixel coordinates to logical screen coordinates
	logicalStartX, logicalStartY, startMode := screen.ConvertModelCoord(args.StartX, args.StartY)
	logicalEndX, logicalEndY, endMode := screen.ConvertModelCoord(args.EndX, args.EndY)

	// Get screen and image sizes for logging
	screenW, screenH := screen.LogicalScreenSize()
	imgW, imgH := screen.ImageSize()

	logging.Info("[drag] from input(%d, %d) → logical(%d, %d) [%s] to input(%d, %d) → logical(%d, %d) [%s] [screen=%dx%d, image=%dx%d]",
		args.StartX, args.StartY, logicalStartX, logicalStartY, startMode,
		args.EndX, args.EndY, logicalEndX, logicalEndY, endMode, screenW, screenH, imgW, imgH)

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
			Description: "Performs a mouse drag operation from start coordinates to end coordinates. Use image pixel coordinates from the screenshot. Useful for moving elements, resizing windows, selecting text, or slider adjustments.",
		},
		performDrag,
	)
}
