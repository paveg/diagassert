package diagassert_test

import (
	"testing"

	"github.com/paveg/diagassert"
	"github.com/paveg/diagassert/internal/testutil"
)

// ValueCaptureDemo demonstrates the enhanced output with value capture
func TestValueCaptureDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping demo test in short mode")
	}

	t.Log("=== Demo of Value Capture API ===")

	// Create a mock to capture output
	mock := testutil.NewMockT()

	// Demo 1: Simple value capture
	t.Log("\n--- Demo 1: Simple Value Capture ---")
	x := 15
	diagassert.Assert(mock, x > 20, diagassert.V("x", x))
	t.Logf("Output:\n%s", mock.GetOutput())

	// Demo 2: Multiple values with custom message
	t.Log("\n--- Demo 2: Multiple Values with Custom Message ---")
	mock = testutil.NewMockT()
	age := 16
	hasLicense := false
	diagassert.Assert(mock, age >= 18 && hasLicense,
		diagassert.Values{
			"age":        age,
			"hasLicense": hasLicense,
		},
		"User permission check failed")
	t.Logf("Output:\n%s", mock.GetOutput())

	// Demo 3: Complex struct with method calls
	t.Log("\n--- Demo 3: Complex Struct with Method Calls ---")
	mock = testutil.NewMockT()
	user := testutil.User{Name: "Alice", Age: 16}
	minAge := 18
	diagassert.Assert(mock, user.Age >= minAge && user.IsAdult(),
		diagassert.V("user", user),
		diagassert.V("user.Age", user.Age),
		diagassert.V("minAge", minAge),
		"Age verification failed")
	t.Logf("Output:\n%s", mock.GetOutput())

	// Demo 4: Mixed values and messages
	t.Log("\n--- Demo 4: Mixed Values and Messages ---")
	mock = testutil.NewMockT()
	a, b, c := 10, 20, 30
	diagassert.Assert(mock, a > b && b > c,
		"First condition check",
		diagassert.V("a", a),
		"Second condition check",
		diagassert.V("b", b),
		diagassert.V("c", c),
		"All values should form descending sequence")
	t.Logf("Output:\n%s", mock.GetOutput())
}