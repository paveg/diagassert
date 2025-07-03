# diagassert

[![Test](https://github.com/paveg/diagassert/actions/workflows/test.yml/badge.svg)](https://github.com/paveg/diagassert/actions/workflows/test.yml)

**No API is the Best API** - A human and AI-friendly assertion library for Go
testing inspired by [power-assert](https://github.com/power-assert-js/power-assert).

## Philosophy: Zero Learning Curve

diagassert has just **2 functions** - use any Go expression directly without
learning dozens of APIs.

See **[Philosophy](./docs/philosophy.md)** for details.

## Installation

```bash
go get github.com/paveg/diagassert
```

## API Reference

### Core Functions

```go
// Assert evaluates any boolean expression
func Assert(t testing.TB, expr bool)

// Require is like Assert but stops test execution on failure
func Require(t testing.TB, expr bool)
```

### Enhanced API (Value Capture)

```go
// Capture individual values
diagassert.Assert(t, expr, diagassert.V("x", x))

// Capture multiple values
diagassert.Assert(t, expr, diagassert.Values{"x": x, "y": y})

// Mix values and custom messages
diagassert.Assert(t, expr, diagassert.V("x", x), "Custom message")
```

### Configuration (Environment Variables)

- `DIAGASSERT_MACHINE_READABLE`: "true" (default) | "false"
- `NO_COLOR`: Set to disable all colors (respects <https://no-color.org/>)
- `FORCE_COLOR`: Set to force enable colors even in non-TTY environments
- `DIAGASSERT_PIPE_COLORS`: "true" (default) | "false" - Enable per-value pipe coloring

## Usage Examples

```go
import (
    "strings"
    "testing"
    "github.com/paveg/diagassert"
)

func TestExample(t *testing.T) {
    x := 10
    y := 20
    
    // Just write any Go expression!
    diagassert.Assert(t, x > y)
    diagassert.Assert(t, x + y == 30)
    diagassert.Assert(t, strings.Contains("hello", "lo"))
}
```

## Output Example

When assertions fail, you get detailed diagnostic information with visual hierarchy:

```text
ASSERTION FAILED at user_test.go:42
Expression: user.Age >= 18 && user.HasLicense()
Result: false

assert(user.Age >= 18 && user.HasLicense())
       |        |  |     |
       16       18 false false
                |        |
                false    |
                         |
                         false

[MACHINE_READABLE_START]
EXPR: user.Age >= 18 && user.HasLicense()
RESULT: false
[MACHINE_READABLE_END]
```

### Visual Features

- **Connecting pipes**: Visual connections between expressions and their values
- **Color-coded output**: Different colors for variables, operators, and results
- **Per-value pipe colors**: Each value gets unique pipe colors for better readability
- **Hierarchical layout**: Clear visual representation of expression evaluation flow

## Features

- **Zero learning curve**: Just use Go expressions directly
- **Power-assert philosophy**: "No API is the best API"
- **Visual hierarchy output**: Power-assert style visual representation with connecting pipes
- **Color support**: Automatic terminal color detection with per-value pipe coloring
- **Machine-readable output**: Perfect for AI tools and CI/CD
- **Environment-based configuration**: No API bloat
- **Standard library compatible**: Works with Go's testing package
- **Unicode support**: Handles international characters and wide characters correctly

## Roadmap

- **Phase 1** ✅: Basic expression extraction and display
- **Phase 2** ✅: Expression evaluation and tree building  
- **Phase 3** ✅: Advanced expression support
  - ✅ AST parsing and tree construction
  - ✅ Expression structure analysis  
  - ✅ Visual hierarchy representation with connecting pipes
  - ✅ Layer-based value positioning algorithm
  - ❌ Runtime variable value extraction (placeholder only)
  - ❌ Struct field value display
  - ❌ Method call result display
- **Phase 4** ✅: Enhanced machine-readable output for AI tools
- **Phase 5** ✅: Value capture API (`V()`, `Values{}`)
- **Phase 6** ✅: Visual enhancements
  - ✅ Color support with terminal detection
  - ✅ Per-value pipe coloring system
  - ✅ Unicode and wide character support
  - ✅ Environment variable configuration
- **Phase 7** 🔄: Performance optimizations (Planned)

## Documentation

- **[Documentation Index](./docs/)** - Complete documentation overview
- **[Examples](./examples/)** - Practical usage examples and demonstrations
- **[Philosophy](./docs/philosophy.md)** - Design principles and "No API"
  approach
- **[API Reference](./doc.go)** - Package-level documentation

## License

[MIT License](./LICENSE)
