// Package diagassert provides assertion utilities for diagnostic testing.
//
// Enhanced API for Value Capture and Custom Messages:
//
// The Assert and Require functions now support variadic arguments for enhanced
// diagnostic output. You can mix and match the following:
//
//  1. Value capture with V():
//     diagassert.Assert(t, expr, diagassert.V("x", x))
//
//  2. Multiple values with Values map:
//     diagassert.Assert(t, expr, diagassert.Values{"x": x, "y": y})
//
//  3. Custom messages:
//     diagassert.Assert(t, expr, "Custom error message")
//
//  4. Mixed usage:
//     diagassert.Assert(t, expr, diagassert.V("x", x), "Error occurred", diagassert.V("y", y))
//
// The original simple API is fully backward compatible:
//
//	diagassert.Assert(t, expr)
package diagassert

import "fmt"

// Value represents a named value for diagnostic output
type Value struct {
	Name  string
	Value interface{}
}

// V creates a Value with the given name and value.
// This is a helper function for capturing individual values.
//
// Usage: diagassert.Assert(t, expr, diagassert.V("x", x))
func V(name string, value interface{}) Value {
	return Value{Name: name, Value: value}
}

// Values represents a map of named values for diagnostic output.
// This allows capturing multiple values at once.
//
// Usage: diagassert.Assert(t, expr, diagassert.Values{"x": x, "y": y})
type Values map[string]interface{}

// AssertionContext holds additional context for assertions
type AssertionContext struct {
	Values   []Value
	Messages []string
}

// NewAssertionContext creates a new assertion context from variadic arguments
func NewAssertionContext(args ...interface{}) *AssertionContext {
	ctx := &AssertionContext{
		Values:   make([]Value, 0),
		Messages: make([]string, 0),
	}

	for _, arg := range args {
		switch v := arg.(type) {
		case Value:
			ctx.Values = append(ctx.Values, v)
		case Values:
			// Convert Values map to individual Value structs
			for name, value := range v {
				ctx.Values = append(ctx.Values, Value{Name: name, Value: value})
			}
		case string:
			ctx.Messages = append(ctx.Messages, v)
		case fmt.Stringer:
			ctx.Messages = append(ctx.Messages, v.String())
		default:
			// Convert other types to string representation
			ctx.Messages = append(ctx.Messages, fmt.Sprintf("%v", v))
		}
	}

	return ctx
}

// HasValues returns true if the context contains any values
func (ctx *AssertionContext) HasValues() bool {
	return len(ctx.Values) > 0
}

// HasMessages returns true if the context contains any messages
func (ctx *AssertionContext) HasMessages() bool {
	return len(ctx.Messages) > 0
}

// GetValuesMap returns all values as a map for easy access
func (ctx *AssertionContext) GetValuesMap() map[string]interface{} {
	result := make(map[string]interface{})
	for _, v := range ctx.Values {
		result[v.Name] = v.Value
	}
	return result
}

// GetCombinedMessage returns all messages joined together
func (ctx *AssertionContext) GetCombinedMessage() string {
	if len(ctx.Messages) == 0 {
		return ""
	}

	combined := ""
	for i, msg := range ctx.Messages {
		if i > 0 {
			combined += " "
		}
		combined += msg
	}
	return combined
}
