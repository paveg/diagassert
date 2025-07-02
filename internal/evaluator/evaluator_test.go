package evaluator

import (
	"go/parser"
	"runtime"
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name       string
		expr       string
		result     bool
		expectExpr string
		expectRes  bool
	}{
		{
			name:       "simple comparison",
			expr:       "x > 20",
			result:     false,
			expectExpr: "x > 20",
			expectRes:  false,
		},
		{
			name:       "logical expression",
			expr:       "age >= 18 && hasLicense",
			result:     false,
			expectExpr: "age >= 18 && hasLicense",
			expectRes:  false,
		},
		{
			name:       "passing expression",
			expr:       "x == x",
			result:     true,
			expectExpr: "x == x",
			expectRes:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get caller PC for the new API
			pc, _, _, _ := runtime.Caller(0)
			result := Evaluate(tt.expr, tt.result, pc)

			if result.Expression != tt.expectExpr {
				t.Errorf("Evaluate().Expression = %q, expected %q", result.Expression, tt.expectExpr)
			}

			if result.Result != tt.expectRes {
				t.Errorf("Evaluate().Result = %v, expected %v", result.Result, tt.expectRes)
			}

			if result.Variables == nil {
				t.Error("Evaluate().Variables should not be nil")
			}

			// In Phase 2, we should have variables populated (even if placeholder)
			// Tree should be created for all expressions
			if result.Tree == nil {
				t.Error("Evaluate().Tree should not be nil in Phase 2")
			}
		})
	}
}

func TestBuildEvaluationTree(t *testing.T) {
	// Test that buildEvaluationTree constructs trees properly
	tests := []struct {
		name string
		expr string
	}{
		{
			name: "simple comparison",
			expr: "x > 20",
		},
		{
			name: "logical expression",
			expr: "age >= 18 && hasLicense",
		},
		{
			name: "method call",
			expr: "user.IsAdult()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variables := make(map[string]interface{})
			tree := buildEvaluationTree(tt.expr, variables)

			if tree == nil {
				t.Errorf("buildEvaluationTree(%q) should not return nil", tt.expr)
				return
			}

			if tree.ID == 0 {
				t.Error("Tree node should have non-zero ID")
			}

			if tree.Type == "" {
				t.Error("Tree node should have a type")
			}
		})
	}
}

func TestExtractVariableNames(t *testing.T) {
	tests := []struct {
		name          string
		expr          string
		expectedCount int // Just check count for now, as order can vary
	}{
		{
			name:          "simple variable",
			expr:          "x > 20",
			expectedCount: 1,
		},
		{
			name:          "multiple variables",
			expr:          "age >= 18 && hasLicense",
			expectedCount: 2,
		},
		{
			name:          "struct field access",
			expr:          "user.Name == \"Alice\"",
			expectedCount: 2, // user and Name are both extracted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse expression using the actual parser
			node, err := parser.ParseExpr(tt.expr)
			if err != nil {
				t.Fatalf("Failed to parse expression: %v", err)
			}

			names := extractVariableNames(node)

			if len(names) != tt.expectedCount {
				t.Errorf("Expected %d variable names, got %d: %v", tt.expectedCount, len(names), names)
			}
		})
	}
}

// Phase 2 specific tests
func TestPhase2EvaluationFeatures(t *testing.T) {
	t.Log("Phase 2 evaluation features are now implemented")

	// Test evaluation tree construction
	pc, _, _, _ := runtime.Caller(0)
	result := Evaluate("x > 20", false, pc)

	if result.Tree == nil {
		t.Error("Phase 2 should create evaluation trees")
	}

	if result.Variables == nil {
		t.Error("Phase 2 should extract variables")
	}
}

func TestEvaluateWithValues(t *testing.T) {
	tests := []struct {
		name           string
		expr           string
		result         bool
		userValues     map[string]interface{}
		expectVarValue interface{}
		expectVarName  string
	}{
		{
			name:           "simple value replacement",
			expr:           "x > 20",
			result:         false,
			userValues:     map[string]interface{}{"x": 15},
			expectVarValue: 15,
			expectVarName:  "x",
		},
		{
			name:           "multiple values",
			expr:           "age >= 18 && hasLicense",
			result:         false,
			userValues:     map[string]interface{}{"age": 16, "hasLicense": false},
			expectVarValue: 16,
			expectVarName:  "age",
		},
		{
			name:           "struct value",
			expr:           "user.Age >= 18",
			result:         false,
			userValues:     map[string]interface{}{"user": struct{ Age int }{Age: 16}},
			expectVarValue: struct{ Age int }{Age: 16},
			expectVarName:  "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc, _, _, _ := runtime.Caller(0)
			result := EvaluateWithValues(tt.expr, tt.result, pc, tt.userValues)

			if result.Expression != tt.expr {
				t.Errorf("EvaluateWithValues().Expression = %q, expected %q", result.Expression, tt.expr)
			}

			if result.Result != tt.result {
				t.Errorf("EvaluateWithValues().Result = %v, expected %v", result.Result, tt.result)
			}

			// Check that user-provided values are properly integrated
			if value, exists := result.Variables[tt.expectVarName]; !exists {
				t.Errorf("Expected variable %q to exist in Variables", tt.expectVarName)
			} else if value != tt.expectVarValue {
				t.Errorf("Expected variable %q to have value %v, got %v", tt.expectVarName, tt.expectVarValue, value)
			}

			// Verify that the evaluation tree uses the correct values
			if result.Tree != nil {
				// For simple expressions, verify the tree structure
				if result.Tree.Type == "comparison" && result.Tree.Left != nil {
					if result.Tree.Left.Type == "identifier" && result.Tree.Left.Text == tt.expectVarName {
						if result.Tree.Left.Value != tt.expectVarValue {
							t.Errorf("Expected tree node for %q to have value %v, got %v",
								tt.expectVarName, tt.expectVarValue, result.Tree.Left.Value)
						}
					}
				}
			}
		})
	}
}

func TestMergeVariables(t *testing.T) {
	tests := []struct {
		name          string
		autoExtracted map[string]interface{}
		userValues    map[string]interface{}
		expected      map[string]interface{}
	}{
		{
			name:          "user values override placeholders",
			autoExtracted: map[string]interface{}{"x": "<x>", "y": "<y>"},
			userValues:    map[string]interface{}{"x": 15},
			expected:      map[string]interface{}{"x": 15, "y": "<y>"},
		},
		{
			name:          "empty user values",
			autoExtracted: map[string]interface{}{"x": "<x>"},
			userValues:    map[string]interface{}{},
			expected:      map[string]interface{}{"x": "<x>"},
		},
		{
			name:          "additional user values",
			autoExtracted: map[string]interface{}{"x": "<x>"},
			userValues:    map[string]interface{}{"x": 15, "z": 25},
			expected:      map[string]interface{}{"x": 15, "z": 25},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeVariables(tt.autoExtracted, tt.userValues)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d variables, got %d", len(tt.expected), len(result))
			}

			for name, expectedValue := range tt.expected {
				if actualValue, exists := result[name]; !exists {
					t.Errorf("Expected variable %q to exist", name)
				} else if actualValue != expectedValue {
					t.Errorf("Expected variable %q to have value %v, got %v", name, expectedValue, actualValue)
				}
			}
		})
	}
}
