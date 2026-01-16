// Package cua provides a cross-platform Computer Use Agent for AI-powered desktop automation.
//
// CUA enables AI models to interact with desktop applications through vision and
// automation tools. It uses agent-sdk-go for the underlying agent infrastructure
// and provides specialized tools for screen capture, mouse, keyboard, and other
// desktop automation capabilities.
//
// Example usage:
//
//	agent, err := cua.New(
//	    cua.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
//	    cua.WithProvider(cua.ProviderAnthropic),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	result, err := agent.Run(ctx, "Open Safari and go to google.com")
package cua

import (
	"context"
	"fmt"

	"github.com/Ingenimax/agent-sdk-go/pkg/agent"
	"github.com/Ingenimax/agent-sdk-go/pkg/interfaces"
	"github.com/Ingenimax/agent-sdk-go/pkg/llm/anthropic"
	"github.com/Ingenimax/agent-sdk-go/pkg/llm/gemini"
	"github.com/Ingenimax/agent-sdk-go/pkg/llm/openai"
	"github.com/Ingenimax/agent-sdk-go/pkg/memory"
	"github.com/Ingenimax/agent-sdk-go/pkg/multitenancy"

	"github.com/anxuanzi/cua/internal/tools"
)

// CUA is the Computer Use Agent that coordinates AI-powered desktop automation.
// It wraps agent-sdk-go's Agent with specialized computer use tools.
type CUA struct {
	config *Config
	agent  *agent.Agent
	tools  []interfaces.Tool
}

// New creates a new CUA instance with the given options.
func New(opts ...Option) (*CUA, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Validate configuration
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Create LLM client based on provider
	var llmClient interfaces.LLM
	var err error

	switch cfg.Provider {
	case ProviderAnthropic:
		model := cfg.Model
		if model == "" {
			model = "claude-sonnet-4-20250514"
		}
		llmClient = anthropic.NewClient(
			cfg.APIKey,
			anthropic.WithModel(model),
		)
	case ProviderOpenAI:
		model := cfg.Model
		if model == "" {
			model = "gpt-4o"
		}
		llmClient = openai.NewClient(
			cfg.APIKey,
			openai.WithModel(model),
		)
	case ProviderGemini:
		model := cfg.Model
		if model == "" {
			model = "gemini-2.5-flash"
		}
		llmClient, err = gemini.NewClient(
			context.Background(),
			gemini.WithAPIKey(cfg.APIKey),
			gemini.WithModel(model),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini client: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}

	// Initialize memory
	mem := memory.NewConversationBuffer()

	// Initialize tools
	toolList := createTools(cfg.ScreenIndex)

	// Create agent with agent-sdk-go
	agentOpts := []agent.Option{
		agent.WithLLM(llmClient),
		agent.WithMemory(mem),
		agent.WithTools(toolList...),
		agent.WithSystemPrompt(systemPrompt),
		agent.WithName("CUA"),
		agent.WithMaxIterations(cfg.MaxIterations),
	}

	// Add LLM config for reasoning if enabled
	if cfg.EnableReasoning {
		agentOpts = append(agentOpts, agent.WithLLMConfig(interfaces.LLMConfig{
			EnableReasoning: true,
			ReasoningBudget: cfg.ReasoningBudget,
		}))
	}

	ag, err := agent.NewAgent(agentOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return &CUA{
		config: cfg,
		agent:  ag,
		tools:  toolList,
	}, nil
}

// createTools initializes all CUA tools.
func createTools(screenIndex int) []interfaces.Tool {
	screenshot := tools.NewScreenshotTool()
	screenshot.ScreenIndex = screenIndex

	click := tools.NewClickTool()
	click.ScreenIndex = screenIndex

	move := tools.NewMoveTool()
	move.ScreenIndex = screenIndex

	drag := tools.NewDragTool()
	drag.ScreenIndex = screenIndex

	scroll := tools.NewScrollTool()
	scroll.ScreenIndex = screenIndex

	return []interfaces.Tool{
		screenshot,
		click,
		move,
		drag,
		scroll,
		tools.NewTypeTool(),
		tools.NewKeyPressTool(),
		tools.NewScreenInfoTool(),
	}
}

// Run executes a task and returns the final result.
// This delegates to agent-sdk-go's agent which handles the ReAct loop.
func (c *CUA) Run(ctx context.Context, task string) (string, error) {
	ctx = multitenancy.WithOrgID(ctx, "CUA-DEFAULT-ORG")
	return c.agent.Run(ctx, task)
}

// RunDetailed executes a task and returns detailed response including token usage.
func (c *CUA) RunDetailed(ctx context.Context, task string) (*interfaces.AgentResponse, error) {
	return c.agent.RunDetailed(ctx, task)
}

// RunEvent represents an event during streaming execution.
type RunEvent struct {
	Type       EventType
	Content    string
	ToolCall   *ToolCallEvent
	ToolResult string
	Thinking   string
	Error      error
}

// ToolCallEvent represents a tool call during streaming.
type ToolCallEvent struct {
	ID        string
	Name      string
	Arguments string
}

// EventType represents the type of run event.
type EventType int

const (
	EventThinking   EventType = iota // Extended thinking/reasoning (Thought phase)
	EventContent                     // Text response generation
	EventToolCall                    // Tool execution initiation (Action phase)
	EventToolResult                  // Tool result (Observation phase)
	EventComplete                    // Completion signal
	EventError                       // Error occurred
)

// RunStream executes a task and streams events back.
// This provides visibility into the ReAct loop: Thought → Action → Observation
func (c *CUA) RunStream(ctx context.Context, task string) (<-chan RunEvent, error) {
	// Create output channel
	events := make(chan RunEvent, 100)

	// Get stream from agent-sdk-go (RunStream is a direct method on Agent)
	agentEvents, err := c.agent.RunStream(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to start stream: %w", err)
	}

	go func() {
		defer close(events)

		for agentEvent := range agentEvents {
			var event RunEvent

			switch agentEvent.Type {
			case interfaces.AgentEventThinking:
				event = RunEvent{
					Type:     EventThinking,
					Thinking: agentEvent.Content,
				}
			case interfaces.AgentEventContent:
				event = RunEvent{
					Type:    EventContent,
					Content: agentEvent.Content,
				}
			case interfaces.AgentEventToolCall:
				if agentEvent.ToolCall != nil {
					event = RunEvent{
						Type: EventToolCall,
						ToolCall: &ToolCallEvent{
							ID:        agentEvent.ToolCall.ID,
							Name:      agentEvent.ToolCall.Name,
							Arguments: agentEvent.ToolCall.Arguments,
						},
					}
				}
			case interfaces.AgentEventToolResult:
				event = RunEvent{
					Type:       EventToolResult,
					ToolResult: agentEvent.Content,
				}
			case interfaces.AgentEventError:
				event = RunEvent{
					Type:  EventError,
					Error: fmt.Errorf("%s", agentEvent.Content),
				}
			case interfaces.AgentEventComplete:
				event = RunEvent{
					Type:    EventComplete,
					Content: agentEvent.Content,
				}
			default:
				continue
			}

			select {
			case events <- event:
			case <-ctx.Done():
				events <- RunEvent{Type: EventError, Error: ctx.Err()}
				return
			}
		}
	}()

	return events, nil
}

// Tools returns the list of available tools.
func (c *CUA) Tools() []interfaces.Tool {
	return c.tools
}

// GetTool returns a specific tool by name.
func (c *CUA) GetTool(name string) (interfaces.Tool, bool) {
	for _, t := range c.tools {
		if t.Name() == name {
			return t, true
		}
	}
	return nil, false
}

// ExecuteTool executes a tool by name with the given arguments.
func (c *CUA) ExecuteTool(ctx context.Context, toolName string, argsJSON string) (string, error) {
	tool, found := c.GetTool(toolName)
	if !found {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}
	return tool.Execute(ctx, argsJSON)
}

// Config returns the current configuration.
func (c *CUA) Config() *Config {
	return c.config
}

// Agent returns the underlying agent-sdk-go agent.
func (c *CUA) Agent() *agent.Agent {
	return c.agent
}

// SystemPrompt returns the system prompt for the CUA agent.
func (c *CUA) SystemPrompt() string {
	return systemPrompt
}

// ToolDefinitions returns JSON-compatible tool definitions for external LLM integration.
func (c *CUA) ToolDefinitions() []map[string]interface{} {
	defs := make([]map[string]interface{}, len(c.tools))
	for i, t := range c.tools {
		params := t.Parameters()
		properties := make(map[string]interface{})
		required := []string{}

		for name, spec := range params {
			prop := map[string]interface{}{
				"type":        spec.Type,
				"description": spec.Description,
			}
			if spec.Enum != nil {
				prop["enum"] = spec.Enum
			}
			if spec.Default != nil {
				prop["default"] = spec.Default
			}
			properties[name] = prop
			if spec.Required {
				required = append(required, name)
			}
		}

		defs[i] = map[string]interface{}{
			"name":        t.Name(),
			"description": t.Description(),
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": properties,
				"required":   required,
			},
		}
	}
	return defs
}

// systemPrompt is the default system prompt for the CUA agent.
// Incorporates Manus context engineering best practices for optimal agent performance.
const systemPrompt = `You are a desktop automation agent that can see and interact with the screen.

COORDINATE SYSTEM:
- All coordinates use a 0-1000 normalized scale
- (0, 0) = top-left corner of the screen
- (1000, 1000) = bottom-right corner of the screen
- (500, 500) = center of the screen
- This scale is resolution-independent and works on any screen size

WORKFLOW (ReAct Pattern):
1. OBSERVE: Take a screenshot to see the current state of the screen
2. THINK: Analyze what you see and plan your next action
3. ACT: Execute ONE action (click, type, scroll, etc.)
4. VERIFY: Take another screenshot to verify the result
5. REPEAT: Continue until the task is complete

AVAILABLE TOOLS:
Screen tools (screen_*):
- screen_capture: Capture the current screen state (USE FREQUENTLY)
- screen_info: Get information about available screens

Mouse tools (mouse_*):
- mouse_click: Click at a normalized position (x, y in 0-1000 range)
- mouse_move: Move cursor without clicking (for hover actions)
- mouse_drag: Drag from one position to another
- mouse_scroll: Scroll at a position in a direction

Keyboard tools (keyboard_*):
- keyboard_type: Type text at the current cursor position
- keyboard_press: Press keyboard keys or combinations (e.g., "cmd+c", "enter")

CONTEXT ENGINEERING (Critical for Long Tasks):

1. TASK RECITATION - Prevent Goal Drift:
   - For multi-step tasks, periodically recite the original objective
   - Before each major action, verify it aligns with the goal
   - If you notice drift, explicitly state: "Refocusing on original task: [task]"

2. ERROR PRESERVATION - Learn from Failures:
   - When an action fails, DO NOT immediately retry the same approach
   - Analyze WHY it failed from the screenshot
   - Document failed attempts mentally to avoid repeating them
   - Try alternative approaches: different coordinates, keyboard shortcuts, or workflows

3. PROGRESSIVE VERIFICATION:
   - After complex sequences, take a screenshot to verify cumulative progress
   - If multiple steps succeeded but the goal isn't achieved, reassess strategy
   - Don't assume success - always verify visually

4. COORDINATE CALIBRATION:
   - Menu bars are typically at y ≈ 20-50
   - Dock (macOS) is at y ≈ 950-1000 (bottom) or x ≈ 0-50 (left)
   - Window title bars are typically at y ≈ 0-30 of the window
   - Buttons and clickable areas should be clicked at their CENTER
   - If a click misses, adjust by ~20-50 units and retry

IMPORTANT GUIDELINES:
1. ALWAYS take a screenshot FIRST to understand the current screen state
2. After each action, take another screenshot to verify the result
3. If an action fails, analyze the screenshot and try a different approach
4. Describe what you see in screenshots before taking actions
5. Use specific coordinates - aim for the CENTER of UI elements
6. For text input, first click to focus the field, then type
7. Wait for UI changes to complete between actions
8. If stuck after 3 attempts, try a completely different approach`
