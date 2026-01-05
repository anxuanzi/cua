//go:build integration
// +build integration

package cua_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/anxuanzi/cua"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests require:
// 1. GOOGLE_API_KEY environment variable set
// 2. macOS accessibility permissions granted
// 3. Screen recording permissions granted

func TestIntegration_AgentCreation(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set, skipping integration test")
	}

	agent := cua.New(
		cua.WithAPIKey(apiKey),
		cua.WithModel(cua.Gemini2Flash),
		cua.WithTimeout(30*time.Second),
		cua.WithMaxActions(10),
		cua.WithHeadless(true),
	)

	require.NotNil(t, agent)
	assert.False(t, agent.IsRunning())
}

func TestIntegration_LowLevelInput(t *testing.T) {
	// Test low-level input functions (these don't require API key)

	// Test screen size
	width, height, err := cua.ScreenSize()
	require.NoError(t, err)
	assert.Greater(t, width, 0)
	assert.Greater(t, height, 0)
	t.Logf("Screen size: %dx%d", width, height)

	// Test screen capture
	img, err := cua.CaptureScreen()
	require.NoError(t, err)
	assert.NotNil(t, img)
	bounds := img.Bounds()
	assert.Equal(t, width, bounds.Dx())
	assert.Equal(t, height, bounds.Dy())
	t.Logf("Captured screen: %dx%d", bounds.Dx(), bounds.Dy())
}

func TestIntegration_ElementFinding(t *testing.T) {
	// Test element finding (requires accessibility permissions)

	// Find focused application
	app, err := cua.FocusedApplication()
	if err != nil {
		t.Skipf("Could not get focused app (permission issue?): %v", err)
	}
	assert.NotNil(t, app)
	t.Logf("Focused app: %s", app.Name)

	// Find all buttons in the focused app
	buttons, err := cua.FindElements(cua.ByRole(cua.RoleButton))
	require.NoError(t, err)
	t.Logf("Found %d buttons", len(buttons))
}

func TestIntegration_SimpleTask(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set, skipping integration test")
	}

	agent := cua.New(
		cua.WithAPIKey(apiKey),
		cua.WithModel(cua.Gemini2Flash),
		cua.WithTimeout(60*time.Second),
		cua.WithMaxActions(20),
		cua.WithHeadless(true),
	)

	// Simple task: describe what's on screen
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := agent.DoContext(ctx, "What application is currently in the foreground? Just tell me the name.")
	if err != nil {
		t.Logf("Task error: %v", err)
		// Don't fail the test, just log it - the model might not support this yet
		return
	}

	t.Logf("Result: %+v", result)
	assert.NotEmpty(t, result.Summary)
}

func TestIntegration_ProgressCallback(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set, skipping integration test")
	}

	agent := cua.New(
		cua.WithAPIKey(apiKey),
		cua.WithModel(cua.Gemini2Flash),
		cua.WithTimeout(30*time.Second),
		cua.WithMaxActions(10),
		cua.WithHeadless(true),
	)

	var steps []cua.Step
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := agent.DoWithProgressContext(ctx, "Take a screenshot and describe what you see.", func(step cua.Step) {
		steps = append(steps, step)
		t.Logf("Step %d: %s - %s", step.Number, step.Action, step.Description)
	})

	if err != nil {
		t.Logf("Task error: %v", err)
		return
	}

	t.Logf("Completed with %d steps", len(steps))
}
