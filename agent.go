package cua

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	internalagent "github.com/anxuanzi/cua/internal/agent"
	"github.com/anxuanzi/cua/internal/memory"
	"github.com/anxuanzi/cua/internal/safety"
	"github.com/anxuanzi/cua/pkg/logging"
	"github.com/anxuanzi/cua/pkg/platform"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// Agent is the main entry point for desktop automation.
// Create an Agent using New() with configuration options.
type Agent struct {
	config *Config

	// mu protects running state
	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc

	// ADK components (initialized lazily)
	initOnce    sync.Once
	initErr     error
	coordinator adkagent.Agent
	adkRunner   *runner.Runner
	sessionSvc  session.Service

	// Safety guardrails
	guardrails *safety.Guardrails
}

// newAgent creates a new Agent with the given configuration.
// This is called by New() in cua.go.
func newAgent(opts ...Option) (*Agent, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Configure logging based on verbose setting
	if cfg.verbose {
		logging.SetLevel(logging.LevelDebug)
	}

	// Log platform info at startup
	info := platform.Current()
	logging.Info("CUA starting on %s (%s)", info.DisplayName, info.Arch)

	// Check for API key
	if cfg.apiKey == "" {
		cfg.apiKey = os.Getenv("GOOGLE_API_KEY")
		if cfg.apiKey == "" {
			cfg.apiKey = os.Getenv("GEMINI_API_KEY")
		}
		if cfg.apiKey == "" {
			return nil, ErrNoAPIKey
		}
	}

	// Map SafetyLevel to internal safety.Level
	var safetyLvl safety.Level
	switch cfg.safetyLevel {
	case SafetyMinimal:
		safetyLvl = safety.LevelMinimal
	case SafetyStrict:
		safetyLvl = safety.LevelStrict
	default:
		safetyLvl = safety.LevelNormal
	}

	// Create guardrails with config-based settings
	guardrailsCfg := safety.GuardrailsConfig{
		Level:                  safetyLvl,
		MaxActionsPerMinute:    cfg.rateLimitPerMinute,
		MaxConsecutiveFailures: 5,
	}
	guardrails, err := safety.NewGuardrails(guardrailsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create guardrails: %w", err)
	}

	return &Agent{
		config:     cfg,
		guardrails: guardrails,
	}, nil
}

// initADK initializes the ADK components lazily.
// This is called once on the first Do/DoContext call.
func (a *Agent) initADK(ctx context.Context) error {
	a.initOnce.Do(func() {
		a.initErr = a.doInitADK(ctx)
	})
	return a.initErr
}

// doInitADK performs the actual initialization.
func (a *Agent) doInitADK(ctx context.Context) error {
	// Create the Gemini models
	// Coordinator uses Pro model for complex reasoning
	// Sub-agents use Flash model for speed
	clientCfg := &genai.ClientConfig{
		APIKey: a.config.apiKey,
	}

	// Determine model names based on config
	// The user's model config is used for sub-agents (Flash by default)
	// Coordinator always uses Pro for reasoning
	subAgentModelName := string(a.config.model)
	coordinatorModelName := string(Gemini2Pro)

	// If user specified Pro, use it for both
	if a.config.model == Gemini2Pro || a.config.model == Gemini3Pro {
		coordinatorModelName = string(a.config.model)
	}

	// Import the gemini package dynamically
	coordinatorModel, err := createGeminiModel(ctx, coordinatorModelName, clientCfg)
	if err != nil {
		return fmt.Errorf("failed to create coordinator model: %w", err)
	}

	subAgentModel, err := createGeminiModel(ctx, subAgentModelName, clientCfg)
	if err != nil {
		return fmt.Errorf("failed to create sub-agent model: %w", err)
	}

	// Create the coordinator agent with sub-agents
	coordinator, err := internalagent.NewCoordinatorAgent(coordinatorModel, subAgentModel)
	if err != nil {
		return fmt.Errorf("failed to create coordinator agent: %w", err)
	}
	a.coordinator = coordinator

	// Create session service (in-memory for now)
	a.sessionSvc = session.InMemoryService()

	// Create the ADK runner
	a.adkRunner, err = runner.New(runner.Config{
		AppName:        "cua",
		Agent:          coordinator,
		SessionService: a.sessionSvc,
	})
	if err != nil {
		return fmt.Errorf("failed to create ADK runner: %w", err)
	}

	return nil
}

// Do executes a task and returns the result.
// This is the simplest way to use CUA.
//
// Example:
//
//	agent := cua.New(cua.WithAPIKey("your-key"))
//	result, err := agent.Do("Open Safari and search for golang")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Summary)
func (a *Agent) Do(task string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.timeout)
	defer cancel()
	return a.DoContext(ctx, task)
}

// DoContext executes a task with the given context.
// Use this when you need cancellation or custom timeout control.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// Run in goroutine for async execution
//	go func() {
//	    result, err := agent.DoContext(ctx, "Monitor dashboard")
//	    // handle result
//	}()
//
//	// Cancel after some condition
//	cancel()
func (a *Agent) DoContext(ctx context.Context, task string) (*Result, error) {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return nil, &TaskError{
			Task: task,
			Err:  ErrAgentBusy,
		}
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

	// Initialize ADK components if not done yet
	if err := a.initADK(ctx); err != nil {
		return &Result{
			Success:  false,
			Summary:  "Failed to initialize agent",
			Duration: time.Since(start),
			Error:    err,
		}, err
	}

	// Run the agent with the task
	result, err := a.runTask(ctx, task, nil)
	result.Duration = time.Since(start)

	return result, err
}

// runTask executes the task through the ADK runner.
func (a *Agent) runTask(ctx context.Context, task string, progressCallback ProgressFunc) (*Result, error) {
	// Generate unique IDs for this session
	userID := "cua-user"

	// Create TaskMemory for context engineering
	taskMem := memory.New(task)

	// Create initial session state with task context
	// The coordinator uses {task_context?} templating to inject this
	initialState := map[string]any{
		"task_context": taskMem.ToPrompt(),
	}

	// Create a new session before running - ADK requires sessions to be created first
	createResp, err := a.sessionSvc.Create(ctx, &session.CreateRequest{
		AppName: "cua",
		UserID:  userID,
		State:   initialState,
	})
	if err != nil {
		return &Result{
			Success: false,
			Summary: "Failed to create session",
			Error:   err,
		}, fmt.Errorf("failed to create session: %w", err)
	}
	sessionID := createResp.Session.ID()

	// Create the user message
	userMessage := genai.NewContentFromText(task, genai.RoleUser)

	// Collect steps and final result
	var steps []Step
	var lastResponse string
	var lastError error
	stepNum := 0
	actionCount := 0
	lastActionTime := time.Now()

	// Run the agent and iterate over events
	runCfg := adkagent.RunConfig{}
	for event, err := range a.adkRunner.Run(ctx, userID, sessionID, userMessage, runCfg) {
		if err != nil {
			lastError = err
			break
		}

		// Check action limit
		actionCount++
		if actionCount > a.config.maxActions {
			lastError = ErrMaxActionsExceeded
			break
		}

		// Process the event and update TaskMemory
		step := a.processEvent(event, &stepNum)
		if step != nil {
			// Validate action with guardrails before counting it
			if err := a.guardrails.ValidateAction(step.Action, step.Target, step.Description); err != nil {
				step.Success = false
				step.Error = err
				a.guardrails.RecordFailure(step.Action, step.Target, err)

				// Check for takeover or critical errors
				if err == safety.ErrTakeoverRequested {
					lastError = err
					steps = append(steps, *step)
					break
				}
				if err == safety.ErrConsecutiveFailures {
					lastError = err
					steps = append(steps, *step)
					break
				}
			} else {
				// Record success with guardrails
				a.guardrails.RecordSuccess(step.Action, step.Target, step.Description)
			}

			steps = append(steps, *step)

			// Record action in TaskMemory
			actionDuration := time.Since(lastActionTime)
			taskMem.RecordAction(
				step.Action,
				nil, // args are in step.Target but not structured
				step.Success,
				step.Description,
				actionDuration,
			)
			lastActionTime = time.Now()

			// Check if agent is stuck and needs help
			if taskMem.NeedsHelp() {
				lastError = fmt.Errorf("agent needs help: %d consecutive failures", taskMem.Summary().ConsecutiveFails)
				break
			}

			// Call progress callback if provided
			if progressCallback != nil {
				progressCallback(*step)
			}
		}

		// Extract final response text
		if event.IsFinalResponse() {
			lastResponse = extractTextFromEvent(event)
		}
	}

	// Determine success based on TaskMemory state
	summary := taskMem.Summary()
	success := lastError == nil && len(steps) > 0 && !summary.IsStuck

	// Build summary from TaskMemory or fallback to lastResponse
	resultSummary := lastResponse
	if resultSummary == "" {
		if success {
			if len(summary.Milestones) > 0 {
				resultSummary = fmt.Sprintf("Task completed in %d steps. Milestones: %s",
					summary.TotalSteps, strings.Join(summary.Milestones, ", "))
			} else {
				resultSummary = fmt.Sprintf("Task completed in %d steps", summary.TotalSteps)
			}
		} else if lastError != nil {
			resultSummary = fmt.Sprintf("Task failed: %v", lastError)
		} else {
			resultSummary = "Task completed with no response"
		}
	}

	return &Result{
		Success: success,
		Summary: resultSummary,
		Steps:   steps,
		Error:   lastError,
	}, lastError
}

// processEvent converts an ADK event to a Step.
func (a *Agent) processEvent(event *session.Event, stepNum *int) *Step {
	// Check for function calls in Content.Parts or agent transfers
	var functionCall *genai.FunctionCall
	if event.Content != nil {
		for _, part := range event.Content.Parts {
			if part.FunctionCall != nil {
				functionCall = part.FunctionCall
				break
			}
		}
	}

	// Skip events without actions
	if functionCall == nil && event.Actions.TransferToAgent == "" {
		return nil
	}

	*stepNum++

	step := &Step{
		Number:  *stepNum,
		Success: true,
	}

	// Handle function calls (tool executions)
	if functionCall != nil {
		step.Action = functionCall.Name
		step.Description = fmt.Sprintf("Executed %s", functionCall.Name)

		// Extract target from args if available
		if args := functionCall.Args; args != nil {
			if x, hasX := args["x"]; hasX {
				if y, hasY := args["y"]; hasY {
					step.Target = fmt.Sprintf("(%v, %v)", x, y)
				}
			}
			if text, hasText := args["text"]; hasText {
				step.Target = fmt.Sprintf("%v", text)
			}
		}
	}

	// Handle agent transfers
	if event.Actions.TransferToAgent != "" {
		step.Action = "transfer"
		step.Description = fmt.Sprintf("Transferred to %s", event.Actions.TransferToAgent)
		step.Target = event.Actions.TransferToAgent
	}

	return step
}

// extractTextFromEvent extracts text content from an event.
func extractTextFromEvent(event *session.Event) string {
	if event.Content == nil {
		return ""
	}

	var texts []string
	for _, part := range event.Content.Parts {
		if part.Text != "" {
			texts = append(texts, part.Text)
		}
	}
	return strings.Join(texts, " ")
}

// DoWithProgress executes a task and calls the callback after each step.
// Use this when you want to monitor progress or display updates.
//
// Example:
//
//	err := agent.DoWithProgress("Fill out the form", func(step cua.Step) {
//	    fmt.Printf("Step %d: %s\n", step.Number, step.Description)
//	    if step.Screenshot != nil {
//	        // Save or display screenshot
//	    }
//	})
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
		return &TaskError{
			Task: task,
			Err:  ErrAgentBusy,
		}
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

	// Initialize ADK components if not done yet
	if err := a.initADK(ctx); err != nil {
		return err
	}

	// Run the task with progress callback
	_, err := a.runTask(ctx, task, callback)
	return err
}

// Stop gracefully stops the currently running task.
// It sends a cancellation signal and waits for the task to stop.
//
// This is safe to call even if no task is running.
func (a *Agent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running || a.cancel == nil {
		return nil // Nothing to stop
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
// This is useful for debugging or logging.
func (a *Agent) Config() Config {
	return *a.config
}
