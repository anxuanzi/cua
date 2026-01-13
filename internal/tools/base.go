// Package tools provides computer use tools for the CUA agent.
// Tools implement the agent-sdk-go interfaces.Tool interface.
package tools

import (
	"encoding/json"
	"fmt"

	"github.com/Ingenimax/agent-sdk-go/pkg/interfaces"
)

// ParameterSpec is an alias to the agent-sdk-go ParameterSpec type.
// This ensures our tools are fully compatible with the SDK.
type ParameterSpec = interfaces.ParameterSpec

// Tool is an alias to the agent-sdk-go Tool interface.
// All CUA tools implement this interface.
type Tool = interfaces.Tool

// BaseTool provides common functionality for CUA tools.
// Embed this in tool implementations. Each tool should override Run().
type BaseTool struct{}

// SuccessResponse creates a standard success response JSON.
func SuccessResponse(data map[string]interface{}) string {
	data["success"] = true
	result, _ := json.Marshal(data)
	return string(result)
}

// ErrorResponse creates a standard error response JSON.
// Errors are returned as observations (not Go errors) so the LLM can learn from them.
func ErrorResponse(message string, suggestion string) string {
	resp := map[string]interface{}{
		"success": false,
		"error":   message,
	}
	if suggestion != "" {
		resp["suggestion"] = suggestion
	}
	result, _ := json.Marshal(resp)
	return string(result)
}

// ParseArgs is a helper to unmarshal JSON arguments into a struct.
func ParseArgs(argsJSON string, dest interface{}) error {
	if argsJSON == "" || argsJSON == "{}" {
		return nil
	}
	return json.Unmarshal([]byte(argsJSON), dest)
}

// ValidateRequired checks if required parameters are present.
func ValidateRequired(args map[string]interface{}, required ...string) error {
	for _, r := range required {
		if _, ok := args[r]; !ok {
			return fmt.Errorf("missing required parameter: %s", r)
		}
	}
	return nil
}
