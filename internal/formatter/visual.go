// Package formatter provides visual formatting for power-assert style output.
package formatter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

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

// CharPosition represents position information for a character in the expression.
type CharPosition struct {
	BytePos   int // バイト位置
	RunePos   int // ルーン（文字）位置
	VisualPos int // 視覚的な位置（全角文字を考慮）
}

// ValuePosition represents a value and its position in the expression.
type ValuePosition struct {
	Expression string
	Value      string
	StartPos   int     // Byte position in expression
	EndPos     int     // End byte position
	VisualPos  int     // Visual position considering wide characters
	VisualEnd  int     // Visual end position
	Depth      int
	Priority   int
}

// PositionMapper helps map AST nodes to accurate positions.
type PositionMapper struct {
	fset         *token.FileSet
	expr         string
	charPositions []CharPosition
}

// formatPowerAssertStyle generates power-assert style visual output.
func (f *VisualFormatter) formatPowerAssertStyle(result *evaluator.ExpressionResult) string {
	expr := result.Expression

	// If no tree, just show the expression and false
	if result.Tree == nil {
		return fmt.Sprintf("  assert(%s)\n         false\n", expr)
	}

	// Create position mapper for precise positioning
	mapper := f.createPositionMapper(expr)

	// Extract positions using AST-based mapping
	positions := f.extractAllPositionsWithAST(result.Tree, expr, mapper)

	// Build visual output
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  assert(%s)\n", expr))

	// Build visual lines with Unicode-aware positioning
	lines := f.buildUnicodeAwareLines(expr, positions, mapper)
	for _, line := range lines {
		b.WriteString("         " + line + "\n")
	}

	return b.String()
}

// isWideRune determines if a rune is a wide character (占2文字分).
func isWideRune(r rune) bool {
	return unicode.In(r,
		unicode.Hiragana,
		unicode.Katakana,
		unicode.Han,
		unicode.Hangul,
	) || (r >= 0xFF00 && r <= 0xFFEF)
}

// visualWidth calculates the visual width of a string considering wide characters.
func visualWidth(s string) int {
	width := 0
	for _, r := range s {
		if isWideRune(r) {
			width += 2 // 全角は2文字分
		} else {
			width++ // 半角は1文字分
		}
	}
	return width
}

// createPositionMapper creates a position mapper for the expression.
func (f *VisualFormatter) createPositionMapper(expr string) *PositionMapper {
	fset := token.NewFileSet()
	charPositions := f.calculateCharPositions(expr)
	
	return &PositionMapper{
		fset:          fset,
		expr:          expr,
		charPositions: charPositions,
	}
}

// calculateCharPositions calculates position information for each character.
func (f *VisualFormatter) calculateCharPositions(s string) []CharPosition {
	positions := make([]CharPosition, 0, len(s))
	
	bytePos := 0
	runePos := 0
	visualPos := 0
	
	for _, r := range s {
		positions = append(positions, CharPosition{
			BytePos:   bytePos,
			RunePos:   runePos,
			VisualPos: visualPos,
		})
		
		// Calculate next positions
		runeLen := utf8.RuneLen(r)
		bytePos += runeLen
		runePos++
		
		if isWideRune(r) {
			visualPos += 2
		} else {
			visualPos++
		}
	}
	
	return positions
}

// extractAllPositionsWithAST extracts positions using AST-based mapping.
func (f *VisualFormatter) extractAllPositionsWithAST(tree *evaluator.EvaluationTree, expr string, mapper *PositionMapper) []ValuePosition {
	var positions []ValuePosition
	
	// Parse expression to get AST
	node, err := parser.ParseExpr(expr)
	if err != nil {
		// Fallback to simple position extraction
		return f.extractAllPositions(tree, expr)
	}
	
	// Use AST to find precise positions
	f.collectPositionsWithAST(tree, node, expr, mapper, &positions, make(map[string]bool))

	// Sort by visual position for consistent output
	sort.Slice(positions, func(i, j int) bool {
		if positions[i].VisualPos != positions[j].VisualPos {
			return positions[i].VisualPos < positions[j].VisualPos
		}
		return positions[i].Priority > positions[j].Priority
	})

	return positions
}

// extractAllPositions extracts all value positions from the tree (fallback method).
func (f *VisualFormatter) extractAllPositions(tree *evaluator.EvaluationTree, expr string) []ValuePosition {
	var positions []ValuePosition
	mapper := f.createPositionMapper(expr)
	f.collectPositions(tree, expr, mapper, &positions, make(map[string]bool))

	// Sort by visual position for consistent output
	sort.Slice(positions, func(i, j int) bool {
		if positions[i].VisualPos != positions[j].VisualPos {
			return positions[i].VisualPos < positions[j].VisualPos
		}
		return positions[i].Priority > positions[j].Priority
	})

	return positions
}

// collectPositionsWithAST collects positions using AST node mapping.
func (f *VisualFormatter) collectPositionsWithAST(tree *evaluator.EvaluationTree, astNode ast.Expr, expr string, mapper *PositionMapper, positions *[]ValuePosition, seen map[string]bool) {
	if tree == nil {
		return
	}

	// Map tree text to AST nodes for precise positioning
	var targetNode ast.Node
	ast.Inspect(astNode, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		
		// Check if this AST node corresponds to our tree node
		if f.nodeMatches(n, tree, expr) {
			targetNode = n
			return false
		}
		
		return true
	})

	if targetNode != nil {
		pos := mapper.fset.Position(targetNode.Pos())
		end := mapper.fset.Position(targetNode.End())
		
		// Convert byte positions to visual positions
		startVisual := f.byteToVisualPos(pos.Offset, mapper.charPositions)
		endVisual := f.byteToVisualPos(end.Offset, mapper.charPositions)
		
		switch tree.Type {
		case "identifier":
			if tree.Value != nil && tree.Text != "" {
				key := fmt.Sprintf("%d-%s", startVisual, tree.Text)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Text,
						Value:      formatValueCompact(tree.Value),
						StartPos:   pos.Offset,
						EndPos:     end.Offset,
						VisualPos:  startVisual,
						VisualEnd:  endVisual,
						Priority:   10,
					})
				}
			}
			
		case "comparison", "logical":
			if tree.Operator != "" {
				// Find operator position within the node
				opPos := f.findOperatorInNode(targetNode, tree.Operator, mapper)
				opVisual := f.byteToVisualPos(opPos, mapper.charPositions)
				
				key := fmt.Sprintf("%d-op-%s", opVisual, tree.Operator)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Operator,
						Value:      fmt.Sprintf("%v", tree.Result),
						StartPos:   opPos,
						EndPos:     opPos + len(tree.Operator),
						VisualPos:  opVisual,
						VisualEnd:  opVisual + visualWidth(tree.Operator),
						Priority:   5,
					})
				}
			}
		}
	}

	// Process children recursively
	f.processChildrenWithAST(tree, astNode, expr, mapper, positions, seen)
}

// collectPositions recursively collects positions from the tree (fallback method).
func (f *VisualFormatter) collectPositions(tree *evaluator.EvaluationTree, expr string, mapper *PositionMapper, positions *[]ValuePosition, seen map[string]bool) {
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
				visualPos := f.byteToVisualPos(pos, mapper.charPositions)
				key := fmt.Sprintf("%d-%s", visualPos, tree.Text)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Text,
						Value:      formatValueCompact(tree.Value),
						StartPos:   pos,
						EndPos:     pos + len(tree.Text),
						VisualPos:  visualPos,
						VisualEnd:  visualPos + visualWidth(tree.Text),
						Priority:   10,
					})
				}
			}
		}

	case "comparison":
		// For comparisons, show the result aligned with the operator
		if tree.Operator != "" {
			if pos := strings.Index(expr, tree.Operator); pos != -1 {
				visualPos := f.byteToVisualPos(pos, mapper.charPositions)
				key := fmt.Sprintf("%d-op-%s", visualPos, tree.Operator)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Operator,
						Value:      fmt.Sprintf("%v", tree.Result),
						StartPos:   pos,
						EndPos:     pos + len(tree.Operator),
						VisualPos:  visualPos,
						VisualEnd:  visualPos + visualWidth(tree.Operator),
						Priority:   5,
					})
				}
			}
		}

	case "logical":
		// For logical operators, show the result aligned with the operator
		if tree.Operator != "" {
			if pos := strings.Index(expr, tree.Operator); pos != -1 {
				visualPos := f.byteToVisualPos(pos, mapper.charPositions)
				key := fmt.Sprintf("%d-log-%s", visualPos, tree.Operator)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Operator,
						Value:      fmt.Sprintf("%v", tree.Result),
						StartPos:   pos,
						EndPos:     pos + len(tree.Operator),
						VisualPos:  visualPos,
						VisualEnd:  visualPos + visualWidth(tree.Operator),
						Priority:   3,
					})
				}
			}
		}
	}

	// Process children
	f.collectPositions(tree.Left, expr, mapper, positions, seen)
	f.collectPositions(tree.Right, expr, mapper, positions, seen)
	for _, child := range tree.Children {
		f.collectPositions(child, expr, mapper, positions, seen)
	}
}

// Helper functions for AST-based positioning

// nodeMatches checks if an AST node matches an evaluation tree node.
func (f *VisualFormatter) nodeMatches(astNode ast.Node, tree *evaluator.EvaluationTree, expr string) bool {
	switch n := astNode.(type) {
	case *ast.Ident:
		return tree.Type == "identifier" && n.Name == tree.Text
	case *ast.BinaryExpr:
		return (tree.Type == "comparison" || tree.Type == "logical") && n.Op.String() == tree.Operator
	case *ast.SelectorExpr:
		return tree.Type == "selector" && strings.Contains(tree.Text, ".")
	}
	return false
}

// findOperatorInNode finds the operator position within an AST node.
func (f *VisualFormatter) findOperatorInNode(astNode ast.Node, operator string, mapper *PositionMapper) int {
	if binExpr, ok := astNode.(*ast.BinaryExpr); ok {
		leftEnd := mapper.fset.Position(binExpr.X.End()).Offset
		rightStart := mapper.fset.Position(binExpr.Y.Pos()).Offset
		
		// Search for operator between left and right operands
		if leftEnd >= 0 && rightStart <= len(mapper.expr) {
			betweenStr := mapper.expr[leftEnd:rightStart]
			opIndex := strings.Index(betweenStr, operator)
			if opIndex >= 0 {
				return leftEnd + opIndex
			}
		}
		return leftEnd
	}
	return 0
}

// processChildrenWithAST processes child nodes recursively.
func (f *VisualFormatter) processChildrenWithAST(tree *evaluator.EvaluationTree, astNode ast.Expr, expr string, mapper *PositionMapper, positions *[]ValuePosition, seen map[string]bool) {
	// Process evaluation tree children
	if tree.Left != nil {
		f.collectPositionsWithAST(tree.Left, astNode, expr, mapper, positions, seen)
	}
	if tree.Right != nil {
		f.collectPositionsWithAST(tree.Right, astNode, expr, mapper, positions, seen)
	}
	for _, child := range tree.Children {
		f.collectPositionsWithAST(child, astNode, expr, mapper, positions, seen)
	}
}

// byteToVisualPos converts byte position to visual position.
func (f *VisualFormatter) byteToVisualPos(bytePos int, charPositions []CharPosition) int {
	if bytePos < 0 || len(charPositions) == 0 {
		return 0
	}
	
	// Find the character at or before this byte position
	for i := len(charPositions) - 1; i >= 0; i-- {
		if charPositions[i].BytePos <= bytePos {
			// Calculate offset within this character
			offset := bytePos - charPositions[i].BytePos
			return charPositions[i].VisualPos + offset
		}
	}
	
	return 0
}

// valuesOverlap checks if two values would overlap visually.
func valuesOverlap(a, b ValuePosition) bool {
	aEnd := a.VisualPos + visualWidth(a.Value)
	bStart := b.VisualPos
	bEnd := b.VisualPos + visualWidth(b.Value)
	aStart := a.VisualPos
	
	// Ensure at least 1 character spacing
	return !(aEnd+1 <= bStart || bEnd+1 <= aStart)
}

// buildUnicodeAwareLines builds visual lines with Unicode support.
func (f *VisualFormatter) buildUnicodeAwareLines(expr string, positions []ValuePosition, mapper *PositionMapper) []string {
	if len(positions) == 0 {
		return []string{"false"}
	}

	// Group positions to avoid overlap using visual positioning
	var lines [][]ValuePosition

	for _, pos := range positions {
		placed := false

		// Try to place in existing line
		for i, line := range lines {
			canPlace := true
			for _, existing := range line {
				if valuesOverlap(pos, existing) {
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

	// Build output lines using visual positioning
	var result []string
	exprVisualWidth := visualWidth(expr)

	for _, linePositions := range lines {
		// Build pipe line and value line
		pipeLine := make([]rune, exprVisualWidth+1)
		valueLine := make([]rune, exprVisualWidth+100) // Extra space for values
		
		// Initialize with spaces
		for i := range pipeLine {
			pipeLine[i] = ' '
		}
		for i := range valueLine {
			valueLine[i] = ' '
		}

		// Place pipes and values for this line
		for _, pos := range linePositions {
			// Place pipe at visual position
			if pos.VisualPos < len(pipeLine) {
				pipeLine[pos.VisualPos] = '|'
			}

			// Place value, ensuring no overlap
			valueRunes := []rune(pos.Value)
			valuePos := pos.VisualPos
			
			// Adjust position if it would cause overlap
			for _, other := range linePositions {
				if other.VisualPos < pos.VisualPos && 
				   other.VisualPos + visualWidth(other.Value) + 1 > valuePos {
					valuePos = other.VisualPos + visualWidth(other.Value) + 1
				}
			}
			
			// Place the value
			for i, r := range valueRunes {
				if valuePos+i < len(valueLine) {
					valueLine[valuePos+i] = r
				}
			}
		}

		// Add pipe line
		pipeStr := strings.TrimRight(string(pipeLine), " ")
		if pipeStr != "" {
			result = append(result, pipeStr)
		}

		// Add value line  
		valueStr := strings.TrimRight(string(valueLine), " ")
		if valueStr != "" {
			result = append(result, valueStr)
		}
	}

	return result
}

// buildVisualLines builds the visual representation lines (fallback method).
func (f *VisualFormatter) buildVisualLines(expr string, positions []ValuePosition) []string {
	if len(positions) == 0 {
		return []string{"false"}
	}

	// Use Unicode-aware line building (fallback to byte positioning)
	return f.buildUnicodeAwareLinesFallback(expr, positions)
}

// buildUnicodeAwareLinesFallback builds lines using visual positioning (fallback for when AST parsing fails).
func (f *VisualFormatter) buildUnicodeAwareLinesFallback(expr string, positions []ValuePosition) []string {
	// Group positions to avoid overlap using visual positioning
	var lines [][]ValuePosition

	for _, pos := range positions {
		placed := false

		// Try to place in existing line
		for i, line := range lines {
			canPlace := true
			for _, existing := range line {
				// Check overlap using visual positions if available, otherwise use byte positions
				if pos.VisualPos > 0 && existing.VisualPos > 0 {
					if valuesOverlap(pos, existing) {
						canPlace = false
						break
					}
				} else {
					// Fallback to byte-based overlap detection
					if pos.StartPos < existing.StartPos+len(existing.Value) &&
						pos.StartPos+len(pos.Value) > existing.StartPos {
						canPlace = false
						break
					}
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

	// Build output lines
	var result []string
	exprVisualWidth := visualWidth(expr)

	for _, linePositions := range lines {
		// Build pipe line and value line using visual positioning when available
		pipeLine := make([]rune, exprVisualWidth+1)
		valueLine := make([]rune, exprVisualWidth+100)
		
		// Initialize with spaces
		for i := range pipeLine {
			pipeLine[i] = ' '
		}
		for i := range valueLine {
			valueLine[i] = ' '
		}

		// Place pipes and values for this line
		for _, pos := range linePositions {
			visualPos := pos.VisualPos
			if visualPos == 0 {
				// Fallback: convert byte position to visual position
				visualPos = f.byteToVisualPosFallback(pos.StartPos, expr)
			}

			// Place pipe at visual position
			if visualPos < len(pipeLine) {
				pipeLine[visualPos] = '|'
			}

			// Place value
			valueRunes := []rune(pos.Value)
			for i, r := range valueRunes {
				if visualPos+i < len(valueLine) {
					valueLine[visualPos+i] = r
				}
			}
		}

		// Add pipe line
		pipeStr := strings.TrimRight(string(pipeLine), " ")
		if pipeStr != "" {
			result = append(result, pipeStr)
		}

		// Add value line  
		valueStr := strings.TrimRight(string(valueLine), " ")
		if valueStr != "" {
			result = append(result, valueStr)
		}
	}

	return result
}

// byteToVisualPosFallback converts byte position to visual position without charPositions array.
func (f *VisualFormatter) byteToVisualPosFallback(bytePos int, expr string) int {
	if bytePos <= 0 {
		return 0
	}
	
	// Count visual width up to the byte position
	visualPos := 0
	currentByte := 0
	
	for _, r := range expr {
		if currentByte >= bytePos {
			break
		}
		
		if isWideRune(r) {
			visualPos += 2
		} else {
			visualPos++
		}
		
		currentByte += utf8.RuneLen(r)
	}
	
	return visualPos
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
