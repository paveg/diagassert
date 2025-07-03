// Package formatter provides visual formatting for power-assert style output.
package formatter

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/paveg/diagassert/internal/evaluator"
)

// VisualFormatter formats evaluation results in power-assert style.
type VisualFormatter struct {
	includeMachineReadable bool
}

// NewVisualFormatter creates a new visual formatter.
func NewVisualFormatter() *VisualFormatter {
	// Respect environment variable for machine-readable output
	includeMachine := os.Getenv("DIAGASSERT_MACHINE_READABLE") != "false"

	return &VisualFormatter{
		includeMachineReadable: includeMachine,
	}
}

// FormatVisual formats the evaluation result in power-assert style.
func (f *VisualFormatter) FormatVisual(result *evaluator.ExpressionResult, file string, line int, customMessage string) string {
	return f.FormatVisualWithContext(result, file, line, customMessage, nil)
}

// FormatVisualWithContext formats the evaluation result with context values.
func (f *VisualFormatter) FormatVisualWithContext(result *evaluator.ExpressionResult, file string, line int, customMessage string, ctx *AssertionContext) string {
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("ASSERTION FAILED at %s:%d\n\n", file, line))

	// Power-assert style visual representation
	b.WriteString(f.formatPowerAssertStyle(result))

	// Custom message section
	if customMessage != "" {
		b.WriteString("\nCUSTOM MESSAGE:\n")
		b.WriteString(customMessage + "\n")
	}

	// Captured values section
	if ctx != nil && len(ctx.Values) > 0 {
		b.WriteString("\nCAPTURED VALUES:\n")
		for _, value := range ctx.Values {
			b.WriteString(fmt.Sprintf("  %s = %v (%T)\n", value.Name, value.Value, value.Value))
		}
	}

	// Machine readable section
	if f.includeMachineReadable {
		b.WriteString("\n[MACHINE_READABLE_START]\n")
		b.WriteString(formatMachineSection(result))

		// Add custom message in machine-readable format
		if customMessage != "" {
			b.WriteString(fmt.Sprintf("CUSTOM_MESSAGE: %s\n", customMessage))
		}

		// Add captured values in machine-readable format
		if ctx != nil && len(ctx.Values) > 0 {
			b.WriteString("CAPTURED_VALUES_START\n")
			for _, value := range ctx.Values {
				b.WriteString(fmt.Sprintf("VALUE: %s = %v (%T)\n", value.Name, value.Value, value.Value))
			}
			b.WriteString("CAPTURED_VALUES_END\n")
		}

		b.WriteString("[MACHINE_READABLE_END]\n")
	}

	return b.String()
}

// ValuePosition represents a value and its position in the expression.
type ValuePosition struct {
	Expression string
	Value      string
	StartPos   int
	EndPos     int
	Depth      int
	Priority   int
}

// formatPowerAssertStyle generates power-assert style visual output.
func (f *VisualFormatter) formatPowerAssertStyle(result *evaluator.ExpressionResult) string {
	expr := result.Expression

	// If no tree, just show the expression and false
	if result.Tree == nil {
		return fmt.Sprintf("  assert(%s)\n         false\n", expr)
	}

	// Extract positions for all meaningful nodes
	positions := f.extractAllPositions(result.Tree, expr)

	// Build visual output
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  assert(%s)\n", expr))

	// Build visual lines - the expression in the output has leading spaces
	lines := f.buildVisualLines(expr, positions)
	for _, line := range lines {
		b.WriteString("         " + line + "\n")
	}

	return b.String()
}

// extractAllPositions extracts all value positions from the tree.
func (f *VisualFormatter) extractAllPositions(tree *evaluator.EvaluationTree, expr string) []ValuePosition {
	var positions []ValuePosition
	f.collectPositions(tree, expr, &positions, make(map[string]bool))

	// Sort by position for consistent output
	sort.Slice(positions, func(i, j int) bool {
		if positions[i].StartPos != positions[j].StartPos {
			return positions[i].StartPos < positions[j].StartPos
		}
		return positions[i].Priority > positions[j].Priority
	})

	return positions
}

// collectPositions recursively collects positions from the tree.
func (f *VisualFormatter) collectPositions(tree *evaluator.EvaluationTree, expr string, positions *[]ValuePosition, seen map[string]bool) {
	if tree == nil {
		return
	}

	// For simple comparisons like "x > 20", we want to show:
	// - The value of x under "x"
	// - The result of the comparison under ">"

	switch tree.Type {
	case "identifier":
		if tree.Value != nil && tree.Text != "" {
			// Find where this identifier appears in the expression
			if pos := strings.Index(expr, tree.Text); pos != -1 {
				key := fmt.Sprintf("%d-%s", pos, tree.Text)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Text,
						Value:      formatValueCompact(tree.Value),
						StartPos:   pos,
						EndPos:     pos + len(tree.Text),
						Priority:   10,
					})
				}
			}
		}

	case "comparison":
		// For comparisons, show the result aligned with the operator
		if tree.Operator != "" {
			if pos := strings.Index(expr, tree.Operator); pos != -1 {
				key := fmt.Sprintf("%d-op-%s", pos, tree.Operator)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Operator,
						Value:      fmt.Sprintf("%v", tree.Result),
						StartPos:   pos,
						EndPos:     pos + len(tree.Operator),
						Priority:   5,
					})
				}
			}
		}

	case "logical":
		// For logical operators, show the result aligned with the operator
		if tree.Operator != "" {
			if pos := strings.Index(expr, tree.Operator); pos != -1 {
				key := fmt.Sprintf("%d-log-%s", pos, tree.Operator)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Operator,
						Value:      fmt.Sprintf("%v", tree.Result),
						StartPos:   pos,
						EndPos:     pos + len(tree.Operator),
						Priority:   3,
					})
				}
			}
		}
	}

	// Process children
	f.collectPositions(tree.Left, expr, positions, seen)
	f.collectPositions(tree.Right, expr, positions, seen)
	for _, child := range tree.Children {
		f.collectPositions(child, expr, positions, seen)
	}
}

// buildVisualLines builds the visual representation lines.
func (f *VisualFormatter) buildVisualLines(expr string, positions []ValuePosition) []string {
	if len(positions) == 0 {
		return []string{"false"}
	}

	// Group positions to avoid overlap
	var lines [][]ValuePosition

	for _, pos := range positions {
		placed := false

		// Try to place in existing line
		for i, line := range lines {
			canPlace := true
			for _, existing := range line {
				// Check if values would overlap
				// For simple cases like "x > 20", we want "10" and "false" on same line
				if pos.StartPos < existing.StartPos+len(existing.Value) &&
					pos.StartPos+len(pos.Value) > existing.StartPos {
					canPlace = false
					break
				}
			}

			if canPlace {
				lines[i] = append(lines[i], pos)
				placed = true
				break
			}
		}

		// Create new line if needed
		if !placed {
			lines = append(lines, []ValuePosition{pos})
		}
	}

	// If everything fits on one line, use simple format
	if len(lines) == 1 && len(lines[0]) == len(positions) {
		// Build pipe line with all positions
		pipeLine := make([]rune, len(expr))
		for i := range pipeLine {
			pipeLine[i] = ' '
		}

		// Build value line
		valueLine := make([]rune, len(expr)+50)
		for i := range valueLine {
			valueLine[i] = ' '
		}

		// First pass: place pipes
		for _, pos := range positions {
			if pos.StartPos < len(pipeLine) {
				pipeLine[pos.StartPos] = '|'
			}
		}

		// Second pass: place values with proper spacing
		lastEnd := -1
		for _, pos := range positions {
			startPos := pos.StartPos

			// Ensure we don't overlap with previous value
			if lastEnd >= 0 && startPos < lastEnd+1 {
				startPos = lastEnd + 1
			}

			for j, ch := range pos.Value {
				if startPos+j < len(valueLine) {
					valueLine[startPos+j] = ch
				}
			}

			lastEnd = startPos + len(pos.Value)
		}

		var result []string
		result = append(result, strings.TrimRight(string(pipeLine), " "))
		result = append(result, strings.TrimRight(string(valueLine), " "))
		return result
	}

	// Multi-line format for overlapping values
	var result []string

	for _, linePositions := range lines {
		// Build pipe line
		pipeLine := make([]rune, len(expr))
		for j := range pipeLine {
			pipeLine[j] = ' '
		}

		// Build value line
		valueLine := make([]rune, len(expr)+50)
		for j := range valueLine {
			valueLine[j] = ' '
		}

		// Place pipes and values for this line
		for _, pos := range linePositions {
			// Place pipe
			if pos.StartPos < len(pipeLine) {
				pipeLine[pos.StartPos] = '|'
			}

			// Place value
			for j, ch := range pos.Value {
				if pos.StartPos+j < len(valueLine) {
					valueLine[pos.StartPos+j] = ch
				}
			}
		}

		// Add both pipe and value lines
		pipeStr := strings.TrimRight(string(pipeLine), " ")
		if pipeStr != "" {
			result = append(result, pipeStr)
		}

		valueStr := strings.TrimRight(string(valueLine), " ")
		if valueStr != "" {
			result = append(result, valueStr)
		}
	}

	return result
}

// formatValueCompact formats a value in a compact way.
func formatValueCompact(v interface{}) string {
	if v == nil {
		return "nil"
	}

	switch val := v.(type) {
	case string:
		if len(val) > 10 {
			return fmt.Sprintf("%q...", val[:10])
		}
		return fmt.Sprintf("%q", val)
	case []int:
		return fmt.Sprintf("%v", val)
	case []string:
		return fmt.Sprintf("%v", val)
	case []interface{}:
		return fmt.Sprintf("%v", val)
	default:
		// For other types, format normally
		s := fmt.Sprintf("%v", val)
		if len(s) > 15 {
			return s[:15] + "..."
		}
		return s
	}
}

// formatMachineSection formats the machine-readable section.
func formatMachineSection(result *evaluator.ExpressionResult) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("EXPR: %s", result.Expression))
	parts = append(parts, fmt.Sprintf("RESULT: %v", result.Result))

	// Add variables
	if len(result.Variables) > 0 {
		var vars []string
		for name, value := range result.Variables {
			vars = append(vars, fmt.Sprintf("%s=%v", name, value))
		}
		sort.Strings(vars)
		parts = append(parts, fmt.Sprintf("VARIABLES: %s", strings.Join(vars, ",")))
	}

	return strings.Join(parts, "\n") + "\n"
}
