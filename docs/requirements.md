# diagassert - TDD Development Plan and Requirements Definition

## Overview

diagassert is an assertion library for Go testing that is friendly to both humans and AI tools. It combines the visual clarity of power-assert with machine readability.

## Requirements Definition

### Functional Requirements

#### 1. Core Features

- **Basic Assertion**: `Assert(t, expr)` evaluates expressions and outputs detailed information on failure
- **Immediate Failure Version**: `Require(t, expr)` terminates test immediately on failure
- **Custom Messages**: `Assert(t, expr, "custom message", args...)`

#### 2. Output Formats

- **Hybrid Output (Default)**: Human-friendly structured text + machine-readable metadata
- **Human-only**: Structured text only
- **Markdown Format**: Optimized for CI/CD display
- **JSON Format**: Complete machine-readable output

#### 3. Expression Evaluation Support

- **Comparison Operators**: `==`, `!=`, `<`, `>`, `<=`, `>=`
- **Logical Operators**: `&&`, `||`, `!`
- **Field Access**: `user.Name`
- **Method Calls**: `user.HasPermission()`
- **Array/Slice Access**: `items[0]`
- **Map Access**: `config["key"]`

#### 4. Context Features

- **Additional Context Information**: `WithContext(key, value)`
- **Variable History Tracking**: `WithVariableHistory(varName)`
- **Snapshots**: `WithSnapshot(varName)`

### Non-Functional Requirements

#### Performance

- Execution time within 10x of standard assertions
- Minimize memory usage

#### Compatibility

- Go 1.18 and above
- Full compatibility with standard testing package
- Easy integration into existing test code

#### Machine Readability

- AI tools can accurately understand failure causes
- Structured evaluation trees
- Expression tracking with unique IDs

## Design Details

### Package Structure

```
diagassert/
├── assert.go          # Main assertion functions
├── assert_test.go     # Tests
├── formatter.go       # Output formatter
├── formatter_test.go
├── evaluator.go       # Expression evaluation engine
├── evaluator_test.go
├── parser.go          # AST analysis
├── parser_test.go
├── types.go           # Type definitions
└── examples/          # Usage examples
```