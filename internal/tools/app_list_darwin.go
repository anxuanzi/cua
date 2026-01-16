//go:build darwin

package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// listApps lists installed applications on macOS.
func listApps(_ context.Context, search string, limit int) (string, error) {
	apps := make([]AppInfo, 0)
	seen := make(map[string]bool)

	// Directories to scan for applications
	appDirs := []string{
		"/Applications",
		"/System/Applications",
		"/System/Applications/Utilities",
		filepath.Join(os.Getenv("HOME"), "Applications"),
	}

	searchLower := strings.ToLower(search)

	for _, dir := range appDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip directories that don't exist or can't be read
		}

		for _, entry := range entries {
			name := entry.Name()

			// Only process .app bundles
			if !strings.HasSuffix(name, ".app") {
				continue
			}

			// Remove .app suffix for display name
			displayName := strings.TrimSuffix(name, ".app")

			// Skip if already seen
			if seen[displayName] {
				continue
			}
			seen[displayName] = true

			// Apply search filter
			if search != "" && !strings.Contains(strings.ToLower(displayName), searchLower) {
				continue
			}

			appPath := filepath.Join(dir, name)

			// Try to read bundle ID from Info.plist
			bundleID := ""
			infoPlistPath := filepath.Join(appPath, "Contents", "Info.plist")
			if _, err := os.Stat(infoPlistPath); err == nil {
				// Could parse plist here, but keeping it simple for now
				bundleID = "" // Would need plist parsing library
			}

			apps = append(apps, AppInfo{
				Name:     displayName,
				Path:     appPath,
				BundleID: bundleID,
			})

			if len(apps) >= limit {
				break
			}
		}

		if len(apps) >= limit {
			break
		}
	}

	// Sort alphabetically
	sort.Slice(apps, func(i, j int) bool {
		return strings.ToLower(apps[i].Name) < strings.ToLower(apps[j].Name)
	})

	// Trim to limit
	if len(apps) > limit {
		apps = apps[:limit]
	}

	// Build response
	appNames := make([]string, len(apps))
	for i, app := range apps {
		appNames[i] = app.Name
	}

	result := map[string]interface{}{
		"success":  true,
		"platform": "darwin",
		"count":    len(apps),
		"apps":     appNames,
	}

	// Include full details if searching
	if search != "" {
		result["details"] = apps
	}

	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}
