# Basic diagassert Examples

This example demonstrates the fundamental usage patterns of diagassert.

## What's Demonstrated

1. **Basic Comparisons** - Simple numeric comparisons
2. **String Operations** - String checks and length validation  
3. **Boolean Logic** - Complex logical expressions
4. **Successful Assertions** - Showing that passing assertions produce no output

## Key Concepts

### Zero Learning Curve

You don't need to learn special assertion APIs. Just use Go expressions:

```go
// Instead of assert.Equal(t, actual, expected)
diagassert.Assert(t, actual == expected)

// Instead of assert.Greater(t, x, y)  
diagassert.Assert(t, x > y)

// Instead of assert.Contains(t, str, substr)
diagassert.Assert(t, strings.Contains(str, substr))
```

### Expression Display

When assertions fail, you see exactly what expression was evaluated:

```text
ASSERTION FAILED at main.go:45
Expression: age >= 18 && hasLicense
Result: false
```

### Machine-Readable Output

By default, machine-readable sections are included for AI tools and automation:

```text
[MACHINE_READABLE_START]
EXPR: age >= 18 && hasLicense
RESULT: false
[MACHINE_READABLE_END]
```

## Running This Example

```bash
cd examples/basic
go run main.go
```

## Expected Output

You'll see assertion failures with detailed diagnostic information, followed by
successful assertions that produce no output (as expected).

### Current Limitations

**Note**: The current implementation shows placeholder values like `<x>` instead
of actual variable values. This is because Phase 3 (Advanced Expression Support)
is still in progress. The AST parsing and tree construction are complete, but
runtime variable value extraction is not yet fully implemented.

For actual variable values, you can use the Value Capture API:

```go
x := 10
diagassert.Assert(t, x > 20, diagassert.V("x", x))
```

This will show the real value of `x` in the diagnostic output.
