package tools

import (
	"github.com/anxuanzi/cua/pkg/logging"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

var needHelpLog = logging.NewToolLogger("need_help")

// NeedHelpArgs defines the arguments for the need_help tool.
type NeedHelpArgs struct {
	// Reason explains why help is needed.
	Reason string `json:"reason" jsonschema:"Explanation of why help is needed and what the agent is stuck on"`

	// AttemptsMade describes what was tried before asking for help.
	AttemptsMade string `json:"attempts_made,omitempty" jsonschema:"Description of attempts made before requesting help"`
}

// NeedHelpResult contains the result of the help request.
type NeedHelpResult struct {
	// HelpRequested indicates the help request was registered.
	HelpRequested bool `json:"help_requested"`

	// Reason is the reason for needing help.
	Reason string `json:"reason"`

	// AttemptsMade is what was tried.
	AttemptsMade string `json:"attempts_made,omitempty"`
}

// performNeedHelp handles help requests and signals the loop to exit.
func performNeedHelp(ctx tool.Context, args NeedHelpArgs) (NeedHelpResult, error) {
	needHelpLog.Start("need_help", args.Reason)

	// Signal the LoopAgent to exit - human intervention needed
	ctx.Actions().Escalate = true

	needHelpLog.Info("need_help", "Human assistance requested: "+args.Reason)
	return NeedHelpResult{
		HelpRequested: true,
		Reason:        args.Reason,
		AttemptsMade:  args.AttemptsMade,
	}, nil
}

// NewNeedHelpTool creates the need_help tool for requesting human assistance.
// When called, it sets Escalate=true which causes the LoopAgent to exit.
// Use this when:
// - Stuck after multiple failed attempts
// - Encountering unexpected states
// - Needing clarification on the task
// - Facing security-sensitive operations requiring confirmation
func NewNeedHelpTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "need_help",
			Description: "Call this tool when you are stuck and need human assistance. Explain what you tried and why you need help. This will pause the automation for human intervention.",
		},
		performNeedHelp,
	)
}
