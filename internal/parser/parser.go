// Package parser provides functionality for parsing Go source code to extract assertion expressions.
package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

// ExtractExpression extracts the expression from source code at the specified line.
// It looks for Assert or Require function calls and returns the expression argument.
func ExtractExpression(filename string, line int) (string, error) {
	// Read the source file
	src, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// Parse the AST
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return "", err
	}

	// Find the expression at the specified line
	var targetExpr string
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		pos := fset.Position(n.Pos())
		if pos.Line != line {
			return true
		}

		// Look for Assert/Require function calls
		if call, ok := n.(*ast.CallExpr); ok {
			if isAssertCall(call) && len(call.Args) >= 2 {
				// Extract the second argument (expression) as string (0=t, 1=expr)
				exprArg := call.Args[1]
				start := fset.Position(exprArg.Pos()).Offset
				end := fset.Position(exprArg.End()).Offset
				if start >= 0 && end <= len(src) && start < end {
					targetExpr = string(src[start:end])
					return false
				}
			}
		}

		return true
	})

	if targetExpr == "" {
		return "", fmt.Errorf("expression not found")
	}

	return targetExpr, nil
}

// isAssertCall determines if a function call is an Assert or Require call.
func isAssertCall(call *ast.CallExpr) bool {
	// Package selector: diagassert.Assert
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		name := sel.Sel.Name
		return name == "Assert" || name == "Require"
	}
	// Direct function call: Assert (within same package)
	if ident, ok := call.Fun.(*ast.Ident); ok {
		name := ident.Name
		return name == "Assert" || name == "Require"
	}
	return false
}
