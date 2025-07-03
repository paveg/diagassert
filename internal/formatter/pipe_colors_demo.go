package formatter

// This file demonstrates the per-value pipe color system implementation.
// It contains example usage and demonstrates the key features.

import (
	"fmt"

	"github.com/paveg/diagassert/internal/evaluator"
)

// DemonstratePerValuePipeColors shows how the per-value pipe color system works
func DemonstratePerValuePipeColors() {
	fmt.Println("=== Per-Value Pipe Color System Demo ===")

	// Create a formatter with per-value pipe colors enabled
	formatter := NewVisualFormatter()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Colors enabled: %v\n", formatter.colorConfig.ColorsEnabled)
	fmt.Printf("  Per-value pipe colors enabled: %v\n", formatter.colorConfig.PipeColorsEnabled)
	fmt.Printf("  Color palette size: %d\n", len(formatter.colorConfig.PipeColorPalette))

	// Example 1: Simple comparison
	result1 := &evaluator.ExpressionResult{
		Expression: "x > 20",
		Result:     false,
		Variables:  map[string]interface{}{"x": 15},
		Tree: &evaluator.EvaluationTree{
			Type:     "comparison",
			Operator: ">",
			Text:     "x > 20",
			Result:   false,
			Left: &evaluator.EvaluationTree{
				Type:  "identifier",
				Text:  "x",
				Value: 15,
			},
			Right: &evaluator.EvaluationTree{
				Type:  "literal",
				Text:  "20",
				Value: 20,
			},
		},
	}

	fmt.Println("\nExample 1: Simple comparison")
	output1 := formatter.FormatVisual(result1, "example.go", 10, "")
	fmt.Println(output1)

	// Example 2: Complex expression with multiple values
	result2 := &evaluator.ExpressionResult{
		Expression: "a > b && c < d",
		Result:     false,
		Variables:  map[string]interface{}{"a": 5, "b": 10, "c": 15, "d": 3},
		Tree: &evaluator.EvaluationTree{
			Type:     "logical",
			Operator: "&&",
			Text:     "a > b && c < d",
			Result:   false,
			Left: &evaluator.EvaluationTree{
				Type:     "comparison",
				Operator: ">",
				Text:     "a > b",
				Result:   false,
				Left: &evaluator.EvaluationTree{
					Type:  "identifier",
					Text:  "a",
					Value: 5,
				},
				Right: &evaluator.EvaluationTree{
					Type:  "identifier",
					Text:  "b",
					Value: 10,
				},
			},
			Right: &evaluator.EvaluationTree{
				Type:     "comparison",
				Operator: "<",
				Text:     "c < d",
				Result:   false,
				Left: &evaluator.EvaluationTree{
					Type:  "identifier",
					Text:  "c",
					Value: 15,
				},
				Right: &evaluator.EvaluationTree{
					Type:  "identifier",
					Text:  "d",
					Value: 3,
				},
			},
		},
	}

	fmt.Println("\nExample 2: Complex logical expression")
	output2 := formatter.FormatVisual(result2, "example.go", 20, "")
	fmt.Println(output2)

	// Demonstrate color consistency
	fmt.Println("\nDemonstrating color consistency:")

	// Create multiple formatters and show that the same expression gets the same color
	formatter2 := NewVisualFormatter()
	formatter3 := NewVisualFormatter()

	// Test expressions
	testExpressions := []string{"x", "y", "z", "count", "value"}

	for _, expr := range testExpressions {
		color1 := formatter.assignPipeColor(expr)
		color2 := formatter2.assignPipeColor(expr)
		color3 := formatter3.assignPipeColor(expr)

		consistent := (color1 == color2) && (color2 == color3)
		fmt.Printf("  Expression '%s': Consistent coloring = %v\n", expr, consistent)
	}

	fmt.Println("\n=== Demo Complete ===")
}

// ShowColorPalette displays the available colors in the pipe color palette
func ShowColorPalette() {
	formatter := NewVisualFormatter()

	fmt.Println("=== Pipe Color Palette ===")
	for i, pipeColor := range formatter.colorConfig.PipeColorPalette {
		coloredPipe := pipeColor.Sprint("|")
		fmt.Printf("Color %d: %s\n", i, coloredPipe)
	}
	fmt.Println("=== End Palette ===")
}
