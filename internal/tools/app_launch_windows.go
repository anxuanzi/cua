//go:build windows

package tools

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// launchApp launches an application on Windows.
func launchApp(ctx context.Context, appName string, wait bool) (string, error) {
	// Common app name mappings for Windows
	appMappings := map[string]string{
		"chrome":             "chrome",
		"google chrome":      "chrome",
		"firefox":            "firefox",
		"edge":               "msedge",
		"microsoft edge":     "msedge",
		"ie":                 "iexplore",
		"internet explorer":  "iexplore",
		"notepad":            "notepad",
		"calculator":         "calc",
		"calc":               "calc",
		"cmd":                "cmd",
		"command prompt":     "cmd",
		"powershell":         "powershell",
		"terminal":           "wt", // Windows Terminal
		"windows terminal":   "wt",
		"explorer":           "explorer",
		"file explorer":      "explorer",
		"paint":              "mspaint",
		"wordpad":            "wordpad",
		"snipping tool":      "snippingtool",
		"task manager":       "taskmgr",
		"control panel":      "control",
		"settings":           "ms-settings:",
		"mail":               "outlookmail:",
		"calendar":           "outlookcal:",
		"photos":             "ms-photos:",
		"maps":               "bingmaps:",
		"weather":            "bingweather:",
		"store":              "ms-windows-store:",
		"xbox":               "xbox:",
		"spotify":            "spotify",
		"slack":              "slack",
		"discord":            "discord",
		"zoom":               "zoom",
		"teams":              "msteams",
		"microsoft teams":    "msteams",
		"vscode":             "code",
		"visual studio code": "code",
		"code":               "code",
		"word":               "winword",
		"excel":              "excel",
		"powerpoint":         "powerpnt",
		"outlook":            "outlook",
	}

	// Check if there's a mapping for the lowercase version
	lowerName := strings.ToLower(appName)
	cmdName := appName
	if mapped, ok := appMappings[lowerName]; ok {
		cmdName = mapped
	}

	// Check if it's a URI scheme (like ms-settings:)
	if strings.Contains(cmdName, ":") {
		cmd := exec.CommandContext(ctx, "cmd", "/c", "start", "", cmdName)
		err := cmd.Run()
		if err == nil {
			time.Sleep(500 * time.Millisecond)
			return SuccessResponse(map[string]interface{}{
				"launched": appName,
				"uri":      cmdName,
				"platform": "windows",
				"waited":   wait,
			}), nil
		}
	}

	// Try to launch using 'start' command
	var cmd *exec.Cmd
	if wait {
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", "/wait", "", cmdName)
	} else {
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", "", cmdName)
	}

	err := cmd.Run()
	if err == nil {
		time.Sleep(500 * time.Millisecond)
		return SuccessResponse(map[string]interface{}{
			"launched": cmdName,
			"platform": "windows",
			"waited":   wait,
		}), nil
	}

	// Try direct execution (for apps in PATH)
	cmd = exec.CommandContext(ctx, cmdName)
	if !wait {
		cmd = exec.CommandContext(ctx, "cmd", "/c", cmdName)
	}
	err = cmd.Start()
	if err == nil {
		if wait {
			cmd.Wait()
		}
		time.Sleep(500 * time.Millisecond)
		return SuccessResponse(map[string]interface{}{
			"launched": cmdName,
			"platform": "windows",
			"waited":   wait,
		}), nil
	}

	return ErrorResponse(
		"failed to launch application: "+appName,
		"Check if the application is installed. Error: "+err.Error(),
	), nil
}
