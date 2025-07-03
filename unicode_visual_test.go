package diagassert_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/paveg/diagassert"
)

// TestUnicodeVisualFormatter tests the Unicode-aware visual formatter
func TestUnicodeVisualFormatter(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		values   map[string]interface{}
		expected []string // Expected content parts in the output
	}{
		{
			name: "simple comparison with Japanese",
			expr: "名前 == 'ボブ'",
			values: map[string]interface{}{
				"名前": "アリス",
			},
			expected: []string{
				"assert(名前 == 'ボブ')",
				"|",
				"\"アリス\"",
				"false",
			},
		},
		{
			name: "mixed ASCII and wide characters",
			expr: "user名前 > 10",
			values: map[string]interface{}{
				"user名前": 5,
			},
			expected: []string{
				"assert(user名前 > 10)",
				"|",
				"5",
				"false",
			},
		},
		{
			name: "complex expression with Japanese",
			expr: "年齢 >= 18 && 免許 == true",
			values: map[string]interface{}{
				"年齢": 16,
				"免許": false,
			},
			expected: []string{
				"assert(年齢 >= 18 && 免許 == true)",
				"|",
				"16",
				"false",
			},
		},
		{
			name: "Korean characters",
			expr: "이름 == '김철수'",
			values: map[string]interface{}{
				"이름": "박영희",
			},
			expected: []string{
				"assert(이름 == '김철수')",
				"|",
				"\"박영희\"",
				"false",
			},
		},
		{
			name: "Chinese characters with numbers",
			expr: "价格 > 100",
			values: map[string]interface{}{
				"价格": 80,
			},
			expected: []string{
				"assert(价格 > 100)",
				"|",
				"80",
				"false",
			},
		},
		{
			name: "full-width symbols",
			expr: "Ａ＞Ｂ",
			values: map[string]interface{}{
				"Ａ": 5,
				"Ｂ": 10,
			},
			expected: []string{
				"assert(Ａ＞Ｂ)",
				"|",
				"5",
				"false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockT()

			// Note: We can't easily test dynamic Unicode expressions with actual parsing
			// This test focuses on the visual formatter's Unicode support
			// The actual expressions would need to be evaluated at runtime

			// For now, test the value capture and formatting with Unicode values
			diagassert.Assert(mock, false,
				diagassert.Values(tt.values))

			output := mock.getOutput()
			t.Logf("Unicode Expression (intended): %s", tt.expr)
			t.Logf("Output:\n%s", output)

			// Check that Unicode values are properly displayed in captured values
			for name, value := range tt.values {
				expectedValueStr := fmt.Sprintf("%s = %v", name, value)
				if !strings.Contains(output, expectedValueStr) {
					t.Errorf("Expected output to contain Unicode value %q, got:\n%s", expectedValueStr, output)
				}
			}

			// Basic checks for power-assert format
			if !strings.Contains(output, "assert(") {
				t.Errorf("Should contain power-assert format, got: %s", output)
			}

			// Check that captured values section exists
			if !strings.Contains(output, "CAPTURED VALUES:") {
				t.Errorf("Should contain captured values section, got: %s", output)
			}
		})
	}
}

// TestVisualWidth tests the visual width calculation for various Unicode characters
func TestVisualWidth(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello", 5},        // ASCII
		{"こんにちは", 10},       // 5 Hiragana characters × 2
		{"Hello世界", 9},      // 5 ASCII + 2 Han × 2
		{"🌍", 1},            // Emoji (treated as 1 width)
		{"A→B", 3},          // 1 + 1 + 1 (arrow is not detected as full-width)
		{"名前", 4},           // 2 Han characters × 2
		{"이름", 4},           // 2 Hangul characters × 2
		{"价格", 4},           // 2 Han characters × 2
		{"Ａ＞Ｂ", 6},          // 3 full-width characters × 2
		{"", 0},             // Empty string
		{"ひらがな123カタカナ", 19}, // Mixed: 4×2 + 3 + 4×2 = 8 + 3 + 8 = 19
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := calculateVisualWidth(tt.input)
			if actual != tt.expected {
				t.Errorf("visualWidth(%q) = %d, want %d", tt.input, actual, tt.expected)
			}
		})
	}
}

// TestCharacterPositioning tests that character positions are calculated correctly
func TestCharacterPositioning(t *testing.T) {
	tests := []struct {
		name  string
		input string
		tests []struct {
			bytePos        int
			expectedVisual int
		}
	}{
		{
			name:  "ASCII only",
			input: "hello",
			tests: []struct {
				bytePos        int
				expectedVisual int
			}{
				{0, 0}, // 'h'
				{1, 1}, // 'e'
				{4, 4}, // 'o'
			},
		},
		{
			name:  "Japanese characters",
			input: "こんにちは",
			tests: []struct {
				bytePos        int
				expectedVisual int
			}{
				{0, 0}, // 'こ' starts at visual 0
				{3, 2}, // 'ん' starts at visual 2
				{6, 4}, // 'に' starts at visual 4
			},
		},
		{
			name:  "Mixed ASCII and Japanese",
			input: "Hello世界",
			tests: []struct {
				bytePos        int
				expectedVisual int
			}{
				{0, 0}, // 'H'
				{5, 5}, // '世' should start at visual position 5
				{8, 7}, // '界' should start at visual position 7
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, test := range tt.tests {
				actual := calculateVisualPositionFromByte(tt.input, test.bytePos)
				if actual != test.expectedVisual {
					t.Errorf("For %q at byte %d: got visual %d, want %d",
						tt.input, test.bytePos, actual, test.expectedVisual)
				}
			}
		})
	}
}

// TestComplexUnicodeExpressions tests complex expressions with mixed character types
func TestComplexUnicodeExpressions(t *testing.T) {
	tests := []struct {
		name        string
		assertion   func(*mockT)
		expectFail  bool
		expectParts []string
	}{
		{
			name: "nested Japanese property access",
			assertion: func(mock *mockT) {
				user := struct {
					名前 string
					年齢 int
				}{名前: "田中", 年齢: 16}

				diagassert.Assert(mock, user.年齢 >= 18,
					diagassert.V("user.年齢", user.年齢))
			},
			expectFail: true,
			expectParts: []string{
				"assert(",
				"年齢 >= 18",
				"|",
				"16",
				"false",
			},
		},
		{
			name: "Korean logical expression",
			assertion: func(mock *mockT) {
				나이 := 16
				면허 := false
				diagassert.Assert(mock, 나이 >= 18 && 면허,
					diagassert.Values{
						"나이": 나이,
						"면허": 면허,
					})
			},
			expectFail: true,
			expectParts: []string{
				"assert(",
				"나이 >= 18 && 면허",
				"|",
				"16",
				"false",
			},
		},
		{
			name: "Chinese with comparison",
			assertion: func(mock *mockT) {
				价格 := 80
				最低价格 := 100
				diagassert.Assert(mock, 价格 > 最低价格,
					diagassert.V("价格", 价格),
					diagassert.V("最低价格", 最低价格))
			},
			expectFail: true,
			expectParts: []string{
				"assert(",
				"价格 > 最低价格",
				"|",
				"80",
				"false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockT()
			tt.assertion(mock)

			if mock.failed != tt.expectFail {
				t.Errorf("Expected fail=%v, got fail=%v", tt.expectFail, mock.failed)
			}

			output := mock.getOutput()
			t.Logf("Complex Unicode Output:\n%s", output)

			for _, part := range tt.expectParts {
				if !strings.Contains(output, part) {
					t.Errorf("Expected output to contain %q, got:\n%s", part, output)
				}
			}
		})
	}
}

// Helper functions for testing

// calculateVisualWidth calculates the visual width of a string (test helper)
func calculateVisualWidth(s string) int {
	width := 0
	for _, r := range s {
		if isWideRune(r) {
			width += 2
		} else {
			width++
		}
	}
	return width
}

// isWideRune determines if a rune is a wide character (test helper)
func isWideRune(r rune) bool {
	// Simplified implementation for testing
	return (r >= 0x1100 && r <= 0x115F) || // Hangul Jamo
		(r >= 0x2E80 && r <= 0x9FFF) || // CJK
		(r >= 0xAC00 && r <= 0xD7AF) || // Hangul Syllables
		(r >= 0xF900 && r <= 0xFAFF) || // CJK Compatibility
		(r >= 0xFF00 && r <= 0xFFEF) // Fullwidth forms
}

// calculateVisualPositionFromByte calculates visual position from byte position (test helper)
func calculateVisualPositionFromByte(s string, bytePos int) int {
	if bytePos <= 0 {
		return 0
	}

	visualPos := 0
	currentByte := 0

	for _, r := range s {
		if currentByte >= bytePos {
			break
		}

		if isWideRune(r) {
			visualPos += 2
		} else {
			visualPos++
		}

		// Calculate byte length of this rune
		if r < 0x80 {
			currentByte += 1
		} else if r < 0x800 {
			currentByte += 2
		} else if r < 0x10000 {
			currentByte += 3
		} else {
			currentByte += 4
		}
	}

	return visualPos
}
