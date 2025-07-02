// Basic examples demonstrating diagassert usage
package main

import (
	"fmt"
	"strings"

	"github.com/paveg/diagassert"
)

// MockT is a simple implementation of diagassert.TestingT for demonstration
type MockT struct {
	name string
}

func (m *MockT) Error(args ...interface{}) {
	fmt.Printf("❌ %s: %v\n", m.name, args[0])
}

func (m *MockT) Fatal(args ...interface{}) {
	fmt.Printf("💥 %s: %v\n", m.name, args[0])
}

func (m *MockT) Helper() {}

func main() {
	fmt.Println("=== diagassert Basic Examples ===")

	// Example 1: Basic comparisons
	fmt.Println("1. Basic Comparisons")
	basicComparisons()

	// Example 2: String operations
	fmt.Println("\n2. String Operations")
	stringOperations()

	// Example 3: Boolean logic
	fmt.Println("\n3. Boolean Logic")
	booleanLogic()

	// Example 4: Successful assertions (no output)
	fmt.Println("\n4. Successful Assertions (no output)")
	successfulAssertions()

	fmt.Println("\n=== Examples Complete ===")
}

func basicComparisons() {
	t := &MockT{name: "BasicComparisons"}

	// These will fail and show diagnostic output
	x := 10
	y := 20

	fmt.Println("Failed assertion: x > y")
	diagassert.Assert(t, x > y)

	fmt.Println("\nFailed assertion: x == 15")
	diagassert.Assert(t, x == 15)
}

func stringOperations() {
	t := &MockT{name: "StringOperations"}

	name := "Alice"
	message := "Hello World"

	fmt.Println("Failed assertion: strings.Contains(name, \"Bob\")")
	diagassert.Assert(t, strings.Contains(name, "Bob"))

	fmt.Println("\nFailed assertion: len(message) < 5")
	diagassert.Assert(t, len(message) < 5)
}

func booleanLogic() {
	t := &MockT{name: "BooleanLogic"}

	age := 16
	hasLicense := false
	isStudent := true

	fmt.Println("Failed assertion: age >= 18 && hasLicense")
	diagassert.Assert(t, age >= 18 && hasLicense)

	fmt.Println("\nFailed assertion: !isStudent || age >= 21")
	diagassert.Assert(t, !isStudent || age >= 21)
}

func successfulAssertions() {
	t := &MockT{name: "SuccessfulAssertions"}

	// These should pass silently
	x := 10
	name := "Alice"

	diagassert.Assert(t, x > 5)
	diagassert.Assert(t, strings.Contains(name, "A"))
	diagassert.Assert(t, x == 10)

	fmt.Println("✅ All assertions passed (no output shown)")
}
