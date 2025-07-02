// Package formatter provides functionality for formatting assertion failure output.
package formatter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// Future implementation will add detailed evaluation here
	// For now, we keep it simple and display the expression

	// Machine-readable section (controlled by environment variable)
	if opts.IncludeMachineReadable {
		b.WriteString("\n[MACHINE_READABLE_START]\n")
		b.WriteString(fmt.Sprintf("EXPR: %s\n", expr))
		b.WriteString("RESULT: false\n")
		b.WriteString("[MACHINE_READABLE_END]\n")
	}

	return b.String()
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