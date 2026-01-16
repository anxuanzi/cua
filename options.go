package cua

// Option is a functional option for configuring the CUA agent.
type Option func(*Config)

// WithProvider sets the LLM provider.
func WithProvider(provider LLMProvider) Option {
	return func(c *Config) {
		c.Provider = provider
	}
}

// WithAPIKey sets the API key for the LLM provider.
func WithAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

// WithModel overrides the default model for the provider.
func WithModel(model string) Option {
	return func(c *Config) {
		c.Model = model
	}
}

// WithScreenIndex sets which screen to use for multi-monitor setups.
func WithScreenIndex(index int) Option {
	return func(c *Config) {
		c.ScreenIndex = index
	}
}

// WithReasoning enables or disables extended thinking mode.
func WithReasoning(enabled bool) Option {
	return func(c *Config) {
		c.EnableReasoning = enabled
	}
}

// WithReasoningBudget sets the token budget for reasoning.
func WithReasoningBudget(budget int) Option {
	return func(c *Config) {
		c.ReasoningBudget = budget
	}
}

// WithMaxIterations sets maximum tool-calling iterations.
func WithMaxIterations(max int) Option {
	return func(c *Config) {
		c.MaxIterations = max
	}
}

// WithTimeout sets the maximum time for a single task in seconds.
func WithTimeout(seconds int) Option {
	return func(c *Config) {
		c.Timeout = seconds
	}
}

// WithOrgID sets the organization ID for multi-tenancy support.
func WithOrgID(orgID string) Option {
	return func(c *Config) {
		c.OrgID = orgID
	}
}

// WithConversationID sets the conversation ID for memory isolation.
func WithConversationID(conversationID string) Option {
	return func(c *Config) {
		c.ConversationID = conversationID
	}
}

// WithBaseURL sets a custom API endpoint URL.
// This allows using custom/proxy endpoints or alternative deployments.
// For Gemini: overrides the default https://generativelanguage.googleapis.com/
// For OpenAI: overrides the default https://api.openai.com/v1
// For Anthropic: overrides the default https://api.anthropic.com
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithTokenLimit sets the maximum number of input tokens allowed.
// When set, the agent will track usage and trigger warnings when approaching the limit.
// This is useful for staying within API rate limits (e.g., Gemini's 1M tokens/minute tier 1 limit).
func WithTokenLimit(limit int) Option {
	return func(c *Config) {
		c.TokenLimit = limit
	}
}

// WithTokenLimitWarning sets the warning threshold and callback for token limit monitoring.
// threshold is a percentage (0-100) at which to trigger warnings (default: 80).
// callback is called when usage reaches the threshold.
func WithTokenLimitWarning(threshold int, callback TokenLimitCallback) Option {
	return func(c *Config) {
		c.TokenLimitWarningThreshold = threshold
		c.OnTokenLimitWarning = callback
	}
}
