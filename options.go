package cua

import "time"

// Config holds the configuration for a CUA agent.
// Use Option functions to modify the default configuration.
type Config struct {
	// APIKey is the Google API key for Gemini access.
	// Required unless GOOGLE_API_KEY environment variable is set.
	apiKey string

	// Model is the Gemini model to use.
	// Default: Gemini3Flash
	model Model

	// SafetyLevel controls how cautious the agent is.
	// Default: SafetyNormal
	safetyLevel SafetyLevel

	// Timeout is the maximum time for a task to complete.
	// Default: 2 minutes
	timeout time.Duration

	// MaxActions is the maximum number of actions per task.
	// Default: 50
	maxActions int

	// Verbose enables detailed logging.
	// Default: false
	verbose bool

	// Headless disables the human takeover UI.
	// Default: false
	headless bool

	// RateLimitPerMinute is the maximum actions per minute.
	// Default: 60
	rateLimitPerMinute int
}

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		model:              Gemini2Flash,
		safetyLevel:        SafetyNormal,
		timeout:            2 * time.Minute,
		maxActions:         50,
		verbose:            false,
		headless:           false,
		rateLimitPerMinute: 60,
	}
}

// Option is a function that modifies the agent configuration.
type Option func(*Config)

// WithAPIKey sets the Google API key for Gemini access.
// If not provided, the agent will look for GOOGLE_API_KEY environment variable.
func WithAPIKey(key string) Option {
	return func(c *Config) {
		c.apiKey = key
	}
}

// WithModel sets the Gemini model to use.
// Default: Gemini2Flash
//
// Use Gemini3Pro for complex tasks requiring advanced reasoning.
// Use Gemini3Flash for faster, cost-effective execution.
func WithModel(model Model) Option {
	return func(c *Config) {
		c.model = model
	}
}

// WithSafetyLevel sets how cautious the agent should be.
// Default: SafetyNormal
//
//   - SafetyNormal: Standard safety checks, pauses for sensitive actions
//   - SafetyStrict: Maximum safety, more frequent pauses
//   - SafetyMinimal: Minimal checks (not recommended for production)
func WithSafetyLevel(level SafetyLevel) Option {
	return func(c *Config) {
		c.safetyLevel = level
	}
}

// WithTimeout sets the maximum duration for a task.
// Default: 2 minutes
//
// The agent will stop and return an error if the task exceeds this duration.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.timeout = d
	}
}

// WithMaxActions sets the maximum number of actions per task.
// Default: 50
//
// The agent will stop and return an error if the action count is exceeded.
// This prevents runaway agents from taking too many actions.
func WithMaxActions(n int) Option {
	return func(c *Config) {
		if n > 0 {
			c.maxActions = n
		}
	}
}

// WithVerbose enables detailed logging to stdout.
// Default: false
//
// When enabled, the agent will log each step and decision.
func WithVerbose(verbose bool) Option {
	return func(c *Config) {
		c.verbose = verbose
	}
}

// WithHeadless disables the human takeover UI.
// Default: false
//
// In headless mode, the agent runs without any UI overlay and
// the global hotkey for human takeover is disabled.
// Use this for programmatic/background automation.
func WithHeadless(headless bool) Option {
	return func(c *Config) {
		c.headless = headless
	}
}

// WithRateLimit sets the maximum actions per minute.
// Default: 60
//
// This prevents the agent from overwhelming the system or
// triggering rate limits on external services.
func WithRateLimit(actionsPerMinute int) Option {
	return func(c *Config) {
		if actionsPerMinute > 0 {
			c.rateLimitPerMinute = actionsPerMinute
		}
	}
}
