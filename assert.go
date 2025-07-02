// Package diagassert provides assertion utilities for diagnostic testing.
package diagassert

import (
	"fmt"
	"path/filepath"
	"runtime"

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
// This is the only API you need to remember.
func Assert(t TestingT, expr bool) {
	t.Helper()

	if expr {
		return
	}

	// On failure: display detailed evaluation of the expression
	output := buildDiagnosticOutput()
	t.Error(output)
}

// Require is the same as Assert, but terminates the test immediately on failure
func Require(t TestingT, expr bool) {
	t.Helper()

	if expr {
		return
	}

	// On failure: display detailed evaluation of the expression and terminate
	output := buildDiagnosticOutput()
	t.Fatal(output)
}

// buildDiagnosticOutput builds diagnostic information
func buildDiagnosticOutput() string {
	// Get caller information
	_, file, line, _ := runtime.Caller(2)

	// Extract expression from source code
	expr, err := parser.ExtractExpression(file, line)
	if err != nil {
		return fmt.Sprintf("ASSERTION FAILED at %s:%d\n(unable to extract expression: %v)",
			filepath.Base(file), line, err)
	}

	// Build diagnostic output using formatter
	opts := formatter.GetDefaultOptions()
	return formatter.BuildDiagnosticOutput(file, line, expr, opts)
}