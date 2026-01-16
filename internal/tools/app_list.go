package tools

import (
	"context"
)

// AppInfo represents information about an installed application.
type AppInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path,omitempty"`
	BundleID string `json:"bundle_id,omitempty"` // macOS only
	Version  string `json:"version,omitempty"`
}

// AppListTool lists installed applications.
type AppListTool struct {
	BaseTool
}

// NewAppListTool creates a new app list tool.
func NewAppListTool() *AppListTool {
	return &AppListTool{}
}

func (t *AppListTool) Name() string {
	return "app_list"
}

func (t *AppListTool) Description() string {
	return `List installed applications on the system.

Returns a list of application names that can be used with the app_launch tool.

Options:
- search: Filter apps by name (case-insensitive substring match)
- limit: Maximum number of apps to return (default: 50)

On macOS: Scans /Applications and ~/Applications
On Windows: Scans Start Menu and Program Files`
}

func (t *AppListTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"search": {
			Type:        "string",
			Description: "Filter apps by name (case-insensitive substring match)",
			Required:    false,
		},
		"limit": {
			Type:        "integer",
			Description: "Maximum number of apps to return (default: 50)",
			Required:    false,
			Default:     50,
		},
	}
}

func (t *AppListTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Search string `json:"search"`
		Limit  int    `json:"limit"`
	}

	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), ""), nil
	}

	if args.Limit <= 0 {
		args.Limit = 50
	}

	// Platform-specific listing
	return listApps(ctx, args.Search, args.Limit)
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *AppListTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
