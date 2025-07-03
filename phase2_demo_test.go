package diagassert_test

import (
	"os"
	"strings"
	"testing"

	"github.com/paveg/diagassert"
	"github.com/paveg/diagassert/internal/testutil"
)

// TestPhase2OutputDemo demonstrates Phase 2 enhanced output
func TestPhase2OutputDemo(t *testing.T) {
	// Enable machine readable for this test
	oldEnv := os.Getenv("DIAGASSERT_MACHINE_READABLE")
	defer os.Setenv("DIAGASSERT_MACHINE_READABLE", oldEnv)
	os.Setenv("DIAGASSERT_MACHINE_READABLE", "true")

	tests := []struct {
		name          string
		assertion     func(*testutil.MockT)
		expectedParts []string
	}{
		{
			name: "simple comparison with enhanced output",
			assertion: func(mock *testutil.MockT) {
				x := 10
				diagassert.Assert(mock, x > 20)
			},
			expectedParts: []string{
				"ASSERTION FAILED at",
				"assert(x > 20)",
				"[MACHINE_READABLE_START]",
				"EXPR: x > 20",
				"RESULT: false",
			},
		},
		{
			name: "logical expression with enhanced output",
			assertion: func(mock *testutil.MockT) {
				age := 16
				hasLicense := false
				diagassert.Assert(mock, age >= 18 && hasLicense)
			},
			expectedParts: []string{
				"ASSERTION FAILED at",
				"assert(age >= 18 && hasLicense)",
				"[MACHINE_READABLE_START]",
				"EXPR: age >= 18 && hasLicense",
				"RESULT: false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := testutil.NewMockT()
			tt.assertion(mock)

			if !mock.Failed() {
				t.Error("Expected assertion to fail")
				return
			}

			t.Logf("Phase 2 Output:\n%s", mock.GetOutput())

			// Verify expected parts are present
			for _, expected := range tt.expectedParts {
				if !strings.Contains(mock.GetOutput(), expected) {
					t.Errorf("Expected output to contain %q", expected)
				}
			}
		})
	}
}

// TestPhase2StructFieldDemo demonstrates struct field evaluation
func TestPhase2StructFieldDemo(t *testing.T) {
	user := testutil.User{Name: "Alice", Age: 16}
	mock := testutil.NewMockT()

	diagassert.Assert(mock, user.Age >= 18)

	if !mock.Failed() {
		t.Error("Expected assertion to fail")
		return
	}

	t.Logf("Struct Field Output:\n%s", mock.GetOutput())

	// Check for Phase 2 features
	expectedParts := []string{
		"ASSERTION FAILED at",
		"assert(user.Age >= 18)",
		"VARIABLES:",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(mock.GetOutput(), expected) {
			t.Errorf("Expected output to contain %q", expected)
		}
	}
}
