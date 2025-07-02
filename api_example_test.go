package diagassert_test

import (
	"testing"

	"github.com/paveg/diagassert"
)

// ExampleAPI demonstrates the new value capture and message API
func TestAPI_Example(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API example test in short mode")
	}

	// Example usage patterns that would fail (commented out to avoid test failure)
	if false {
		// 1. Single value capture with V()
		x := 10
		diagassert.Assert(t, x > 20, diagassert.V("x", x))

		// 2. Multiple values with Values map
		y := 5
		diagassert.Assert(t, x < y, diagassert.Values{"x": x, "y": y})

		// 3. Custom messages as strings
		condition := false
		diagassert.Assert(t, condition, "This condition should be true")

		// 4. Mixed usage: values + messages
		diagassert.Assert(t, x < y,
			diagassert.V("x", x),
			"Values should be ordered",
			diagassert.V("y", y))

		// 5. Complex mixed usage
		user := struct {
			Name string
			Age  int
		}{Name: "Alice", Age: 16}
		minAge := 18

		diagassert.Assert(t, user.Age >= minAge,
			"User age validation failed",
			diagassert.V("user.Age", user.Age),
			diagassert.V("minAge", minAge),
			diagassert.Values{"user.Name": user.Name, "required": "adult"},
			"Expected user to be an adult")
	}

	// The original API still works perfectly
	diagassert.Assert(t, true)
	x := 1
	diagassert.Assert(t, x == 1)
	str := "hello"
	diagassert.Assert(t, str == "hello")
}

// Manual test to see the actual output
func TestAPI_ManualExample(t *testing.T) {
	if testing.Short() {
		t.Skip("Manual test - run with go test -v to see output")
	}

	t.Log("Manual test to see the new API output")

	// Uncomment and run to see the enhanced output:
	// x := 10
	// diagassert.Assert(t, x > 20, diagassert.V("x", x), "Expected x to be greater than 20")
}
