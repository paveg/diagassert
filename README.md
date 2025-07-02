# diagassert

**No API is the Best API** - A human and AI-friendly assertion library for Go testing inspired by power-assert.

## Philosophy: Zero Learning Curve

diagassert has just **2 functions** - use any Go expression directly without learning dozens of APIs. See **[Philosophy](./docs/philosophy.md)** for details.

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

### Configuration (Environment Variables)

- `DIAGASSERT_MACHINE_READABLE`: "true" (default) | "false"

## Usage Examples

```go
import (
    "strings"
    "testing"
    "github.com/paveg/diagassert"
)

func TestExample(t *testing.T) {
    user := User{Name: "Alice", Age: 16}
    
    // Just write any Go expression!
    diagassert.Assert(t, user.Age >= 18)
    diagassert.Assert(t, user.HasLicense())
    diagassert.Assert(t, user.Age >= 18 && user.HasLicense())
    
    // Works with any expression
    diagassert.Assert(t, strings.Contains(user.Name, "A"))
    diagassert.Assert(t, len(user.Name) > 0)
}
```

## Output Example

When assertions fail, you get detailed diagnostic information:

```
ASSERTION FAILED at user_test.go:42
Expression: user.Age >= 18 && user.HasLicense()
Result: false

[MACHINE_READABLE_START]
EXPR: user.Age >= 18 && user.HasLicense()
RESULT: false
[MACHINE_READABLE_END]
```

## Features

- **Zero learning curve**: Just use Go expressions directly
- **Power-assert philosophy**: "No API is the best API"
- **Machine-readable output**: Perfect for AI tools and CI/CD
- **Environment-based configuration**: No API bloat
- **Standard library compatible**: Works with Go's testing package

## Roadmap

- **Phase 1** ✅: Basic expression extraction and display
- **Phase 2**: Variable value display and evaluation trees
- **Phase 3**: Advanced expression support (methods, fields)
- **Phase 4**: Enhanced machine-readable output for AI tools

## Documentation

- **[Documentation Index](./docs/)** - Complete documentation overview
- **[Examples](./examples/)** - Practical usage examples and demonstrations
- **[Philosophy](./docs/philosophy.md)** - Design principles and "No API" approach
- **[API Reference](./doc.go)** - Package-level documentation

## License

MIT License