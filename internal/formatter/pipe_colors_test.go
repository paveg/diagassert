package formatter

import (
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/paveg/diagassert/internal/evaluator"
)

func TestPipeColorsConfiguration(t *testing.T) {
	tests := []struct {
		name                    string
		envVars                 map[string]string
		expectPipeColorsEnabled bool
		expectPipeColorPalette  bool
	}{
		{
			name:                    "pipe colors enabled by default",
			envVars:                 map[string]string{},
			expectPipeColorsEnabled: true,
			expectPipeColorPalette:  true,
		},
		{
			name:                    "pipe colors disabled by DIAGASSERT_PIPE_COLORS=false",
			envVars:                 map[string]string{"DIAGASSERT_PIPE_COLORS": "false"},
			expectPipeColorsEnabled: false,
			expectPipeColorPalette:  true, // Palette should still be created
		},
		{
			name:                    "pipe colors enabled explicitly",
			envVars:                 map[string]string{"DIAGASSERT_PIPE_COLORS": "true"},
			expectPipeColorsEnabled: true,
			expectPipeColorPalette:  true,
		},
		{
			name:                    "pipe colors enabled with other value",
			envVars:                 map[string]string{"DIAGASSERT_PIPE_COLORS": "yes"},
			expectPipeColorsEnabled: true,
			expectPipeColorPalette:  true,
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

			if formatter.colorConfig.PipeColorsEnabled != tt.expectPipeColorsEnabled {
				t.Errorf("expected pipe colors enabled: %v, got: %v", tt.expectPipeColorsEnabled, formatter.colorConfig.PipeColorsEnabled)
			}

			if tt.expectPipeColorPalette && len(formatter.colorConfig.PipeColorPalette) == 0 {
				t.Error("expected pipe color palette to be created")
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

func TestPipeColorPalette(t *testing.T) {
	formatter := NewVisualFormatter()

	// Test that the palette is created
	if len(formatter.colorConfig.PipeColorPalette) == 0 {
		t.Fatal("expected pipe color palette to be created")
	}

	// Test that we have the expected number of colors
	expectedColors := 8
	if len(formatter.colorConfig.PipeColorPalette) != expectedColors {
		t.Errorf("expected %d colors in palette, got %d", expectedColors, len(formatter.colorConfig.PipeColorPalette))
	}

	// Test that all colors are different from each other
	for i := 0; i < len(formatter.colorConfig.PipeColorPalette); i++ {
		for j := i + 1; j < len(formatter.colorConfig.PipeColorPalette); j++ {
			if formatter.colorConfig.PipeColorPalette[i] == formatter.colorConfig.PipeColorPalette[j] {
				t.Errorf("duplicate colors found at indices %d and %d", i, j)
			}
		}
	}
}

func TestAssignPipeColor(t *testing.T) {
	formatter := NewVisualFormatter()

	tests := []struct {
		name              string
		expression        string
		pipeColorsEnabled bool
		expectSameColor   bool
	}{
		{
			name:              "consistent color assignment",
			expression:        "x",
			pipeColorsEnabled: true,
			expectSameColor:   true,
		},
		{
			name:              "different expressions get different colors",
			expression:        "y",
			pipeColorsEnabled: true,
			expectSameColor:   false,
		},
		{
			name:              "pipe colors disabled falls back to default",
			expression:        "x",
			pipeColorsEnabled: false,
			expectSameColor:   false, // Should use default pipe color
		},
	}

	// Get color for "x" as reference
	formatter.colorConfig.PipeColorsEnabled = true
	referenceColor := formatter.assignPipeColor("x")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter.colorConfig.PipeColorsEnabled = tt.pipeColorsEnabled

			color1 := formatter.assignPipeColor(tt.expression)
			color2 := formatter.assignPipeColor(tt.expression)

			// Same expression should always return the same color
			if color1 != color2 {
				t.Error("same expression should return consistent color")
			}

			// Test against reference color
			if tt.expectSameColor && color1 != referenceColor {
				t.Error("expected same color as reference")
			}
			if !tt.expectSameColor && tt.expression != "x" && color1 == referenceColor {
				t.Error("expected different color from reference")
			}
		})
	}
}

func TestGetPipeColorForValue(t *testing.T) {
	formatter := NewVisualFormatter()

	position := ValuePosition{
		Expression: "test",
		Value:      "42",
	}

	color1 := formatter.getPipeColorForValue(position)
	color2 := formatter.getPipeColorForValue(position)

	if color1 != color2 {
		t.Error("same value position should return consistent color")
	}
}

func TestSimpleHash(t *testing.T) {
	formatter := NewVisualFormatter()

	tests := []struct {
		name     string
		input    string
		expected bool // whether hash should be consistent
	}{
		{
			name:     "consistent hash for same input",
			input:    "test",
			expected: true,
		},
		{
			name:     "different hash for different input",
			input:    "different",
			expected: false,
		},
	}

	referenceHash := formatter.simpleHash("test")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := formatter.simpleHash(tt.input)
			hash2 := formatter.simpleHash(tt.input)

			// Same input should always return the same hash
			if hash1 != hash2 {
				t.Error("same input should return consistent hash")
			}

			// Hash should be non-negative
			if hash1 < 0 || hash2 < 0 {
				t.Error("hash should be non-negative")
			}

			// Test against reference hash
			if tt.expected && hash1 != referenceHash {
				t.Error("expected same hash as reference")
			}
			if !tt.expected && tt.input != "test" && hash1 == referenceHash {
				t.Error("expected different hash from reference")
			}
		})
	}
}

func TestColorizePerValuePipe(t *testing.T) {
	formatter := NewVisualFormatter()

	position := ValuePosition{
		Expression: "x",
		Value:      "10",
	}

	tests := []struct {
		name          string
		colorsEnabled bool
		pipeColors    bool
		expectAnsi    bool
	}{
		{
			name:          "colors enabled with pipe colors",
			colorsEnabled: true,
			pipeColors:    true,
			expectAnsi:    true,
		},
		{
			name:          "colors enabled without pipe colors",
			colorsEnabled: true,
			pipeColors:    false,
			expectAnsi:    true,
		},
		{
			name:          "colors disabled",
			colorsEnabled: false,
			pipeColors:    true,
			expectAnsi:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter.colorConfig.ColorsEnabled = tt.colorsEnabled
			formatter.colorConfig.PipeColorsEnabled = tt.pipeColors

			result := formatter.colorizePerValuePipe("|", position)

			hasAnsi := strings.Contains(result, "\x1b[")
			if hasAnsi != tt.expectAnsi {
				t.Errorf("expected ANSI codes: %v, got: %v, result: %q", tt.expectAnsi, hasAnsi, result)
			}

			// The pipe character should always be preserved
			if !strings.Contains(result, "|") {
				t.Errorf("expected result to contain pipe character, got: %q", result)
			}
		})
	}
}

func TestPerValuePipeColorIntegration(t *testing.T) {
	// Save original env vars
	originalNoColor := os.Getenv("NO_COLOR")
	originalPipeColors := os.Getenv("DIAGASSERT_PIPE_COLORS")
	defer func() {
		if originalNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", originalNoColor)
		}
		if originalPipeColors == "" {
			os.Unsetenv("DIAGASSERT_PIPE_COLORS")
		} else {
			os.Setenv("DIAGASSERT_PIPE_COLORS", originalPipeColors)
		}
	}()

	// Enable colors and per-value pipe colors
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("DIAGASSERT_PIPE_COLORS")

	// Reset color package state
	color.NoColor = false

	formatter := NewVisualFormatter()

	// Create a test result with a simple expression
	result := &evaluator.ExpressionResult{
		Expression: "x > 10",
		Result:     false,
		Variables:  map[string]interface{}{"x": 5},
		Tree: &evaluator.EvaluationTree{
			ID:       1,
			Type:     "comparison",
			Operator: ">",
			Text:     "x > 10",
			Result:   false,
			Left: &evaluator.EvaluationTree{
				ID:    2,
				Type:  "identifier",
				Text:  "x",
				Value: 5,
			},
			Right: &evaluator.EvaluationTree{
				ID:    3,
				Type:  "literal",
				Text:  "10",
				Value: 10,
			},
		},
	}

	// Generate output
	output := formatter.FormatVisual(result, "test.go", 42, "")

	// Check that output contains ANSI escape sequences (indicating colors are applied)
	if !strings.Contains(output, "\x1b[") {
		t.Error("expected colored output to contain ANSI escape sequences")
	}

	// Check that the basic structure is maintained
	if !strings.Contains(output, "ASSERTION FAILED") {
		t.Error("expected output to contain assertion header")
	}

	if !strings.Contains(output, "x > 10") {
		t.Error("expected output to contain expression")
	}
}

func TestForceColorPipe(t *testing.T) {
	formatter := NewVisualFormatter()

	// Test with different colors from the palette
	for i, pipeColor := range formatter.colorConfig.PipeColorPalette {
		result := formatter.forceColorPipe("|", pipeColor)

		// Should contain ANSI escape sequences
		if !strings.Contains(result, "\x1b[") {
			t.Errorf("color %d: expected ANSI escape sequences", i)
		}

		// Should contain the pipe character
		if !strings.Contains(result, "|") {
			t.Errorf("color %d: expected pipe character", i)
		}

		// Should contain reset sequence
		if !strings.Contains(result, "\x1b[0m") {
			t.Errorf("color %d: expected reset sequence", i)
		}
	}
}
