# diagassert - No API is the Best API

## Design Philosophy

The most important principle of power-assert: **No need to memorize APIs**

### Problems with Traditional Assertion Libraries

```go
// Need to memorize many APIs
assert.Equal(t, actual, expected)
assert.NotEqual(t, actual, expected)
assert.Greater(t, x, y)
assert.GreaterOrEqual(t, x, y)
assert.Contains(t, str, substr)
assert.NotContains(t, str, substr)
assert.True(t, condition)
assert.False(t, condition)
assert.Nil(t, obj)
assert.NotNil(t, obj)
// ... dozens of APIs ...
```

### diagassert's Approach

```go
// This is all you need to remember!
diagassert.Assert(t, actual == expected)
diagassert.Assert(t, actual != expected)
diagassert.Assert(t, x > y)
diagassert.Assert(t, x >= y)
diagassert.Assert(t, strings.Contains(str, substr))
diagassert.Assert(t, !strings.Contains(str, substr))
diagassert.Assert(t, condition)
diagassert.Assert(t, !condition)
diagassert.Assert(t, obj == nil)
diagassert.Assert(t, obj != nil)
```

## API Specification

### Public API (That's all!)

```go
// Assert - evaluates any expression
func Assert(t testing.TB, expr bool)

// Require - terminates immediately on failure
func Require(t testing.TB, expr bool)
```

### Configuration (Environment Variables)

Control via environment variables without increasing API surface:

- `DIAGASSERT_MACHINE_READABLE`: "true" (default) | "false"

## Output Format

### Basic Failure Output

```
ASSERTION FAILED at user_test.go:42
Expression: age >= 18 && hasLicense
Result: false
```

### Future Enhancement (Phase 2 onwards)

```
ASSERTION FAILED at user_test.go:42
Expression: age >= 18 && hasLicense
Result: false

EVALUATION TRACE:
├─ age >= 18
│  ├─ age = 16 (int)
│  ├─ 18 = 18 (int)
│  └─ RESULT: false
└─ hasLicense = false (bool)

[MACHINE_READABLE_START]
EXPR: age >= 18 && hasLicense
RESULT: false
FAIL_REASON: left_operand_false
[MACHINE_READABLE_END]
```

## Implementation Plan (Revised)

### Phase 1: Minimal Implementation (Current)
- ✅ Only `Assert` and `Require`
- ✅ Extract expressions from source code
- ✅ Basic output
- ✅ Configuration via environment variables

### Phase 2: Expression Evaluation
- AST analysis of expressions
- Runtime variable value retrieval
- Build and display evaluation trees

### Phase 3: Advanced Expression Support
- Display method call results
- Display struct field values
- Display array/slice element values

### Phase 4: Enhanced Machine Readability
- Structured evaluation trees
- Failure path tracking
- AI tool hint generation

## Usage Example

```go
func TestUserPermissions(t *testing.T) {
    user := User{Name: "Alice", Age: 16}
    
    // That's all! No special APIs needed
    diagassert.Assert(t, user.Age >= 18)
    diagassert.Assert(t, user.HasLicense())
    diagassert.Assert(t, user.Age >= 18 && user.HasLicense())
}
```

## Why Simple API Matters

1. **Zero Learning Cost**: Just write Go expressions directly
2. **No IDE Completion Needed**: No need to search for special methods
3. **Easy Migration**: Simple migration from existing `if` statements
4. **Expressiveness**: Can use any Go expression

## Technical Challenges and Solutions

### Challenge 1: Retrieving Expression Values at Runtime
- **Solution**: Source code AST analysis + runtime reflection

### Challenge 2: Balancing Machine Readability and Simplicity
- **Solution**: Hybrid output by default, controlled via environment variables

### Challenge 3: Performance
- **Solution**: Detailed analysis only on failure

## Summary

diagassert is a simple yet powerful assertion library that inherits power-assert's "No API is the best API" philosophy while adding machine readability for the AI era.

This design provides:

- **Only 2 APIs**: Assert and Require
- **Intuitive Usage**: Pass any Go expression directly
- **Configuration via Environment Variables**: No API bloat
- **Machine Readability**: Enabled by default, but non-intrusive