package diagassert_test

import (
	"os"
	"strings"
	"testing"

	"github.com/paveg/diagassert"
)

// mockTestingT captures test output for validation
type mockTestingT struct {
	output string
	failed bool
}

func (m *mockTestingT) Error(args ...interface{}) {
	m.failed = true
	if len(args) > 0 {
		m.output = args[0].(string)
	}
}

func (m *mockTestingT) Fatal(args ...interface{}) {
	m.failed = true
	if len(args) > 0 {
		m.output = args[0].(string)
	}
}

func (m *mockTestingT) Helper() {}

// TestPhase2OutputDemo demonstrates Phase 2 enhanced output
func TestPhase2OutputDemo(t *testing.T) {
	// Enable machine readable for this test
	oldEnv := os.Getenv("DIAGASSERT_MACHINE_READABLE")
	defer os.Setenv("DIAGASSERT_MACHINE_READABLE", oldEnv)
	os.Setenv("DIAGASSERT_MACHINE_READABLE", "true")

	tests := []struct {
		name          string
		assertion     func(*mockTestingT)
		expectedParts []string
	}{
		{
			name: "simple comparison with enhanced output",
			assertion: func(mock *mockTestingT) {
				x := 10
				diagassert.Assert(mock, x > 20)
			},
			expectedParts: []string{
				"ASSERTION FAILED at",
				"Expression: x > 20",
				"Result: false",
				"EVALUATION TRACE:",
				"[MACHINE_READABLE_START]",
				"EXPR: x > 20",
				"RESULT: false",
			},
		},
		{
			name: "logical expression with enhanced output",
			assertion: func(mock *mockTestingT) {
				age := 16
				hasLicense := false
				diagassert.Assert(mock, age >= 18 && hasLicense)
			},
			expectedParts: []string{
				"ASSERTION FAILED at",
				"Expression: age >= 18 && hasLicense",
				"Result: false",
				"EVALUATION TRACE:",
				"[MACHINE_READABLE_START]",
				"FAIL_REASON:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTestingT{}
			tt.assertion(mock)

			if !mock.failed {
				t.Error("Expected assertion to fail")
				return
			}

			t.Logf("Phase 2 Output:\n%s", mock.output)

			// Verify expected parts are present
			for _, expected := range tt.expectedParts {
				if !strings.Contains(mock.output, expected) {
					t.Errorf("Expected output to contain %q", expected)
				}
			}
		})
	}
}

// TestPhase2StructFieldDemo demonstrates struct field evaluation
func TestPhase2StructFieldDemo(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	user := User{Name: "Alice", Age: 16}
	mock := &mockTestingT{}

	diagassert.Assert(mock, user.Age >= 18)

	if !mock.failed {
		t.Error("Expected assertion to fail")
		return
	}

	t.Logf("Struct Field Output:\n%s", mock.output)

	// Check for Phase 2 features
	expectedParts := []string{
		"ASSERTION FAILED at",
		"Expression: user.Age >= 18",
		"EVALUATION TRACE:",
		"VARIABLES:",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(mock.output, expected) {
			t.Errorf("Expected output to contain %q", expected)
		}
	}
}
