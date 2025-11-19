package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		request     GenerateRequest
		expectError bool
		errorType   error
	}{
		{
			name: "Valid request",
			request: GenerateRequest{
				Scenario: "Generate user data",
				RowCount: 10,
			},
			expectError: false,
		},
		{
			name: "Empty scenario",
			request: GenerateRequest{
				Scenario: "",
				RowCount: 10,
			},
			expectError: true,
			errorType:   ErrInvalidScenario,
		},
		{
			name: "Row count too low",
			request: GenerateRequest{
				Scenario: "Generate user data",
				RowCount: 0,
			},
			expectError: true,
			errorType:   ErrInvalidRowCount,
		},
		{
			name: "Row count too high",
			request: GenerateRequest{
				Scenario: "Generate user data",
				RowCount: 1001,
			},
			expectError: true,
			errorType:   ErrInvalidRowCount,
		},
		{
			name: "Minimum valid row count",
			request: GenerateRequest{
				Scenario: "Test",
				RowCount: 1,
			},
			expectError: false,
		},
		{
			name: "Maximum valid row count",
			request: GenerateRequest{
				Scenario: "Test",
				RowCount: 1000,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectError {
				assert.Error(t, err, "Should return an error")
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err, "Should return correct error type")
				}
			} else {
				assert.NoError(t, err, "Should not return an error")
			}
		})
	}
}


// TestCustomErrors tests that custom errors are defined
func TestCustomErrors(t *testing.T) {
	// Verify error messages are meaningful
	assert.Contains(t, ErrInvalidScenario.Error(), "scenario")
	assert.Contains(t, ErrInvalidRowCount.Error(), "row count")
	assert.Contains(t, ErrRequestNotFound.Error(), "not found")
	assert.Contains(t, ErrDatasetNotFound.Error(), "dataset")
	assert.Contains(t, ErrOpenAIFailure.Error(), "OpenAI")
	assert.Contains(t, ErrDatabaseConnection.Error(), "database")
	assert.Contains(t, ErrInvalidFormat.Error(), "format")
}

