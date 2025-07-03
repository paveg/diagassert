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
	case *ast.SliceExpr:
		return buildSliceTree(n, variables, fset)
	case *ast.ParenExpr:
		return buildTreeFromAST(n.X, variables, fset)
	case *ast.ArrayType:
		return buildArrayTypeTree(n, variables, fset)
	case *ast.CompositeLit:
		return buildCompositeLitTree(n, variables, fset)
	case *ast.FuncLit:
		return buildFuncLitTree(n, variables, fset)
	case *ast.TypeAssertExpr:
		return buildTypeAssertTree(n, variables, fset)
	case *ast.StarExpr:
		return buildStarExprTree(n, variables, fset)
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
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array && val.Kind() != reflect.Map {
		return nil
	}

	// Handle map indexing
	if val.Kind() == reflect.Map {
		indexVal := reflect.ValueOf(index)
		if !indexVal.Type().AssignableTo(val.Type().Key()) {
			return nil
		}
		result := val.MapIndex(indexVal)
		if !result.IsValid() {
			return nil
		}
		return result.Interface()
	}

	// Handle slice/array indexing
	var idx int
	switch i := index.(type) {
	case int:
		idx = i
	case int64:
		idx = int(i)
	case int32:
		idx = int(i)
	default:
		return nil
	}

	if idx < 0 || idx >= val.Len() {
		return nil
	}

	return val.Index(idx).Interface()
}

// buildSliceTree builds tree for slice expressions like "arr[1:3]".
func buildSliceTree(slice *ast.SliceExpr, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	baseTree := buildTreeFromAST(slice.X, variables, fset)

	var lowTree, highTree, maxTree *EvaluationTree
	var text strings.Builder

	text.WriteString(baseTree.Text)
	text.WriteString("[")

	if slice.Low != nil {
		lowTree = buildTreeFromAST(slice.Low, variables, fset)
		text.WriteString(lowTree.Text)
	}

	text.WriteString(":")

	if slice.High != nil {
		highTree = buildTreeFromAST(slice.High, variables, fset)
		text.WriteString(highTree.Text)
	}

	if slice.Max != nil {
		text.WriteString(":")
		maxTree = buildTreeFromAST(slice.Max, variables, fset)
		text.WriteString(maxTree.Text)
	}

	text.WriteString("]")

	var value interface{}
	var result bool

	if baseTree.Value != nil {
		var low, high, max interface{}
		if lowTree != nil {
			low = lowTree.Value
		}
		if highTree != nil {
			high = highTree.Value
		}
		if maxTree != nil {
			max = maxTree.Value
		}

		if sliceValue := getSliceValue(baseTree.Value, low, high, max); sliceValue != nil {
			value = sliceValue
			result = true
		}
	}

	children := []*EvaluationTree{baseTree}
	if lowTree != nil {
		children = append(children, lowTree)
	}
	if highTree != nil {
		children = append(children, highTree)
	}
	if maxTree != nil {
		children = append(children, maxTree)
	}

	return &EvaluationTree{
		ID:       getNextNodeID(),
		Type:     "slice",
		Children: children,
		Value:    value,
		Result:   result,
		Text:     text.String(),
	}
}

// getSliceValue extracts slice value from an object.
func getSliceValue(obj, low, high, max interface{}) interface{} {
	if obj == nil {
		return nil
	}

	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array && val.Kind() != reflect.String {
		return nil
	}

	length := val.Len()

	// Default slice bounds
	lowIdx := 0
	highIdx := length
	maxIdx := length

	// Parse low bound
	if low != nil {
		if idx, ok := convertToInt(low); ok {
			lowIdx = idx
		} else {
			return nil
		}
	}

	// Parse high bound
	if high != nil {
		if idx, ok := convertToInt(high); ok {
			highIdx = idx
		} else {
			return nil
		}
	}

	// Parse max bound (for 3-index slices)
	if max != nil {
		if idx, ok := convertToInt(max); ok {
			maxIdx = idx
		} else {
			return nil
		}
	}

	// Validate bounds
	if lowIdx < 0 || lowIdx > length || highIdx < lowIdx || highIdx > length {
		return nil
	}
	if max != nil && (maxIdx < highIdx || maxIdx > length) {
		return nil
	}

	// Perform the slice operation
	if val.Kind() == reflect.String {
		str := val.String()
		return str[lowIdx:highIdx]
	}

	if max != nil {
		// 3-index slice
		return val.Slice3(lowIdx, highIdx, maxIdx).Interface()
	}

	// 2-index slice
	return val.Slice(lowIdx, highIdx).Interface()
}

// convertToInt converts various numeric types to int.
func convertToInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	default:
		return 0, false
	}
}

// buildArrayTypeTree builds tree for array type expressions.
func buildArrayTypeTree(arrayType *ast.ArrayType, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	var text strings.Builder
	text.WriteString("[")

	if arrayType.Len != nil {
		lengthTree := buildTreeFromAST(arrayType.Len, variables, fset)
		text.WriteString(lengthTree.Text)
	}

	text.WriteString("]")

	if arrayType.Elt != nil {
		eltTree := buildTreeFromAST(arrayType.Elt, variables, fset)
		text.WriteString(eltTree.Text)
	}

	return &EvaluationTree{
		ID:   getNextNodeID(),
		Type: "array_type",
		Text: text.String(),
	}
}

// buildCompositeLitTree builds tree for composite literals like "[]int{1, 2, 3}".
func buildCompositeLitTree(comp *ast.CompositeLit, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	var children []*EvaluationTree
	var text strings.Builder

	if comp.Type != nil {
		typeTree := buildTreeFromAST(comp.Type, variables, fset)
		children = append(children, typeTree)
		text.WriteString(typeTree.Text)
	}

	text.WriteString("{")

	for i, elt := range comp.Elts {
		if i > 0 {
			text.WriteString(", ")
		}
		eltTree := buildTreeFromAST(elt, variables, fset)
		children = append(children, eltTree)
		text.WriteString(eltTree.Text)
	}

	text.WriteString("}")

	return &EvaluationTree{
		ID:       getNextNodeID(),
		Type:     "composite_lit",
		Children: children,
		Text:     text.String(),
	}
}

// buildFuncLitTree builds tree for function literals.
func buildFuncLitTree(funcLit *ast.FuncLit, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	return &EvaluationTree{
		ID:   getNextNodeID(),
		Type: "func_lit",
		Text: "func(...) {...}",
	}
}

// buildTypeAssertTree builds tree for type assertions like "x.(int)".
func buildTypeAssertTree(typeAssert *ast.TypeAssertExpr, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	baseTree := buildTreeFromAST(typeAssert.X, variables, fset)

	var typeText string
	if typeAssert.Type != nil {
		typeTree := buildTreeFromAST(typeAssert.Type, variables, fset)
		typeText = typeTree.Text
	} else {
		typeText = "type" // for x.(type) in type switches
	}

	text := fmt.Sprintf("%s.(%s)", baseTree.Text, typeText)

	return &EvaluationTree{
		ID:   getNextNodeID(),
		Type: "type_assert",
		Left: baseTree,
		Text: text,
	}
}

// buildStarExprTree builds tree for pointer dereference expressions like "*ptr".
func buildStarExprTree(star *ast.StarExpr, variables map[string]interface{}, fset *token.FileSet) *EvaluationTree {
	baseTree := buildTreeFromAST(star.X, variables, fset)

	var value interface{}
	var result bool

	if baseTree.Value != nil {
		val := reflect.ValueOf(baseTree.Value)
		if val.Kind() == reflect.Ptr && !val.IsNil() {
			value = val.Elem().Interface()
			result = isTruthy(value)
		}
	}

	return &EvaluationTree{
		ID:     getNextNodeID(),
		Type:   "dereference",
		Left:   baseTree,
		Value:  value,
		Result: result,
		Text:   fmt.Sprintf("*%s", baseTree.Text),
	}
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

// VariableContext holds information about a variable's usage context.
type VariableContext struct {
	Name       string
	Type       string      // "identifier", "field", "index", "slice", "method"
	Parent     string      // For nested access like "obj.field"
	Index      interface{} // For array/slice access
	SliceStart interface{} // For slice expressions
	SliceEnd   interface{} // For slice expressions
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
