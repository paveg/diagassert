# diagassert Documentation

This directory contains comprehensive documentation for the diagassert project.

## Documentation Files

### Core Documentation

- **[Philosophy](./philosophy.md)** - Design principles and the "No API is the
  best API" approach
- **[TDD Plan](./tdd-plan.md)** - Test-driven development roadmap and
  implementation phases

### Additional Resources

- **[Examples](../examples/)** - Practical usage examples and demonstrations
- **[API Reference](../doc.go)** - Package-level API documentation
- **[Source Code](../)** - Main source code and implementation

## Quick Start

For a quick overview of diagassert:

1. **Philosophy**: Read [philosophy.md](./philosophy.md) to understand the "Zero
   Learning Curve" approach
2. **Examples**: Check [../examples/basic/](../examples/basic/) for hands-on
   demonstrations
3. **Implementation**: See [tdd-plan.md](./tdd-plan.md) for development roadmap

## Project Status

- ✅ **Phase 1**: Basic expression extraction and display
- ✅ **Phase 2**: Expression evaluation and tree building
- 🔄 **Phase 3**: Advanced expression support (In Progress)
  - ✅ AST parsing and tree construction
  - ✅ Expression structure analysis  
  - ❌ Runtime variable value extraction (placeholder only)
  - ❌ Struct field value display
  - ❌ Method call result display
- ✅ **Phase 4**: Enhanced machine-readable output for AI tools
- ✅ **Phase 5**: Value capture API (`V()`, `Values{}`)
- 🔄 **Phase 6**: Output format extensions (Planned)
- 🔄 **Phase 7**: Performance optimizations (Planned)

## Contributing

See the [TDD Plan](./tdd-plan.md) for implementation roadmap and development
guidelines.
