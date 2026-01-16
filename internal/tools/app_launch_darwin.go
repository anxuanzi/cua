//go:build darwin

package tools

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// launchApp launches an application on macOS using the 'open' command.
func launchApp(ctx context.Context, appName string, wait bool) (string, error) {
	// Try different variations of the app name
	variations := []string{
		appName,
		appName + ".app",
	}

	// Common app name mappings
	appMappings := map[string]string{
		"chrome":             "Google Chrome",
		"google chrome":      "Google Chrome",
		"firefox":            "Firefox",
		"safari":             "Safari",
		"terminal":           "Terminal",
		"iterm":              "iTerm",
		"iterm2":             "iTerm",
		"vscode":             "Visual Studio Code",
		"code":               "Visual Studio Code",
		"sublime":            "Sublime Text",
		"atom":               "Atom",
		"finder":             "Finder",
		"mail":               "Mail",
		"notes":              "Notes",
		"reminders":          "Reminders",
		"calendar":           "Calendar",
		"messages":           "Messages",
		"facetime":           "FaceTime",
		"photos":             "Photos",
		"music":              "Music",
		"podcasts":           "Podcasts",
		"tv":                 "TV",
		"news":               "News",
		"stocks":             "Stocks",
		"maps":               "Maps",
		"weather":            "Weather",
		"calculator":         "Calculator",
		"preview":            "Preview",
		"textedit":           "TextEdit",
		"activity monitor":   "Activity Monitor",
		"disk utility":       "Disk Utility",
		"system preferences": "System Preferences",
		"system settings":    "System Settings",
		"app store":          "App Store",
		"xcode":              "Xcode",
		"slack":              "Slack",
		"discord":            "Discord",
		"zoom":               "zoom.us",
		"notion":             "Notion",
		"spotify":            "Spotify",
	}

	// Check if there's a mapping for the lowercase version
	lowerName := strings.ToLower(appName)
	if mapped, ok := appMappings[lowerName]; ok {
		variations = append([]string{mapped}, variations...)
	}

	var lastErr error
	for _, name := range variations {
		args := []string{"-a", name}
		if wait {
			args = append(args, "-W")
		}

		cmd := exec.CommandContext(ctx, "open", args...)
		err := cmd.Run()
		if err == nil {
			// Give the app a moment to launch
			time.Sleep(500 * time.Millisecond)

			return SuccessResponse(map[string]interface{}{
				"launched": name,
				"platform": "darwin",
				"waited":   wait,
			}), nil
		}
		lastErr = err
	}

	// If all variations failed, try mdfind to find the app
	mdfindCmd := exec.CommandContext(ctx, "mdfind", "kMDItemKind == 'Application' && kMDItemDisplayName == '"+appName+"'")
	output, err := mdfindCmd.Output()
	if err == nil && len(output) > 0 {
		// Found the app path, try to open it directly
		appPath := strings.TrimSpace(strings.Split(string(output), "\n")[0])
		if appPath != "" {
			cmd := exec.CommandContext(ctx, "open", appPath)
			if wait {
				cmd = exec.CommandContext(ctx, "open", "-W", appPath)
			}
			if err := cmd.Run(); err == nil {
				time.Sleep(500 * time.Millisecond)
				return SuccessResponse(map[string]interface{}{
					"launched": appName,
					"path":     appPath,
					"platform": "darwin",
					"waited":   wait,
				}), nil
			}
		}
	}

	return ErrorResponse(
		"failed to launch application: "+appName,
		"Check if the application is installed. Error: "+lastErr.Error(),
	), nil
}
