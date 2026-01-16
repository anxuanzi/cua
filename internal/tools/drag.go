package tools

import (
	"context"
	"time"

	"github.com/anxuanzi/cua/internal/coords"
	"github.com/go-vgo/robotgo"
)

// DragTool performs mouse drag operations using normalized coordinates (0-1000 scale).
type DragTool struct {
	BaseTool
	// ScreenIndex specifies which screen to use (default: 0 = primary).
	ScreenIndex int
}

// NewDragTool creates a new drag tool.
func NewDragTool() *DragTool {
	return &DragTool{ScreenIndex: 0}
}

func (t *DragTool) Name() string {
	return "mouse_drag"
}

func (t *DragTool) Description() string {
	return `Drag from one position to another. Coordinates are NORMALIZED to 0-1000 scale. (0,0) is top-left, (1000,1000) is bottom-right. This performs a mouse press at the start position, moves to the end position, then releases.`
}

func (t *DragTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"start_x": {
			Type:        "integer",
			Description: "Starting X coordinate normalized 0-1000",
			Required:    true,
		},
		"start_y": {
			Type:        "integer",
			Description: "Starting Y coordinate normalized 0-1000",
			Required:    true,
		},
		"end_x": {
			Type:        "integer",
			Description: "Ending X coordinate normalized 0-1000",
			Required:    true,
		},
		"end_y": {
			Type:        "integer",
			Description: "Ending Y coordinate normalized 0-1000",
			Required:    true,
		},
		"button": {
			Type:        "string",
			Description: "Mouse button to use for dragging",
			Required:    false,
			Default:     "left",
			Enum:        []interface{}{"left", "right", "center"},
		},
		"screen_index": {
			Type:        "integer",
			Description: "Screen index for multi-monitor setups (0 = primary)",
			Required:    false,
			Default:     0,
		},
	}
}

func (t *DragTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		StartX      int    `json:"start_x"`
		StartY      int    `json:"start_y"`
		EndX        int    `json:"end_x"`
		EndY        int    `json:"end_y"`
		Button      string `json:"button"`
		ScreenIndex int    `json:"screen_index"`
	}
	args.Button = "left" // default

	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide start_x, start_y, end_x, end_y coordinates in 0-1000 scale"), nil
	}

	// Validate normalized coordinates
	for _, coord := range []struct {
		name string
		val  int
	}{
		{"start_x", args.StartX}, {"start_y", args.StartY},
		{"end_x", args.EndX}, {"end_y", args.EndY},
	} {
		if coord.val < 0 || coord.val > 1000 {
			return ErrorResponse(coord.name+" coordinate out of range", "Use normalized 0-1000 scale"), nil
		}
	}

	// Get screen info
	screenIndex := args.ScreenIndex
	if screenIndex == 0 && t.ScreenIndex != 0 {
		screenIndex = t.ScreenIndex
	}
	screen := coords.GetScreen(screenIndex)

	// Convert normalized coordinates (0-1000) to absolute screen coordinates
	// Standard mapping: 0=left/top, 1000=right/bottom (matches TuriX-CUA)
	startScreenX := screen.X + int(float64(args.StartX)/1000.0*float64(screen.Width))
	startScreenY := screen.Y + int(float64(args.StartY)/1000.0*float64(screen.Height))
	endScreenX := screen.X + int(float64(args.EndX)/1000.0*float64(screen.Width))
	endScreenY := screen.Y + int(float64(args.EndY)/1000.0*float64(screen.Height))

	// Perform drag: move to start, press, move to end, release
	robotgo.Move(startScreenX, startScreenY)
	time.Sleep(50 * time.Millisecond)

	robotgo.Toggle(args.Button, "down")
	time.Sleep(50 * time.Millisecond)

	// Smooth drag with intermediate steps for better reliability
	steps := 10
	for i := 1; i <= steps; i++ {
		x := startScreenX + (endScreenX-startScreenX)*i/steps
		y := startScreenY + (endScreenY-startScreenY)*i/steps
		robotgo.Move(x, y)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)
	robotgo.Toggle(args.Button, "up")

	return SuccessResponse(map[string]interface{}{
		"dragged_from_screen":    map[string]int{"x": startScreenX, "y": startScreenY},
		"dragged_to_screen":      map[string]int{"x": endScreenX, "y": endScreenY},
		"normalized_coords_from": map[string]int{"x": args.StartX, "y": args.StartY},
		"normalized_coords_to":   map[string]int{"x": args.EndX, "y": args.EndY},
		"screen_dimensions":      map[string]int{"width": screen.Width, "height": screen.Height},
		"button":                 args.Button,
		"screen_index":           screenIndex,
	}), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *DragTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
