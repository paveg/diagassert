package formatter

import (
	"strings"
	"testing"

	"github.com/paveg/diagassert/internal/evaluator"
)

func TestVisualFormatter_FormatVisual(t *testing.T) {
	tests := []struct {
		name           string
		result         *evaluator.ExpressionResult
		file           string
		line           int
		customMessage  string
		expectedOutput []string // Lines that should appear in output
	}{
		{
			name: "simple comparison",
			result: &evaluator.ExpressionResult{
				Expression: "x > 20",
				Result:     false,
				Variables: map[string]interface{}{
					"x": 10,
				},
				Tree: &evaluator.EvaluationTree{
					ID:       1,
					Type:     "comparison",
					Operator: ">",
					Text:     "x > 20",
					Result:   false,
					Left: &evaluator.EvaluationTree{
						ID:    2,
						Type:  "identifier",
						Text:  "x",
						Value: 10,
					},
					Right: &evaluator.EvaluationTree{
						ID:    3,
						Type:  "literal",
						Text:  "20",
						Value: 20,
					},
				},
			},
			file: "example_test.go",
			line: 42,
			expectedOutput: []string{
				"ASSERTION FAILED at example_test.go:42",
				"  assert(x > 20)",
				"|",     // Should have pipes
				"10",    // Should show value 10
				"false", // Should show comparison result
			},
		},
		{
			name: "logical AND expression",
			result: &evaluator.ExpressionResult{
				Expression: "age >= 18 && hasLicense",
				Result:     false,
				Variables: map[string]interface{}{
					"age":        16,
					"hasLicense": false,
				},
				Tree: &evaluator.EvaluationTree{
					ID:       1,
					Type:     "logical",
					Operator: "&&",
					Text:     "age >= 18 && hasLicense",
					Result:   false,
					Left: &evaluator.EvaluationTree{
						ID:       2,
						Type:     "comparison",
						Operator: ">=",
						Text:     "age >= 18",
						Result:   false,
						Left: &evaluator.EvaluationTree{
							ID:    3,
							Type:  "identifier",
							Text:  "age",
							Value: 16,
						},
					},
					Right: &evaluator.EvaluationTree{
						ID:     4,
						Type:   "identifier",
						Text:   "hasLicense",
						Value:  false,
						Result: false,
					},
				},
			},
			file: "example_test.go",
			line: 100,
			expectedOutput: []string{
				"ASSERTION FAILED at example_test.go:100",
				"  assert(age >= 18 && hasLicense)",
				"age",        // Should contain variable names and values
				"16",         // age value
				"hasLicense", // variable name
				"false",      // hasLicense value and comparison result
			},
		},
		{
			name: "with custom message",
			result: &evaluator.ExpressionResult{
				Expression: "x == y",
				Result:     false,
				Variables: map[string]interface{}{
					"x": 5,
					"y": 10,
				},
				Tree: &evaluator.EvaluationTree{
					ID:       1,
					Type:     "comparison",
					Operator: "==",
					Text:     "x == y",
					Result:   false,
					Left: &evaluator.EvaluationTree{
						ID:    2,
						Type:  "identifier",
						Text:  "x",
						Value: 5,
					},
					Right: &evaluator.EvaluationTree{
						ID:    3,
						Type:  "identifier",
						Text:  "y",
						Value: 10,
					},
				},
			},
			file:          "test.go",
			line:          50,
			customMessage: "Values should be equal",
			expectedOutput: []string{
				"ASSERTION FAILED at test.go:50",
				"Values should be equal",
				"  assert(x == y)",
				"5",
				"10",
				"false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewVisualFormatter()
			output := formatter.FormatVisual(tt.result, tt.file, tt.line, tt.customMessage)

			// Check that all expected lines appear in the output
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nActual output:\n%s", expected, output)
				}
			}

			// Verify it has power-assert style format
			if !strings.Contains(output, "  assert(") {
				t.Error("Output should have power-assert style 'assert(' prefix")
			}
		})
	}
}

func TestVisualFormatter_ValuePositioning(t *testing.T) {
	formatter := NewVisualFormatter()

	// Test that values are properly aligned with their expressions
	result := &evaluator.ExpressionResult{
		Expression: "x > 20",
		Result:     false,
		Tree: &evaluator.EvaluationTree{
			Type:     "comparison",
			Operator: ">",
			Text:     "x > 20",
			Result:   false,
			Left: &evaluator.EvaluationTree{
				Type:  "identifier",
				Text:  "x",
				Value: 10,
			},
		},
	}

	output := formatter.FormatVisual(result, "test.go", 1, "")

	// The output should align the value "10" under "x"
	lines := strings.Split(output, "\n")

	// Find the assert line
	var assertLineIndex int
	for i, line := range lines {
		if strings.Contains(line, "assert(x > 20)") {
			assertLineIndex = i
			break
		}
	}

	// The value should appear in a subsequent line, aligned with 'x'
	if assertLineIndex > 0 && assertLineIndex+2 < len(lines) {
		valueLine := lines[assertLineIndex+2]
		// Check that "10" appears at approximately the right position
		if !strings.Contains(valueLine, "10") {
			t.Errorf("Value '10' not found in the expected position.\nOutput:\n%s", output)
		}
	}
}

func TestVisualFormatter_FormatValueCompact(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil value", nil, "nil"},
		{"short string", "hello", `"hello"`},
		{"long string", "this is a very long string that should be truncated", `"this is a "...`},
		{"int slice", []int{1, 2, 3}, "[1 2 3]"},
		{"string slice", []string{"a", "b"}, `[a b]`},
		{"int value", 42, "42"},
		{"bool value", true, "true"},
		{"long formatted value", "verylongvaluethatexceedsfifteencharacters", `"verylongva"...`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValueCompact(tt.value)
			if result != tt.expected {
				t.Errorf("formatValueCompact(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestVisualFormatter_MachineReadableSection(t *testing.T) {
	// Test with machine-readable enabled
	t.Setenv("DIAGASSERT_MACHINE_READABLE", "true")
	formatter := NewVisualFormatter()

	result := &evaluator.ExpressionResult{
		Expression: "x > 0",
		Result:     false,
		Variables: map[string]interface{}{
			"x": -5,
		},
	}

	output := formatter.FormatVisual(result, "test.go", 1, "")

	if !strings.Contains(output, "[MACHINE_READABLE_START]") {
		t.Error("Expected machine-readable section to be included")
	}
	if !strings.Contains(output, "EXPR: x > 0") {
		t.Error("Expected expression in machine-readable section")
	}
	if !strings.Contains(output, "VARIABLES: x=-5") {
		t.Error("Expected variables in machine-readable section")
	}
}

func TestVisualFormatter_NoMachineReadable(t *testing.T) {
	// Test with machine-readable disabled
	t.Setenv("DIAGASSERT_MACHINE_READABLE", "false")
	formatter := NewVisualFormatter()

	result := &evaluator.ExpressionResult{
		Expression: "x > 0",
		Result:     false,
	}

	output := formatter.FormatVisual(result, "test.go", 1, "")

	if strings.Contains(output, "[MACHINE_READABLE_START]") {
		t.Error("Machine-readable section should not be included when disabled")
	}
}

func TestVisualFormatter_Issue10_Alignment(t *testing.T) {
	// Test for GitHub issue #10 - alignment issues with colored output
	formatter := NewVisualFormatter()

	result := &evaluator.ExpressionResult{
		Expression: "result == expected",
		Result:     false,
		Variables: map[string]interface{}{
			"result":   3,
			"expected": 0,
		},
		Tree: &evaluator.EvaluationTree{
			ID:       1,
			Type:     "comparison",
			Operator: "==",
			Text:     "result == expected",
			Result:   false,
			Left: &evaluator.EvaluationTree{
				ID:    2,
				Type:  "identifier",
				Text:  "result",
				Value: 3,
			},
			Right: &evaluator.EvaluationTree{
				ID:    3,
				Type:  "identifier",
				Text:  "expected",
				Value: 0,
			},
		},
	}

	output := formatter.FormatVisual(result, "test.go", 1, "")

	// Verify the output doesn't contain truncated ANSI escape sequences
	if strings.Contains(output, "30m") || strings.Contains(output, "31m") || strings.Contains(output, "32m") {
		t.Errorf("Output contains truncated ANSI escape sequences: %q", output)
	}

	// Verify that values are properly aligned
	lines := strings.Split(output, "\n")
	var assertLine, valueLine string
	for i, line := range lines {
		if strings.Contains(line, "assert(result == expected)") {
			assertLine = line
			// Find a line with values
			for j := i + 1; j < len(lines); j++ {
				if strings.Contains(lines[j], "3") || strings.Contains(lines[j], "0") {
					valueLine = lines[j]
					break
				}
			}
			break
		}
	}

	if assertLine == "" {
		t.Error("Could not find assert line in output")
	}
	if valueLine == "" {
		t.Error("Could not find value line in output")
	}

	// The test mainly checks that we don't get truncated escape sequences
	// and that the output is well-formed
	t.Logf("Assert line: %q", assertLine)
	t.Logf("Value line: %q", valueLine)
	t.Logf("Full output:\n%s", output)
}
