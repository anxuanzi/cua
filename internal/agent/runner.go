package agent

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/anxuanzi/cua/internal/memory"
	"github.com/anxuanzi/cua/internal/safety"
	"github.com/anxuanzi/cua/pkg/logging"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// RunConfig holds configuration for running a task.
type RunConfig struct {
	// MaxActions is the maximum number of actions allowed.
	MaxActions int

	// Guardrails provides safety validation.
	Guardrails *safety.Guardrails

	// ProgressCallback is called after each step (optional).
	ProgressCallback func(step Step)
}

// Step represents a single action taken by the agent.
type Step struct {
	Number      int
	Action      string
	Description string
	Target      string
	Success     bool
	Duration    time.Duration
	Error       error
}

// Result contains the outcome of a completed task.
type Result struct {
	Success  bool
	Summary  string
	Steps    []Step
	Duration time.Duration
	Error    error
}

// Runner executes tasks through the ADK agent system.
type Runner struct {
	adkRunner  *runner.Runner
	sessionSvc session.Service
}

// NewRunner creates a new task runner.
func NewRunner(coordinator adkagent.Agent) (*Runner, error) {
	sessionSvc := session.InMemoryService()

	adkRunner, err := runner.New(runner.Config{
		AppName:        "cua",
		Agent:          coordinator,
		SessionService: sessionSvc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ADK runner: %w", err)
	}

	return &Runner{
		adkRunner:  adkRunner,
		sessionSvc: sessionSvc,
	}, nil
}

// getPlatformName returns the platform identifier for the instruction template.
func getPlatformName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	default:
		return runtime.GOOS
	}
}

// Run executes a task and returns the result.
func (r *Runner) Run(ctx context.Context, task string, cfg RunConfig) (*Result, error) {
	userID := "cua-user"

	// Create TaskMemory for context engineering
	taskMem := memory.New(task)

	// Create session with task context and platform info
	// These values are injected into the instruction template via {task_context} and {platform}
	initialState := map[string]any{
		"task_context": taskMem.ToPrompt(),
		"platform":     getPlatformName(),
	}

	createResp, err := r.sessionSvc.Create(ctx, &session.CreateRequest{
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
	sess := createResp.Session
	sessionID := sess.ID()

	// Execute task
	var steps []Step
	var lastResponse string
	var lastError error
	stepNum := 0
	actionCount := 0
	lastActionTime := time.Now()

	taskMessage := genai.NewContentFromText(task, genai.RoleUser)
	runCfg := adkagent.RunConfig{}

	logging.Info("Starting task: %s", task)

	for event, err := range r.adkRunner.Run(ctx, userID, sessionID, taskMessage, runCfg) {
		if err != nil {
			// Log and record the error
			logging.Error("Runner error: %v", err)

			// Unknown tool errors are non-fatal - model tried to call a tool we don't have
			if strings.Contains(err.Error(), "unknown tool") {
				logging.Warn("Model called unknown tool, continuing...")
				// Don't break - let the loop continue to next event
				// But we need a valid event to process, so skip this iteration
				if event == nil {
					continue
				}
			} else if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") {
				// Rate limit error - fatal, but with a helpful message
				logging.Warn("Rate limit hit. Please wait and try again.")
				lastError = fmt.Errorf("API rate limit exceeded - please wait 1 minute and retry")
				break
			} else {
				// Other errors are fatal
				lastError = err
				break
			}
		}

		// Skip nil events
		if event == nil {
			continue
		}

		actionCount++
		if cfg.MaxActions > 0 && actionCount > cfg.MaxActions {
			lastError = fmt.Errorf("exceeded maximum actions: %d", cfg.MaxActions)
			logging.Warn("Max actions exceeded: %d", cfg.MaxActions)
			break
		}

		// Process event into step
		step := processEvent(event, &stepNum)
		if step != nil {
			// Validate with guardrails
			if cfg.Guardrails != nil {
				if err := cfg.Guardrails.ValidateAction(step.Action, step.Target, step.Description); err != nil {
					step.Success = false
					step.Error = err
					cfg.Guardrails.RecordFailure(step.Action, step.Target, err)

					if err == safety.ErrTakeoverRequested || err == safety.ErrConsecutiveFailures {
						lastError = err
						steps = append(steps, *step)
						break
					}
				} else {
					cfg.Guardrails.RecordSuccess(step.Action, step.Target, step.Description)
				}
			}

			steps = append(steps, *step)

			// Update TaskMemory
			actionDuration := time.Since(lastActionTime)
			taskMem.RecordAction(step.Action, nil, step.Success, step.Description, actionDuration)
			lastActionTime = time.Now()

			// Update session state with new context (for next iteration)
			// This ensures the {task_context} placeholder gets fresh data
			if err := sess.State().Set("task_context", taskMem.ToPrompt()); err != nil {
				logging.Warn("Failed to update session state: %v", err)
			}

			// Check if stuck
			if taskMem.NeedsHelp() {
				lastError = fmt.Errorf("agent needs help: %d consecutive failures", taskMem.Summary().ConsecutiveFails)
				break
			}

			// Progress callback
			if cfg.ProgressCallback != nil {
				cfg.ProgressCallback(*step)
			}
		}

		// Extract response
		if event.IsFinalResponse() {
			lastResponse = extractTextFromEvent(event)
			logging.Debug("Final response: %s", lastResponse)
		}
	}

	// Build result
	summary := taskMem.Summary()
	success := lastError == nil && len(steps) > 0 && !summary.IsStuck

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

	logging.Info("Task finished: success=%v, steps=%d", success, len(steps))

	return &Result{
		Success: success,
		Summary: resultSummary,
		Steps:   steps,
		Error:   lastError,
	}, lastError
}

// processEvent converts an ADK event to a Step.
func processEvent(event *session.Event, stepNum *int) *Step {
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

	if functionCall != nil {
		step.Action = functionCall.Name
		step.Description = fmt.Sprintf("Executed %s", functionCall.Name)

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
