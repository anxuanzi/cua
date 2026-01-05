// Package platform provides operating system detection and platform-specific information.
package platform

import (
	"fmt"
	"runtime"
	"strings"
)

// OS represents an operating system.
type OS string

const (
	// Darwin is macOS.
	Darwin OS = "darwin"
	// Windows is Microsoft Windows.
	Windows OS = "windows"
	// Linux is Linux.
	Linux OS = "linux"
	// Unknown is an unrecognized OS.
	Unknown OS = "unknown"
)

// Info contains platform-specific information.
type Info struct {
	// OS is the operating system (darwin, windows, linux).
	OS OS

	// Arch is the CPU architecture (amd64, arm64, etc.).
	Arch string

	// Version is the OS version string (if available).
	Version string

	// DisplayName is a human-readable OS name.
	DisplayName string
}

// Current returns the current platform information.
func Current() Info {
	os := OS(runtime.GOOS)
	info := Info{
		OS:   os,
		Arch: runtime.GOARCH,
	}

	switch os {
	case Darwin:
		info.DisplayName = "macOS"
	case Windows:
		info.DisplayName = "Windows"
	case Linux:
		info.DisplayName = "Linux"
	default:
		info.DisplayName = string(os)
	}

	return info
}

// IsMacOS returns true if running on macOS.
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux returns true if running on Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// KeyboardInfo contains platform-specific keyboard information.
type KeyboardInfo struct {
	// PrimaryModifier is the main modifier key (Cmd on macOS, Ctrl on Windows/Linux).
	PrimaryModifier string

	// SecondaryModifier is the secondary modifier (Ctrl on macOS, Alt on Windows).
	SecondaryModifier string

	// AppLauncher describes how to open the app launcher.
	AppLauncher AppLauncherInfo

	// CommonShortcuts are platform-specific keyboard shortcuts.
	CommonShortcuts map[string]Shortcut
}

// AppLauncherInfo describes how to open applications on this platform.
type AppLauncherInfo struct {
	// Name is the launcher name (e.g., "Spotlight", "Start Menu").
	Name string

	// OpenMethod describes how to open the launcher.
	OpenMethod string

	// Key is the key to press.
	Key string

	// Modifiers are the modifier keys to hold.
	Modifiers []string
}

// Shortcut represents a keyboard shortcut.
type Shortcut struct {
	// Description is what the shortcut does.
	Description string

	// Key is the key to press.
	Key string

	// Modifiers are the modifier keys to hold.
	Modifiers []string
}

// GetKeyboardInfo returns keyboard information for the current platform.
func GetKeyboardInfo() KeyboardInfo {
	switch OS(runtime.GOOS) {
	case Darwin:
		return KeyboardInfo{
			PrimaryModifier:   "cmd",
			SecondaryModifier: "ctrl",
			AppLauncher: AppLauncherInfo{
				Name:       "Spotlight",
				OpenMethod: "Press Cmd+Space to open Spotlight search",
				Key:        "space",
				Modifiers:  []string{"cmd"},
			},
			CommonShortcuts: map[string]Shortcut{
				"copy":       {Description: "Copy", Key: "c", Modifiers: []string{"cmd"}},
				"paste":      {Description: "Paste", Key: "v", Modifiers: []string{"cmd"}},
				"cut":        {Description: "Cut", Key: "x", Modifiers: []string{"cmd"}},
				"undo":       {Description: "Undo", Key: "z", Modifiers: []string{"cmd"}},
				"redo":       {Description: "Redo", Key: "z", Modifiers: []string{"cmd", "shift"}},
				"save":       {Description: "Save", Key: "s", Modifiers: []string{"cmd"}},
				"select_all": {Description: "Select All", Key: "a", Modifiers: []string{"cmd"}},
				"find":       {Description: "Find", Key: "f", Modifiers: []string{"cmd"}},
				"quit":       {Description: "Quit Application", Key: "q", Modifiers: []string{"cmd"}},
				"close":      {Description: "Close Window", Key: "w", Modifiers: []string{"cmd"}},
				"new_tab":    {Description: "New Tab", Key: "t", Modifiers: []string{"cmd"}},
				"switch_app": {Description: "Switch Application", Key: "tab", Modifiers: []string{"cmd"}},
			},
		}

	case Windows:
		return KeyboardInfo{
			PrimaryModifier:   "ctrl",
			SecondaryModifier: "alt",
			AppLauncher: AppLauncherInfo{
				Name:       "Start Menu",
				OpenMethod: "Press Windows key to open Start Menu",
				Key:        "cmd", // robotgo uses "cmd" for Windows key too
				Modifiers:  []string{},
			},
			CommonShortcuts: map[string]Shortcut{
				"copy":       {Description: "Copy", Key: "c", Modifiers: []string{"ctrl"}},
				"paste":      {Description: "Paste", Key: "v", Modifiers: []string{"ctrl"}},
				"cut":        {Description: "Cut", Key: "x", Modifiers: []string{"ctrl"}},
				"undo":       {Description: "Undo", Key: "z", Modifiers: []string{"ctrl"}},
				"redo":       {Description: "Redo", Key: "y", Modifiers: []string{"ctrl"}},
				"save":       {Description: "Save", Key: "s", Modifiers: []string{"ctrl"}},
				"select_all": {Description: "Select All", Key: "a", Modifiers: []string{"ctrl"}},
				"find":       {Description: "Find", Key: "f", Modifiers: []string{"ctrl"}},
				"quit":       {Description: "Quit Application", Key: "f4", Modifiers: []string{"alt"}},
				"close":      {Description: "Close Window", Key: "w", Modifiers: []string{"ctrl"}},
				"new_tab":    {Description: "New Tab", Key: "t", Modifiers: []string{"ctrl"}},
				"switch_app": {Description: "Switch Application", Key: "tab", Modifiers: []string{"alt"}},
			},
		}

	default: // Linux and others
		return KeyboardInfo{
			PrimaryModifier:   "ctrl",
			SecondaryModifier: "alt",
			AppLauncher: AppLauncherInfo{
				Name:       "Application Menu",
				OpenMethod: "Press Super/Meta key to open application menu",
				Key:        "cmd",
				Modifiers:  []string{},
			},
			CommonShortcuts: map[string]Shortcut{
				"copy":       {Description: "Copy", Key: "c", Modifiers: []string{"ctrl"}},
				"paste":      {Description: "Paste", Key: "v", Modifiers: []string{"ctrl"}},
				"cut":        {Description: "Cut", Key: "x", Modifiers: []string{"ctrl"}},
				"undo":       {Description: "Undo", Key: "z", Modifiers: []string{"ctrl"}},
				"redo":       {Description: "Redo", Key: "z", Modifiers: []string{"ctrl", "shift"}},
				"save":       {Description: "Save", Key: "s", Modifiers: []string{"ctrl"}},
				"select_all": {Description: "Select All", Key: "a", Modifiers: []string{"ctrl"}},
				"find":       {Description: "Find", Key: "f", Modifiers: []string{"ctrl"}},
				"quit":       {Description: "Quit Application", Key: "q", Modifiers: []string{"ctrl"}},
				"close":      {Description: "Close Window", Key: "w", Modifiers: []string{"ctrl"}},
				"new_tab":    {Description: "New Tab", Key: "t", Modifiers: []string{"ctrl"}},
				"switch_app": {Description: "Switch Application", Key: "tab", Modifiers: []string{"alt"}},
			},
		}
	}
}

// FormatShortcut formats a shortcut for display.
func FormatShortcut(key string, modifiers []string) string {
	if len(modifiers) == 0 {
		return key
	}
	return strings.Join(modifiers, "+") + "+" + key
}

// ToPromptContext generates platform context for inclusion in agent prompts.
func ToPromptContext() string {
	info := Current()
	kb := GetKeyboardInfo()

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<platform>\n"))
	sb.WriteString(fmt.Sprintf("  <os>%s</os>\n", info.DisplayName))
	sb.WriteString(fmt.Sprintf("  <os_id>%s</os_id>\n", info.OS))
	sb.WriteString(fmt.Sprintf("  <arch>%s</arch>\n", info.Arch))
	sb.WriteString(fmt.Sprintf("  <primary_modifier>%s</primary_modifier>\n", kb.PrimaryModifier))
	sb.WriteString(fmt.Sprintf("  \n"))
	sb.WriteString(fmt.Sprintf("  <app_launcher>\n"))
	sb.WriteString(fmt.Sprintf("    <name>%s</name>\n", kb.AppLauncher.Name))
	sb.WriteString(fmt.Sprintf("    <how_to_open>%s</how_to_open>\n", kb.AppLauncher.OpenMethod))
	sb.WriteString(fmt.Sprintf("    <key>%s</key>\n", kb.AppLauncher.Key))
	if len(kb.AppLauncher.Modifiers) > 0 {
		sb.WriteString(fmt.Sprintf("    <modifiers>%s</modifiers>\n", strings.Join(kb.AppLauncher.Modifiers, ", ")))
	}
	sb.WriteString(fmt.Sprintf("  </app_launcher>\n"))
	sb.WriteString(fmt.Sprintf("  \n"))
	sb.WriteString(fmt.Sprintf("  <common_shortcuts>\n"))
	for name, shortcut := range kb.CommonShortcuts {
		sb.WriteString(fmt.Sprintf("    <%s>%s</%s>\n", name, FormatShortcut(shortcut.Key, shortcut.Modifiers), name))
	}
	sb.WriteString(fmt.Sprintf("  </common_shortcuts>\n"))
	sb.WriteString(fmt.Sprintf("</platform>"))

	return sb.String()
}

// GetAppOpenInstructions returns instructions for opening an app on this platform.
func GetAppOpenInstructions(appName string) string {
	info := Current()
	kb := GetKeyboardInfo()

	switch info.OS {
	case Darwin:
		return fmt.Sprintf(`To open %s on macOS:
1. Press Cmd+Space to open Spotlight
2. Type "%s"
3. Press Enter when the app appears in results

Alternative: Look for the app icon in the Dock at the bottom of the screen.`, appName, appName)

	case Windows:
		return fmt.Sprintf(`To open %s on Windows:
1. Press the Windows key to open Start Menu
2. Type "%s"
3. Press Enter when the app appears in results

Alternative: Look for the app in the Start Menu or taskbar.`, appName, appName)

	default:
		return fmt.Sprintf(`To open %s:
1. Open the application menu using %s
2. Search for "%s"
3. Click or press Enter to launch`, appName, kb.AppLauncher.OpenMethod, appName)
	}
}
