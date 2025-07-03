package formatter

import (
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/paveg/diagassert/internal/evaluator"
)

func TestColorConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectColor bool
	}{
		{
			name:        "colors enabled by default",
			envVars:     map[string]string{},
			expectColor: true,
		},
		{
			name:        "NO_COLOR disables colors",
			envVars:     map[string]string{"NO_COLOR": "1"},
			expectColor: false,
		},
		{
			name:        "FORCE_COLOR enables colors",
			envVars:     map[string]string{"FORCE_COLOR": "1"},
			expectColor: true,
		},
		{
			name:        "FORCE_COLOR overrides NO_COLOR",
			envVars:     map[string]string{"NO_COLOR": "1", "FORCE_COLOR": "1"},
			expectColor: true, // FORCE_COLOR should override NO_COLOR
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			originalEnvVars := make(map[string]string)
			for key := range tt.envVars {
				originalEnvVars[key] = os.Getenv(key)
			}

			// Set test env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Reset color package state
			color.NoColor = false

			// Create formatter and test
			formatter := NewVisualFormatter()

			if formatter.colorConfig.ColorsEnabled != tt.expectColor {
				t.Errorf("expected colors enabled: %v, got: %v", tt.expectColor, formatter.colorConfig.ColorsEnabled)
			}

			// Restore original env vars
			for key, value := range originalEnvVars {
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}
		})
	}
}

func TestColorOutput(t *testing.T) {
	// Save original NO_COLOR state
	originalNoColor := os.Getenv("NO_COLOR")
	defer func() {
		if originalNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", originalNoColor)
		}
	}()

	tests := []struct {
		name     string
		noColor  bool
		expected map[string]bool // map of strings we expect to find (with/without ANSI codes)
	}{
		{
			name:    "with colors enabled",
			noColor: false,
			expected: map[string]bool{
				"\x1b[":     true, // ANSI escape sequences should be present
				"m":         true, // ANSI color codes end with 'm'
				"false":     true, // Content should still be there
				"ASSERTION": true, // Header text should be there
			},
		},
		{
			name:    "with colors disabled",
			noColor: true,
			expected: map[string]bool{
				"\x1b[":     false, // No ANSI escape sequences
				"false":     true,  // Content should still be there
				"ASSERTION": true,  // Header text should be there
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set NO_COLOR environment variable
			if tt.noColor {
				os.Setenv("NO_COLOR", "1")
			} else {
				os.Unsetenv("NO_COLOR")
			}

			// Reset color package state
			color.NoColor = false

			// Create formatter
			formatter := NewVisualFormatter()

			// Create a test result
			result := &evaluator.ExpressionResult{
				Expression: "x > 10",
				Result:     false,
				Variables:  map[string]interface{}{"x": 5},
			}

			// Generate output
			output := formatter.FormatVisual(result, "test.go", 42, "")

			// Check expectations
			for substring, shouldExist := range tt.expected {
				contains := strings.Contains(output, substring)
				if contains != shouldExist {
					if shouldExist {
						t.Errorf("expected output to contain %q, but it didn't. Output:\n%s", substring, output)
					} else {
						t.Errorf("expected output to NOT contain %q, but it did. Output:\n%s", substring, output)
					}
				}
			}
		})
	}
}

func TestColorHelperFunctions(t *testing.T) {
	formatter := NewVisualFormatter()

	tests := []struct {
		name       string
		function   func(string) string
		input      string
		noColor    bool
		expectAnsi bool
	}{
		{
			name:       "colorizeHeader with colors",
			function:   formatter.colorizeHeader,
			input:      "ASSERTION FAILED",
			noColor:    false,
			expectAnsi: true,
		},
		{
			name:       "colorizeHeader without colors",
			function:   formatter.colorizeHeader,
			input:      "ASSERTION FAILED",
			noColor:    true,
			expectAnsi: false,
		},
		{
			name:       "colorizePipe with colors",
			function:   formatter.colorizePipe,
			input:      "|",
			noColor:    false,
			expectAnsi: true,
		},
		{
			name:       "colorizePipe without colors",
			function:   formatter.colorizePipe,
			input:      "|",
			noColor:    true,
			expectAnsi: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set color configuration
			formatter.colorConfig.ColorsEnabled = !tt.noColor
			color.NoColor = tt.noColor

			result := tt.function(tt.input)

			hasAnsi := strings.Contains(result, "\x1b[")
			if hasAnsi != tt.expectAnsi {
				t.Errorf("expected ANSI codes: %v, got: %v, result: %q", tt.expectAnsi, hasAnsi, result)
			}

			// The input text should always be preserved
			if !strings.Contains(result, tt.input) {
				t.Errorf("expected result to contain input text %q, got: %q", tt.input, result)
			}
		})
	}
}

func TestColorizeValue(t *testing.T) {
	formatter := NewVisualFormatter()

	tests := []struct {
		name       string
		value      string
		isOperator bool
		expected   string // expected color type in the output
	}{
		{
			name:       "true value",
			value:      "true",
			isOperator: false,
			expected:   "green", // Should use TrueColor (green)
		},
		{
			name:       "false value",
			value:      "false",
			isOperator: false,
			expected:   "red", // Should use FalseColor (red)
		},
		{
			name:       "operator value",
			value:      "false",
			isOperator: true,
			expected:   "yellow", // Should use OperatorColor (yellow)
		},
		{
			name:       "variable value",
			value:      "42",
			isOperator: false,
			expected:   "blue", // Should use VariableColor (blue)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter.colorConfig.ColorsEnabled = true
			color.NoColor = false

			result := formatter.colorizeValue(tt.value, tt.isOperator)

			// The result should contain ANSI escape sequences
			if !strings.Contains(result, "\x1b[") {
				t.Errorf("expected colored output to contain ANSI escape sequences, got: %q", result)
			}

			// The original value should be preserved
			if !strings.Contains(result, tt.value) {
				t.Errorf("expected result to contain original value %q, got: %q", tt.value, result)
			}
		})
	}
}

func TestIsOperatorValue(t *testing.T) {
	formatter := NewVisualFormatter()

	tests := []struct {
		name       string
		expression string
		value      string
		expected   bool
	}{
		{
			name:       "comparison operator",
			expression: ">",
			value:      "false",
			expected:   true,
		},
		{
			name:       "equality operator",
			expression: "==",
			value:      "false",
			expected:   true,
		},
		{
			name:       "logical operator",
			expression: "&&",
			value:      "true",
			expected:   true,
		},
		{
			name:       "variable identifier",
			expression: "x",
			value:      "42",
			expected:   false,
		},
		{
			name:       "literal value",
			expression: "10",
			value:      "10",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.isOperatorValue(tt.expression, tt.value)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for expression %q, value %q", tt.expected, result, tt.expression, tt.value)
			}
		})
	}
}
