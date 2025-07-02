package diagassert

import (
	"strings"
	"testing"
)

// Note: Using mockT and User from assert_test.go to avoid redeclaration

// TestAssert_WithValueCapture tests assertion failures with comprehensive value capture functionality
func TestAssert_WithValueCapture(t *testing.T) {
	t.Run("single_value_with_V", func(t *testing.T) {
		mock := newMockT()
		x := 10

		// Test the exact API pattern: diagassert.V("x", x)
		Assert(mock, x > 20, V("x", x))

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "x = 10 (int)") {
			t.Errorf("Should show captured value with type, got: %s", output)
		}
	})

	t.Run("multiple_values_with_Values_map", func(t *testing.T) {
		mock := newMockT()
		x := 10
		y := 5

		// Test the exact API pattern: diagassert.Values{"x": x, "y": y}
		Assert(mock, x < y, Values{"x": x, "y": y})

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "x = 10") {
			t.Errorf("Should show x value, got: %s", output)
		}
		if !strings.Contains(output, "y = 5") {
			t.Errorf("Should show y value, got: %s", output)
		}
	})

	t.Run("mixed_V_and_Values", func(t *testing.T) {
		mock := newMockT()
		a := 10
		b := 20
		c := 30

		// Test mixing V() and Values{} in the same assertion
		Assert(mock, a > b, V("a", a), Values{"b": b, "c": c})

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "a = 10") {
			t.Errorf("Should show V() captured value, got: %s", output)
		}
		if !strings.Contains(output, "b = 20") {
			t.Errorf("Should show Values{} captured value, got: %s", output)
		}
		if !strings.Contains(output, "c = 30") {
			t.Errorf("Should show Values{} captured value, got: %s", output)
		}
	})

	t.Run("custom_message_with_values", func(t *testing.T) {
		mock := newMockT()
		score := 65
		passingScore := 70

		// Test custom message combined with value capture
		Assert(mock, score >= passingScore,
			"Student failed the exam",
			V("score", score),
			V("passingScore", passingScore))

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should contain custom message section, got: %s", output)
		}
		if !strings.Contains(output, "Student failed the exam") {
			t.Errorf("Should show custom message, got: %s", output)
		}
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "score = 65") {
			t.Errorf("Should show score value, got: %s", output)
		}
		if !strings.Contains(output, "passingScore = 70") {
			t.Errorf("Should show passingScore value, got: %s", output)
		}
	})

	t.Run("complex_struct_values", func(t *testing.T) {
		mock := newMockT()
		user := User{Name: "Alice", Age: 16}
		minAge := 18

		// Test capturing complex struct values
		Assert(mock, user.Age >= minAge,
			"User age validation",
			V("user", user),
			V("user.Age", user.Age),
			V("minAge", minAge))

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "user = {Alice 16}") {
			t.Errorf("Should show struct value, got: %s", output)
		}
		if !strings.Contains(output, "user.Age = 16") {
			t.Errorf("Should show struct field value, got: %s", output)
		}
		if !strings.Contains(output, "minAge = 18") {
			t.Errorf("Should show minAge value, got: %s", output)
		}
	})

	t.Run("multiple_messages_with_values", func(t *testing.T) {
		mock := newMockT()
		x := 5
		y := 10

		// Test multiple custom messages with value capture
		Assert(mock, x > y,
			"First validation message",
			V("x", x),
			"Second validation message",
			V("y", y),
			"Final validation message")

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should contain custom message section, got: %s", output)
		}
		if !strings.Contains(output, "First validation message Second validation message Final validation message") {
			t.Errorf("Should combine all messages, got: %s", output)
		}
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "x = 5") {
			t.Errorf("Should show x value, got: %s", output)
		}
		if !strings.Contains(output, "y = 10") {
			t.Errorf("Should show y value, got: %s", output)
		}
	})

	t.Run("type_information_in_output", func(t *testing.T) {
		mock := newMockT()
		intVal := 42
		strVal := "hello"
		boolVal := true
		floatVal := 3.14

		// Test that type information is included in output
		Assert(mock, false,
			V("intVal", intVal),
			V("strVal", strVal),
			V("boolVal", boolVal),
			V("floatVal", floatVal))

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "intVal = 42 (int)") {
			t.Errorf("Should show int type, got: %s", output)
		}
		if !strings.Contains(output, "strVal = hello (string)") {
			t.Errorf("Should show string type, got: %s", output)
		}
		if !strings.Contains(output, "boolVal = true (bool)") {
			t.Errorf("Should show bool type, got: %s", output)
		}
		if !strings.Contains(output, "floatVal = 3.14 (float64)") {
			t.Errorf("Should show float64 type, got: %s", output)
		}
	})
}

// TestAssert_NoValueCapture tests assertions without any value capture (backward compatibility)
func TestAssert_NoValueCapture(t *testing.T) {
	t.Run("original_API_still_works", func(t *testing.T) {
		mock := newMockT()

		// Test the original simple API
		Assert(mock, false)

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "ASSERTION FAILED") {
			t.Errorf("Should contain assertion failed message, got: %s", output)
		}
		if strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should not contain captured values section when none provided, got: %s", output)
		}
		if strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should not contain custom message section when none provided, got: %s", output)
		}
	})

	t.Run("expression_evaluation_without_capture", func(t *testing.T) {
		mock := newMockT()
		x := 10
		y := 20

		// Test complex expression without value capture
		Assert(mock, x > y && y < 15)

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "ASSERTION FAILED") {
			t.Errorf("Should contain assertion failed message, got: %s", output)
		}
		if !strings.Contains(output, "x > y && y < 15") {
			t.Errorf("Should show the expression, got: %s", output)
		}
		// Should still show evaluated variables if the evaluator finds them
		if !strings.Contains(output, "VARIABLES:") && !strings.Contains(output, "EVALUATION TRACE:") {
			t.Errorf("Should show some evaluation information, got: %s", output)
		}
	})
}

// TestAssert_OptionalValueCapture tests various combinations of optional value capture
func TestAssert_OptionalValueCapture(t *testing.T) {
	t.Run("message_only_no_values", func(t *testing.T) {
		mock := newMockT()

		// Test custom message without value capture
		Assert(mock, false, "This is just a custom message")

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should contain custom message section, got: %s", output)
		}
		if !strings.Contains(output, "This is just a custom message") {
			t.Errorf("Should show the custom message, got: %s", output)
		}
		if strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should not contain captured values section when none provided, got: %s", output)
		}
	})

	t.Run("values_only_no_message", func(t *testing.T) {
		mock := newMockT()
		x := 100
		y := 50

		// Test value capture without custom message
		Assert(mock, x < y, V("x", x), V("y", y))

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "x = 100") {
			t.Errorf("Should show x value, got: %s", output)
		}
		if !strings.Contains(output, "y = 50") {
			t.Errorf("Should show y value, got: %s", output)
		}
		if strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should not contain custom message section when none provided, got: %s", output)
		}
	})

	t.Run("empty_Values_map", func(t *testing.T) {
		mock := newMockT()

		// Test with empty Values map
		Assert(mock, false, Values{}, "Message with empty values")

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should contain custom message section, got: %s", output)
		}
		if strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should not contain captured values section when Values{} is empty, got: %s", output)
		}
	})

	t.Run("nil_and_zero_values", func(t *testing.T) {
		mock := newMockT()
		var nilPtr *string
		zeroInt := 0
		emptyString := ""

		// Test capturing nil and zero values
		Assert(mock, false,
			"Testing nil and zero values",
			V("nilPtr", nilPtr),
			V("zeroInt", zeroInt),
			V("emptyString", emptyString))

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "nilPtr = <nil>") {
			t.Errorf("Should show nil value, got: %s", output)
		}
		if !strings.Contains(output, "zeroInt = 0") {
			t.Errorf("Should show zero int value, got: %s", output)
		}
		if !strings.Contains(output, "emptyString =  (string)") {
			t.Errorf("Should show empty string value, got: %s", output)
		}
	})

	t.Run("interleaved_messages_and_values", func(t *testing.T) {
		mock := newMockT()
		a := 1
		b := 2
		c := 3

		// Test interleaving messages and values in various orders
		Assert(mock, false,
			"Start message",
			V("a", a),
			"Middle message",
			Values{"b": b, "c": c},
			"End message")

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should contain custom message section, got: %s", output)
		}
		if !strings.Contains(output, "Start message Middle message End message") {
			t.Errorf("Should combine all messages in order, got: %s", output)
		}
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "a = 1") {
			t.Errorf("Should show a value, got: %s", output)
		}
		if !strings.Contains(output, "b = 2") {
			t.Errorf("Should show b value, got: %s", output)
		}
		if !strings.Contains(output, "c = 3") {
			t.Errorf("Should show c value, got: %s", output)
		}
	})
}

// TestRequire_WithValueCapture tests the Require function with value capture
func TestRequire_WithValueCapture(t *testing.T) {
	t.Run("require_with_value_capture_should_panic", func(t *testing.T) {
		mock := newMockT()
		x := 5

		defer func() {
			if r := recover(); r == nil {
				t.Error("Require should panic on failure")
			}
		}()

		// Test that Require panics with value capture
		Require(mock, x > 10,
			"Critical condition failed",
			V("x", x),
			V("threshold", 10))

		if !mock.failed {
			t.Error("Require should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should contain custom message section, got: %s", output)
		}
		if !strings.Contains(output, "Critical condition failed") {
			t.Errorf("Should show custom message, got: %s", output)
		}
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "x = 5") {
			t.Errorf("Should show x value, got: %s", output)
		}
		if !strings.Contains(output, "threshold = 10") {
			t.Errorf("Should show threshold value, got: %s", output)
		}
	})
}

// TestValueCapture_MachineReadableOutput tests that machine-readable output includes captured values
func TestValueCapture_MachineReadableOutput(t *testing.T) {
	t.Run("machine_readable_includes_captured_values", func(t *testing.T) {
		mock := newMockT()
		x := 42
		message := "Test message"

		// Test that machine-readable output includes captured values and messages
		Assert(mock, false,
			message,
			V("x", x))

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "[MACHINE_READABLE_START]") {
			t.Errorf("Should contain machine readable section, got: %s", output)
		}
		if !strings.Contains(output, "CUSTOM_MESSAGE: Test message") {
			t.Errorf("Should include custom message in machine readable section, got: %s", output)
		}
		if !strings.Contains(output, "CAPTURED_VALUES_START") {
			t.Errorf("Should include captured values start marker, got: %s", output)
		}
		if !strings.Contains(output, "VALUE: x = 42 (int)") {
			t.Errorf("Should include captured value in machine readable format, got: %s", output)
		}
		if !strings.Contains(output, "CAPTURED_VALUES_END") {
			t.Errorf("Should include captured values end marker, got: %s", output)
		}
		if !strings.Contains(output, "[MACHINE_READABLE_END]") {
			t.Errorf("Should contain machine readable end marker, got: %s", output)
		}
	})
}
