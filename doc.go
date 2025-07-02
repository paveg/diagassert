// Package diagassert provides assertion utilities for diagnostic testing.
//
// API Functions:
//   - Assert(t testing.TB, expr bool) - evaluates any Go expression
//   - Require(t testing.TB, expr bool) - like Assert but stops test execution on failure
//
// Configuration:
//   - DIAGASSERT_MACHINE_READABLE: "true" (default) | "false"
//
// Example:
//
//	diagassert.Assert(t, user.Age >= 18)
//	diagassert.Assert(t, strings.Contains(name, "test"))
//	diagassert.Assert(t, x > y && y > 0)
//
// For detailed documentation, see: https://github.com/paveg/diagassert
package diagassert
