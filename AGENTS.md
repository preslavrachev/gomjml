# gomjml - Native Go MJML Compiler

## Project Overview
A native Go implementation of the MJML email framework for compiling MJML markup to responsive HTML. 100% feature-complete with all 26 MJML components implemented and thoroughly tested against reference implementation.

## Current Development Status
**Branch**: `dev/switch-from-mrml-to-mjml` - Transitioning from MRML-based testing to native reference implementation

The project is currently transitioning away from using MRML (Rust MJML implementation) as the reference for testing and validation. This shift enables:
- **Independence**: No external dependencies for testing and validation
- **Native Control**: Full control over reference implementation and test generation
- **Performance**: Improved development workflow without external tool dependencies
- **Consistency**: Direct alignment with MJML specification rather than third-party interpretation

## Key Architecture
- **CLI Application**: `cmd/gomjml/` with compile and test commands
- **Core Library**: `mjml/` package (importable) with component system
- **Parser**: `parser/` package for XML/AST processing
- **Components**: Individual MJML component implementations in `mjml/components/`

## Build & Test Commands
```bash
# Build CLI (for debugging, always use debug build)
go build -tags debug -o bin/gomjml ./cmd/gomjml

# Build production version (no debug output)
go build -o bin/gomjml ./cmd/gomjml

# Run integration tests (against reference implementation)
./bin/gomjml test

# Run tests with verbose output
./bin/gomjml test -v

# Run tests matching pattern
./bin/gomjml test -pattern "basic"

# Run integration tests with debug build for detailed output
go test -tags debug -v ./mjml -run TestMJMLAgainstExpected

# Run benchmarks
./bench.sh

# Test case comparison with htmlcompare utility
# First build the utility:
go build -o bin/htmlcompare ./cmd/htmlcompare

# Option 1: Run from mjml/testdata/ (auto-detects location):
cd mjml/testdata
../../bin/htmlcompare basic                    # Compare basic.mjml vs basic.html
../../bin/htmlcompare basic --verbose          # Show more diff context

# Option 2: Run from project root (specify testdata directory):
./bin/htmlcompare basic --testdata-dir mjml/testdata         # Compare from root
./bin/htmlcompare mj-button-align --testdata-dir mjml/testdata
./bin/htmlcompare basic -v --testdata-dir mjml/testdata      # Verbose output

# The utility auto-detects project root, builds debug gomjml, and performs semantic HTML comparison

# Compile MJML to HTML
./bin/gomjml compile input.mjml -o output.html

# Compile with debug attributes for troubleshooting
./bin/gomjml compile input.mjml -o output.html --debug

# Debug compile with verbose logging (requires debug build)
./bin/gomjml-debug compile input.mjml -o output.html
```

## Development Guidelines

### Component Interface & Architecture
- **Component Interface**: All components implement `Render(w io.StringWriter) error` and `GetTagName() string`
- **Base Component**: Extend `*BaseComponent` which provides common functionality and attribute handling
- **Testing**: Integration tests validate against reference implementation
- **Performance Focus**: Recent commits show memory optimizations and performance improvements
- **Email Compatibility**: Generates MSO-compatible HTML for Outlook/Gmail/Apple Mail

### Component Implementation Standards

#### HTML Generation
- **Use HTMLTag Builder**: Use `html.NewHTMLTag()` for generating HTML elements instead of string concatenation
- **StringWriter Pattern**: All internal rendering methods must use `io.StringWriter` interface for performance
- **Constants Usage**: Use constants from `mjml/constants` package instead of hardcoded strings:
  - CSS properties: `constants.CSSFontSize`, `constants.CSSPadding`, etc.
  - HTML attributes: `constants.AttrClass`, `constants.AttrCellSpacing`, etc.
  - MJML attributes: `constants.MJMLFontFamily`, `constants.MJMLPadding`, etc.
  - Common values: `constants.VAlignMiddle`, `constants.AlignCenter`, etc.
  - Language/Direction: `constants.LangUndetermined`, `constants.DirAuto`, etc.
  - **🚨 Code Review Check**: Always look for magic strings that should be constants!

#### Code Structure
- **Font Tracking**: Always call `c.TrackFontFamily(value)` when processing font-family attributes
- **Attribute Handling**: 
  - Use `c.GetAttributeWithDefault(c, name)` directly instead of creating wrapper `getAttribute` methods
  - Only create custom attribute methods when implementing specific inheritance patterns (like accordion element)
  - For explicit-only attributes like font-family, use `c.Node.GetAttribute(name)` directly
- **Individual Padding**: Support individual padding properties (`padding-top`, `padding-bottom`, etc.) alongside general `padding`
- **MSO Compatibility**: Use MSO conditional comments for Outlook-specific elements: `<!--[if !mso | IE]><!--> ... <!--<![endif]-->`
- **CSS Class Support**: ALL components MUST support the `css-class` attribute using `c.BuildClassAttribute()` on their main container element

#### Example Pattern
```go
func (c *MyComponent) Render(w io.StringWriter) error {
    // Get attributes using constants
    padding := c.getAttribute(constants.MJMLPadding)
    fontFamily := c.getAttribute(constants.MJMLFontFamily)
    
    // Use HTMLTag builder
    tdTag := html.NewHTMLTag("td").
        AddStyle(constants.CSSPadding, padding).
        AddStyle(constants.CSSFontFamily, fontFamily)
    
    if err := tdTag.RenderOpen(w); err != nil {
        return err
    }
    
    // Render content...
    
    if _, err := w.WriteString("</td>"); err != nil {
        return err
    }
    return nil
}

func (c *MyComponent) GetTagName() string {
    return "mj-my-component"
}
```

## Implementation Status
✅ **Implemented (26 components)**: Core layout (mjml, mj-head, mj-body, mj-section, mj-column, mj-wrapper, mj-group), content (mj-text, mj-button, mj-image, mj-divider, mj-social*, mj-raw, mj-hero, mj-navbar, mj-spacer, mj-table, mj-carousel, mj-carousel-image), head components (mj-title, mj-font, mj-preview, mj-style, mj-attributes, mj-all), accordion components (mj-accordion, mj-accordion-element, mj-accordion-text, mj-accordion-title)

✅ **Complete MJML Coverage**: All 26 major MJML specification components implemented

## AST Caching (Performance Feature)

### Overview
The AST cache improves performance by storing parsed MJML templates in memory. This is an **opt-in feature** disabled by default.

### Usage
**CLI:**
```bash
# Enable caching
./bin/gomjml compile input.mjml --cache

# Configure cache TTL
./bin/gomjml compile input.mjml --cache --cache-ttl=10m
```

**Library:**
```go
// Enable caching for performance
html, err := mjml.Render(template, mjml.WithCache())

// Configure cache before first use (call only once)
mjml.SetASTCacheTTLOnce(10 * time.Minute)

// Graceful shutdown for long-running applications
defer mjml.StopASTCacheCleanup()
```

### Configuration
- **Default TTL**: 5 minutes
- **Default cleanup interval**: 2.5 minutes (TTL/2) 
- **Thread-safe**: All operations safe for concurrent usage
- **Memory growth**: No size limits - monitor memory usage in production

### When to Use Caching
- High-volume applications with template reuse
- Web servers rendering the same templates repeatedly  
- Batch processing with repeated templates
- Applications where parsing time > rendering time

### When NOT to Use Caching
- Single-use template rendering
- Memory-constrained environments
- Applications with constantly changing templates
- Short-lived processes where cache warmup overhead > benefits

### Memory Considerations
- **Cache growth**: Unbounded between cleanup cycles
- **Memory usage**: ~5-50KB per cached template (varies by complexity)
- **Cleanup strategy**: Background goroutine removes expired entries
- **Production monitoring**: Watch memory usage for cache size estimation

### Technical Implementation
- **Concurrency**: Uses singleflight pattern to prevent duplicate parsing under load
- **Storage**: `sync.Map` optimized for high read/low write workloads
- **Hashing**: DoS-resistant template hashing with `maphash` for fast cache keys
- **Expiration**: Fixed TTL with background cleanup (no LRU complexity)

## Recent Development Focus
- **MSO Compatibility Improvements**: Enhanced Outlook rendering with better conditional comments and section transitions
- **Test Stabilization**: Systematic testing and validation of components against reference implementation
- **Transition from MRML**: Moving away from MRML dependencies towards native reference implementation
- **Component Refinements**: Fine-tuning existing components for better email client compatibility
- **Testing Infrastructure**: Improved htmlcompare utility and test validation processes

## Additional Documentation Sources

### AGENTS.md Files
When working on this project, always search for `AGENTS.md` files that may contain valuable context:
- Check project root: `./AGENTS.md`
- Search nested folders: `find . -name "AGENTS*.md" -type f`
- Use glob patterns: `**/*AGENTS*.md`

These files often contain:
- Project documentation and reference information
- Test case documentation and validation procedures
- Implementation decisions and rationale
- Development guidelines and conventions
- Component-specific implementation notes