package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDragTool(t *testing.T) {
	t.Parallel()

	tool, err := NewDragTool()

	require.NoError(t, err)
	assert.NotNil(t, tool)
}

func TestDragArgs_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        DragArgs
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid drag - horizontal",
			args: DragArgs{
				StartX: 100,
				StartY: 100,
				EndX:   200,
				EndY:   100,
			},
			expectError: false,
		},
		{
			name: "valid drag - vertical",
			args: DragArgs{
				StartX: 100,
				StartY: 100,
				EndX:   100,
				EndY:   200,
			},
			expectError: false,
		},
		{
			name: "valid drag - diagonal",
			args: DragArgs{
				StartX: 100,
				StartY: 100,
				EndX:   200,
				EndY:   200,
			},
			expectError: false,
		},
		{
			name: "invalid drag - same start and end",
			args: DragArgs{
				StartX: 100,
				StartY: 100,
				EndX:   100,
				EndY:   100,
			},
			expectError: true,
			errorMsg:    "start and end coordinates cannot be the same",
		},
		{
			name: "invalid drag - negative start X",
			args: DragArgs{
				StartX: -10,
				StartY: 100,
				EndX:   200,
				EndY:   100,
			},
			expectError: true,
			errorMsg:    "start_x cannot be negative",
		},
		{
			name: "invalid drag - negative start Y",
			args: DragArgs{
				StartX: 100,
				StartY: -10,
				EndX:   200,
				EndY:   100,
			},
			expectError: true,
			errorMsg:    "start_y cannot be negative",
		},
		{
			name: "invalid drag - negative end X",
			args: DragArgs{
				StartX: 100,
				StartY: 100,
				EndX:   -10,
				EndY:   100,
			},
			expectError: true,
			errorMsg:    "end_x cannot be negative",
		},
		{
			name: "invalid drag - negative end Y",
			args: DragArgs{
				StartX: 100,
				StartY: 100,
				EndX:   200,
				EndY:   -10,
			},
			expectError: true,
			errorMsg:    "end_y cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDragArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDragResult_Fields(t *testing.T) {
	t.Parallel()

	// Test successful result
	successResult := DragResult{
		Success: true,
		StartX:  100,
		StartY:  100,
		EndX:    200,
		EndY:    200,
	}

	assert.True(t, successResult.Success)
	assert.Equal(t, 100, successResult.StartX)
	assert.Equal(t, 100, successResult.StartY)
	assert.Equal(t, 200, successResult.EndX)
	assert.Equal(t, 200, successResult.EndY)
	assert.Empty(t, successResult.Error)

	// Test failed result
	failedResult := DragResult{
		Success: false,
		StartX:  100,
		StartY:  100,
		EndX:    200,
		EndY:    200,
		Error:   "drag operation failed",
	}

	assert.False(t, failedResult.Success)
	assert.Equal(t, "drag operation failed", failedResult.Error)
}

func TestPerformDrag_ValidationFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     DragArgs
		errorMsg string
	}{
		{
			name: "same coordinates",
			args: DragArgs{
				StartX: 100,
				StartY: 100,
				EndX:   100,
				EndY:   100,
			},
			errorMsg: "start and end coordinates cannot be the same",
		},
		{
			name: "negative start X",
			args: DragArgs{
				StartX: -1,
				StartY: 100,
				EndX:   200,
				EndY:   100,
			},
			errorMsg: "start_x cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := performDrag(nil, tt.args)

			// performDrag should return validation errors in result, not as error
			assert.NoError(t, err)
			assert.False(t, result.Success)
			assert.Contains(t, result.Error, tt.errorMsg)
		})
	}
}
