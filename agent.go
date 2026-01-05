package cua

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	internalagent "github.com/anxuanzi/cua/internal/agent"
	"github.com/anxuanzi/cua/internal/safety"
	"github.com/anxuanzi/cua/pkg/logging"
	"github.com/anxuanzi/cua/pkg/platform"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// Agent is the main entry point for desktop automation.
// Create an Agent using New() with configuration options.
type Agent struct {
	config *Config

	// Protects running state
	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc

	// Initialized lazily
	initOnce   sync.Once
	initErr    error
	runner     *internalagent.Runner
	guardrails *safety.Guardrails
}

// newAgent creates a new Agent with the given configuration.
func newAgent(opts ...Option) (*Agent, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Configure logging
	if cfg.verbose {
		logging.SetLevel(logging.LevelDebug)
	}

	// Log platform info
	info := platform.Current()
	logging.Info("CUA starting on %s (%s)", info.DisplayName, info.Arch)

	// Get API key
	if cfg.apiKey == "" {
		cfg.apiKey = os.Getenv("GOOGLE_API_KEY")
		if cfg.apiKey == "" {
			cfg.apiKey = os.Getenv("GEMINI_API_KEY")
		}
		if cfg.apiKey == "" {
			return nil, ErrNoAPIKey
		}
	}

	return &Agent{config: cfg}, nil
}

// initOnce initializes ADK components lazily.
func (a *Agent) init(ctx context.Context) error {
	a.initOnce.Do(func() {
		a.initErr = a.doInit(ctx)
	})
	return a.initErr
}

// doInit performs the actual initialization.
func (a *Agent) doInit(ctx context.Context) error {
	// Create Gemini models
	clientCfg := &genai.ClientConfig{APIKey: a.config.apiKey}

	subAgentModelName := string(a.config.model)
	coordinatorModelName := string(Gemini2Pro)

	if a.config.model == Gemini2Pro || a.config.model == Gemini3Pro {
		coordinatorModelName = string(a.config.model)
	}

	coordinatorModel, err := gemini.NewModel(ctx, coordinatorModelName, clientCfg)
	if err != nil {
		return fmt.Errorf("failed to create coordinator model: %w", err)
	}

	subAgentModel, err := gemini.NewModel(ctx, subAgentModelName, clientCfg)
	if err != nil {
		return fmt.Errorf("failed to create sub-agent model: %w", err)
	}

	// Create coordinator agent
	coordinator, err := internalagent.NewCoordinatorAgent(coordinatorModel, subAgentModel)
	if err != nil {
		return fmt.Errorf("failed to create coordinator: %w", err)
	}

	// Create runner
	a.runner, err = internalagent.NewRunner(coordinator)
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	// Create guardrails
	var safetyLvl safety.Level
	switch a.config.safetyLevel {
	case SafetyMinimal:
		safetyLvl = safety.LevelMinimal
	case SafetyStrict:
		safetyLvl = safety.LevelStrict
	default:
		safetyLvl = safety.LevelNormal
	}

	a.guardrails, err = safety.NewGuardrails(safety.GuardrailsConfig{
		Level:                  safetyLvl,
		MaxActionsPerMinute:    a.config.rateLimitPerMinute,
		MaxConsecutiveFailures: 5,
	})
	if err != nil {
		return fmt.Errorf("failed to create guardrails: %w", err)
	}

	return nil
}

// Do executes a task and returns the result.
//
// Example:
//
//	agent := cua.New(cua.WithAPIKey("your-key"))
//	result, err := agent.Do("Open Safari and search for golang")
func (a *Agent) Do(task string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.timeout)
	defer cancel()
	return a.DoContext(ctx, task)
}

// DoContext executes a task with the given context.
func (a *Agent) DoContext(ctx context.Context, task string) (*Result, error) {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return nil, &TaskError{Task: task, Err: ErrAgentBusy}
	}
	a.running = true
	ctx, a.cancel = context.WithCancel(ctx)
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.running = false
		a.cancel = nil
		a.mu.Unlock()
	}()

	start := time.Now()

	// Initialize if needed
	if err := a.init(ctx); err != nil {
		return &Result{
			Success:  false,
			Summary:  "Failed to initialize agent",
			Duration: time.Since(start),
			Error:    err,
		}, err
	}

	// Run the task
	result, err := a.runner.Run(ctx, task, internalagent.RunConfig{
		MaxActions: a.config.maxActions,
		Guardrails: a.guardrails,
	})

	// Convert internal result to public result
	return a.convertResult(result, time.Since(start)), err
}

// DoWithProgress executes a task and calls the callback after each step.
func (a *Agent) DoWithProgress(task string, callback ProgressFunc) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.timeout)
	defer cancel()
	return a.DoWithProgressContext(ctx, task, callback)
}

// DoWithProgressContext executes a task with progress callbacks and custom context.
func (a *Agent) DoWithProgressContext(ctx context.Context, task string, callback ProgressFunc) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return &TaskError{Task: task, Err: ErrAgentBusy}
	}
	a.running = true
	ctx, a.cancel = context.WithCancel(ctx)
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.running = false
		a.cancel = nil
		a.mu.Unlock()
	}()

	if err := a.init(ctx); err != nil {
		return err
	}

	// Convert callback
	var internalCallback func(internalagent.Step)
	if callback != nil {
		internalCallback = func(s internalagent.Step) {
			callback(Step{
				Number:      s.Number,
				Action:      s.Action,
				Description: s.Description,
				Target:      s.Target,
				Success:     s.Success,
				Duration:    s.Duration,
				Error:       s.Error,
			})
		}
	}

	_, err := a.runner.Run(ctx, task, internalagent.RunConfig{
		MaxActions:       a.config.maxActions,
		Guardrails:       a.guardrails,
		ProgressCallback: internalCallback,
	})
	return err
}

// Stop gracefully stops the currently running task.
func (a *Agent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running || a.cancel == nil {
		return nil
	}

	a.cancel()
	return nil
}

// IsRunning returns true if a task is currently executing.
func (a *Agent) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}

// Config returns a copy of the agent's configuration.
func (a *Agent) Config() Config {
	return *a.config
}

// convertResult converts internal result to public Result.
func (a *Agent) convertResult(r *internalagent.Result, duration time.Duration) *Result {
	if r == nil {
		return &Result{
			Success:  false,
			Summary:  "No result",
			Duration: duration,
		}
	}

	steps := make([]Step, len(r.Steps))
	for i, s := range r.Steps {
		steps[i] = Step{
			Number:      s.Number,
			Action:      s.Action,
			Description: s.Description,
			Target:      s.Target,
			Success:     s.Success,
			Duration:    s.Duration,
			Error:       s.Error,
		}
	}

	return &Result{
		Success:  r.Success,
		Summary:  r.Summary,
		Steps:    steps,
		Duration: duration,
		Error:    r.Error,
	}
}
