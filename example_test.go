package diagassert_test

import (
	"testing"

	"github.com/paveg/diagassert"
	"github.com/paveg/diagassert/internal/testutil"
)

func TestExample_ActualOutput(t *testing.T) {
	// Test to verify actual output
	if testing.Short() {
		t.Skip("Skipping example test in short mode")
	}

	// This test is expected to fail (for output verification)
	x := 10

	// This should pass
	diagassert.Assert(t, x > 5)

	// This should fail (to verify output)
	if false { // Don't actually execute
		diagassert.Assert(t, x > 20)
	}
}

func TestExample_Demo(t *testing.T) {
	t.Log("Demo of diagassert output")

	user := testutil.User{Name: "Alice", Age: 16}

	// Demo to see actual output (expecting failures)
	if false {
		diagassert.Assert(t, user.Age >= 18)
		diagassert.Assert(t, user.HasLicense())
		diagassert.Assert(t, user.Age >= 18 && user.HasLicense())
	}
}



// Manual test to see actual failure output
func TestExample_ManualFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Manual test - run with go test -v to see output")
	}

	t.Log("Manual test to see actual failure output")

	// Manually uncomment and run to see the output
	// x := 10
	// diagassert.Assert(t, x > 20)
}
