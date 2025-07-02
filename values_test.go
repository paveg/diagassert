package diagassert

import (
	"strings"
	"testing"
)

// TestAPI_ValueCapture tests the new API for value capture and custom messages
func TestAPI_ValueCapture(t *testing.T) {
	t.Run("single value capture with V()", func(t *testing.T) {
		mock := newMockT()
		x := 10

		// Test the exact usage pattern: diagassert.V("x", x)
		Assert(mock, x > 20, V("x", x))

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "x = 10") {
			t.Errorf("Should show captured value x = 10, got: %s", output)
		}
	})

	t.Run("multiple values with Values map", func(t *testing.T) {
		mock := newMockT()
		x := 10
		y := 5

		// Test the exact usage pattern: diagassert.Values{"x": x, "y": y}
		Assert(mock, x < y, Values{"x": x, "y": y})

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "x = 10") {
			t.Errorf("Should show captured value x = 10, got: %s", output)
		}
		if !strings.Contains(output, "y = 5") {
			t.Errorf("Should show captured value y = 5, got: %s", output)
		}
	})

	t.Run("custom message as string", func(t *testing.T) {
		mock := newMockT()
		condition := false

		// Test custom messages as strings
		Assert(mock, condition, "This is a custom error message")

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should contain custom message section, got: %s", output)
		}
		if !strings.Contains(output, "This is a custom error message") {
			t.Errorf("Should show custom message, got: %s", output)
		}
	})

	t.Run("mixed usage: values + messages", func(t *testing.T) {
		mock := newMockT()
		x := 10
		y := 5

		// Test mixing values and messages
		Assert(mock, x < y, V("x", x), "Values should be ordered", V("y", y))

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should contain custom message section, got: %s", output)
		}
		if !strings.Contains(output, "Values should be ordered") {
			t.Errorf("Should show custom message, got: %s", output)
		}
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "x = 10") {
			t.Errorf("Should show captured value x = 10, got: %s", output)
		}
		if !strings.Contains(output, "y = 5") {
			t.Errorf("Should show captured value y = 5, got: %s", output)
		}
	})

	t.Run("complex mixed usage", func(t *testing.T) {
		mock := newMockT()
		user := struct {
			Name string
			Age  int
		}{Name: "Alice", Age: 16}
		minAge := 18

		// Test complex mixed usage combining all patterns
		Assert(mock, user.Age >= minAge,
			"User age validation failed",
			V("user.Age", user.Age),
			V("minAge", minAge),
			Values{"user.Name": user.Name, "required": "adult"},
			"Expected user to be an adult")

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()

		// Check for custom messages
		if !strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should contain custom message section, got: %s", output)
		}
		if !strings.Contains(output, "User age validation failed") {
			t.Errorf("Should show first custom message, got: %s", output)
		}
		if !strings.Contains(output, "Expected user to be an adult") {
			t.Errorf("Should show second custom message, got: %s", output)
		}

		// Check for captured values
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "user.Age = 16") {
			t.Errorf("Should show user.Age value, got: %s", output)
		}
		if !strings.Contains(output, "minAge = 18") {
			t.Errorf("Should show minAge value, got: %s", output)
		}
		if !strings.Contains(output, "user.Name = Alice") {
			t.Errorf("Should show user.Name value, got: %s", output)
		}
		if !strings.Contains(output, "required = adult") {
			t.Errorf("Should show required value, got: %s", output)
		}
	})
}

// TestAPI_BackwardCompatibility ensures the original API still works
func TestAPI_BackwardCompatibility(t *testing.T) {
	t.Run("original API without additional args", func(t *testing.T) {
		mock := newMockT()

		// Test that the original API still works
		Assert(mock, false)

		if !mock.failed {
			t.Error("Assert should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "ASSERTION FAILED") {
			t.Errorf("Should contain assertion failed message, got: %s", output)
		}

		// Should not contain new sections when no args provided
		if strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should not contain custom message section, got: %s", output)
		}
		if strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should not contain captured values section, got: %s", output)
		}
	})
}

// TestRequire_ValueCapture tests the Require function with value capture
func TestRequire_ValueCapture(t *testing.T) {
	t.Run("require with value capture", func(t *testing.T) {
		mock := newMockT()
		x := 10

		defer func() {
			if r := recover(); r == nil {
				t.Error("Require should panic on failure")
			}
		}()

		// Test Require with value capture
		Require(mock, x > 20, V("x", x), "Critical condition failed")

		if !mock.failed {
			t.Error("Require should have failed")
		}

		output := mock.getOutput()
		if !strings.Contains(output, "CAPTURED VALUES:") {
			t.Errorf("Should contain captured values section, got: %s", output)
		}
		if !strings.Contains(output, "CUSTOM MESSAGE:") {
			t.Errorf("Should contain custom message section, got: %s", output)
		}
	})
}

// TestV_Helper tests the V helper function
func TestV_Helper(t *testing.T) {
	t.Run("V helper creates correct Value", func(t *testing.T) {
		value := V("test", 42)

		if value.Name != "test" {
			t.Errorf("Expected name 'test', got %s", value.Name)
		}
		if value.Value != 42 {
			t.Errorf("Expected value 42, got %v", value.Value)
		}
	})
}

// TestValues_Type tests the Values type
func TestValues_Type(t *testing.T) {
	t.Run("Values type works as map", func(t *testing.T) {
		values := Values{
			"x": 10,
			"y": "hello",
			"z": true,
		}

		if values["x"] != 10 {
			t.Errorf("Expected x=10, got %v", values["x"])
		}
		if values["y"] != "hello" {
			t.Errorf("Expected y='hello', got %v", values["y"])
		}
		if values["z"] != true {
			t.Errorf("Expected z=true, got %v", values["z"])
		}
	})
}

// TestAssertionContext tests the AssertionContext functionality
func TestAssertionContext(t *testing.T) {
	t.Run("NewAssertionContext with mixed args", func(t *testing.T) {
		ctx := NewAssertionContext(
			V("x", 10),
			"message1",
			Values{"y": 20, "z": 30},
			"message2",
		)

		if !ctx.HasValues() {
			t.Error("Should have values")
		}
		if !ctx.HasMessages() {
			t.Error("Should have messages")
		}

		if len(ctx.Values) != 3 { // x, y, z
			t.Errorf("Expected 3 values, got %d", len(ctx.Values))
		}
		if len(ctx.Messages) != 2 { // message1, message2
			t.Errorf("Expected 2 messages, got %d", len(ctx.Messages))
		}

		valuesMap := ctx.GetValuesMap()
		if valuesMap["x"] != 10 {
			t.Errorf("Expected x=10, got %v", valuesMap["x"])
		}
		if valuesMap["y"] != 20 {
			t.Errorf("Expected y=20, got %v", valuesMap["y"])
		}
		if valuesMap["z"] != 30 {
			t.Errorf("Expected z=30, got %v", valuesMap["z"])
		}

		combined := ctx.GetCombinedMessage()
		if !strings.Contains(combined, "message1") {
			t.Errorf("Should contain message1, got: %s", combined)
		}
		if !strings.Contains(combined, "message2") {
			t.Errorf("Should contain message2, got: %s", combined)
		}
	})
}

// Note: Using mockT and newMockT from assert_test.go
