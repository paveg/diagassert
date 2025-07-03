// Package formatter provides functionality for formatting assertion failure output.
package formatter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paveg/diagassert/internal/evaluator"
)

// Options contains configuration options for formatting output.
type Options struct {
	IncludeMachineReadable bool
	Format                 string // "hybrid", "human", "machine"
}

// BuildDiagnosticOutput constructs a formatted diagnostic message for assertion failures.
func BuildDiagnosticOutput(file string, line int, expr string, opts Options) string {
	var b strings.Builder

	// Build the basic failure message
	b.WriteString(fmt.Sprintf("ASSERTION FAILED at %s:%d\n", filepath.Base(file), line))
	b.WriteString(fmt.Sprintf("Expression: %s\n", expr))
	b.WriteString("Result: false\n")

	// Machine-readable section (controlled by environment variable)
	if opts.IncludeMachineReadable {
		b.WriteString("\n[MACHINE_READABLE_START]\n")
		b.WriteString(fmt.Sprintf("EXPR: %s\n", expr))
		b.WriteString("RESULT: false\n")
		b.WriteString("[MACHINE_READABLE_END]\n")
	}

	return b.String()
}

// AssertionContext represents the context information for assertions (imported from main package).
// This is defined here to avoid circular imports while allowing the formatter to handle context.
type AssertionContext struct {
	Values   []Value  // Captured values using V() or Values{}
	Messages []string // Custom messages
}

// Value represents a named value for diagnostic output.
type Value struct {
	Name  string
	Value interface{}
}

// BuildDiagnosticOutputWithEvaluator constructs enhanced diagnostic output using evaluator results.
func BuildDiagnosticOutputWithEvaluator(file string, line int, result *evaluator.ExpressionResult, opts Options) string {
	return BuildDiagnosticOutputWithEvaluatorAndContext(file, line, result, nil, opts)
}

// BuildDiagnosticOutputWithEvaluatorAndContext constructs enhanced diagnostic output using evaluator results and assertion context.
func BuildDiagnosticOutputWithEvaluatorAndContext(file string, line int, result *evaluator.ExpressionResult, ctx *AssertionContext, opts Options) string {
	// Use visual formatter for power-assert style output
	visualFormatter := NewVisualFormatter()

	// Extract custom message from context
	var customMessage string
	if ctx != nil && hasMessages(ctx) {
		customMessage = getCombinedMessage(ctx)
	}

	// Merge context values into result variables if available
	if ctx != nil && hasValues(ctx) {
		if result.Variables == nil {
			result.Variables = make(map[string]interface{})
		}
		for _, value := range ctx.Values {
			result.Variables[value.Name] = value.Value
		}
	}

	return visualFormatter.FormatVisualWithContext(result, filepath.Base(file), line, customMessage, ctx)
}

// hasValues returns true if the context contains any values
func hasValues(ctx *AssertionContext) bool {
	return ctx != nil && len(ctx.Values) > 0
}

// hasMessages returns true if the context contains any messages
func hasMessages(ctx *AssertionContext) bool {
	return ctx != nil && len(ctx.Messages) > 0
}

// getCombinedMessage returns all messages joined together
func getCombinedMessage(ctx *AssertionContext) string {
	if ctx == nil || len(ctx.Messages) == 0 {
		return ""
	}

	combined := ""
	for i, msg := range ctx.Messages {
		if i > 0 {
			combined += " "
		}
		combined += msg
	}
	return combined
}

// formatEvaluationTree formats the evaluation tree in human-readable format.
func formatEvaluationTree(tree *evaluator.EvaluationTree, prefix string, isLast bool) string {
	if tree == nil {
		return ""
	}

	var b strings.Builder

	// Draw tree structure
	connector := "├─ "
	if isLast {
		connector = "└─ "
	}

	b.WriteString(prefix + connector)

	// Format the node based on its type
	switch tree.Type {
	case "comparison", "logical":
		b.WriteString(fmt.Sprintf("%s\n", tree.Text))
		if tree.Left != nil {
			b.WriteString(formatEvaluationTree(tree.Left, prefix+getChildPrefix(isLast), false))
		}
		if tree.Right != nil {
			b.WriteString(formatEvaluationTree(tree.Right, prefix+getChildPrefix(isLast), true))
		}
		b.WriteString(fmt.Sprintf("%s%s RESULT: %t\n", prefix, getChildPrefix(isLast)+"└─", tree.Result))

	case "identifier":
		if tree.Value != nil {
			b.WriteString(fmt.Sprintf("%s = %v (%T)\n", tree.Text, tree.Value, tree.Value))
		} else {
			b.WriteString(fmt.Sprintf("%s = <undefined>\n", tree.Text))
		}

	case "literal":
		b.WriteString(fmt.Sprintf("%s = %v (%T)\n", tree.Text, tree.Value, tree.Value))

	case "selector":
		b.WriteString(fmt.Sprintf("%s\n", tree.Text))
		if tree.Left != nil {
			b.WriteString(formatEvaluationTree(tree.Left, prefix+getChildPrefix(isLast), false))
		}
		if tree.Value != nil {
			b.WriteString(fmt.Sprintf("%s%s VALUE: %v (%T)\n", prefix, getChildPrefix(isLast)+"└─", tree.Value, tree.Value))
		}

	case "method_call":
		b.WriteString(fmt.Sprintf("%s\n", tree.Text))
		if tree.Left != nil {
			b.WriteString(formatEvaluationTree(tree.Left, prefix+getChildPrefix(isLast), false))
		}
		if tree.Value != nil {
			b.WriteString(fmt.Sprintf("%s%s RETURNS: %v (%T)\n", prefix, getChildPrefix(isLast)+"└─", tree.Value, tree.Value))
		}

	case "index":
		b.WriteString(fmt.Sprintf("%s\n", tree.Text))
		if tree.Left != nil {
			b.WriteString(formatEvaluationTree(tree.Left, prefix+getChildPrefix(isLast), false))
		}
		if tree.Right != nil {
			b.WriteString(formatEvaluationTree(tree.Right, prefix+getChildPrefix(isLast), false))
		}
		if tree.Value != nil {
			b.WriteString(fmt.Sprintf("%s%s VALUE: %v (%T)\n", prefix, getChildPrefix(isLast)+"└─", tree.Value, tree.Value))
		}

	case "unary":
		b.WriteString(fmt.Sprintf("%s\n", tree.Text))
		if tree.Left != nil {
			b.WriteString(formatEvaluationTree(tree.Left, prefix+getChildPrefix(isLast), true))
		}
		b.WriteString(fmt.Sprintf("%s%s RESULT: %t\n", prefix, getChildPrefix(isLast)+"└─", tree.Result))

	default:
		b.WriteString(fmt.Sprintf("%s (type: %s)\n", tree.Text, tree.Type))
	}

	return b.String()
}

// getChildPrefix returns the appropriate prefix for child nodes.
func getChildPrefix(isLast bool) string {
	if isLast {
		return "   "
	}
	return "│  "
}

// formatMachineReadableTree formats the evaluation tree for machine parsing.
func formatMachineReadableTree(tree *evaluator.EvaluationTree) string {
	if tree == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("TREE_START\n")
	b.WriteString(formatTreeNodeMachine(tree, 0))
	b.WriteString("TREE_END\n")
	return b.String()
}

// formatTreeNodeMachine formats a single tree node for machine reading.
func formatTreeNodeMachine(tree *evaluator.EvaluationTree, depth int) string {
	if tree == nil {
		return ""
	}

	var b strings.Builder
	indent := strings.Repeat("  ", depth)

	b.WriteString(fmt.Sprintf("%sNODE_ID: %d\n", indent, tree.ID))
	b.WriteString(fmt.Sprintf("%sTYPE: %s\n", indent, tree.Type))
	b.WriteString(fmt.Sprintf("%sTEXT: %s\n", indent, tree.Text))
	b.WriteString(fmt.Sprintf("%sRESULT: %t\n", indent, tree.Result))

	if tree.Operator != "" {
		b.WriteString(fmt.Sprintf("%sOPERATOR: %s\n", indent, tree.Operator))
	}

	if tree.Value != nil {
		b.WriteString(fmt.Sprintf("%sVALUE: %v\n", indent, tree.Value))
		b.WriteString(fmt.Sprintf("%sVALUE_TYPE: %T\n", indent, tree.Value))
	}

	if tree.Left != nil {
		b.WriteString(fmt.Sprintf("%sLEFT_CHILD:\n", indent))
		b.WriteString(formatTreeNodeMachine(tree.Left, depth+1))
	}

	if tree.Right != nil {
		b.WriteString(fmt.Sprintf("%sRIGHT_CHILD:\n", indent))
		b.WriteString(formatTreeNodeMachine(tree.Right, depth+1))
	}

	for i, child := range tree.Children {
		b.WriteString(fmt.Sprintf("%sCHILD_%d:\n", indent, i))
		b.WriteString(formatTreeNodeMachine(child, depth+1))
	}

	return b.String()
}

// analyzeFallureReason analyzes the evaluation tree to determine the primary failure reason.
func analyzeFallureReason(tree *evaluator.EvaluationTree) string {
	if tree == nil {
		return "unknown"
	}

	switch tree.Type {
	case "logical":
		if tree.Operator == "&&" {
			if tree.Left != nil && !tree.Left.Result {
				return "left_operand_false"
			}
			if tree.Right != nil && !tree.Right.Result {
				return "right_operand_false"
			}
		} else if tree.Operator == "||" {
			return "both_operands_false"
		}

	case "comparison":
		return "comparison_failed"

	case "method_call":
		return "method_returned_false"

	case "identifier":
		if tree.Value == nil {
			return "variable_undefined"
		}
		return "variable_falsy"

	case "unary":
		if tree.Operator == "!" {
			return "negation_true"
		}
	}

	return "expression_false"
}

// ShouldIncludeMachineReadable determines whether to include machine-readable sections.
func ShouldIncludeMachineReadable() bool {
	// Controlled by environment variable (default is true)
	env := os.Getenv("DIAGASSERT_MACHINE_READABLE")
	return env != "false"
}

// GetDefaultOptions returns the default formatting options.
func GetDefaultOptions() Options {
	return Options{
		IncludeMachineReadable: ShouldIncludeMachineReadable(),
		Format:                 "hybrid",
	}
}
