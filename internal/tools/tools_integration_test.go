//go:build integration

package tools

import (
	"context"
	"testing"
	"time"
)

func TestClickToolActualExecution(t *testing.T) {
	clickTool, err := NewClickTool()
	if err != nil {
		t.Fatalf("Failed to create click tool: %v", err)
	}

	t.Logf("Tool definition: %+v", clickTool.Definition())

	// Actually click at 100, 100
	ctx := context.Background()
	result, err := clickTool.Call(ctx, map[string]any{
		"x": 100,
		"y": 100,
	})
	if err != nil {
		t.Fatalf("Tool call error: %v", err)
	}
	t.Logf("Click result: %+v", result)
}

func TestKeyPressToolActualExecution(t *testing.T) {
	keyPressTool, err := NewKeyPressTool()
	if err != nil {
		t.Fatalf("Failed to create key press tool: %v", err)
	}

	t.Logf("Tool definition: %+v", keyPressTool.Definition())

	// Wait a moment then press Cmd+Space
	time.Sleep(2 * time.Second)

	ctx := context.Background()
	result, err := keyPressTool.Call(ctx, map[string]any{
		"key":       "space",
		"modifiers": []string{"cmd"},
	})
	if err != nil {
		t.Fatalf("Tool call error: %v", err)
	}
	t.Logf("Key press result: %+v", result)
}
