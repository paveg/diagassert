# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is diagassert, a Go assertion library that implements the power-assert philosophy of "No API is the best API". The core innovation is providing detailed diagnostic output while maintaining an extremely simple 2-function API.

## Core Philosophy & Design Constraints

**Critical Design Principle**: This library deliberately contradicts traditional assertion library design. Instead of providing dozens of specialized assertion functions, it provides only 2 functions that work with any Go expression:

- `Assert(t, expr bool)` - fails test on false
- `Require(t, expr bool)` - terminates test immediately on false

**Never add more public API functions**. Any new functionality must be achieved through:

1. Environment variable configuration
2. Internal package enhancements  
3. Output format improvements

## Key Architecture

### AST-Based Expression Extraction

The core technical innovation is runtime AST parsing to extract and display the actual Go expressions that failed:

```go
// Instead of: assert.Greater(t, x, y)
diagassert.Assert(t, x > y)  // Shows "x > y" with actual values
```

This requires:

1. `runtime.Caller()` to get source file and line number
2. `go/parser` to parse the source file into AST
3. AST traversal to find the Assert/Require call and extract the expression argument

### Internal Package Structure

- **`internal/parser/`** - AST parsing to extract expressions from source code
- **`internal/formatter/`** - Output formatting including machine-readable sections
- **`internal/evaluator/`** - Future Phase 2+ functionality for variable value extraction

### Development Phases

**Phase 1 (Current)**: Basic expression extraction and display

- ✅ AST parsing to show "x > y" instead of just "false"
- ✅ Machine-readable output for AI tools
- ✅ Environment variable configuration

**Phase 2-4 (Planned)**: Advanced evaluation

- Variable value extraction: "x = 10, y = 20"
- Evaluation trees for complex expressions
- Enhanced machine-readable output

## Common Development Commands

```bash
# Run all tests
make test
# OR: go test -v -race ./...

# Run tests with coverage
make coverage

# Run single test
go test -v -run TestAssert_SimpleAPI

# Run tests for specific internal package
go test -v ./internal/parser

# Format code
make fmt

# Lint code
make lint

# Install development tools
make install-tools
```

## Testing Strategy

### Mock Testing Interface

Tests use a custom `mockT` type that implements the minimal `TestingT` interface rather than embedding `testing.T`. This allows capturing assertion output for verification.

### Expression Testing

Test cases verify that:

1. Expressions are correctly extracted from source code
2. Output format matches expected structure
3. Machine-readable sections are properly formatted
4. Environment variable configuration works

## Configuration

**Environment Variables** (the only form of configuration):

- `DIAGASSERT_MACHINE_READABLE`: "true" (default) | "false"

**Never add command-line flags or configuration files** - this would violate the simplicity principle.

## Key Implementation Details

### Runtime Caller Detection

Uses `runtime.Caller(2)` to get the file/line of the Assert/Require call, then parses that source file to extract the exact expression.

### AST Parsing Approach

1. Read source file at assertion location
2. Parse into AST using `go/parser`
3. Find CallExpr nodes matching Assert/Require
4. Extract the second argument (the expression) as string
5. Support both package-qualified (`diagassert.Assert`) and unqualified (`Assert`) calls

### Output Format Design

Balances human readability with machine parsability:

- Human section: file location, expression, result
- Machine section: structured data for AI tools (wrapped in `[MACHINE_READABLE_START/END]`)

## Development Guidelines

### Adding New Features

- Phase 2+ features go in `internal/evaluator/`
- Output enhancements go in `internal/formatter/`
- Parser improvements go in `internal/parser/`
- **Never add new public API functions**

### Testing New Features

- Write failing tests first (TDD approach)
- Test both human and machine-readable output
- Verify AST parsing edge cases
- Test environment variable interactions

### Performance Considerations

- AST parsing only happens on assertion failure
- File reading is cached within single assertion
- Detailed evaluation (Phase 2+) should be lazy

## Anti-Patterns to Avoid

1. **Adding specialized assertion functions** - contradicts core philosophy
2. **Complex configuration APIs** - use environment variables only
3. **Breaking the 2-function public API** - this is the primary constraint
4. **Removing machine-readable output** - essential for AI tooling integration
