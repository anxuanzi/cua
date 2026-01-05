package tools

import (
	"github.com/anxuanzi/cua/pkg/logging"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

var completeTaskLog = logging.NewToolLogger("complete_task")

// CompleteTaskArgs defines the arguments for the complete_task tool.
type CompleteTaskArgs struct {
	// Summary is a brief description of what was accomplished.
	Summary string `json:"summary" jsonschema:"A brief summary of what was accomplished in completing the task"`
}

// CompleteTaskResult contains the result of task completion.
type CompleteTaskResult struct {
	// Success indicates the task was marked complete.
	Success bool `json:"success"`

	// Summary is the completion summary.
	Summary string `json:"summary"`
}

// performCompleteTask handles task completion and signals the loop to exit.
func performCompleteTask(ctx tool.Context, args CompleteTaskArgs) (CompleteTaskResult, error) {
	completeTaskLog.Start("complete_task", args.Summary)

	// Signal the LoopAgent to exit
	ctx.Actions().Escalate = true

	completeTaskLog.Success("complete_task", "Task completed successfully")
	return CompleteTaskResult{
		Success: true,
		Summary: args.Summary,
	}, nil
}

// NewCompleteTaskTool creates the complete_task tool for signaling successful task completion.
// When called, it sets Escalate=true which causes the LoopAgent to exit.
func NewCompleteTaskTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "complete_task",
			Description: "Call this tool ONLY when the user's task has been FULLY completed. Provide a summary of what was accomplished. This will end the automation session.",
		},
		performCompleteTask,
	)
}
