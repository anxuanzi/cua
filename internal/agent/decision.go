package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// ExitLoopArgs defines the arguments for the exit_loop tool.
type ExitLoopArgs struct {
	Summary string `json:"summary" jsonschema:"A brief summary of what was accomplished"`
}

// ExitLoopResult defines the result of the exit_loop tool.
type ExitLoopResult struct {
	Success bool   `json:"success"`
	Summary string `json:"summary"`
}

// exitLoop is the tool that signals the ReAct loop to terminate.
func exitLoop(ctx tool.Context, args ExitLoopArgs) (ExitLoopResult, error) {
	// Signal to exit the loop
	ctx.Actions().Escalate = true
	return ExitLoopResult{
		Success: true,
		Summary: args.Summary,
	}, nil
}

// NewDecisionAgent creates the Decision Agent for task reasoning.
// It uses the coordinator model (Pro) for complex reasoning.
// The decision agent reads screen state and decides the next action.
func NewDecisionAgent(m model.LLM) (agent.Agent, error) {
	// Create the exit_loop tool
	exitLoopTool, err := functiontool.New(
		functiontool.Config{
			Name:        "exit_loop",
			Description: "Call this tool ONLY when the user's task has been fully completed. Provide a summary of what was accomplished.",
		},
		exitLoop,
	)
	if err != nil {
		return nil, err
	}

	instruction := BuildDecisionInstruction()

	return llmagent.New(llmagent.Config{
		Name:        "decision",
		Model:       m,
		Description: "Decides the next action based on screen state and task goal.",
		Instruction: instruction,
		Tools: []tool.Tool{
			exitLoopTool,
		},
		OutputKey: "next_action",
	})
}
