package tools

import (
	"context"
	"encoding/json"

	"github.com/anxuanzi/cua/internal/coords"
)

// ScreenInfoTool provides information about available screens.
type ScreenInfoTool struct {
	BaseTool
}

// NewScreenInfoTool creates a new screen info tool.
func NewScreenInfoTool() *ScreenInfoTool {
	return &ScreenInfoTool{}
}

func (t *ScreenInfoTool) Name() string {
	return "screen_info"
}

func (t *ScreenInfoTool) Description() string {
	return `Get information about available screens/monitors. Returns dimensions, positions, and scale factors for all connected displays. Use this to understand the screen layout before performing actions.`
}

func (t *ScreenInfoTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"screen_index": {
			Type:        "integer",
			Description: "Specific screen index to get info for (-1 for all screens)",
			Required:    false,
			Default:     -1,
		},
	}
}

func (t *ScreenInfoTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		ScreenIndex int `json:"screen_index"`
	}
	args.ScreenIndex = -1 // default: all screens

	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), ""), nil
	}

	if args.ScreenIndex >= 0 {
		// Get info for specific screen
		screen := coords.GetScreen(args.ScreenIndex)
		result := map[string]interface{}{
			"screen_index": screen.Index,
			"x":            screen.X,
			"y":            screen.Y,
			"width":        screen.Width,
			"height":       screen.Height,
			"scale_factor": screen.ScaleFactor,
			"is_primary":   screen.IsPrimary,
		}
		resultJSON, _ := json.Marshal(result)
		return string(resultJSON), nil
	}

	// Get info for all screens
	screens := coords.GetAllScreens()
	screenInfos := make([]map[string]interface{}, len(screens))

	for i, screen := range screens {
		screenInfos[i] = map[string]interface{}{
			"screen_index": screen.Index,
			"x":            screen.X,
			"y":            screen.Y,
			"width":        screen.Width,
			"height":       screen.Height,
			"scale_factor": screen.ScaleFactor,
			"is_primary":   screen.IsPrimary,
		}
	}

	result := map[string]interface{}{
		"screen_count": len(screens),
		"screens":      screenInfos,
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *ScreenInfoTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
