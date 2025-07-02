package diagassert

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// mockT is a mock implementation of the TestingT interface
type mockT struct {
	failed   bool
	messages []string
}

func newMockT() *mockT {
	return &mockT{
		messages: make([]string, 0),
	}
}

func (m *mockT) Fatal(args ...interface{}) {
	for _, arg := range args {
		m.messages = append(m.messages, fmt.Sprint(arg))
	}
	m.failed = true
	panic("FailNow called")
}

func (m *mockT) Error(args ...interface{}) {
	for _, arg := range args {
		m.messages = append(m.messages, fmt.Sprint(arg))
	}
	m.failed = true
}

func (m *mockT) Helper() {}

func (m *mockT) Failed() bool {
	return m.failed
}

func (m *mockT) getOutput() string {
	return strings.Join(m.messages, "\n")
}

// **Simple API: Use only Assert(t, expression)**

func TestAssert_SimpleAPI(t *testing.T) {
	t.Run("just pass expression - true case", func(t *testing.T) {
		// That's it! No need to learn other APIs
		Assert(t, true)
		Assert(t, 1 == 1)
		Assert(t, "hello" == "hello")
	})

	t.Run("just pass expression - false case", func(t *testing.T) {
		mock := newMockT()

		// Just pass a simple expression
		Assert(mock, false)

		if !mock.failed {
			t.Error("Assert(false) should fail")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "ASSERTION FAILED") {
			t.Errorf("Should contain failure message, got: %s", output)
		}
	})
}

func TestAssert_Expressions(t *testing.T) {
	// power-assert philosophy: you can pass any expression as-is

	t.Run("comparison", func(t *testing.T) {
		mock := newMockT()
		x := 10
		Assert(mock, x > 20)

		output := mock.getOutput()
		// The expression is displayed as-is
		if !strings.Contains(output, "x > 20") {
			t.Errorf("Should show expression, got: %s", output)
		}
	})

	t.Run("logical operations", func(t *testing.T) {
		mock := newMockT()
		age := 16
		hasLicense := false

		// Complex expressions are also displayed as-is
		Assert(mock, age >= 18 && hasLicense)

		output := mock.getOutput()
		if !strings.Contains(output, "age >= 18 && hasLicense") {
			t.Errorf("Should show full expression, got: %s", output)
		}
	})

	t.Run("method calls", func(t *testing.T) {
		mock := newMockT()
		user := User{Name: "Alice", Age: 16}

		// Method calls are also displayed as-is
		Assert(mock, user.IsAdult())

		output := mock.getOutput()
		if !strings.Contains(output, "user.IsAdult()") {
			t.Errorf("Should show method call, got: %s", output)
		}
	})
}

func TestAssert_NoLearningCurve(t *testing.T) {
	// Zero learning cost: no need to learn special matchers or APIs

	mock := newMockT()

	// Traditional assertion libraries require:
	// assert.Equal(t, actual, expected)
	// assert.Greater(t, x, y)
	// assert.Contains(t, str, substr)
	// assert.True(t, condition)
	// etc... many APIs to learn

	// With diagassert, it's all just this:
	actual, expected := 10, 20
	Assert(mock, actual == expected)

	x, y := 5, 10
	Assert(mock, x > y)

	str := "hello world"
	Assert(mock, strings.Contains(str, "world"))

	condition := false
	Assert(mock, condition)

	// Check the output
	output := mock.getOutput()
	if !strings.Contains(output, "actual == expected") {
		t.Error("Should show the exact expression used")
	}
}

func TestAssert_MachineReadable(t *testing.T) {
	t.Run("with machine readable section", func(t *testing.T) {
		// Machine-readable section is included by default
		mock := newMockT()
		Assert(mock, false)

		output := mock.getOutput()
		if !strings.Contains(output, "[MACHINE_READABLE_START]") {
			t.Error("Should include machine readable section by default")
		}
	})

	t.Run("disable machine readable", func(t *testing.T) {
		// Disable via environment variable
		os.Setenv("DIAGASSERT_MACHINE_READABLE", "false")
		defer os.Unsetenv("DIAGASSERT_MACHINE_READABLE")

		mock := newMockT()
		Assert(mock, false)

		output := mock.getOutput()
		if strings.Contains(output, "[MACHINE_READABLE_START]") {
			t.Error("Should not include machine readable section when disabled")
		}
	})
}

func TestRequire(t *testing.T) {
	t.Run("should panic on failure", func(t *testing.T) {
		mock := newMockT()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Require should panic on failure")
			}
		}()

		Require(mock, false)
	})
}

// Test struct
type User struct {
	Name string
	Age  int
}

func (u User) IsAdult() bool {
	return u.Age >= 18
}

// Future enhancement tests (Phase 2 and beyond)
func TestAssert_FutureEnhancements(t *testing.T) {
	t.Skip("Future enhancements - showing variable values in output")

	// In the future, we aim for output like this:
	// Expression: x > 20
	// ├─ x = 10 (int)
	// ├─ 20 = 20 (int)
	// └─ RESULT: false

	mock := newMockT()
	x := 10
	Assert(mock, x > 20)

	output := mock.getOutput()
	// Verify that variable values are displayed
	if !strings.Contains(output, "x = 10") {
		t.Error("Should show variable value")
	}
}
