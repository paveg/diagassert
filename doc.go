// Package diagassert provides assertion utilities for diagnostic testing.
//
// Philosophy: Zero Learning Curve
//
// diagassert has just 2 functions - Assert and Require.
// Use any Go expression directly without learning dozens of APIs.
//
//	diagassert.Assert(t, user.Age >= 18)
//	diagassert.Assert(t, strings.Contains(name, "test"))
//	diagassert.Assert(t, x > y && y > 0)
//
// When assertions fail, you get detailed diagnostic information:
//
//	ASSERTION FAILED at user_test.go:42
//	Expression: user.Age >= 18 && user.HasLicense()
//	Result: false
//
//	[MACHINE_READABLE_START]
//	EXPR: user.Age >= 18 && user.HasLicense()
//	RESULT: false
//	[MACHINE_READABLE_END]
//
// Configuration is done via environment variables:
//   - DIAGASSERT_MACHINE_READABLE: "true" (default) | "false"
//
// For detailed documentation, see: https://github.com/paveg/diagassert
package diagassert