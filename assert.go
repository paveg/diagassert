// Package diagassert provides assertion utilities for diagnostic testing.
package diagassert

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/paveg/diagassert/internal/evaluator"
	"github.com/paveg/diagassert/internal/formatter"
	"github.com/paveg/diagassert/internal/parser"
)

// TestingT is a minimal testing interface
type TestingT interface {
	Error(args ...interface{})
	Fatal(args ...interface{})
	Helper()
}

// Assert evaluates any expression and outputs detailed diagnostic information if false.
// This is the primary API you need to remember.
//
// Basic usage:
//
//	Assert(t, expr)
//
// Enhanced usage with values and messages:
//
//	Assert(t, expr, V("x", x), "custom message")
//	Assert(t, expr, Values{"x": x, "y": y})
//	Assert(t, expr, "custom message", V("z", z))
func Assert(t TestingT, expr bool, args ...interface{}) {
	t.Helper()

	if expr {
		return
	}

	// On failure: display detailed evaluation of the expression
	ctx := NewAssertionContext(args...)
	output := buildDiagnosticOutputWithContext(expr, ctx)
	t.Error(output)
}

// Require is the same as Assert, but terminates the test immediately on failure
func Require(t TestingT, expr bool, args ...interface{}) {
	t.Helper()

	if expr {
		return
	}

	// On failure: display detailed evaluation of the expression and terminate
	ctx := NewAssertionContext(args...)
	output := buildDiagnosticOutputWithContext(expr, ctx)
	t.Fatal(output)
}


// buildDiagnosticOutputWithContext builds diagnostic information with enhanced evaluation and context
func buildDiagnosticOutputWithContext(exprResult bool, ctx *AssertionContext) string {
	// Get caller information
	pc, file, line, ok := runtime.Caller(2) // Same as original since we're called from Assert/Require
	if !ok {
		return "ASSERTION FAILED (unable to get caller information)"
	}

	// Extract expression from source code
	expr, err := parser.ExtractExpression(file, line)
	if err != nil {
		return fmt.Sprintf("ASSERTION FAILED at %s:%d\n(unable to extract expression: %v)",
			filepath.Base(file), line, err)
	}

	// Perform enhanced evaluation with variable extraction
	var result *evaluator.ExpressionResult
	if ctx.HasValues() {
		// Use user-provided values when available
		userValues := ctx.GetValuesMap()
		result = evaluator.EvaluateWithValues(expr, exprResult, pc, userValues)
	} else {
		// Use standard evaluation without user values
		result = evaluator.Evaluate(expr, exprResult, pc)
	}

	// Build diagnostic output using enhanced formatter with context
	opts := formatter.GetDefaultOptions()

	// Convert our AssertionContext to formatter.AssertionContext
	var formatterCtx *formatter.AssertionContext
	if ctx.HasMessages() || ctx.HasValues() {
		formatterCtx = &formatter.AssertionContext{
			Messages: ctx.Messages,
			Values:   make([]formatter.Value, len(ctx.Values)),
		}

		// Convert Value types
		for i, v := range ctx.Values {
			formatterCtx.Values[i] = formatter.Value{
				Name:  v.Name,
				Value: v.Value,
			}
		}
	}

	output := formatter.BuildDiagnosticOutputWithEvaluatorAndContext(file, line, result, formatterCtx, opts)

	return output
}
