// Package evaluator provides functionality for evaluating assertion expressions and building evaluation trees.
package evaluator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"runtime"
	"strings"
)

// ExpressionResult represents the result of evaluating an expression.
type ExpressionResult struct {
	Expression string
	Result     bool
	Variables  map[string]interface{}
	Tree       *EvaluationTree
}

// EvaluationTree represents the tree structure of expression evaluation.
type EvaluationTree struct {
	ID       int
	Type     string // "comparison", "logical", "method_call", "identifier", "literal"
	Operator string // ">", "&&", "||", etc.
	Left     *EvaluationTree
	Right    *EvaluationTree
	Value    interface{}
	Result   bool
	Text     string // Original expression text
	Children []*EvaluationTree
}

var nodeCounter int

// Evaluate performs expression evaluation with variable value extraction and tree building.
func Evaluate(expr string, result bool, callerFrame uintptr) *ExpressionResult {
	variables := extractVariableValuesFromFrame(expr, callerFrame)
	tree := buildEvaluationTree(expr, variables)

	return &ExpressionResult{
		Expression: expr,
		Result:     result,
		Variables:  variables,
		Tree:       tree,
	}
}

// EvaluateWithValues performs expression evaluation with user-provided values merged with auto-extracted variables.
func EvaluateWithValues(expr string, result bool, callerFrame uintptr, userValues map[string]interface{}) *ExpressionResult {
	// Extract variables from frame (returns placeholders like "<x>")
	autoExtracted := extractVariableValuesFromFrame(expr, callerFrame)

	// Merge user-provided values with auto-extracted, giving precedence to user values
	variables := mergeVariables(autoExtracted, userValues)

	tree := buildEvaluationTree(expr, variables)

	return &ExpressionResult{
		Expression: expr,
		Result:     result,
		Variables:  variables,
		Tree:       tree,
	}
}

// buildEvaluationTree constructs a detailed evaluation tree for the expression.
func buildEvaluationTree(expr string, variables map[string]interface{}) *EvaluationTree {
	nodeCounter = 0

	fset := token.NewFileSet()
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return &EvaluationTree{
			ID:     getNextNodeID(),
			Type:   "error",
			Text:   expr,
			Result: false,
		}
	}

	return buildTreeFromAST(node, variables, fset)
}

// buildTreeFromAST recursively builds evaluation tree from AST node.
func buildTreeFromAST(node ast.Expr, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		return buildBinaryExprTree(n, variables, fset)
	case *ast.UnaryExpr:
		return buildUnaryExprTree(n, variables, fset)
	case *ast.Ident:
		return buildIdentTree(n, variables)
	case *ast.BasicLit:
		return buildLiteralTree(n)
	case *ast.SelectorExpr:
		return buildSelectorTree(n, variables, fset)
	case *ast.CallExpr:
		return buildCallTree(n, variables, fset)
	case *ast.IndexExpr:
		return buildIndexTree(n, variables, fset)
	case *ast.ParenExpr:
		return buildTreeFromAST(n.X, variables, fset)
	default:
		return &EvaluationTree{
			ID:   getNextNodeID(),
			Type: "unknown",
			Text: fmt.Sprintf("%T", node),
		}
	}
}

// buildBinaryExprTree builds tree for binary expressions like "x > y" or "a && b".
func buildBinaryExprTree(expr *ast.BinaryExpr, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	left := buildTreeFromAST(expr.X, variables, fset)
	right := buildTreeFromAST(expr.Y, variables, fset)

	operator := expr.Op.String()
	result := evaluateBinaryExpr(left, right, operator)

	return &EvaluationTree{
		ID:       getNextNodeID(),
		Type:     getBinaryExprType(operator),
		Operator: operator,
		Left:     left,
		Right:    right,
		Result:   result,
		Text:     fmt.Sprintf("%s %s %s", left.Text, operator, right.Text),
	}
}

// buildUnaryExprTree builds tree for unary expressions like "!condition".
func buildUnaryExprTree(expr *ast.UnaryExpr, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	operand := buildTreeFromAST(expr.X, variables, fset)
	operator := expr.Op.String()

	var result bool
	if operator == "!" {
		result = !operand.Result
	} else {
		result = operand.Result
	}

	return &EvaluationTree{
		ID:       getNextNodeID(),
		Type:     "unary",
		Operator: operator,
		Left:     operand,
		Result:   result,
		Text:     fmt.Sprintf("%s%s", operator, operand.Text),
	}
}

// buildIdentTree builds tree for identifiers like "x", "user".
func buildIdentTree(ident *ast.Ident, variables map[string]interface{}) *EvaluationTree {
	value, exists := variables[ident.Name]

	return &EvaluationTree{
		ID:     getNextNodeID(),
		Type:   "identifier",
		Value:  value,
		Result: exists && isTruthy(value),
		Text:   ident.Name,
	}
}

// buildLiteralTree builds tree for literals like "18", "true", "\"hello\"".
func buildLiteralTree(lit *ast.BasicLit) *EvaluationTree {
	value := parseLiteral(lit)

	return &EvaluationTree{
		ID:     getNextNodeID(),
		Type:   "literal",
		Value:  value,
		Result: isTruthy(value),
		Text:   lit.Value,
	}
}

// buildSelectorTree builds tree for selector expressions like "user.Age".
func buildSelectorTree(sel *ast.SelectorExpr, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	baseTree := buildTreeFromAST(sel.X, variables, fset)
	fieldName := sel.Sel.Name
	text := fmt.Sprintf("%s.%s", baseTree.Text, fieldName)

	var value interface{}
	var result bool

	if baseTree.Value != nil {
		if fieldValue := getFieldValue(baseTree.Value, fieldName); fieldValue != nil {
			value = fieldValue
			result = isTruthy(value)
		}
	}

	return &EvaluationTree{
		ID:     getNextNodeID(),
		Type:   "selector",
		Left:   baseTree,
		Value:  value,
		Result: result,
		Text:   text,
	}
}

// buildCallTree builds tree for method calls like "user.IsAdult()".
func buildCallTree(call *ast.CallExpr, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	// Simplified implementation - full method call evaluation would be more complex
	var text strings.Builder

	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		baseTree := buildTreeFromAST(fun.X, variables, fset)
		methodName := fun.Sel.Name
		text.WriteString(fmt.Sprintf("%s.%s()", baseTree.Text, methodName))

		// Try to call the method if possible
		var value interface{}
		var result bool
		if baseTree.Value != nil {
			if methodResult := callMethod(baseTree.Value, methodName); methodResult != nil {
				value = methodResult
				result = isTruthy(value)
			}
		}

		return &EvaluationTree{
			ID:     getNextNodeID(),
			Type:   "method_call",
			Left:   baseTree,
			Value:  value,
			Result: result,
			Text:   text.String(),
		}
	default:
		return &EvaluationTree{
			ID:   getNextNodeID(),
			Type: "call",
			Text: "function_call",
		}
	}
}

// buildIndexTree builds tree for index expressions like "arr[0]".
func buildIndexTree(index *ast.IndexExpr, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	baseTree := buildTreeFromAST(index.X, variables, fset)
	indexTree := buildTreeFromAST(index.Index, variables, fset)
	text := fmt.Sprintf("%s[%s]", baseTree.Text, indexTree.Text)

	var value interface{}
	var result bool

	if baseTree.Value != nil && indexTree.Value != nil {
		if indexValue := getIndexValue(baseTree.Value, indexTree.Value); indexValue != nil {
			value = indexValue
			result = isTruthy(value)
		}
	}

	return &EvaluationTree{
		ID:     getNextNodeID(),
		Type:   "index",
		Left:   baseTree,
		Right:  indexTree,
		Value:  value,
		Result: result,
		Text:   text,
	}
}

// Helper functions

func getNextNodeID() int {
	nodeCounter++
	return nodeCounter
}

func getBinaryExprType(operator string) string {
	switch operator {
	case "&&", "||":
		return "logical"
	case "==", "!=", "<", "<=", ">", ">=":
		return "comparison"
	default:
		return "binary"
	}
}

func evaluateBinaryExpr(left, right *EvaluationTree, operator string) bool {
	switch operator {
	case "&&":
		return left.Result && right.Result
	case "||":
		return left.Result || right.Result
	case "==":
		return compareValues(left.Value, right.Value, "==")
	case "!=":
		return compareValues(left.Value, right.Value, "!=")
	case "<":
		return compareValues(left.Value, right.Value, "<")
	case "<=":
		return compareValues(left.Value, right.Value, "<=")
	case ">":
		return compareValues(left.Value, right.Value, ">")
	case ">=":
		return compareValues(left.Value, right.Value, ">=")
	default:
		return false
	}
}

func compareValues(left, right interface{}, operator string) bool {
	if left == nil || right == nil {
		switch operator {
		case "==":
			return left == right
		case "!=":
			return left != right
		default:
			return false
		}
	}

	// Convert to comparable types and compare
	leftVal := reflect.ValueOf(left)
	rightVal := reflect.ValueOf(right)

	if !leftVal.Type().Comparable() || !rightVal.Type().Comparable() {
		return false
	}

	switch operator {
	case "==":
		return reflect.DeepEqual(left, right)
	case "!=":
		return !reflect.DeepEqual(left, right)
	case "<", "<=", ">", ">=":
		return compareNumeric(left, right, operator)
	default:
		return false
	}
}

func compareNumeric(left, right interface{}, operator string) bool {
	leftVal := getNumericValue(left)
	rightVal := getNumericValue(right)

	if leftVal == nil || rightVal == nil {
		return false
	}

	l := *leftVal
	r := *rightVal

	switch operator {
	case "<":
		return l < r
	case "<=":
		return l <= r
	case ">":
		return l > r
	case ">=":
		return l >= r
	default:
		return false
	}
}

func getNumericValue(v interface{}) *float64 {
	switch val := v.(type) {
	case int:
		f := float64(val)
		return &f
	case int8:
		f := float64(val)
		return &f
	case int16:
		f := float64(val)
		return &f
	case int32:
		f := float64(val)
		return &f
	case int64:
		f := float64(val)
		return &f
	case uint:
		f := float64(val)
		return &f
	case uint8:
		f := float64(val)
		return &f
	case uint16:
		f := float64(val)
		return &f
	case uint32:
		f := float64(val)
		return &f
	case uint64:
		f := float64(val)
		return &f
	case float32:
		f := float64(val)
		return &f
	case float64:
		return &val
	default:
		return nil
	}
}

func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0.0
	case string:
		return v != ""
	default:
		return !reflect.ValueOf(v).IsZero()
	}
}

func parseLiteral(lit *ast.BasicLit) interface{} {
	switch lit.Kind {
	case token.INT:
		// Simple int parsing - more sophisticated parsing could be added
		var val int
		if _, err := fmt.Sscanf(lit.Value, "%d", &val); err != nil {
			return lit.Value // Return original string if parsing fails
		}
		return val
	case token.FLOAT:
		var val float64
		if _, err := fmt.Sscanf(lit.Value, "%f", &val); err != nil {
			return lit.Value // Return original string if parsing fails
		}
		return val
	case token.STRING:
		// Remove quotes
		if len(lit.Value) >= 2 {
			return lit.Value[1 : len(lit.Value)-1]
		}
		return lit.Value
	case token.CHAR:
		if len(lit.Value) >= 3 {
			return rune(lit.Value[1])
		}
		return rune(0)
	default:
		return lit.Value
	}
}

func getFieldValue(obj interface{}, fieldName string) interface{} {
	if obj == nil {
		return nil
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	field := val.FieldByName(fieldName)
	if !field.IsValid() || !field.CanInterface() {
		return nil
	}

	return field.Interface()
}

func callMethod(obj interface{}, methodName string) interface{} {
	if obj == nil {
		return nil
	}

	val := reflect.ValueOf(obj)
	method := val.MethodByName(methodName)
	if !method.IsValid() {
		return nil
	}

	// Call method with no arguments (simplified)
	results := method.Call(nil)
	if len(results) > 0 {
		return results[0].Interface()
	}

	return nil
}

func getIndexValue(obj, index interface{}) interface{} {
	if obj == nil || index == nil {
		return nil
	}

	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil
	}

	var idx int
	switch i := index.(type) {
	case int:
		idx = i
	default:
		return nil
	}

	if idx < 0 || idx >= val.Len() {
		return nil
	}

	return val.Index(idx).Interface()
}

// mergeVariables merges user-provided values with auto-extracted variables.
// User-provided values take precedence over auto-extracted placeholder values.
func mergeVariables(autoExtracted, userValues map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Start with auto-extracted values (placeholders)
	for name, value := range autoExtracted {
		merged[name] = value
	}

	// Override with user-provided values
	for name, value := range userValues {
		merged[name] = value
	}

	return merged
}

// extractVariableValuesFromFrame extracts variable values from the caller's stack frame.
func extractVariableValuesFromFrame(expr string, callerFrame uintptr) map[string]interface{} {
	variables := make(map[string]interface{})

	// Parse expression to find variable names
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return variables
	}

	// Extract variable names from AST
	varNames := extractVariableNames(node)

	// Get function info
	fn := runtime.FuncForPC(callerFrame)
	if fn == nil {
		return variables
	}

	// Try to extract variable values using runtime introspection
	// This is a simplified implementation - real stack inspection is very complex
	for _, name := range varNames {
		// For demonstration, we'll use a placeholder approach
		// In a real implementation, this would require deep runtime introspection
		variables[name] = fmt.Sprintf("<%s>", name)
	}

	return variables
}

// extractVariableNames recursively extracts variable names from AST.
func extractVariableNames(node ast.Expr) []string {
	var names []string

	ast.Inspect(node, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			// Skip built-in identifiers
			if ident.Name != "true" && ident.Name != "false" && ident.Name != "nil" {
				names = append(names, ident.Name)
			}
		}
		return true
	})

	return names
}
