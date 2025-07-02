package diagassert_test

import (
	"strings"
	"testing"

	"github.com/paveg/diagassert"
	"github.com/paveg/diagassert/internal/testutil"
)

// Test for value capture functionality - Original Request Test Cases

func TestAssert_WithValueCapture(t *testing.T) {
	t.Run("single value capture", func(t *testing.T) {
		mock := testutil.NewMockT()
		x := 10

		// Provide values using V() helper
		diagassert.Assert(mock, x > 20, diagassert.V("x", x))

		output := mock.GetOutput()

		// Verify that values are displayed
		if !strings.Contains(output, "x = 10") {
			t.Errorf("Should show captured value, got: %s", output)
		}
	})

	t.Run("multiple value captures", func(t *testing.T) {
		mock := testutil.NewMockT()
		x, y := 10, 20

		// Capture multiple values
		diagassert.Assert(mock, x > y,
			diagassert.V("x", x),
			diagassert.V("y", y))

		output := mock.GetOutput()

		if !strings.Contains(output, "x = 10") || !strings.Contains(output, "y = 20") {
			t.Errorf("Should show all captured values, got: %s", output)
		}
	})

	t.Run("values map", func(t *testing.T) {
		mock := testutil.NewMockT()
		age := 16
		hasLicense := false

		// Use Values map
		diagassert.Assert(mock, age >= 18 && hasLicense,
			diagassert.Values{
				"age":        age,
				"hasLicense": hasLicense,
			})

		output := mock.GetOutput()

		if !strings.Contains(output, "age = 16") {
			t.Errorf("Should show age value, got: %s", output)
		}
		if !strings.Contains(output, "hasLicense = false") {
			t.Errorf("Should show hasLicense value, got: %s", output)
		}
	})

	t.Run("with custom message", func(t *testing.T) {
		mock := testutil.NewMockT()
		x := 10

		// Combination of value capture and custom message
		diagassert.Assert(mock, x > 20,
			diagassert.V("x", x),
			"Expected x to be greater than 20")

		output := mock.GetOutput()

		if !strings.Contains(output, "x = 10") {
			t.Errorf("Should show value, got: %s", output)
		}
		if !strings.Contains(output, "Expected x to be greater than 20") {
			t.Errorf("Should show custom message, got: %s", output)
		}
	})

	t.Run("complex expression with values", func(t *testing.T) {
		mock := testutil.NewMockT()
		user := testutil.User{Name: "Alice", Age: 16}
		minAge := 18

		diagassert.Assert(mock, user.Age >= minAge && user.IsAdult(),
			diagassert.Values{
				"user.Age": user.Age,
				"minAge":   minAge,
				"user":     user,
			})

		output := mock.GetOutput()

		// Structured evaluation tree is displayed
		if !strings.Contains(output, "EVALUATION TRACE:") {
			t.Errorf("Should show evaluation trace, got: %s", output)
		}

		// Provided values are used
		if !strings.Contains(output, "user.Age = 16") {
			t.Errorf("Should use provided value for user.Age, got: %s", output)
		}
	})
}

func TestAssert_NoValueCapture(t *testing.T) {
	// Basic functionality works even without value capture
	t.Run("basic functionality without value capture", func(t *testing.T) {
		mock := testutil.NewMockT()
		diagassert.Assert(mock, false)

		output := mock.GetOutput()

		// Basic failure message is displayed
		if !strings.Contains(output, "ASSERTION FAILED") {
			t.Errorf("Should show basic failure message, got: %s", output)
		}
	})

	t.Run("expression shown without values", func(t *testing.T) {
		mock := testutil.NewMockT()
		x := 10
		diagassert.Assert(mock, x > 20)

		output := mock.GetOutput()

		// Expression is displayed
		if !strings.Contains(output, "x > 20") {
			t.Errorf("Should show expression, got: %s", output)
		}

		// Values are not displayed (since not provided)
		// But expression structure is displayed
		if !strings.Contains(output, "EVALUATION TRACE:") {
			t.Errorf("Should show evaluation trace structure, got: %s", output)
		}
	})
}

func TestAssert_OptionalValueCapture(t *testing.T) {
	// Value capture is completely optional
	mock := testutil.NewMockT()
	x, y, z := 10, 20, 30

	// Provide only some values
	diagassert.Assert(mock, x > y && y < z,
		diagassert.V("y", y)) // Provide only y value

	output := mock.GetOutput()

	// Provided values are displayed
	if !strings.Contains(output, "y = 20") {
		t.Errorf("Should show provided value, got: %s", output)
	}

	// Full expression is displayed
	if !strings.Contains(output, "x > y && y < z") {
		t.Errorf("Should show full expression, got: %s", output)
	}
}

// Usage patterns demonstration
func Example_usagePatterns() {
	t := &testing.T{}

	// Pattern 1: Simple (no values)
	x := 10
	diagassert.Assert(t, x > 20)

	// Pattern 2: Provide only important values
	diagassert.Assert(t, x > 20, diagassert.V("x", x))

	// Pattern 3: Provide all values for complex expression
	y := 30
	diagassert.Assert(t, x > 20 && y < 40,
		diagassert.Values{"x": x, "y": y})

	// Pattern 4: With message
	diagassert.Assert(t, x > 20,
		diagassert.V("x", x),
		"Value check failed")
}
