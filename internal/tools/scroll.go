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
	// X is the X coordinate where scrolling occurs (in image pixels).
	// Use the pixel position from the screenshot image.
	X int `json:"x" jsonschema:"X coordinate in image pixels (from the screenshot)"`

	// Y is the Y coordinate where scrolling occurs (in image pixels).
	// Use the pixel position from the screenshot image.
	Y int `json:"y" jsonschema:"Y coordinate in image pixels (from the screenshot)"`

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

	// Convert image pixel coordinates to logical screen coordinates
	logicalX, logicalY, coordMode := screen.ConvertModelCoord(args.X, args.Y)

	// Get screen and image sizes for logging
	screenW, screenH := screen.LogicalScreenSize()
	imgW, imgH := screen.ImageSize()

	logging.Info("[scroll] at input(%d, %d) → logical(%d, %d), delta=(%d, %d) [mode=%s, screen=%dx%d, image=%dx%d]",
		args.X, args.Y, logicalX, logicalY, args.DeltaX, args.DeltaY, coordMode, screenW, screenH, imgW, imgH)

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
			Description: "Scrolls the mouse wheel at the specified coordinates. Use image pixel coordinates from the screenshot. Positive delta_y scrolls down, negative scrolls up. Positive delta_x scrolls right, negative scrolls left.",
		},
		performScroll,
	)
}
