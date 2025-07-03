# Per-Value Pipe Color System Implementation

## Overview

This document describes the implementation of a per-value pipe color system for the diagassert visual formatter. The system assigns unique colors to each value's pipes in deep expression hierarchies, making them more readable while maintaining the existing value coloring system.

## Design Decisions

### 1. Environment Variable Configuration

- **Variable**: `DIAGASSERT_PIPE_COLORS`
- **Default**: Enabled (only disabled when explicitly set to "false")
- **Rationale**: Follows the library's philosophy of minimal API with environment variable configuration

### 2. Color Palette Design

The system uses 8 carefully chosen colors that are:
- **Distinguishable**: Each color is visually distinct from others
- **Accessible**: Good contrast for readability
- **Different from existing colors**: Avoids conflicts with current color scheme

**Color Palette**:
1. Cyan (`color.FgCyan`)
2. Magenta (`color.FgMagenta`) 
3. Bright Green (`color.FgHiGreen`)
4. Bright Yellow (`color.FgHiYellow`)
5. Bright Blue (`color.FgHiBlue`)
6. Bright Magenta (`color.FgHiMagenta`)
7. Bright Cyan (`color.FgHiCyan`)
8. White (`color.FgWhite`)

### 3. Deterministic Color Assignment

- Uses a simple hash function based on expression text
- Ensures consistent color assignment across runs
- Hash formula: `hash = hash*31 + int(char) + position`
- Color index: `hash % palette_size`

### 4. Backward Compatibility

- Existing pipe coloring system remains unchanged when per-value colors are disabled
- Falls back gracefully to default gray pipes
- Maintains all existing color helper functions

## Implementation Details

### Core Components

#### 1. Extended ColorConfig Structure

```go
type ColorConfig struct {
    // Existing fields...
    PipeColorPalette  []*color.Color // Color palette for per-value pipes
    PipeColorsEnabled bool           // Enable per-value pipe colors
    // Existing fields...
}
```

#### 2. Key Functions

- `createPipeColorPalette()`: Creates the 8-color palette
- `assignPipeColor(expression string)`: Deterministic color assignment
- `getPipeColorForValue(ValuePosition)`: Gets color for specific value
- `simpleHash(string)`: Consistent hash function
- `colorizePerValuePipe(text, position)`: Applies per-value colors
- `colorizePerValuePipeLine(line, layerAssignment, layerIdx)`: Colors entire pipe lines

#### 3. Integration Points

- Updated `setupColorConfig()` to initialize per-value pipe colors
- Modified `buildPowerAssertTreeWithLayers()` to use new coloring system
- Enhanced `colorizePipeLine()` with per-value color support

### Algorithm Flow

1. **Configuration Phase**:
   - Check `DIAGASSERT_PIPE_COLORS` environment variable
   - Create color palette if colors are enabled
   - Initialize color configuration

2. **Color Assignment Phase**:
   - For each value position in the expression tree
   - Generate hash based on expression text
   - Map hash to color index in palette
   - Return corresponding color

3. **Rendering Phase**:
   - Build layer assignment with pipe positions
   - For each layer, map pipe positions to their values
   - Apply appropriate colors to each pipe character
   - Fall back to default pipe color when needed

### Performance Considerations

- Color assignment only happens during assertion failure (failure path)
- Hash calculation is O(n) where n is expression length
- Color palette lookup is O(1)
- Per-value pipe coloring adds minimal overhead to existing rendering

## Usage Examples

### Basic Usage (Default - Enabled)

```go
// Per-value pipe colors are enabled by default
formatter := NewVisualFormatter()
// Each value gets its own pipe color automatically
```

### Disable Per-Value Pipe Colors

```bash
export DIAGASSERT_PIPE_COLORS=false
```

```go
// Will use traditional single-color pipes
formatter := NewVisualFormatter()
```

### Test Color Consistency

```go
formatter1 := NewVisualFormatter()
formatter2 := NewVisualFormatter()

color1 := formatter1.assignPipeColor("x")
color2 := formatter2.assignPipeColor("x")
// color1 == color2 (consistent across instances)
```

## Visual Impact

### Before (Single Color Pipes)
```
  assert(x > y && z < w)
         |   |    |   |
         5   10   15  12
              |    |
              false false
                   |
                   false
```

### After (Per-Value Pipe Colors)
```
  assert(x > y && z < w)
         |   |    |   |     <- Each | has a different color
         5   10   15  12    <- Values keep existing colors
              |    |        <- Different colors for each pipe
              false false
                   |        <- Unique color for final result
                   false
```

## Testing Strategy

### Test Coverage

1. **Configuration Tests**:
   - Environment variable handling
   - Color palette creation
   - Default behavior

2. **Color Assignment Tests**:
   - Deterministic assignment
   - Hash consistency
   - Palette rotation

3. **Integration Tests**:
   - End-to-end formatting
   - Backward compatibility
   - FORCE_COLOR handling

4. **Visual Tests**:
   - ANSI code generation
   - Color consistency
   - Fallback behavior

### Test Files

- `pipe_colors_test.go`: Comprehensive test suite
- `pipe_colors_demo.go`: Visual demonstration
- Integration with existing color tests

## Compatibility

### Environment Variables
- Respects existing `NO_COLOR` and `FORCE_COLOR` variables
- New `DIAGASSERT_PIPE_COLORS` variable follows same patterns

### Existing Color System
- Value colors unchanged (blue for variables, red for false, etc.)
- Header colors unchanged (bold red)
- Operator colors unchanged (yellow)
- Only pipe colors get per-value treatment

### Terminal Support
- Uses fatih/color package for terminal detection
- Graceful fallback on terminals without color support
- ANSI code generation for FORCE_COLOR scenarios

## Future Enhancements

### Potential Improvements
1. **Custom Color Palettes**: Allow users to define their own color palettes
2. **Expression-Based Coloring**: Different color schemes for different expression types
3. **Color Intensity**: Use color intensity to indicate expression depth
4. **High Contrast Mode**: Alternative palette for accessibility

### Configuration Expansion
- `DIAGASSERT_PIPE_PALETTE`: Custom color specification
- `DIAGASSERT_PIPE_INTENSITY`: Enable depth-based intensity
- `DIAGASSERT_HIGH_CONTRAST`: Accessibility mode

## Conclusion

The per-value pipe color system enhances readability of complex expression hierarchies while maintaining the library's core philosophy of minimal API with environmental configuration. The implementation is backward compatible, performant, and provides a solid foundation for future enhancements.