package evaluator

import (
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
			result := Evaluate(tt.expr, tt.result)

			if result.Expression != tt.expectExpr {
				t.Errorf("Evaluate().Expression = %q, expected %q", result.Expression, tt.expectExpr)
			}

			if result.Result != tt.expectRes {
				t.Errorf("Evaluate().Result = %v, expected %v", result.Result, tt.expectRes)
			}

			if result.Variables == nil {
				t.Error("Evaluate().Variables should not be nil")
			}

			// Variables map should be empty in Phase 1
			if len(result.Variables) != 0 {
				t.Errorf("Evaluate().Variables should be empty in Phase 1, got %d items", len(result.Variables))
			}

			// Tree should be nil in Phase 1
			if result.Tree != nil {
				t.Error("Evaluate().Tree should be nil in Phase 1")
			}
		})
	}
}

func TestBuildEvaluationTree(t *testing.T) {
	// Test that BuildEvaluationTree returns nil in Phase 1
	tests := []string{
		"x > 20",
		"age >= 18 && hasLicense",
		"user.IsAdult()",
	}

	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			tree := BuildEvaluationTree(expr)
			if tree != nil {
				t.Errorf("BuildEvaluationTree(%q) should return nil in Phase 1, got %+v", expr, tree)
			}
		})
	}
}

func TestExtractVariableValues(t *testing.T) {
	// Test that ExtractVariableValues returns empty map in Phase 1
	tests := []string{
		"x > 20",
		"age >= 18 && hasLicense",
		"user.Name == \"Alice\"",
	}

	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			variables := ExtractVariableValues(expr)
			if variables == nil {
				t.Errorf("ExtractVariableValues(%q) should not return nil", expr)
			}
			if len(variables) != 0 {
				t.Errorf("ExtractVariableValues(%q) should return empty map in Phase 1, got %d items", expr, len(variables))
			}
		})
	}
}

// Future tests for Phase 2 implementation
func TestFutureEvaluationFeatures(t *testing.T) {
	t.Skip("Future Phase 2 tests - variable value extraction and evaluation trees")
	
	// Future tests will include:
	// - Variable value extraction
	// - Evaluation tree construction
	// - Method call result tracking
	// - Struct field access
	// - Array/slice element access
}