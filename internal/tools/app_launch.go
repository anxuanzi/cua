package tools

import (
	"context"
)

// AppLaunchTool launches applications by name.
type AppLaunchTool struct {
	BaseTool
}

// NewAppLaunchTool creates a new app launch tool.
func NewAppLaunchTool() *AppLaunchTool {
	return &AppLaunchTool{}
}

func (t *AppLaunchTool) Name() string {
	return "app_launch"
}

func (t *AppLaunchTool) Description() string {
	return `Launch an application by name. This is more reliable than using Spotlight/Start menu for opening apps.

Examples:
- "Calculator" → Opens Calculator app
- "Safari" or "Google Chrome" → Opens browser
- "Terminal" or "cmd" → Opens terminal

On macOS: Uses 'open -a' command
On Windows: Uses 'start' command or direct execution

Returns success with the launched app name, or error if app not found.`
}

func (t *AppLaunchTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"app_name": {
			Type:        "string",
			Description: "Name of the application to launch (e.g., 'Calculator', 'Safari', 'Notepad')",
			Required:    true,
		},
		"wait": {
			Type:        "boolean",
			Description: "Wait for the app to launch before returning (default: false)",
			Required:    false,
			Default:     false,
		},
	}
}

func (t *AppLaunchTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		AppName string `json:"app_name"`
		Wait    bool   `json:"wait"`
	}

	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide app_name"), nil
	}

	if args.AppName == "" {
		return ErrorResponse("app_name cannot be empty", "Provide the application name to launch"), nil
	}

	// Platform-specific launch
	return launchApp(ctx, args.AppName, args.Wait)
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *AppLaunchTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
