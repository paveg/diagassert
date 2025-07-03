# diagassert Examples

This directory contains practical examples demonstrating different usage
patterns of diagassert.

## Examples

### [Basic Usage](./basic/)

Simple usage patterns showing the core functionality:

- Basic assertions with different expression types
- Comparison operations
- String operations
- Boolean logic

### Advanced Usage (Coming Soon)

Complex expressions and advanced patterns will be added as Phase 3 development progresses:

- Nested expressions with real variable values
- Method calls with result display
- Struct field access with actual values
- Complex logical combinations

### Integration Examples (Planned)

Real-world usage scenarios to be added:

- Unit testing patterns
- Integration with existing test suites
- Performance considerations
- Best practices

## Running Examples

Currently available examples:

```bash
# Run basic examples
cd examples/basic && go run main.go
```

Future examples will be added as development progresses.

## Testing Examples

Examples include their own tests:

```bash
# Test all examples
go test ./examples/...

# Test basic example
cd examples/basic && go test
```
