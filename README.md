# gomjml - Native Go MJML Compiler

A native Go implementation of the MJML email framework, providing fast compilation of [MJML](https://mjml.io/) markup to responsive HTML. This implementation has been inspired by and tested against [MRML](https://github.com/jdrouet/mrml), the Rust implementation of MJML. See some [performance benchmarks](docs/benchmarks.md) for detailed comparison with other MJML implementations. 

![status](https://img.shields.io/badge/status-in_active_development-blueviolet)
![Tests](https://github.com/preslavrachev/gomjml/actions/workflows/test.yml/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/preslavrachev/gomjml)

> **Full Disclosure**: This project has been created in cooperation with [Claude Code](https://www.anthropic.com/claude-code). I wouldn't have been able to achieve such a feat without Claude's help in turning my bizarre requirements into Go code. Still, it wasn't all smooth sailing. While Claude was able to generate a plausible MVP relatively quickly, bringing it something even remotely usable took a lot more human guidance, going back and forth, throwing away a bunch of code and starting over. There's lots I have learned in the process, and I will soon write a series of blog posts addressing my experience.
>
> ![](https://img.shields.io/badge/Claude-D97757?style=for-the-badge&logo=claude&logoColor=white)

## 🚀 Features

- **Complete MJML Implementation**: 100% feature-complete with all 26 MJML components implemented and tested against MRML (the Rust implementation of MJML). A well-structured Go library with clean package separation
- **Enhanced Email Compatibility**: Generates HTML that works reliably across all email clients with robust Microsoft Outlook support and VML background rendering for legacy versions
- **Fast Performance**: Native Go performance, comparable to Rust MRML implementation
- **Optional AST Caching**: Opt-in template caching for speedup on repeated renders
- **Complete Component System**: Support for essential MJML components with proper inheritance
- **CLI & Library**: Use as command-line tool or importable Go package
- **Tested Against MRML**: Integration tests validate output compatibility with reference implementation

## 📦 Installation

### Install CLI

```bash
# Clone and build
git clone https://github.com/preslavrachev/gomjml
cd gomjml
go build -o bin/gomjml ./cmd/gomjml

# Add to PATH (optional)
export PATH=$PATH:$(pwd)/bin
```

### Install as Go Package

```bash
# Import as library
go get github.com/preslavrachev/gomjml
```

## 🔧 Usage

### Command Line Interface

The CLI provides a structured command system with individual commands:

```bash
# Basic compilation
./bin/gomjml compile input.mjml -o output.html

# Output to stdout
./bin/gomjml compile input.mjml -s

# Include debug attributes for component traceability
./bin/gomjml compile input.mjml -s --debug

# Enable caching for better performance on repeated renders
./bin/gomjml compile input.mjml -o output.html --cache

# Configure cache with custom TTL
./bin/gomjml compile input.mjml -o output.html --cache --cache-ttl=10m

# Run test suite
./bin/gomjml test

# Get help
./bin/gomjml --help
./bin/gomjml compile --help
```

#### CLI Commands

- **`compile [input]`** - Compile MJML to HTML (main command)
- **`test`** - Run test suite against MRML reference implementation
- **`help`** - Show help information

#### Compile Command Options

- `-o, --output string`: Output file path
- `-s, --stdout`: Output to stdout  
- `--debug`: Include debug attributes for component traceability (default: false)
- `--cache`: Enable AST caching for performance (default: false)
- `--cache-ttl`: Cache TTL duration (default: 5m)
- `--cache-cleanup-interval`: Cache cleanup interval (default: `cache-ttl/2`)

### Go Package API

The implementation provides clean, importable packages:

```go
package main

import (
	"fmt"
	"log"

	"github.com/preslavrachev/gomjml/mjml"
	"github.com/preslavrachev/gomjml/parser"
)

func main() {
	mjmlContent := `<mjml>
        <mj-head>
          <mj-title>My Newsletter</mj-title>
        </mj-head>
        <mj-body>
          <mj-section>
            <mj-column>
              <mj-text>Hello World!</mj-text>
              <mj-button href="https://example.com">Click Me</mj-button>
            </mj-column>
          </mj-section>
        </mj-body>
      </mjml>`

	// Method 1: Direct rendering (recommended)
	html, err := mjml.Render(mjmlContent)
	if err != nil {
		log.Fatal("Render error:", err)
	}
	fmt.Println(html)

	// Method 1b: Direct rendering with debug attributes
	htmlWithDebug, err := mjml.Render(mjmlContent, mjml.WithDebugTags(true))
	if err != nil {
		log.Fatal("Render error:", err)
	}
	fmt.Println(htmlWithDebug) // Includes data-mj-debug-* attributes

	// Method 1c: Enable caching for performance (opt-in feature)
	htmlWithCache, err := mjml.Render(mjmlContent, mjml.WithCache())
	if err != nil {
		log.Fatal("Render error:", err)
	}
	fmt.Println(htmlWithCache) // Uses cached AST if available

	// For long-running applications, configure cache TTL before first use
	mjml.SetASTCacheTTLOnce(10 * time.Minute)
	
	// For graceful shutdown in long-running applications (optional)
	// Not needed for CLI tools or short-lived processes
	defer mjml.StopASTCacheCleanup()

	// Method 2: Step-by-step processing
	ast, err := parser.ParseMJML(mjmlContent)
	if err != nil {
		log.Fatal("Parse error:", err)
	}

	component, err := mjml.NewFromAST(ast)
	if err != nil {
		log.Fatal("Component creation error:", err)
	}

	html, err = mjml.RenderComponentString(component)
	if err != nil {
		log.Fatal("Render error:", err)
	}
	fmt.Println(html)
}
```

### Adding New Components

While it is not recommended to do so, because it will break the compatibility with the MJML specification, you can fork the repository and add new components by following these steps:

```go
// 1. Create component file in mjml/components/
package components

import (
    "io"
    "strings"
    
    "github.com/preslavrachev/gomjml/mjml/options"
    "github.com/preslavrachev/gomjml/parser"
)

type MJNewComponent struct {
    *BaseComponent
}

func NewMJNewComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJNewComponent {
    return &MJNewComponent{
        BaseComponent: NewBaseComponent(node, opts),
    }
}

// Note: RenderString() is no longer part of the Component interface
// Use mjml.RenderComponentString(component) helper function instead

func (c *MJNewComponent) Render(w io.Writer) error {
    // Implementation here - write HTML directly to Writer
    // Use c.AddDebugAttribute(tag, "new") for debug traceability
    
    // Example implementation:
    // if _, err := w.Write([]byte("<div>Hello World</div>")); err != nil {
    //     return err
    // }
    return nil
}

func (c *MJNewComponent) GetTagName() string {
    return "mj-new"
}

// 2. Add to component factory in mjml/component.go
case "mj-new":
    return components.NewMJNewComponent(node, opts), nil

// 3. Add test cases in mjml/integration_test.go
// 4. Update README.md documentation
```

#### Component Interface Requirements

All MJML components must implement the `Component` interface, which requires:

- **`Render(w io.Writer) error`**: Primary rendering method that writes HTML directly to a Writer for optimal performance
- **`GetTagName() string`**: Returns the component's MJML tag name

For string-based rendering, use the helper function `mjml.RenderComponentString(component)` instead of a component method.

#### Delaying Component Implementation

If you need to register a component but won't implement its functionality right away, use the `NotImplementedError` pattern:

```go
func (c *MJNewComponent) Render(w io.Writer) error {
    // TODO: Implement mj-new component functionality
    return &NotImplementedError{ComponentName: "mj-new"}
}

func (c *MJNewComponent) GetTagName() string {
    return "mj-new"
}
```

## 📋 Component Implementation Status

| Component | Status | Description |
|-----------|--------|-------------|
| **Core Layout** | | |
| `mjml` | ✅ **Implemented** | Root document container with DOCTYPE and HTML structure |
| `mj-head` | ✅ **Implemented** | Document metadata container |
| `mj-body` | ✅ **Implemented** | Email body container with responsive layout |
| `mj-section` | ✅ **Implemented** | Layout sections with background support |
| `mj-column` | ✅ **Implemented** | Responsive columns with automatic width calculation |
| `mj-wrapper` | ✅ **Implemented** | Wrapper component with border, background-color, and padding support |
| `mj-group` | ✅ **Implemented** | Group multiple columns in a section |
| **Content Components** | | |
| `mj-text` | ✅ **Implemented** | Text content with full styling support |
| `mj-button` | ✅ **Implemented** | Email-safe buttons with customizable styling and links |
| `mj-image` | ✅ **Implemented** | Responsive images with link wrapping and alt text |
| `mj-divider` | ✅ **Implemented** | Visual separators and spacing elements |
| `mj-social` | ✅ **Implemented** | Social media icons container |
| `mj-social-element` | ✅ **Implemented** | Individual social media icons |
| `mj-navbar` | ✅ **Implemented** | Navigation bar component |
| `mj-navbar-link` | ✅ **Implemented** | Navigation links within navbar |
| `mj-raw` | ✅ **Implemented** | Raw HTML content insertion |
| **Head Components** | | |
| `mj-title` | ✅ **Implemented** | Document title for email clients |
| `mj-font` | ✅ **Implemented** | Custom font imports with Google Fonts support |
| `mj-preview` | ✅ **Implemented** | Preview text for email clients |
| `mj-style` | ✅ **Implemented** | Custom CSS styles |
| `mj-attributes` | ✅ **Implemented** | Global attribute definitions |
| `mj-all` | ✅ **Implemented** | Global attributes for all components |
| **Other Components** | | |
| `mj-accordion` | ✅ **Implemented** | Collapsible content sections |
| `mj-accordion-text` | ✅ **Implemented** | Text content within accordion |
| `mj-accordion-title` | ✅ **Implemented** | Title for accordion sections |
| `mj-carousel` | ✅ **Implemented** | Interactive image carousel component |
| `mj-carousel-image` | ✅ **Implemented** | Images within carousel |
| `mj-hero` | ✅ **Implemented** | Header/banner sections with background images |
| `mj-spacer` | ✅ **Implemented** | Layout spacing control |
| `mj-table` | ✅ **Implemented** | Email-safe table component with border and styling support |

### Implementation Summary
- **✅ Implemented: 26 components** - All essential layout, content, head components, accordion, navbar, hero, spacer, table, and carousel components with enhanced rendering robustness
- **❌ Not Implemented: 0 components** - Full MJML specification coverage achieved
- **Total MJML Components: 26** - Complete coverage of all major MJML specification components

### Integration Test Status
Based on the integration test suite in `mjml/integration_test.go`, the implemented components are thoroughly tested against the MRML (Rust) reference implementation to ensure compatibility and correctness.

### Performance Benchmarks

| Benchmark                                        |  Time   | Memory  | Allocs |
| :----------------------------------------------- | :-----: | :-----: | :----: |
| BenchmarkMJMLRender_10_Sections-8                | 0.48ms  | 0.60MB  |  4.2K  |
| BenchmarkMJMLRender_10_Sections_Cache-8          | 0.23ms  | 0.47MB  |  2.0K  |
| BenchmarkMJMLRender_100_Sections-8               | 5.21ms  | 6.26MB  | 39.6K  |
| BenchmarkMJMLRender_100_Sections_Cache-8         | 2.72ms  | 5.03MB  | 19.4K  |
| BenchmarkMJMLRender_1000_Sections-8              | 47.47ms | 63.78MB | 393.4K |
| BenchmarkMJMLRender_1000_Sections_Cache-8        | 23.83ms | 50.44MB | 192.2K |
| BenchmarkMJMLRender_10_Sections_Memory-8         | 0.54ms  | 0.60MB  |  4.2K  |
| BenchmarkMJMLRender_10_Sections_Memory_Cache-8   | 0.27ms  | 0.47MB  |  2.0K  |
| BenchmarkMJMLRender_100_Sections_Memory-8        | 5.06ms  | 6.20MB  | 39.6K  |
| BenchmarkMJMLRender_100_Sections_Memory_Cache-8  | 2.56ms  | 4.97MB  | 19.3K  |
| BenchmarkMJMLRender_1000_Sections_Memory-8       | 48.59ms | 63.81MB | 393.4K |
| BenchmarkMJMLRender_1000_Sections_Memory_Cache-8 | 23.96ms | 50.44MB | 192.2K |
| BenchmarkMJMLRender_100_Sections_Writer-8        | 2.35ms  | 4.53MB  | 15.0K  |

For comprehensive performance analysis including comparisons with other MJML implementations, see our dedicated [performance benchmarks documentation](docs/benchmarks.md).

```bash
# Run comparative benchmarks
./bench-austin.sh --markdown

# Run internal Go benchmarks
./bench.sh
```

## 🏗️ Architecture

The Go implementation follows a clean, modular architecture inspired by Go best practices:

### Project Structure

```
go/
├── cmd/gomjml/              # CLI application
│   ├── main.go             # Minimal entry point
│   └── command/            # Individual CLI commands
│       ├── root.go         # Root command setup
│       ├── compile.go      # MJML compilation command
│       └── test.go         # Test runner command
│
├── mjml/                   # Core MJML library (importable)
│   ├── component.go        # Component factory and interfaces
│   ├── render.go          # Main rendering logic and MJMLComponent
│   ├── mjml_test.go       # Library unit tests
│   ├── integration_test.go # MRML comparison tests
│   │
│   ├── components/        # Individual component implementations
│   │   ├── base.go        # Shared Component interface and BaseComponent
│   │   ├── head.go        # mj-head, mj-title, mj-font components
│   │   ├── body.go        # mj-body component
│   │   ├── section.go     # mj-section component
│   │   ├── column.go      # mj-column component
│   │   ├── text.go        # mj-text component
│   │   ├── button.go      # mj-button component
│   │   └── image.go       # mj-image component
│   │
│   └── testdata/          # Test MJML files
│       ├── basic.mjml
│       ├── with-head.mjml
│       └── complex-layout.mjml
│
└── parser/                # MJML parsing package (importable)
    ├── parser.go          # XML parsing logic with MJMLNode AST
    └── parser_test.go     # Parser unit tests
```

### Processing Pipeline

```
MJML Input → XML Parser → AST → Component Tree → HTML Output
                ↓              ↓           ↓
            Validation    Attribute     CSS Generation
                         Processing    & Email Compatibility
```

### Key Design Principles

1. **Package Separation**: Clean separation between CLI, library, and parsing concerns
2. **Component System**: Consistent Component interface with embedded BaseComponent
3. **Email Compatibility**: MSO/Outlook conditional comments and email-safe CSS
4. **Responsive Design**: Mobile-first CSS with media queries
5. **Testing Strategy**: Direct library testing without subprocess dependencies

## 🧪 Testing & MRML Compatibility

### What is MRML?

[MRML](https://github.com/jdrouet/mrml) is a Rust implementation of the MJML email framework that provides a fast, native alternative to the original JavaScript implementation. This Go implementation uses MRML as its reference for testing compatibility and correctness.

**Why did I choose MRML as a reference, rather than the default MJML compiler?**
- **Performance**: Native Rust performance comparable to our Go implementation
- **Compatibility**: Produces the same MJML-compliant HTML output as the JavaScript version
- **Reliability**: Well-tested, production-ready implementation
- **Accessibility**: Already installed and working in our development environment

### Test Suite

The comprehensive test suite validates output against MRML:

```bash
# Run all tests via CLI
./bin/gomjml test

# Run with verbose output
./bin/gomjml test -v

# Run specific test pattern
./bin/gomjml test -pattern "basic"

# Direct Go testing
cd mjml && go test -v
```

## 📊 Performance & Compatibility

### Performance Characteristics

- **Fast Compilation**: Native Go performance, typically sub-millisecond for basic templates
- **Memory Efficient**: Minimal allocations during parsing and rendering
- **Scalable**: Handles complex MJML documents with multiple sections and components

### AST Caching (Opt-in Performance Feature)

**When to Enable Caching:**
- High-volume applications rendering the same templates repeatedly
- Web servers with template reuse patterns
- Batch processing where templates are rendered multiple times
- Applications where parsing time > rendering time

**Memory Management:**
- **Default TTL**: 5 minutes per cached template
- **Memory Usage**: ~5-50KB per cached template (varies by complexity)
- **Growth Pattern**: Cache grows between cleanup cycles, shrinks during cleanup
- **No Size Limits**: Monitor memory usage in production environments

**Thread Safety:**
- All cache operations are safe for concurrent use
- Singleflight pattern prevents duplicate parsing under high load
- Background cleanup runs automatically every 2.5 minutes (default)

**Configuration:**
```go
// Set cache TTL before first use (call only once)
mjml.SetASTCacheTTLOnce(10 * time.Minute)

// Set cleanup interval (call only once) 
mjml.SetASTCacheCleanupIntervalOnce(5 * time.Minute)

// For graceful shutdown in long-running applications (optional)
// Not needed for CLI tools or short-lived processes
defer mjml.StopASTCacheCleanup()
```

**When NOT to Use Caching:**
- Single-use template rendering
- Memory-constrained environments  
- Applications with constantly changing templates
- Short-lived processes where cache warmup overhead > benefits

### Email Client Compatibility

Generated HTML works across all major email clients:

- **Microsoft Outlook** (2007, 2010, 2013, 2016, 365)
- **Gmail** (Web, iOS, Android)
- **Apple Mail** (macOS, iOS)
- **Thunderbird**
- **Yahoo Mail**
- **Outlook.com / Hotmail**

### Email-Specific Features

- **Enhanced MSO Conditional Comments**: Comprehensive Outlook-specific styling and layout fixes
- **VML Background Support**: Legacy Outlook compatibility with Vector Markup Language backgrounds
- **CSS Inlining Ready**: Structure compatible with CSS inlining tools
- **Mobile Responsive**: Automatic mobile breakpoints and media queries
- **Web Font Support**: Google Fonts integration with fallbacks

## 🔗 Related Projects

- **[MJML](https://mjml.io/)** - Original JavaScript implementation and framework specification
- **[MRML](https://github.com/jdrouet/mrml)** - Rust implementation used as reference for testing
- **[MJML Documentation](https://documentation.mjml.io/)** - Official MJML component specification
- **[MJML Try It Live](https://mjml.io/try-it-live)** - Online MJML editor and tester
