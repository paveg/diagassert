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
