// Package formatter provides visual formatting for power-assert style output.
//
// Color Support:
// The visual formatter now includes comprehensive color support using github.com/fatih/color.
// Colors are automatically detected based on terminal capabilities and environment variables:
//
// Environment Variables:
//   - NO_COLOR: Set to any value to disable colors (respects https://no-color.org/)
//   - FORCE_COLOR: Set to any value to force enable colors
//   - DIAGASSERT_PIPE_COLORS: Set to "false" to disable per-value pipe colors (default: enabled)
//
// Color Scheme:
//   - Header ("ASSERTION FAILED"): Bold Red
//   - Pipes (|): Gray/Dim (default) or per-value colors when enabled
//   - Variable values: Blue
//   - Boolean true: Green
//   - Boolean false: Red
//   - Operators (>, ==, &&, etc.): Yellow
//
// Per-Value Pipe Colors:
// When DIAGASSERT_PIPE_COLORS is enabled (default), each value in deep expression hierarchies
// gets its own unique pipe color to improve readability. Colors are assigned deterministically
// based on the expression text, ensuring consistent coloring across runs.
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

	"github.com/fatih/color"
	"github.com/paveg/diagassert/internal/evaluator"
)

// ColorConfig holds color configuration for different output elements
type ColorConfig struct {
	// Element colors
	HeaderColor   *color.Color // "ASSERTION FAILED" header
	PipeColor     *color.Color // Visual pipes (|) - default color
	VariableColor *color.Color // Variable values (blue)
	TrueColor     *color.Color // Boolean true values (green)
	FalseColor    *color.Color // Boolean false values (red)
	OperatorColor *color.Color // Operators like >, <, == (yellow)

	// Per-value pipe colors
	PipeColorPalette  []*color.Color // Color palette for per-value pipes
	PipeColorsEnabled bool           // Enable per-value pipe colors

	// Color detection
	ColorsEnabled bool
}

// VisualFormatter formats evaluation results in power-assert style.
type VisualFormatter struct {
	includeMachineReadable bool
	colorConfig            *ColorConfig
}

// NewVisualFormatter creates a new visual formatter.
func NewVisualFormatter() *VisualFormatter {
	// Respect environment variable for machine-readable output
	includeMachine := os.Getenv("DIAGASSERT_MACHINE_READABLE") != "false"

	return &VisualFormatter{
		includeMachineReadable: includeMachine,
		colorConfig:            setupColorConfig(),
	}
}

// setupColorConfig creates and configures the color system
func setupColorConfig() *ColorConfig {
	// Detect if colors should be enabled
	colorsEnabled := shouldEnableColors()

	// Handle FORCE_COLOR override by temporarily clearing NO_COLOR
	var originalNoColor string
	var hadNoColor bool
	if colorsEnabled && os.Getenv("FORCE_COLOR") != "" && os.Getenv("NO_COLOR") != "" {
		originalNoColor = os.Getenv("NO_COLOR")
		hadNoColor = true
		os.Unsetenv("NO_COLOR")
	}

	// Set the global color.NoColor flag based on our detection
	color.NoColor = !colorsEnabled

	// Check if per-value pipe colors should be enabled
	pipeColorsEnabled := os.Getenv("DIAGASSERT_PIPE_COLORS") != "false"

	config := &ColorConfig{
		ColorsEnabled: colorsEnabled,
		HeaderColor:   color.New(color.FgRed, color.Bold), // Bold red for "ASSERTION FAILED"
		PipeColor:     color.New(color.FgHiBlack),         // Gray/dim for pipes
		VariableColor: color.New(color.FgBlue),            // Blue for variables
		TrueColor:     color.New(color.FgGreen),           // Green for true
		FalseColor:    color.New(color.FgRed),             // Red for false
		OperatorColor: color.New(color.FgYellow),          // Yellow for operators

		// Per-value pipe colors
		PipeColorPalette:  createPipeColorPalette(),
		PipeColorsEnabled: pipeColorsEnabled,
	}

	// Restore NO_COLOR if it was set
	if hadNoColor {
		os.Setenv("NO_COLOR", originalNoColor)
	}

	return config
}

// createPipeColorPalette creates a palette of colors for per-value pipes
// Colors are chosen to be distinguishable, accessible, and different from existing colors
func createPipeColorPalette() []*color.Color {
	return []*color.Color{
		color.New(color.FgCyan),      // Cyan - distinguishable from blue
		color.New(color.FgMagenta),   // Magenta - distinct color
		color.New(color.FgHiGreen),   // Bright green - different from regular green
		color.New(color.FgHiYellow),  // Bright yellow - different from regular yellow
		color.New(color.FgHiBlue),    // Bright blue - different from regular blue
		color.New(color.FgHiMagenta), // Bright magenta - vibrant
		color.New(color.FgHiCyan),    // Bright cyan - vivid
		color.New(color.FgWhite),     // White - good contrast
	}
}

// shouldEnableColors detects if colors should be enabled based on environment and terminal capabilities
func shouldEnableColors() bool {
	// Check FORCE_COLOR environment variable first (it should override NO_COLOR)
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Respect NO_COLOR environment variable (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Default to enabling colors in most cases
	// The fatih/color package will handle terminal detection automatically
	// We enable colors by default and let the color package decide whether to apply them
	return true
}

// GetColorConfig returns the current color configuration (for testing purposes)
func (f *VisualFormatter) GetColorConfig() *ColorConfig {
	return f.colorConfig
}

// FormatVisual formats the evaluation result in power-assert style.
func (f *VisualFormatter) FormatVisual(result *evaluator.ExpressionResult, file string, line int, customMessage string) string {
	return f.FormatVisualWithContext(result, file, line, customMessage, nil)
}

// FormatVisualWithContext formats the evaluation result with context values.
func (f *VisualFormatter) FormatVisualWithContext(result *evaluator.ExpressionResult, file string, line int, customMessage string, ctx *AssertionContext) string {
	var b strings.Builder

	// Header with color
	header := fmt.Sprintf("ASSERTION FAILED at %s:%d", file, line)
	b.WriteString(f.colorizeHeader(header) + "\n\n")

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

// Color helper functions

// colorizeHeader applies color to the header text
func (f *VisualFormatter) colorizeHeader(text string) string {
	if !f.colorConfig.ColorsEnabled {
		return text
	}
	// If FORCE_COLOR is set, manually apply colors even if NO_COLOR is set
	if os.Getenv("FORCE_COLOR") != "" && os.Getenv("NO_COLOR") != "" {
		return "\033[31;1m" + text + "\033[0;22m"
	}
	return f.colorConfig.HeaderColor.Sprint(text)
}

// colorizePipe applies color to pipe characters
func (f *VisualFormatter) colorizePipe(text string) string {
	if !f.colorConfig.ColorsEnabled {
		return text
	}
	// If FORCE_COLOR is set, manually apply colors even if NO_COLOR is set
	if os.Getenv("FORCE_COLOR") != "" && os.Getenv("NO_COLOR") != "" {
		return "\033[90m" + text + "\033[0m"
	}
	return f.colorConfig.PipeColor.Sprint(text)
}

// colorizeValue applies appropriate color to a value based on its type
func (f *VisualFormatter) colorizeValue(value string, isOperator bool) string {
	if !f.colorConfig.ColorsEnabled {
		return value
	}

	// If FORCE_COLOR is set, manually apply colors even if NO_COLOR is set
	if os.Getenv("FORCE_COLOR") != "" && os.Getenv("NO_COLOR") != "" {
		if isOperator {
			return "\033[33m" + value + "\033[0m" // Yellow for operators
		}
		switch value {
		case "true":
			return "\033[32m" + value + "\033[0m" // Green for true
		case "false":
			return "\033[31m" + value + "\033[0m" // Red for false
		default:
			return "\033[34m" + value + "\033[0m" // Blue for variables
		}
	}

	// Special handling for operators
	if isOperator {
		return f.colorConfig.OperatorColor.Sprint(value)
	}

	// Color based on value content
	switch value {
	case "true":
		return f.colorConfig.TrueColor.Sprint(value)
	case "false":
		return f.colorConfig.FalseColor.Sprint(value)
	default:
		return f.colorConfig.VariableColor.Sprint(value)
	}
}

// colorizePipeLine applies color to pipe characters in a line
func (f *VisualFormatter) colorizePipeLine(line string) string {
	if !f.colorConfig.ColorsEnabled {
		return line
	}

	// If FORCE_COLOR is set, manually apply colors even if NO_COLOR is set
	if os.Getenv("FORCE_COLOR") != "" && os.Getenv("NO_COLOR") != "" {
		return strings.ReplaceAll(line, "|", "\033[90m|\033[0m")
	}

	// Replace pipe characters with colored ones
	return strings.ReplaceAll(line, "|", f.colorConfig.PipeColor.Sprint("|"))
}

// colorizePerValuePipeLine applies per-value colors to pipe characters in a line
// layerAssignment contains the layer information needed to determine which values correspond to which pipes
// layerIdx is the current layer being processed
func (f *VisualFormatter) colorizePerValuePipeLine(line string, layerAssignment LayerAssignment, layerIdx int) string {
	if !f.colorConfig.ColorsEnabled || !f.colorConfig.PipeColorsEnabled {
		return f.colorizePipeLine(line) // Fall back to standard pipe coloring
	}

	// Build a map of pipe positions to their corresponding ValuePosition
	pipeToValue := make(map[int]ValuePosition)

	// Collect all value positions that have pipes in this layer or deeper layers
	for currentLayerIdx := layerIdx; currentLayerIdx < len(layerAssignment.Layers); currentLayerIdx++ {
		for _, node := range layerAssignment.Layers[currentLayerIdx] {
			pipeToValue[node.PipePosition] = node.Position
		}
	}

	// If no per-value mapping is available, use default coloring
	if len(pipeToValue) == 0 {
		return f.colorizePipeLine(line)
	}

	// Build the result string with per-value pipe colors
	var result strings.Builder
	lineRunes := []rune(line)

	for pos, char := range lineRunes {
		if char == '|' {
			if valuePos, exists := pipeToValue[pos]; exists {
				// Get the color for this specific value
				pipeColor := f.getPipeColorForValue(valuePos)

				// Handle FORCE_COLOR case
				if os.Getenv("FORCE_COLOR") != "" && os.Getenv("NO_COLOR") != "" {
					result.WriteString(f.forceColorPipe("|", pipeColor))
				} else {
					result.WriteString(pipeColor.Sprint("|"))
				}
			} else {
				// Use default pipe color for pipes without specific value mapping
				result.WriteString(f.colorConfig.PipeColor.Sprint("|"))
			}
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// buildColoredValueLine builds a value line with appropriate colors for each value
func (f *VisualFormatter) buildColoredValueLine(layer []VisualNode, maxWidth int) string {
	// Build a line of spaces
	line := strings.Repeat(" ", maxWidth)
	lineRunes := []rune(line)

	// Track which positions have been colored to avoid overlaps
	coloredPositions := make(map[int]bool)

	// Place each value with its color
	for _, node := range layer {
		value := node.Position.Value
		startPos := node.PipePosition

		// Place the raw value in the line first
		if startPos < len(lineRunes) {
			// Clear the space for this value
			valueRunes := []rune(value)
			for i, r := range valueRunes {
				if startPos+i < len(lineRunes) && !coloredPositions[startPos+i] {
					lineRunes[startPos+i] = r
					coloredPositions[startPos+i] = true
				}
			}
		}
	}

	// Convert back to string and apply colors if enabled
	plainLine := string(lineRunes)

	if !f.colorConfig.ColorsEnabled {
		return plainLine
	}

	// Apply colors to the line by replacing values with colored versions
	result := plainLine
	for _, node := range layer {
		value := node.Position.Value
		isOperator := f.isOperatorValue(node.Position.Expression, value)
		coloredValue := f.colorizeValue(value, isOperator)

		// Replace the first occurrence of the value with its colored version
		result = strings.Replace(result, value, coloredValue, 1)
	}

	return result
}

// isOperatorValue determines if a value represents an operator result
func (f *VisualFormatter) isOperatorValue(expression, value string) bool {
	// Check if the expression is an operator
	operators := []string{">", "<", ">=", "<=", "==", "!=", "&&", "||", "!", "+", "-", "*", "/", "%"}

	for _, op := range operators {
		if expression == op {
			return true
		}
	}

	return false
}

// Per-value pipe color functions

// assignPipeColor assigns a color to a pipe based on the expression text
// Uses deterministic hashing to ensure consistent color assignment
func (f *VisualFormatter) assignPipeColor(expression string) *color.Color {
	if !f.colorConfig.ColorsEnabled || !f.colorConfig.PipeColorsEnabled {
		return f.colorConfig.PipeColor // Fall back to default pipe color
	}

	if len(f.colorConfig.PipeColorPalette) == 0 {
		return f.colorConfig.PipeColor // Fall back to default pipe color
	}

	// Use a simple hash to deterministically assign colors
	hash := f.simpleHash(expression)
	colorIndex := hash % len(f.colorConfig.PipeColorPalette)
	return f.colorConfig.PipeColorPalette[colorIndex]
}

// getPipeColorForValue gets the appropriate pipe color for a specific value position
func (f *VisualFormatter) getPipeColorForValue(position ValuePosition) *color.Color {
	return f.assignPipeColor(position.Expression)
}

// simpleHash provides a simple hash function for consistent color assignment
func (f *VisualFormatter) simpleHash(s string) int {
	hash := 0
	for i, char := range s {
		hash = hash*31 + int(char) + i
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// colorizePerValuePipe applies per-value pipe colors to a pipe character
func (f *VisualFormatter) colorizePerValuePipe(text string, position ValuePosition) string {
	if !f.colorConfig.ColorsEnabled {
		return text
	}

	pipeColor := f.getPipeColorForValue(position)

	// If FORCE_COLOR is set, manually apply colors even if NO_COLOR is set
	if os.Getenv("FORCE_COLOR") != "" && os.Getenv("NO_COLOR") != "" {
		// Map color.Color to ANSI codes for force color mode
		return f.forceColorPipe(text, pipeColor)
	}

	return pipeColor.Sprint(text)
}

// forceColorPipe applies pipe colors manually when FORCE_COLOR is set
func (f *VisualFormatter) forceColorPipe(text string, pipeColor *color.Color) string {
	// Map the fatih/color.Color to ANSI codes for force color mode
	// This is a simple mapping for the colors we use in our palette

	// Get the color by comparing with known colors from our palette
	for i, paletteColor := range f.colorConfig.PipeColorPalette {
		if pipeColor == paletteColor {
			// Map each palette color to its ANSI code
			switch i {
			case 0: // Cyan
				return "\033[36m" + text + "\033[0m"
			case 1: // Magenta
				return "\033[35m" + text + "\033[0m"
			case 2: // Bright green
				return "\033[92m" + text + "\033[0m"
			case 3: // Bright yellow
				return "\033[93m" + text + "\033[0m"
			case 4: // Bright blue
				return "\033[94m" + text + "\033[0m"
			case 5: // Bright magenta
				return "\033[95m" + text + "\033[0m"
			case 6: // Bright cyan
				return "\033[96m" + text + "\033[0m"
			case 7: // White
				return "\033[97m" + text + "\033[0m"
			}
		}
	}

	// Default to cyan if no mapping found
	return "\033[36m" + text + "\033[0m"
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
	Position        ValuePosition
	EvaluationDepth int   // Semantic depth in the evaluation tree
	VisualLayer     int   // Visual layer for rendering (may differ from evaluation depth)
	PipePosition    int   // Fixed pipe position in the expression
	ConflictsWith   []int // Indices of other nodes this conflicts with
}

// LayerAssignment manages the assignment of values to visual layers
type LayerAssignment struct {
	MaxLayer      int            // Maximum layer number assigned
	Layers        [][]VisualNode // Nodes grouped by visual layer
	PipePositions map[int]bool   // Track fixed pipe positions
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
	pipe := f.colorizePipe("|")
	falseValue := f.colorizeValue("false", false)
	b.WriteString(fmt.Sprintf("         %s%s\n", padding, pipe))
	b.WriteString(fmt.Sprintf("         %s%s\n", padding, falseValue))

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

		// Add pipe line with color (use per-value coloring if enabled)
		pipeStr := f.colorizePerValuePipeLine(string(pipeRunes), layerAssignment, layerIdx)
		pipeStr = strings.TrimRight(pipeStr, " ")
		if pipeStr != "" {
			result = append(result, pipeStr)
		}

		// Build value line with colors
		valueStr := f.buildColoredValueLine(layer, exprWidth+100)
		valueStr = strings.TrimRight(valueStr, " ")
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

// rangesOverlap checks if two ranges overlap
func (f *VisualFormatter) rangesOverlap(start1, end1, start2, end2 int) bool {
	return !(end1 <= start2 || end2 <= start1)
}

// formatValueCompact formats a value in a compact way.
func formatValueCompact(v interface{}) string {
	if v == nil {
		return "nil"
	}

	switch val := v.(type) {
	case string:
		// Improve string truncation with better length limits
		if len(val) > 10 {
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
		if len(s) > 15 {
			return s[:15] + "..."
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
	if len(s) > 10 {
		return s[:10] + "..."
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
