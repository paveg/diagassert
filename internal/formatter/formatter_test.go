package formatter

import (
	"os"
	"strings"
	"testing"
)

func TestBuildDiagnosticOutput(t *testing.T) {
	tests := []struct {
		name           string
		file           string
		line           int
		expr           string
		opts           Options
		expectedParts  []string
		notExpected    []string
	}{
		{
			name: "basic output without machine readable",
			file: "/path/to/test.go",
			line: 42,
			expr: "x > 20",
			opts: Options{
				IncludeMachineReadable: false,
				Format:                 "human",
			},
			expectedParts: []string{
				"ASSERTION FAILED at test.go:42",
				"Expression: x > 20",
				"Result: false",
			},
			notExpected: []string{
				"[MACHINE_READABLE_START]",
			},
		},
		{
			name: "output with machine readable section",
			file: "/path/to/test.go",
			line: 42,
			expr: "x > 20",
			opts: Options{
				IncludeMachineReadable: true,
				Format:                 "hybrid",
			},
			expectedParts: []string{
				"ASSERTION FAILED at test.go:42",
				"Expression: x > 20",
				"Result: false",
				"[MACHINE_READABLE_START]",
				"EXPR: x > 20",
				"RESULT: false",
				"[MACHINE_READABLE_END]",
			},
		},
		{
			name: "complex expression",
			file: "/path/to/user_test.go",
			line: 123,
			expr: "user.Age >= 18 && user.HasLicense()",
			opts: Options{
				IncludeMachineReadable: true,
				Format:                 "hybrid",
			},
			expectedParts: []string{
				"ASSERTION FAILED at user_test.go:123",
				"Expression: user.Age >= 18 && user.HasLicense()",
				"Result: false",
				"[MACHINE_READABLE_START]",
				"EXPR: user.Age >= 18 && user.HasLicense()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDiagnosticOutput(tt.file, tt.line, tt.expr, tt.opts)

			// Check that all expected parts are present
			for _, expected := range tt.expectedParts {
				if !strings.Contains(result, expected) {
					t.Errorf("BuildDiagnosticOutput() missing expected part: %q\nFull output:\n%s", expected, result)
				}
			}

			// Check that none of the not-expected parts are present
			for _, notExpected := range tt.notExpected {
				if strings.Contains(result, notExpected) {
					t.Errorf("BuildDiagnosticOutput() contains unexpected part: %q\nFull output:\n%s", notExpected, result)
				}
			}
		})
	}
}

func TestShouldIncludeMachineReadable(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "default (no env var)",
			envValue: "",
			expected: true,
		},
		{
			name:     "explicitly true",
			envValue: "true",
			expected: true,
		},
		{
			name:     "explicitly false",
			envValue: "false",
			expected: false,
		},
		{
			name:     "other value",
			envValue: "something",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment variable
			originalValue := os.Getenv("DIAGASSERT_MACHINE_READABLE")
			defer func() {
				if originalValue == "" {
					os.Unsetenv("DIAGASSERT_MACHINE_READABLE")
				} else {
					os.Setenv("DIAGASSERT_MACHINE_READABLE", originalValue)
				}
			}()

			// Set test environment value
			if tt.envValue == "" {
				os.Unsetenv("DIAGASSERT_MACHINE_READABLE")
			} else {
				os.Setenv("DIAGASSERT_MACHINE_READABLE", tt.envValue)
			}

			result := ShouldIncludeMachineReadable()
			if result != tt.expected {
				t.Errorf("ShouldIncludeMachineReadable() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetDefaultOptions(t *testing.T) {
	opts := GetDefaultOptions()

	if opts.Format != "hybrid" {
		t.Errorf("GetDefaultOptions().Format = %q, expected %q", opts.Format, "hybrid")
	}

	// Machine readable should be determined by environment
	expected := ShouldIncludeMachineReadable()
	if opts.IncludeMachineReadable != expected {
		t.Errorf("GetDefaultOptions().IncludeMachineReadable = %v, expected %v", opts.IncludeMachineReadable, expected)
	}
}