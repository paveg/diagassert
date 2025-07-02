package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractExpression(t *testing.T) {
	// Create a temporary test file
	testContent := `package main

import "github.com/paveg/diagassert"

func TestExample(t *testing.T) {
	x := 10
	diagassert.Assert(t, x > 20)  // This is line 7
	y := 5
	diagassert.Assert(t, x > y && y < 10)  // This is line 9
}
`

	// Create temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		line     int
		expected string
		wantErr  bool
	}{
		{
			name:     "simple comparison",
			line:     7,
			expected: "x > 20",
			wantErr:  false,
		},
		{
			name:     "complex expression",
			line:     9,
			expected: "x > y && y < 10",
			wantErr:  false,
		},
		{
			name:     "non-existent line",
			line:     100,
			expected: "",
			wantErr:  true,
		},
		{
			name:     "line without assert",
			line:     6,
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractExpression(testFile, tt.line)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractExpression() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ExtractExpression() unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("ExtractExpression() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestIsAssertCall(t *testing.T) {
	// This function is tested indirectly through ExtractExpression
	// Additional unit tests could be added here if needed
	t.Skip("isAssertCall is tested indirectly through ExtractExpression")
}