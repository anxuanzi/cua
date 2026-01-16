//go:build windows

package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// listApps lists installed applications on Windows.
func listApps(_ context.Context, search string, limit int) (string, error) {
	apps := make([]AppInfo, 0)
	seen := make(map[string]bool)

	// Common locations for applications
	programFiles := os.Getenv("ProgramFiles")
	programFilesX86 := os.Getenv("ProgramFiles(x86)")
	localAppData := os.Getenv("LOCALAPPDATA")
	appData := os.Getenv("APPDATA")
	userProfile := os.Getenv("USERPROFILE")

	// Start Menu locations
	startMenuDirs := []string{
		filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs"),
	}

	// Scan Start Menu for shortcuts (.lnk files)
	searchLower := strings.ToLower(search)

	for _, dir := range startMenuDirs {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			if d.IsDir() {
				return nil
			}

			name := d.Name()

			// Only process .lnk files
			if !strings.HasSuffix(strings.ToLower(name), ".lnk") {
				return nil
			}

			// Remove .lnk suffix for display name
			displayName := strings.TrimSuffix(name, ".lnk")
			displayName = strings.TrimSuffix(displayName, ".LNK")

			// Skip common non-app entries
			lowerName := strings.ToLower(displayName)
			if strings.Contains(lowerName, "uninstall") ||
				strings.Contains(lowerName, "readme") ||
				strings.Contains(lowerName, "help") ||
				strings.Contains(lowerName, "license") ||
				strings.Contains(lowerName, "documentation") {
				return nil
			}

			// Skip if already seen
			if seen[displayName] {
				return nil
			}
			seen[displayName] = true

			// Apply search filter
			if search != "" && !strings.Contains(strings.ToLower(displayName), searchLower) {
				return nil
			}

			apps = append(apps, AppInfo{
				Name: displayName,
				Path: path,
			})

			if len(apps) >= limit {
				return filepath.SkipAll
			}

			return nil
		})

		if err != nil || len(apps) >= limit {
			break
		}
	}

	// Also scan Program Files for .exe files (top-level only)
	programDirs := []string{programFiles, programFilesX86, localAppData, filepath.Join(userProfile, "AppData", "Local", "Programs")}

	for _, dir := range programDirs {
		if dir == "" {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()

			// Skip common non-app directories
			lowerName := strings.ToLower(name)
			if strings.HasPrefix(lowerName, "common") ||
				strings.HasPrefix(lowerName, "windows") ||
				strings.HasPrefix(lowerName, "microsoft") && !strings.Contains(lowerName, "vscode") {
				continue
			}

			// Skip if already seen
			if seen[name] {
				continue
			}

			// Apply search filter
			if search != "" && !strings.Contains(lowerName, searchLower) {
				continue
			}

			// Check if directory contains an .exe file
			appDir := filepath.Join(dir, name)
			hasExe := false
			dirEntries, _ := os.ReadDir(appDir)
			for _, de := range dirEntries {
				if strings.HasSuffix(strings.ToLower(de.Name()), ".exe") {
					hasExe = true
					break
				}
			}

			if !hasExe {
				continue
			}

			seen[name] = true
			apps = append(apps, AppInfo{
				Name: name,
				Path: appDir,
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
		"platform": "windows",
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
