// Package formatter provides visual formatting for power-assert style output.
package formatter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
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
	Expression  string
	Value       string
	StartPos    int // Byte position in expression
	EndPos      int // End byte position
	VisualPos   int // Visual position considering wide characters
	VisualEnd   int // Visual end position
	Depth       int // Depth in the AST tree for proper layering
	Priority    int // Priority for positioning (higher = more important)
	VisualLayer int // Visual layer assigned for rendering (separate from semantic depth)
}

// VisualNode separates evaluation depth from visual layers for rendering
type VisualNode struct {
	Position      ValuePosition
	EvaluationDepth int      // Semantic depth in the evaluation tree
	VisualLayer   int      // Visual layer for rendering (may differ from evaluation depth)
	PipePosition  int      // Fixed pipe position in the expression
	ConflictsWith []int    // Indices of other nodes this conflicts with
}

// LayerAssignment manages the assignment of values to visual layers
type LayerAssignment struct {
	MaxLayer int                    // Maximum layer number assigned
	Layers   [][]VisualNode         // Nodes grouped by visual layer
	PipePositions map[int]bool      // Track fixed pipe positions
}

// Range represents an occupied interval in a visual layer
type Range struct {
	Start, End int
}

// ConflictMap tracks which positions conflict with each other
type ConflictMap struct {
	Conflicts map[int][]int // Map from position index to list of conflicting positions
}

// PositionMapper helps map AST nodes to accurate positions.
type PositionMapper struct {
	fset          *token.FileSet
	expr          string
	charPositions []CharPosition
}

// formatPowerAssertStyle generates power-assert style visual output.
func (f *VisualFormatter) formatPowerAssertStyle(result *evaluator.ExpressionResult) string {
	expr := result.Expression

	// If no tree, show the expression with proper pipe alignment
	if result.Tree == nil {
		return f.formatSimpleAssertStyle(expr)
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

// formatSimpleAssertStyle formats basic assert style when no tree is available.
func (f *VisualFormatter) formatSimpleAssertStyle(expr string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  assert(%s)\n", expr))

	// Add a simple pipe under the end of the expression to show false
	exprVisualWidth := visualWidth(expr)
	padding := strings.Repeat(" ", exprVisualWidth)
	b.WriteString(fmt.Sprintf("         %s|\n", padding))
	b.WriteString(fmt.Sprintf("         %sfalse\n", padding))

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

	// If AST-based approach didn't find positions, fallback to simple method
	if len(positions) == 0 {
		return f.extractAllPositions(tree, expr)
	}

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
	f.collectPositionsWithASTDepth(tree, astNode, expr, mapper, positions, seen, 0)
}

// collectPositionsWithASTDepth collects positions using AST node mapping with depth tracking.
func (f *VisualFormatter) collectPositionsWithASTDepth(tree *evaluator.EvaluationTree, astNode ast.Expr, expr string, mapper *PositionMapper, positions *[]ValuePosition, seen map[string]bool, depth int) {
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
		// Get accurate positions using AST node positions
		startPos, endPos := f.getASTNodePosition(targetNode, mapper)
		startVisual := f.byteToVisualPos(startPos, mapper.charPositions)
		endVisual := f.byteToVisualPos(endPos, mapper.charPositions)

		switch tree.Type {
		case "identifier":
			if tree.Value != nil && tree.Text != "" {
				key := fmt.Sprintf("%d-%s", startVisual, tree.Text)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Text,
						Value:      formatValueCompact(tree.Value),
						StartPos:   startPos,
						EndPos:     endPos,
						VisualPos:  startVisual,
						VisualEnd:  endVisual,
						Depth:      depth,
						Priority:   20, // Higher priority for identifier values
					})
				}
			}

		case "literal":
			if tree.Value != nil && tree.Text != "" {
				key := fmt.Sprintf("%d-lit-%s", startVisual, tree.Text)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Text,
						Value:      formatValueCompact(tree.Value),
						StartPos:   startPos,
						EndPos:     endPos,
						VisualPos:  startVisual,
						VisualEnd:  endVisual,
						Depth:      depth,
						Priority:   15, // Show literal values too
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
						Depth:      depth + 1, // Operator result at deeper level than operands
						Priority:   5,
					})
				}
			}
		}
	}

	// Process children with proper depth hierarchy for power-assert style display
	// For binary expressions like "x > 20", we want values (15, 20) at one level,
	// then operation result (false) at the next level
	var nextDepth int
	if tree.Type == "comparison" || tree.Type == "logical" {
		// For operators, their operands should be at a deeper level to show values first,
		// then the operation result at the current level
		nextDepth = depth + 1
	} else {
		// For other types, increase depth for children
		nextDepth = depth + 1
	}

	f.processChildrenWithASTDepth(tree, astNode, expr, mapper, positions, seen, nextDepth)
}

// collectPositions recursively collects positions from the tree (fallback method).
func (f *VisualFormatter) collectPositions(tree *evaluator.EvaluationTree, expr string, mapper *PositionMapper, positions *[]ValuePosition, seen map[string]bool) {
	f.collectPositionsDepth(tree, expr, mapper, positions, seen, 0)
}

// collectPositionsDepth recursively collects positions from the tree with depth tracking (fallback method).
func (f *VisualFormatter) collectPositionsDepth(tree *evaluator.EvaluationTree, expr string, mapper *PositionMapper, positions *[]ValuePosition, seen map[string]bool, depth int) {
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
				// Fallback to byte position if visual position calculation fails
				if visualPos == 0 && pos != 0 {
					visualPos = pos
				}
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
						Depth:      depth,
						Priority:   20, // Higher priority for identifier values
					})
				}
			}
		}

	case "literal":
		if tree.Value != nil && tree.Text != "" {
			// Find where this literal appears in the expression
			if pos := strings.Index(expr, tree.Text); pos != -1 {
				visualPos := f.byteToVisualPos(pos, mapper.charPositions)
				// Fallback to byte position if visual position calculation fails
				if visualPos == 0 && pos != 0 {
					visualPos = pos
				}
				key := fmt.Sprintf("%d-lit-%s", visualPos, tree.Text)
				if !seen[key] {
					seen[key] = true
					*positions = append(*positions, ValuePosition{
						Expression: tree.Text,
						Value:      formatValueCompact(tree.Value),
						StartPos:   pos,
						EndPos:     pos + len(tree.Text),
						VisualPos:  visualPos,
						VisualEnd:  visualPos + visualWidth(tree.Text),
						Depth:      depth,
						Priority:   15, // Show literal values too
					})
				}
			}
		}

	case "comparison":
		// For comparisons, show the result aligned with the operator
		if tree.Operator != "" {
			if pos := strings.Index(expr, tree.Operator); pos != -1 {
				visualPos := f.byteToVisualPos(pos, mapper.charPositions)
				// Fallback to byte position if visual position calculation fails
				if visualPos == 0 && pos != 0 {
					visualPos = pos
				}
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
						Depth:      depth + 1, // Operator result at deeper level than operands
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
				// Fallback to byte position if visual position calculation fails
				if visualPos == 0 && pos != 0 {
					visualPos = pos
				}
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
						Depth:      depth + 1, // Operator result at deeper level than operands
						Priority:   3,
					})
				}
			}
		}
	}

	// Process children with proper depth hierarchy for power-assert style display
	// For binary expressions like "x > 20", we want values (15, 20) at one level,
	// then operation result (false) at the next level
	var nextDepth int
	if tree.Type == "comparison" || tree.Type == "logical" {
		// For operators, their operands should be at a deeper level to show values first,
		// then the operation result at the current level
		nextDepth = depth + 1
	} else {
		// For other types, increase depth for children
		nextDepth = depth + 1
	}

	f.collectPositionsDepth(tree.Left, expr, mapper, positions, seen, nextDepth)
	f.collectPositionsDepth(tree.Right, expr, mapper, positions, seen, nextDepth)
	for _, child := range tree.Children {
		f.collectPositionsDepth(child, expr, mapper, positions, seen, nextDepth)
	}
}

// Helper functions for AST-based positioning

// nodeMatches checks if an AST node matches an evaluation tree node.
func (f *VisualFormatter) nodeMatches(astNode ast.Node, tree *evaluator.EvaluationTree, expr string) bool {
	switch n := astNode.(type) {
	case *ast.Ident:
		return tree.Type == "identifier" && n.Name == tree.Text
	case *ast.BasicLit:
		return tree.Type == "literal" && n.Value == tree.Text
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

// processChildrenWithASTDepth processes child nodes recursively with depth tracking.
func (f *VisualFormatter) processChildrenWithASTDepth(tree *evaluator.EvaluationTree, astNode ast.Expr, expr string, mapper *PositionMapper, positions *[]ValuePosition, seen map[string]bool, depth int) {
	// Process evaluation tree children
	if tree.Left != nil {
		f.collectPositionsWithASTDepth(tree.Left, astNode, expr, mapper, positions, seen, depth)
	}
	if tree.Right != nil {
		f.collectPositionsWithASTDepth(tree.Right, astNode, expr, mapper, positions, seen, depth)
	}
	for _, child := range tree.Children {
		f.collectPositionsWithASTDepth(child, astNode, expr, mapper, positions, seen, depth)
	}
}

// getASTNodePosition gets the byte position range of an AST node.
func (f *VisualFormatter) getASTNodePosition(node ast.Node, mapper *PositionMapper) (int, int) {
	if node == nil {
		return 0, 0
	}

	startPos := mapper.fset.Position(node.Pos())
	endPos := mapper.fset.Position(node.End())

	// Return byte positions relative to the expression
	return startPos.Offset, endPos.Offset
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

	// Check for actual overlap (no spacing requirement for power-assert style)
	return !(aEnd <= bStart || bEnd <= aStart)
}

// buildUnicodeAwareLines builds visual lines with Unicode support and the new layer architecture.
func (f *VisualFormatter) buildUnicodeAwareLines(expr string, positions []ValuePosition, mapper *PositionMapper) []string {
	if len(positions) == 0 {
		return []string{"false"}
	}

	// Fix visual positions by calculating them correctly based on expression content
	correctedPositions := f.correctVisualPositions(positions, expr)

	// Use the new layer-based architecture instead of depth-based grouping
	return f.buildPowerAssertTreeWithLayers(expr, correctedPositions)
}

// correctVisualPositions fixes visual positions by finding actual positions in the expression
func (f *VisualFormatter) correctVisualPositions(positions []ValuePosition, expr string) []ValuePosition {
	corrected := make([]ValuePosition, len(positions))
	copy(corrected, positions)

	for i := range corrected {
		pos := &corrected[i]

		// Find the actual position of this expression element in the text
		actualPos := f.findActualPosition(pos.Expression, expr)
		if actualPos >= 0 {
			pos.VisualPos = actualPos
			pos.VisualEnd = actualPos + visualWidth(pos.Expression)
		}
	}

	return corrected
}

// findActualPosition finds the position of an expression element in the source text
func (f *VisualFormatter) findActualPosition(element string, expr string) int {
	// For simple identifiers and operators, use string search
	if pos := strings.Index(expr, element); pos != -1 {
		return pos
	}

	// For operators that might have whitespace variations, try with spaces
	if len(element) == 1 || (len(element) == 2 && strings.Contains(element, "=")) {
		spaced := " " + element + " "
		if pos := strings.Index(expr, spaced); pos != -1 {
			return pos + 1 // Return position of the operator, not the leading space
		}
	}

	return -1
}

// assignVisualLayers assigns values to visual layers using greedy algorithm to minimize layers
func (f *VisualFormatter) assignVisualLayers(positions []ValuePosition) LayerAssignment {
	assignment := LayerAssignment{
		MaxLayer:      0,
		Layers:        make([][]VisualNode, 0),
		PipePositions: make(map[int]bool),
	}
	
	// Convert positions to visual nodes
	nodes := make([]VisualNode, len(positions))
	for i, pos := range positions {
		nodes[i] = VisualNode{
			Position:        pos,
			EvaluationDepth: pos.Depth,
			VisualLayer:     -1, // Unassigned
			PipePosition:    pos.VisualPos,
		}
		assignment.PipePositions[pos.VisualPos] = true
	}
	
	// Sort nodes by priority (higher priority gets better layer assignment)
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Position.Priority != nodes[j].Position.Priority {
			return nodes[i].Position.Priority > nodes[j].Position.Priority
		}
		return nodes[i].PipePosition < nodes[j].PipePosition
	})
	
	// Assign each node to the lowest available layer
	for i := range nodes {
		layerAssigned := false
		
		// Try to place in existing layers
		for layerIdx := 0; layerIdx < len(assignment.Layers); layerIdx++ {
			if f.canPlaceInLayer(nodes[i], assignment.Layers[layerIdx]) {
				assignment.Layers[layerIdx] = append(assignment.Layers[layerIdx], nodes[i])
				nodes[i].VisualLayer = layerIdx
				layerAssigned = true
				break
			}
		}
		
		// Create new layer if needed
		if !layerAssigned {
			newLayer := []VisualNode{nodes[i]}
			assignment.Layers = append(assignment.Layers, newLayer)
			nodes[i].VisualLayer = len(assignment.Layers) - 1
			assignment.MaxLayer = nodes[i].VisualLayer
		}
	}
	
	return assignment
}

// canPlaceInLayer checks if a node can be placed in the given layer without conflicts
func (f *VisualFormatter) canPlaceInLayer(node VisualNode, layer []VisualNode) bool {
	nodeRange := f.getValueRange(node)
	
	for _, existing := range layer {
		existingRange := f.getValueRange(existing)
		if f.rangesOverlap(nodeRange.Start, nodeRange.End, existingRange.Start, existingRange.End) {
			return false
		}
	}
	return true
}

// getValueRange calculates the display range for a value
func (f *VisualFormatter) getValueRange(node VisualNode) Range {
	valueWidth := visualWidth(node.Position.Value)
	startPos := node.PipePosition
	endPos := startPos + valueWidth
	return Range{Start: startPos, End: endPos}
}

// buildPowerAssertTreeWithLayers builds power-assert tree using visual layers
func (f *VisualFormatter) buildPowerAssertTreeWithLayers(expr string, positions []ValuePosition) []string {
	if len(positions) == 0 {
		return []string{"false"}
	}
	
	// Assign values to visual layers
	layerAssignment := f.assignVisualLayers(positions)
	
	var result []string
	exprWidth := visualWidth(expr)
	
	// Build each visual layer
	for layerIdx := 0; layerIdx <= layerAssignment.MaxLayer; layerIdx++ {
		if layerIdx >= len(layerAssignment.Layers) {
			continue
		}
		
		layer := layerAssignment.Layers[layerIdx]
		if len(layer) == 0 {
			continue
		}
		
		// Build pipe line with all relevant pipe positions
		pipeLine := strings.Repeat(" ", exprWidth+100)
		pipeRunes := []rune(pipeLine)
		
		// Place pipes for values in this layer AND pipes that continue from deeper layers
		for pipePos := range layerAssignment.PipePositions {
			showPipe := false
			
			// Show pipe if there's a value at this layer
			for _, node := range layer {
				if node.PipePosition == pipePos {
					showPipe = true
					break
				}
			}
			
			// Show pipe if there are values at deeper layers
			if !showPipe && layerIdx < layerAssignment.MaxLayer {
				for futureLayerIdx := layerIdx + 1; futureLayerIdx < len(layerAssignment.Layers); futureLayerIdx++ {
					for _, futureNode := range layerAssignment.Layers[futureLayerIdx] {
						if futureNode.PipePosition == pipePos {
							showPipe = true
							break
						}
					}
					if showPipe {
						break
					}
				}
			}
			
			if showPipe && pipePos < len(pipeRunes) {
				pipeRunes[pipePos] = '|'
			}
		}
		
		// Add pipe line
		pipeStr := strings.TrimRight(string(pipeRunes), " ")
		if pipeStr != "" {
			result = append(result, pipeStr)
		}
		
		// Build value line
		valueLine := strings.Repeat(" ", exprWidth+100)
		valueRunes := []rune(valueLine)
		
		// Place values at their exact pipe positions (no smart centering)
		for _, node := range layer {
			valueText := []rune(node.Position.Value)
			startPos := node.PipePosition
			
			// Place value directly under its pipe
			for i, r := range valueText {
				if startPos+i < len(valueRunes) {
					valueRunes[startPos+i] = r
				}
			}
		}
		
		// Add value line
		valueStr := strings.TrimRight(string(valueRunes), " ")
		if valueStr != "" {
			result = append(result, valueStr)
		}
		
		// Add spacing between layers (except for the last layer)
		if layerIdx < layerAssignment.MaxLayer {
			result = append(result, "")
		}
	}
	
	return result
}

// buildPowerAssertTree builds the classic power-assert tree structure (DEPRECATED - use buildPowerAssertTreeWithLayers)
func (f *VisualFormatter) buildPowerAssertTree(expr string, depthGroups [][]ValuePosition) []string {
	if len(depthGroups) == 0 {
		return []string{"false"}
	}

	var result []string
	exprWidth := visualWidth(expr)

	// Collect all unique positions across all depths for connecting pipes
	allPositions := make(map[int]bool)
	for _, positions := range depthGroups {
		for _, pos := range positions {
			allPositions[pos.VisualPos] = true
		}
	}

	// Process each depth level separately to show clear hierarchy
	for depthIdx, positions := range depthGroups {
		if len(positions) == 0 {
			continue
		}

		// Sort positions by their visual position for consistent ordering
		sort.Slice(positions, func(i, j int) bool {
			return positions[i].VisualPos < positions[j].VisualPos
		})

		// Build pipe line with pipes for all relevant positions
		pipeLine := strings.Repeat(" ", exprWidth+100)
		pipeRunes := []rune(pipeLine)

		// Place pipes at positions that have values at this depth OR deeper depths
		for pipePos := range allPositions {
			// Show pipe if there's a value at this depth OR if there are deeper values at this position
			showPipe := false
			
			// Check if this position has a value at current depth
			for _, pos := range positions {
				if pos.VisualPos == pipePos {
					showPipe = true
					break
				}
			}
			
			// Check if this position has values at deeper depths
			if !showPipe && depthIdx < len(depthGroups)-1 {
				for futureDepthIdx := depthIdx + 1; futureDepthIdx < len(depthGroups); futureDepthIdx++ {
					for _, futurePos := range depthGroups[futureDepthIdx] {
						if futurePos.VisualPos == pipePos {
							showPipe = true
							break
						}
					}
					if showPipe {
						break
					}
				}
			}
			
			if showPipe && pipePos < len(pipeRunes) {
				pipeRunes[pipePos] = '|'
			}
		}

		// Add the pipe line
		pipeStr := strings.TrimRight(string(pipeRunes), " ")
		if pipeStr != "" {
			result = append(result, pipeStr)
		}

		// Build value line with non-overlapping positioning
		valueLine := strings.Repeat(" ", exprWidth+100)
		valueRunes := []rune(valueLine)

		// Place values with smart positioning and guaranteed no overlaps
		usedRanges := make([][]int, 0) // Keep track of used ranges [start, end]
		for _, pos := range positions {
			valueText := []rune(pos.Value)
			valueWidth := visualWidth(pos.Value) // Use visualWidth for accurate width calculation
			
			// Find the best position for this value with smart alignment
			bestPos := f.findSmartValuePosition(pos, valueWidth, usedRanges, exprWidth)
			
			// Mark range as used
			usedRanges = append(usedRanges, []int{bestPos, bestPos + valueWidth})

			// Place the value
			for i, r := range valueText {
				if bestPos+i < len(valueRunes) {
					valueRunes[bestPos+i] = r
				}
			}
		}

		// Add value line
		valueStr := strings.TrimRight(string(valueRunes), " ")
		if valueStr != "" {
			result = append(result, valueStr)
		}

		// Add spacing between depth levels (except for the last one)
		if depthIdx < len(depthGroups)-1 {
			// Add blank line only if next depth has content
			if depthIdx+1 < len(depthGroups) && len(depthGroups[depthIdx+1]) > 0 {
				result = append(result, "")
			}
		}
	}

	return result
}

// findSmartValuePosition finds the best position for a value with smart alignment
func (f *VisualFormatter) findSmartValuePosition(pos ValuePosition, valueWidth int, usedRanges [][]int, exprWidth int) int {
	// Calculate expression element width for smart centering
	exprElementWidth := visualWidth(pos.Expression)
	preferredPos := pos.VisualPos
	
	// For long values or long expressions, try to center the value under the expression
	if valueWidth > 3 || exprElementWidth > 3 {
		// Try to center the value under the expression element
		centerOffset := (exprElementWidth - valueWidth) / 2
		centeredPos := preferredPos + centerOffset
		
		// Ensure centered position is not negative
		if centeredPos >= 0 && f.isRangeAvailable(centeredPos, centeredPos+valueWidth, usedRanges) {
			return centeredPos
		}
	}
	
	// Fallback to original logic
	return f.findBestValuePositionWithRanges(preferredPos, valueWidth, usedRanges, exprWidth)
}

// findBestValuePositionWithRanges finds the best position for a value to avoid overlaps using ranges
func (f *VisualFormatter) findBestValuePositionWithRanges(preferredPos, valueWidth int, usedRanges [][]int, exprWidth int) int {
	// First, try the preferred position (directly under the pipe)
	if f.isRangeAvailable(preferredPos, preferredPos+valueWidth, usedRanges) {
		return preferredPos
	}

	// If preferred position is not available, try positions to the right with adequate spacing
	for offset := 1; offset <= 30; offset++ {
		candidatePos := preferredPos + offset
		if candidatePos+valueWidth < exprWidth+80 && f.isRangeAvailable(candidatePos, candidatePos+valueWidth, usedRanges) {
			return candidatePos
		}
	}

	// If no position to the right works, try positions to the left
	for offset := 1; offset <= 20; offset++ {
		candidatePos := preferredPos - offset
		if candidatePos >= 0 && f.isRangeAvailable(candidatePos, candidatePos+valueWidth, usedRanges) {
			return candidatePos
		}
	}

	// Fallback: find any available position
	for pos := 0; pos < exprWidth+50; pos++ {
		if f.isRangeAvailable(pos, pos+valueWidth, usedRanges) {
			return pos
		}
	}

	// Ultimate fallback: use the preferred position anyway
	return preferredPos
}

// isRangeAvailable checks if a range doesn't overlap with any used ranges
func (f *VisualFormatter) isRangeAvailable(start, end int, usedRanges [][]int) bool {
	for _, usedRange := range usedRanges {
		usedStart, usedEnd := usedRange[0], usedRange[1]
		// Check if ranges overlap (with 1-character padding for safety)
		if !(end+1 <= usedStart || start >= usedEnd+1) {
			return false
		}
	}
	return true
}

// findBestValuePosition finds the best position for a value to avoid overlaps
func (f *VisualFormatter) findBestValuePosition(preferredPos, valueWidth int, usedPositions map[int]bool, exprWidth int) int {
	// First, try the preferred position (directly under the pipe)
	if f.isPositionAvailable(preferredPos, valueWidth, usedPositions) {
		return preferredPos
	}

	// If preferred position is not available, try positions to the right
	for offset := 1; offset <= 20; offset++ {
		candidatePos := preferredPos + offset
		if candidatePos+valueWidth < exprWidth+50 && f.isPositionAvailable(candidatePos, valueWidth, usedPositions) {
			return candidatePos
		}
	}

	// If no position to the right works, try positions to the left
	for offset := 1; offset <= 20; offset++ {
		candidatePos := preferredPos - offset
		if candidatePos >= 0 && f.isPositionAvailable(candidatePos, valueWidth, usedPositions) {
			return candidatePos
		}
	}

	// Fallback: use the preferred position anyway (should rarely happen)
	return preferredPos
}

// isPositionAvailable checks if a position range is available for placing a value
func (f *VisualFormatter) isPositionAvailable(pos, width int, usedPositions map[int]bool) bool {
	for i := 0; i < width; i++ {
		if usedPositions[pos+i] {
			return false
		}
	}
	return true
}

// hasSignificantOverlaps checks if positions would significantly overlap when placed directly under pipes
func (f *VisualFormatter) hasSignificantOverlaps(positions []ValuePosition) bool {
	for i, pos1 := range positions {
		for j, pos2 := range positions {
			if i >= j {
				continue
			}

			// Check if values would overlap when placed directly under their pipes
			pos1End := pos1.VisualPos + visualWidth(pos1.Value)
			pos2Start := pos2.VisualPos

			if pos1End > pos2Start {
				return true
			}
		}
	}
	return false
}

// calculateOptimalValuePosition calculates the best position for a value to minimize overlaps
func (f *VisualFormatter) calculateOptimalValuePosition(pos ValuePosition, linePositions []ValuePosition, exprWidth int) int {
	// Start with the position directly under the pipe
	basePos := pos.VisualPos
	valueWidth := visualWidth(pos.Value)

	// Check for conflicts with other values in the same line
	for _, other := range linePositions {
		if other.VisualPos == pos.VisualPos {
			continue // Skip self
		}

		// Calculate where this other value will be placed
		otherPos := other.VisualPos
		otherWidth := visualWidth(other.Value)

		// If there's an overlap, try to find a better position
		if f.rangesOverlap(basePos, basePos+valueWidth, otherPos, otherPos+otherWidth) {
			// Try moving to the right of the conflicting value
			candidatePos := otherPos + otherWidth + 1
			if candidatePos < exprWidth+50 { // Don't go too far right
				conflict := false
				// Check if this new position conflicts with other values
				for _, check := range linePositions {
					if check.VisualPos == pos.VisualPos || check.VisualPos == other.VisualPos {
						continue
					}
					checkPos := check.VisualPos
					checkWidth := visualWidth(check.Value)
					if f.rangesOverlap(candidatePos, candidatePos+valueWidth, checkPos, checkPos+checkWidth) {
						conflict = true
						break
					}
				}
				if !conflict {
					basePos = candidatePos
				}
			}
		}
	}

	return basePos
}

// rangesOverlap checks if two ranges overlap
func (f *VisualFormatter) rangesOverlap(start1, end1, start2, end2 int) bool {
	return !(end1 <= start2 || end2 <= start1)
}

// groupPositionsByDepth groups positions by their depth level.
func (f *VisualFormatter) groupPositionsByDepth(positions []ValuePosition) [][]ValuePosition {
	if len(positions) == 0 {
		return nil
	}

	var groups [][]ValuePosition
	currentDepth := positions[0].Depth
	currentGroup := []ValuePosition{}

	for _, pos := range positions {
		if pos.Depth != currentDepth {
			if len(currentGroup) > 0 {
				groups = append(groups, currentGroup)
			}
			currentGroup = []ValuePosition{pos}
			currentDepth = pos.Depth
		} else {
			currentGroup = append(currentGroup, pos)
		}
	}

	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}

// groupPositionsToAvoidOverlap groups positions within a depth level to avoid overlap.
func (f *VisualFormatter) groupPositionsToAvoidOverlap(positions []ValuePosition) [][]ValuePosition {
	if len(positions) == 0 {
		return nil
	}

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

	return lines
}

// formatValueCompact formats a value in a compact way.
func formatValueCompact(v interface{}) string {
	if v == nil {
		return "nil"
	}

	switch val := v.(type) {
	case string:
		// Improve string truncation with better length limits
		if len(val) > 15 {
			return fmt.Sprintf("%q...", val[:10])
		}
		return fmt.Sprintf("%q", val)
	case []int:
		return formatSliceCompact(val)
	case []string:
		return formatStringSliceCompact(val)
	case []interface{}:
		return formatInterfaceSliceCompact(val)
	case bool:
		return fmt.Sprintf("%v", val)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%v", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	default:
		// For structs and other complex types, try to format them nicely
		s := formatStructCompact(val)
		if len(s) > 20 {
			return s[:17] + "..."
		}
		return s
	}
}

// formatSliceCompact formats an int slice in a compact way.
func formatSliceCompact(slice []int) string {
	if len(slice) == 0 {
		return "[]"
	}
	if len(slice) <= 3 {
		return fmt.Sprintf("%v", slice)
	}
	return fmt.Sprintf("[%d,%d,...]", slice[0], slice[1])
}

// formatStringSliceCompact formats a string slice in a compact way.
func formatStringSliceCompact(slice []string) string {
	if len(slice) == 0 {
		return "[]"
	}
	if len(slice) <= 2 {
		// For short slices, use Go's default representation to match test expectations
		return fmt.Sprintf("%v", slice)
	}
	first := slice[0]
	if len(first) > 5 {
		first = first[:5] + "..."
	}
	return fmt.Sprintf("[%q,...]", first)
}

// formatInterfaceSliceCompact formats an interface slice in a compact way.
func formatInterfaceSliceCompact(slice []interface{}) string {
	if len(slice) == 0 {
		return "[]"
	}
	if len(slice) <= 2 {
		return fmt.Sprintf("%v", slice)
	}
	return fmt.Sprintf("[%v,...]", slice[0])
}

// formatStructCompact formats a struct in a compact way.
func formatStructCompact(v interface{}) string {
	val := reflect.ValueOf(v)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return "nil"
		}
		val = val.Elem()
	}

	// Handle structs
	if val.Kind() == reflect.Struct {
		typ := val.Type()
		var fields []string

		// Show first 2 fields
		for i := 0; i < val.NumField() && i < 2; i++ {
			field := val.Field(i)
			if field.CanInterface() {
				fieldName := typ.Field(i).Name
				fieldValue := formatValueCompact(field.Interface())
				fields = append(fields, fmt.Sprintf("%s:%s", fieldName, fieldValue))
			}
		}

		if val.NumField() > 2 {
			fields = append(fields, "...")
		}

		return fmt.Sprintf("{%s}", strings.Join(fields, ","))
	}

	// Fallback to regular formatting
	s := fmt.Sprintf("%v", v)
	if len(s) > 15 {
		return s[:15] + "..."
	}
	return s
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

	// Add step-by-step evaluation if tree is available
	if result.Tree != nil {
		parts = append(parts, "EVALUATION_STEPS:")
		steps := extractEvaluationSteps(result.Tree)
		for i, step := range steps {
			parts = append(parts, fmt.Sprintf("  Step %d: %s", i+1, step))
		}
	}

	return strings.Join(parts, "\n") + "\n"
}

// extractEvaluationSteps traverses the evaluation tree and returns step-by-step evaluation
func extractEvaluationSteps(tree *evaluator.EvaluationTree) []string {
	var steps []string
	var nodeCounter int

	// Helper function to traverse the tree in evaluation order
	var traverse func(node *evaluator.EvaluationTree)
	traverse = func(node *evaluator.EvaluationTree) {
		if node == nil {
			return
		}

		// First traverse children (post-order traversal for evaluation order)
		if node.Left != nil {
			traverse(node.Left)
		}
		if node.Right != nil {
			traverse(node.Right)
		}

		// Then process this node
		nodeCounter++
		step := formatEvaluationStep(node)
		if step != "" {
			steps = append(steps, step)
		}
	}

	traverse(tree)
	return steps
}

// formatEvaluationStep formats a single evaluation step
func formatEvaluationStep(node *evaluator.EvaluationTree) string {
	switch node.Type {
	case "identifier":
		if node.Value != nil {
			return fmt.Sprintf("`%s` => %v", node.Text, node.Value)
		}
		return fmt.Sprintf("`%s` => <%s>", node.Text, node.Text)
	
	case "literal":
		return fmt.Sprintf("`%s` => %v", node.Text, node.Value)
	
	case "comparison":
		if node.Left != nil && node.Right != nil {
			leftVal := formatNodeValue(node.Left)
			rightVal := formatNodeValue(node.Right)
			return fmt.Sprintf("`%s` with %s %s %s => %v", 
				node.Text, leftVal, node.Operator, rightVal, node.Result)
		}
		return fmt.Sprintf("`%s` => %v", node.Text, node.Result)
	
	case "logical":
		if node.Left != nil && node.Right != nil {
			leftVal := formatNodeResult(node.Left)
			rightVal := formatNodeResult(node.Right)
			return fmt.Sprintf("`%s` with %s %s %s => %v",
				node.Text, leftVal, node.Operator, rightVal, node.Result)
		}
		return fmt.Sprintf("`%s` => %v", node.Text, node.Result)
	
	case "unary":
		if node.Right != nil {
			rightVal := formatNodeValue(node.Right)
			return fmt.Sprintf("`%s` with %s%s => %v",
				node.Text, node.Operator, rightVal, node.Result)
		}
		return fmt.Sprintf("`%s` => %v", node.Text, node.Result)
	
	case "call":
		return fmt.Sprintf("`%s` => %v", node.Text, node.Value)
	
	case "index":
		return fmt.Sprintf("`%s` => %v", node.Text, node.Value)
	
	case "selector":
		return fmt.Sprintf("`%s` => %v", node.Text, node.Value)
	
	default:
		// For any other types, show the expression and its result if available
		if node.Value != nil {
			return fmt.Sprintf("`%s` => %v", node.Text, node.Value)
		}
		// Result is always available (bool type)
		return fmt.Sprintf("`%s` => %v", node.Text, node.Result)
	}
}

// formatNodeValue returns a string representation of a node's value
func formatNodeValue(node *evaluator.EvaluationTree) string {
	if node.Value != nil {
		return fmt.Sprintf("%v", node.Value)
	}
	return fmt.Sprintf("<%s>", node.Text)
}

// formatNodeResult returns a string representation of a node's result
func formatNodeResult(node *evaluator.EvaluationTree) string {
	// Result is always available (bool type)
	return fmt.Sprintf("%v", node.Result)
}

// === NEW VISUAL LAYER ARCHITECTURE ===

// createVisualNodes converts ValuePositions to VisualNodes with pipe positions
func (f *VisualFormatter) createVisualNodes(positions []ValuePosition) []VisualNode {
	nodes := make([]VisualNode, len(positions))
	
	for i, pos := range positions {
		nodes[i] = VisualNode{
			Position:        pos,
			EvaluationDepth: pos.Depth,
			VisualLayer:     -1, // Not yet assigned
			PipePosition:    pos.VisualPos, // Fixed pipe position
			ConflictsWith:   []int{},
		}
	}
	
	return nodes
}

// detectConflicts builds a conflict map between visual nodes
func (f *VisualFormatter) detectConflicts(nodes []VisualNode) ConflictMap {
	conflictMap := ConflictMap{
		Conflicts: make(map[int][]int),
	}
	
	for i, nodeA := range nodes {
		for j, nodeB := range nodes {
			if i >= j {
				continue // Avoid duplicates and self-comparison
			}
			
			// Check if the values would overlap if placed at the same layer
			if f.nodesWouldOverlap(nodeA, nodeB) {
				conflictMap.Conflicts[i] = append(conflictMap.Conflicts[i], j)
				conflictMap.Conflicts[j] = append(conflictMap.Conflicts[j], i)
			}
		}
	}
	
	return conflictMap
}

// nodesWouldOverlap checks if two nodes would overlap if placed on the same layer
func (f *VisualFormatter) nodesWouldOverlap(nodeA, nodeB VisualNode) bool {
	// Calculate the space needed for each value
	valueAStart := nodeA.PipePosition
	valueAEnd := valueAStart + visualWidth(nodeA.Position.Value)
	
	valueBStart := nodeB.PipePosition
	valueBEnd := valueBStart + visualWidth(nodeB.Position.Value)
	
	// Check for overlap with minimal spacing (1 character)
	return !(valueAEnd+1 <= valueBStart || valueBEnd+1 <= valueAStart)
}


