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
