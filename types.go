// Package cua provides a cross-platform Computer Use Agent for AI-powered desktop automation.
package cua

// LLMProvider represents the LLM provider to use.
type LLMProvider string

const (
	// ProviderAnthropic uses Anthropic's Claude models.
	ProviderAnthropic LLMProvider = "anthropic"
	// ProviderOpenAI uses OpenAI's GPT models.
	ProviderOpenAI LLMProvider = "openai"
	// ProviderGemini uses Google's Gemini models.
	ProviderGemini LLMProvider = "gemini"
)

// Config holds the configuration for the CUA agent.
type Config struct {
	// Provider specifies which LLM provider to use.
	Provider LLMProvider

	// APIKey is the API key for the selected provider.
	APIKey string

	// Model overrides the default model for the provider.
	Model string

	// ScreenIndex specifies which screen to use for multi-monitor setups.
	ScreenIndex int

	// EnableReasoning enables extended thinking/reasoning mode.
	EnableReasoning bool

	// ReasoningBudget sets the token budget for reasoning (default: 4096).
	ReasoningBudget int

	// MaxIterations sets maximum tool-calling iterations (default: 50).
	MaxIterations int

	// Timeout sets the maximum time for a single task in seconds (default: 120).
	Timeout int

	// OrgID is the organization ID for multi-tenancy support.
	OrgID string

	// ConversationID is the conversation ID for memory isolation.
	ConversationID string
}

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		Provider:        ProviderAnthropic,
		Model:           "",
		ScreenIndex:     0,
		EnableReasoning: true,
		ReasoningBudget: 4096,
		MaxIterations:   50,
		Timeout:         120,
	}
}
