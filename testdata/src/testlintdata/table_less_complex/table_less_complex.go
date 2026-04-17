package tablelesscomplex

import (
	"testing"
)

// BAD: Conditional logic in subtest based on table fields
func TestBadConditionalLogic(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOutput string
		shouldErr  bool
		expectCall bool
	}{
		{
			name:       "success case",
			input:      "valid",
			wantOutput: "result",
			shouldErr:  false,
			expectCall: true,
		},
		{
			name:      "error case",
			input:     "invalid",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processInput(tt.input)

			// This is BAD: multiple conditional checks on table fields
			if tt.shouldErr { // want "table-driven test contains conditional logic on table fields; consider splitting into separate tests"
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectCall { // want "table-driven test contains conditional logic on table fields; consider splitting into separate tests"
				if result != tt.wantOutput {
					t.Errorf("got %s, want %s", result, tt.wantOutput)
				}
			}
		})
	}
}

// GOOD: Simple table test with conditional wantErr is acceptable
func TestGoodSimple(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "simple case",
			input:      "test",
			wantOutput: "output",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processInput(tt.input)

			// This pattern is GOOD: single wantErr branch for success/failure
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.wantOutput {
				t.Errorf("got %s, want %s", result, tt.wantOutput)
			}
		})
	}
}

// BAD: Conditional with expected result checking
func TestBadExpectResult(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult string
		skipValidation bool
	}{
		{
			name:           "case 1",
			input:          "test",
			expectedResult: "output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processInput(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.skipValidation { // want "table-driven test contains conditional logic on table fields; consider splitting into separate tests"
				return
			}

			if tt.expectedResult != "" {
				if result != tt.expectedResult {
					t.Errorf("got %s, want %s", result, tt.expectedResult)
				}
			}
		})
	}
}

// GOOD: Separate tests instead of conditional logic
func TestGoodSeparateTests(t *testing.T) {
	successCases := []struct {
		name       string
		input      string
		wantOutput string
	}{
		{
			name:       "case 1",
			input:      "test",
			wantOutput: "output",
		},
	}

	for _, tt := range successCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processInput(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.wantOutput {
				t.Errorf("got %s, want %s", result, tt.wantOutput)
			}
		})
	}
}

func TestGoodErrorCases(t *testing.T) {
	errorCases := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid input",
			input: "invalid",
		},
	}

	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processInput(tt.input)
			if err == nil {
				t.Errorf("expected error but got none")
			}
		})
	}
}

// Helper function for testing
func processInput(input string) (string, error) {
	if input == "invalid" {
		return "", newTestError("invalid input")
	}
	return "output", nil
}

type testError string

func newTestError(msg string) error {
	return testError(msg)
}

func (e testError) Error() string {
	return string(e)
}
