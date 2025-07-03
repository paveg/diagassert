package diagassert_test

import (
	"testing"

	"github.com/paveg/diagassert"
)

// This file demonstrates the power-assert style visual output

func TestVisualDemo_SimpleComparison(t *testing.T) {
	x := 10

	// This will show:
	// ASSERTION FAILED at visual_demo_test.go:XX
	//
	//   assert(x > 20)
	//          |   |
	//          10  false
	diagassert.Assert(t, x > 20)
}

func TestVisualDemo_LogicalAND(t *testing.T) {
	age := 16
	hasLicense := false

	// This will show:
	// ASSERTION FAILED at visual_demo_test.go:XX
	//
	//   assert(age >= 18 && hasLicense)
	//          |   |  |  |  |
	//          16  |  18 |  false
	//              false false
	diagassert.Assert(t, age >= 18 && hasLicense)
}

func TestVisualDemo_ArrayAccess(t *testing.T) {
	scores := []int{65, 70, 75}

	// This will show:
	// ASSERTION FAILED at visual_demo_test.go:XX
	//
	//   assert(scores[0] >= 80)
	//          |       |  |  |
	//          |       65 |  80
	//          |          false
	//          [65 70 75]
	diagassert.Assert(t, scores[0] >= 80,
		diagassert.V("scores", scores),
		diagassert.V("scores[0]", scores[0]))
}

func TestVisualDemo_MethodCall(t *testing.T) {
	// Skip this test in normal runs - it's just for demonstration
	t.Skip("Demonstration test - enable to see visual output")

	user := struct {
		Name string
		Age  int
	}{Name: "Alice", Age: 16}

	// This will show a nice visual representation with values aligned
	diagassert.Assert(t, user.Age >= 18,
		diagassert.Values{"user": user})
}

func TestVisualDemo_ComplexExpression(t *testing.T) {
	// Skip this test in normal runs - it's just for demonstration
	t.Skip("Demonstration test - enable to see visual output")

	a, b, c := 5, 10, 15
	x, y := 20, 30

	// This will show multiple levels of evaluation
	diagassert.Assert(t, (a+b) > c && x < y,
		diagassert.V("a", a),
		diagassert.V("b", b),
		diagassert.V("c", c),
		diagassert.V("x", x),
		diagassert.V("y", y))
}
