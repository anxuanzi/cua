// Package cua provides a cross-platform Computer Use Agent for AI-powered desktop automation.
package cua

import "sync"

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

// TokenUsage represents token usage statistics.
type TokenUsage struct {
	// InputTokens is the number of input/prompt tokens used.
	InputTokens int `json:"input_tokens"`
	// OutputTokens is the number of output/completion tokens used.
	OutputTokens int `json:"output_tokens"`
	// TotalTokens is the total number of tokens used.
	TotalTokens int `json:"total_tokens"`
	// ReasoningTokens is the number of tokens used for reasoning (if applicable).
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}

// UsageStats represents cumulative token usage statistics across multiple runs.
type UsageStats struct {
	mu sync.RWMutex

	// Cumulative token usage
	TotalInputTokens     int `json:"total_input_tokens"`
	TotalOutputTokens    int `json:"total_output_tokens"`
	TotalTokens          int `json:"total_tokens"`
	TotalReasoningTokens int `json:"total_reasoning_tokens,omitempty"`

	// Execution statistics
	TotalRuns      int   `json:"total_runs"`
	TotalLLMCalls  int   `json:"total_llm_calls"`
	TotalToolCalls int   `json:"total_tool_calls"`
	TotalTimeMs    int64 `json:"total_time_ms"`
}

// Add adds token usage to the cumulative statistics.
func (s *UsageStats) Add(usage *TokenUsage, llmCalls, toolCalls int, timeMs int64) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if usage != nil {
		s.TotalInputTokens += usage.InputTokens
		s.TotalOutputTokens += usage.OutputTokens
		s.TotalTokens += usage.TotalTokens
		s.TotalReasoningTokens += usage.ReasoningTokens
	}
	s.TotalRuns++
	s.TotalLLMCalls += llmCalls
	s.TotalToolCalls += toolCalls
	s.TotalTimeMs += timeMs
}

// Get returns a copy of the current usage statistics.
func (s *UsageStats) Get() UsageStats {
	if s == nil {
		return UsageStats{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return UsageStats{
		TotalInputTokens:     s.TotalInputTokens,
		TotalOutputTokens:    s.TotalOutputTokens,
		TotalTokens:          s.TotalTokens,
		TotalReasoningTokens: s.TotalReasoningTokens,
		TotalRuns:            s.TotalRuns,
		TotalLLMCalls:        s.TotalLLMCalls,
		TotalToolCalls:       s.TotalToolCalls,
		TotalTimeMs:          s.TotalTimeMs,
	}
}

// Reset resets all usage statistics to zero.
func (s *UsageStats) Reset() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalInputTokens = 0
	s.TotalOutputTokens = 0
	s.TotalTokens = 0
	s.TotalReasoningTokens = 0
	s.TotalRuns = 0
	s.TotalLLMCalls = 0
	s.TotalToolCalls = 0
	s.TotalTimeMs = 0
}

// TokenLimitCallback is called when token usage approaches or exceeds limits.
type TokenLimitCallback func(current, limit int, percentUsed float64)

// Config holds the configuration for the CUA agent.
type Config struct {
	// Provider specifies which LLM provider to use.
	Provider LLMProvider

	// APIKey is the API key for the selected provider.
	APIKey string

	// Model overrides the default model for the provider.
	Model string

	// BaseURL is the custom API endpoint URL (optional).
	// For Gemini: overrides the default https://generativelanguage.googleapis.com/
	// For OpenAI: overrides the default https://api.openai.com/v1
	// For Anthropic: overrides the default https://api.anthropic.com
	BaseURL string

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

	// TokenLimit is the maximum number of input tokens allowed per minute (optional).
	// When set, the agent will track usage and call OnTokenLimitWarning when approaching the limit.
	TokenLimit int

	// TokenLimitWarningThreshold is the percentage (0-100) at which to trigger warnings.
	// Default is 80 (warn at 80% of limit).
	TokenLimitWarningThreshold int

	// OnTokenLimitWarning is called when token usage approaches the limit.
	OnTokenLimitWarning TokenLimitCallback
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
