# TDD Implementation Plan

**Status Update**: This document shows the original development plan. Many
phases have been completed or are in progress.

## Phase 1: Basic Features ✅ COMPLETED

### Step 1.1: Minimal Assert Implementation

```go
// Test cases
func TestAssert_BasicBool(t *testing.T) {
    // Should pass for true
    diagassert.Assert(t, true)
    
    // Should fail for false
    mockT := new(testing.T)
    diagassert.Assert(mockT, false)
    // Verify mockT.Failed() == true
}
```

### Step 1.2: Custom Message Support

```go
func TestAssert_WithMessage(t *testing.T) {
    mockT := new(testing.T)
    diagassert.Assert(mockT, false, "Expected true but got false")
    // Verify output contains custom message
}
```

### Step 1.3: Basic Output Format

```go
func TestAssert_BasicOutput(t *testing.T) {
    mockT := new(testing.T)
    diagassert.Assert(mockT, false)
    // Verify output matches expected format
    // Should contain "ASSERTION FAILED at ..."
}
```

## Phase 2: Expression Evaluation ✅ COMPLETED

### Step 2.1: Comparison Operator Support

```go
func TestAssert_Comparison(t *testing.T) {
    x := 10
    mockT := new(testing.T)
    diagassert.Assert(mockT, x > 20)
    // Verify output contains "x > 20" and "x = 10"
}
```

### Step 2.2: Expressions with Multiple Variables

```go
func TestAssert_MultipleVariables(t *testing.T) {
    x, y := 10, 20
    mockT := new(testing.T)
    diagassert.Assert(mockT, x > y)
    // Verify output contains values of both variables
}
```

### Step 2.3: Logical Operator Support

```go
func TestAssert_LogicalOperators(t *testing.T) {
    age := 16
    hasLicense := false
    mockT := new(testing.T)
    diagassert.Assert(mockT, age >= 18 && hasLicense)
    // Verify evaluation tree is correctly displayed
}
```

## Phase 3: Advanced Expression Support 🔄 IN PROGRESS

**Current Status**: AST parsing and tree construction completed, but runtime
variable value extraction is using placeholders only.

### Step 3.1: Struct Field Access

```go
func TestAssert_StructField(t *testing.T) {
    user := User{Name: "Alice", Age: 16}
    mockT := new(testing.T)
    diagassert.Assert(mockT, user.Age >= 18)
    // Verify output contains "user.Age = 16"
}
```

### Step 3.2: Method Calls

```go
func TestAssert_MethodCall(t *testing.T) {
    user := User{Name: "Alice", Age: 16}
    mockT := new(testing.T)
    diagassert.Assert(mockT, user.IsAdult())
    // Verify method call and result are displayed
}
```

### Step 3.3: Array/Slice Access

```go
func TestAssert_SliceAccess(t *testing.T) {
    scores := []int{65, 70, 75}
    mockT := new(testing.T)
    diagassert.Assert(mockT, scores[0] >= 80)
    // Verify output contains "scores[0] = 65"
}
```

## Phase 4: Machine-Readable Output ✅ COMPLETED

### Step 4.1: Machine-Readable Section

```go
func TestAssert_MachineReadableSection(t *testing.T) {
    mockT := new(testing.T)
    diagassert.EnableMachineReadable()
    diagassert.Assert(mockT, false)
    // Verify [MACHINE_READABLE_START] section is included
}
```

### Step 4.2: Expression Tree IDs

```go
func TestAssert_ExpressionIDs(t *testing.T) {
    x := 10
    mockT := new(testing.T)
    diagassert.Assert(mockT, x > 5 && x < 20)
    // Verify each expression has an assigned ID
}
```

### Step 4.3: Failure Path Tracking

```go
func TestAssert_FailurePath(t *testing.T) {
    // Verify failure path is correctly recorded for complex expressions
}
```

## Phase 5: Value Capture API ✅ COMPLETED

**Note**: This phase was implemented differently than originally planned, using
a more flexible approach.

### Step 5.1: Individual Value Capture

```go
func TestAssert_ValueCapture(t *testing.T) {
    x := 10
    mockT := new(testing.T)
    diagassert.Assert(mockT, x > 20, diagassert.V("x", x))
    // Verify output contains actual value of x
}
```

### Step 5.2: Multiple Value Capture

```go
func TestAssert_MultipleValues(t *testing.T) {
    x, y := 10, 20
    mockT := new(testing.T)
    diagassert.Assert(mockT, x > y, diagassert.Values{"x": x, "y": y})
    // Verify output contains actual values of both variables
}
```

### Step 5.3: Mixed Usage with Messages

```go
func TestAssert_MixedUsage(t *testing.T) {
    x := 10
    mockT := new(testing.T)
    diagassert.Assert(mockT, x > 20, diagassert.V("x", x), "Custom message")
    // Verify both values and messages are displayed
}
```

## Phase 6: Output Format Extensions 🔄 PLANNED

### Step 6.1: Markdown Output

```go
func TestAssert_MarkdownFormat(t *testing.T) {
    mockT := new(testing.T)
    diagassert.AssertWithFormat(mockT, false, diagassert.FormatMarkdown)
    // Verify output is in Markdown format
}
```

### Step 6.2: JSON Output

```go
func TestAssert_JSONFormat(t *testing.T) {
    mockT := new(testing.T)
    diagassert.AssertWithFormat(mockT, false, diagassert.FormatJSON)
    // Verify valid JSON is output
}
```

## Implementation Priority

### Essential Features (Phase 1-2)

- Basic assertions
- Comparison and logical operator support
- Basic output format

### Important Features (Phase 3-4)

- Struct and method support
- Machine-readable section

### Additional Features (Phase 5-6)

- Context features
- Multiple output formats

## Success Criteria

### Functionality

- All test cases pass
- Documentation examples work

### Quality

- Code coverage 90% or higher
- Performance within acceptable range in benchmarks

### Usability

- Integration into existing tests requires only one line change
- Output is intuitive and easy to understand

## Risks and Countermeasures

| Risk | Countermeasure |
|------|----------------|
| AST analysis complexity | Gradually increase supported expressions |
| Performance degradation | Implement benchmarks early |
| Verbose output | Make detail level configurable |

## Next Steps

1. Start with Phase 1.1 minimal implementation
2. Write tests first for each step (TDD)
3. Continuously refactor
4. Collect user feedback early
