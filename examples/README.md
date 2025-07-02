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

### [Advanced Usage](./advanced/)

Complex expressions and advanced patterns:

- Nested expressions
- Method calls
- Struct field access
- Complex logical combinations

### [Integration Examples](./integration/)

Real-world usage scenarios:

- Unit testing patterns
- Integration with existing test suites
- Performance considerations
- Best practices

## Running Examples

Each example can be run independently:

```bash
# Run basic examples
cd examples/basic && go run main.go

# Run advanced examples  
cd examples/advanced && go run main.go

# Run integration examples
cd examples/integration && go run main.go
```

## Testing Examples

Examples include their own tests:

```bash
# Test all examples
go test ./examples/...

# Test specific example
cd examples/basic && go test
```
