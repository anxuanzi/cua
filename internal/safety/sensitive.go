package safety

import (
	"regexp"
	"strings"
)

// SensitivePattern defines a pattern that may require user confirmation.
type SensitivePattern struct {
	// Name is a descriptive name for the pattern.
	Name string

	// Pattern is the regex pattern to match.
	Pattern *regexp.Regexp

	// Level indicates severity (warning, block).
	Level SensitiveLevel

	// Description explains why this is sensitive.
	Description string
}

// SensitiveLevel indicates how to handle sensitive matches.
type SensitiveLevel int

const (
	// SensitiveLevelWarning logs a warning but allows the action.
	SensitiveLevelWarning SensitiveLevel = iota

	// SensitiveLevelConfirm requires user confirmation before proceeding.
	SensitiveLevelConfirm

	// SensitiveLevelBlock blocks the action entirely.
	SensitiveLevelBlock
)

// SensitiveDetector checks for potentially dangerous actions.
type SensitiveDetector struct {
	patterns []SensitivePattern
}

// NewSensitiveDetector creates a detector with default patterns.
func NewSensitiveDetector() *SensitiveDetector {
	return &SensitiveDetector{
		patterns: defaultSensitivePatterns(),
	}
}

// defaultSensitivePatterns returns the built-in sensitive patterns.
func defaultSensitivePatterns() []SensitivePattern {
	return []SensitivePattern{
		// Password and credential patterns
		{
			Name:        "password_field",
			Pattern:     regexp.MustCompile(`(?i)(password|passwd|pwd|credential|secret|token)`),
			Level:       SensitiveLevelConfirm,
			Description: "Interacting with password or credential fields",
		},
		{
			Name:        "api_key",
			Pattern:     regexp.MustCompile(`(?i)(api[_-]?key|access[_-]?token|auth[_-]?token|bearer)`),
			Level:       SensitiveLevelBlock,
			Description: "Interacting with API keys or tokens",
		},

		// System settings
		{
			Name:        "system_preferences",
			Pattern:     regexp.MustCompile(`(?i)(system preferences|system settings|control panel|admin|administrator)`),
			Level:       SensitiveLevelConfirm,
			Description: "Accessing system settings",
		},
		{
			Name:        "security_privacy",
			Pattern:     regexp.MustCompile(`(?i)(security|privacy|firewall|permissions|accessibility)`),
			Level:       SensitiveLevelConfirm,
			Description: "Accessing security or privacy settings",
		},

		// Financial and payment
		{
			Name:        "payment",
			Pattern:     regexp.MustCompile(`(?i)(credit card|debit card|payment|checkout|purchase|buy now|billing)`),
			Level:       SensitiveLevelConfirm,
			Description: "Interacting with payment or financial information",
		},
		{
			Name:        "banking",
			Pattern:     regexp.MustCompile(`(?i)(bank|account number|routing number|wire transfer|cryptocurrency|wallet)`),
			Level:       SensitiveLevelBlock,
			Description: "Interacting with banking information",
		},

		// Personal information
		{
			Name:        "ssn",
			Pattern:     regexp.MustCompile(`(?i)(ssn|social security|national id|passport|driver.?s? license)`),
			Level:       SensitiveLevelBlock,
			Description: "Interacting with government ID or SSN fields",
		},

		// Destructive actions
		{
			Name:        "delete",
			Pattern:     regexp.MustCompile(`(?i)(delete|remove|erase|clear all|reset|format|wipe)`),
			Level:       SensitiveLevelConfirm,
			Description: "Performing destructive action",
		},
		{
			Name:        "shutdown",
			Pattern:     regexp.MustCompile(`(?i)(shutdown|restart|reboot|power off|force quit|kill)`),
			Level:       SensitiveLevelConfirm,
			Description: "Shutting down or restarting system/application",
		},

		// Email and communication
		{
			Name:        "send_email",
			Pattern:     regexp.MustCompile(`(?i)(send|submit|post|publish|broadcast).*?(email|message|mail)`),
			Level:       SensitiveLevelConfirm,
			Description: "Sending email or message",
		},

		// Terminal and code execution
		{
			Name:        "terminal",
			Pattern:     regexp.MustCompile(`(?i)(terminal|command prompt|powershell|bash|shell|sudo|su\s)`),
			Level:       SensitiveLevelConfirm,
			Description: "Interacting with terminal or command line",
		},
	}
}

// SensitiveMatch represents a match result.
type SensitiveMatch struct {
	Pattern     SensitivePattern
	MatchedText string
}

// Check analyzes the action and target for sensitive patterns.
// Returns nil if no sensitive patterns matched.
func (d *SensitiveDetector) Check(action, target, description string) []SensitiveMatch {
	text := strings.ToLower(action + " " + target + " " + description)

	var matches []SensitiveMatch
	for _, pattern := range d.patterns {
		if loc := pattern.Pattern.FindStringIndex(text); loc != nil {
			matches = append(matches, SensitiveMatch{
				Pattern:     pattern,
				MatchedText: text[loc[0]:loc[1]],
			})
		}
	}

	return matches
}

// IsSensitive returns true if any sensitive pattern matches.
func (d *SensitiveDetector) IsSensitive(action, target, description string) bool {
	return len(d.Check(action, target, description)) > 0
}

// GetHighestLevel returns the highest sensitivity level among matches.
func (d *SensitiveDetector) GetHighestLevel(matches []SensitiveMatch) SensitiveLevel {
	highest := SensitiveLevelWarning
	for _, match := range matches {
		if match.Pattern.Level > highest {
			highest = match.Pattern.Level
		}
	}
	return highest
}

// AddPattern adds a custom sensitive pattern.
func (d *SensitiveDetector) AddPattern(pattern SensitivePattern) {
	d.patterns = append(d.patterns, pattern)
}

// RemovePattern removes a pattern by name.
func (d *SensitiveDetector) RemovePattern(name string) {
	for i, p := range d.patterns {
		if p.Name == name {
			d.patterns = append(d.patterns[:i], d.patterns[i+1:]...)
			return
		}
	}
}
