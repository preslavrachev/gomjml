# gomjml - Native Go MJML Compiler

A native Go implementation of the MJML email framework, providing fast compilation of [MJML](https://mjml.io/) markup to responsive HTML. This implementation has been inspired by and tested against [MRML](https://github.com/jdrouet/mrml), the Rust implementation of MJML.

> **Full Disclaimer**: This project has been created in some cooperation with [Claude Code](https://www.anthropic.com/claude-code). I wouldn't have been able to achieve such a feat without Claude's help in turning my bizarre requirements into Go code. Still, it wasn't all smooth sailing. While Claude was able to generate a plausible MVP relatively quickly, bringing it somethign even remotely usable took a lot more human guidance, goign back and forth, throwing away a bunch of code and starting over. There's lots I have learned in the process, and I will soon write a series of blog posts addressing my experience.

## 🚀 Features

- **NOT Production Ready Yet!!!**: The current code implementation is about 80% feature-complete with MRML (the Rust implementation of MJML). When it is done, it will be a professionally structured Go library with clean package separation
- **Email Compatible**: Generates HTML that works across email clients (Outlook, Gmail, Apple Mail, etc.)
- **Fast Performance**: Native Go performance, comparable to Rust MRML implementation
- **Complete Component System**: Support for essential MJML components with proper inheritance
- **CLI & Library**: Use as command-line tool or importable Go package
- **Tested Against MRML**: Integration tests validate output compatibility with reference implementation

## 📦 Installation

### Install CLI

```bash
# Clone and build
git clone https://github.com/preslavrachev/gomjml
cd go
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
- `-w, --watch`: Watch file for changes (placeholder)
- `--beautify`: Beautify HTML output (default: true)
- `--minify`: Minify HTML output (default: false)
- `--validation-level string`: Validation level - strict, soft, or skip (default: "soft")
- `--debug`: Include debug attributes for component traceability (default: false)

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

	// Method 2: Step-by-step processing
	ast, err := parser.ParseMJML(mjmlContent)
	if err != nil {
		log.Fatal("Parse error:", err)
	}

	component, err := mjml.NewFromAST(ast)
	if err != nil {
		log.Fatal("Component creation error:", err)
	}

	html, err = component.Render()
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

func (c *MJNewComponent) Render() (string, error) {
    // Implementation here
    // Use c.AddDebugAttribute(tag, "new") for debug traceability
    return "", nil
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

#### Delaying Component Implementation

If you need to register a component but won't implement its functionality right away, use the `NotImplementedError` pattern:

```go
func (c *MJNewComponent) Render() (string, error) {
    // TODO: Implement mj-new component functionality
    return "", &NotImplementedError{ComponentName: "mj-new"}
}
```

## 📋 Component Implementation Status

| Component | Status | Description | Notes |
|-----------|--------|-------------|-------|
| **Core Layout** | | | |
| `mjml` | ✅ **Implemented** | Root document container with DOCTYPE and HTML structure | |
| `mj-head` | ✅ **Implemented** | Document metadata container | |
| `mj-body` | ✅ **Implemented** | Email body container with responsive layout | |
| `mj-section` | ✅ **Implemented** | Layout sections with background support | |
| `mj-column` | ✅ **Implemented** | Responsive columns with automatic width calculation | |
| `mj-wrapper` | ✅ **Implemented** | Wrapper component with border, background-color, and padding support | |
| `mj-group` | ✅ **Implemented** | Group multiple columns in a section | |
| **Content Components** | | | |
| `mj-text` | ✅ **Implemented** | Text content with full styling support | |
| `mj-button` | ✅ **Implemented** | Email-safe buttons with customizable styling and links | |
| `mj-image` | ✅ **Implemented** | Responsive images with link wrapping and alt text | |
| `mj-divider` | ✅ **Implemented** | Visual separators and spacing elements | |
| `mj-social` | ✅ **Implemented** | Social media icons container | |
| `mj-social-element` | ✅ **Implemented** | Individual social media icons | |
| `mj-raw` | ✅ **Implemented** | Raw HTML content insertion | |
| **Head Components** | | | |
| `mj-title` | ✅ **Implemented** | Document title for email clients | |
| `mj-font` | ✅ **Implemented** | Custom font imports with Google Fonts support | |
| `mj-preview` | ✅ **Implemented** | Preview text for email clients | |
| `mj-style` | ✅ **Implemented** | Custom CSS styles | |
| `mj-attributes` | ✅ **Implemented** | Global attribute definitions | |
| `mj-all` | ✅ **Implemented** | Global attributes for all components | |
| **Interactive Components** | | | |
| `mj-accordion` | ❌ **Not Implemented** | Collapsible content sections | |
| `mj-accordion-text` | ❌ **Not Implemented** | Text content within accordion | |
| `mj-accordion-title` | ❌ **Not Implemented** | Title for accordion sections | |
| `mj-carousel` | ❌ **Not Implemented** | Image carousel component | |
| `mj-carousel-image` | ❌ **Not Implemented** | Images within carousel | |
| `mj-hero` | ❌ **Not Implemented** | Header/banner sections with background images | |
| `mj-navbar` | ❌ **Not Implemented** | Navigation bar component | |
| `mj-navbar-link` | ❌ **Not Implemented** | Navigation links within navbar | |
| `mj-spacer` | ❌ **Not Implemented** | Layout spacing control | |
| `mj-table` | ❌ **Not Implemented** | Email-safe table component | |

### Implementation Summary
- **✅ Implemented: 16 components** - Core layout, content, and head components are fully functional
- **❌ Not Implemented: 10 components** - Advanced interactive components return `NotImplementedError`
- **Total MJML Components: 26** - Covers all major MJML specification components

### Integration Test Status
Based on the integration test suite in `mjml/integration_test.go`, the implemented components are thoroughly tested against the MRML (Rust) reference implementation to ensure compatibility and correctness.

### Baseline Benchmark

The following benchmarks were run on a MacBook Pro M1 Pro with 16GB RAM, Go 1.21.4. As you can see, for a MVP, the performance is will eb quite sufficient for most use cases, however, there is something to be desired in terms of memory usage and allocations. I am hopign to make significant refactoring and improvement in future releases.

```bash
./bench.sh  # You can also add --markdown for a markdown table output
```

| Benchmark                          |  Time   | Memory  | Allocs |
| :--------------------------------- | :-----: | :-----: | :----: |
| BenchmarkMJMLRender_Small-8        | 0.48ms  | 0.93MB  |  7.9K  |
| BenchmarkMJMLRender_Medium-8       | 5.83ms  | 9.75MB  | 76.5K  |
| BenchmarkMJMLRender_Large-8        | 52.59ms | 98.51MB | 762.4K |
| BenchmarkMJMLRender_SmallMemory-8  | 0.49ms  | 0.93MB  |  7.9K  |
| BenchmarkMJMLRender_MediumMemory-8 | 5.85ms  | 9.75MB  | 76.5K  |
| BenchmarkMJMLRender_LargeMemory-8  | 52.37ms | 98.51MB | 762.4K |
| BenchmarkMJMLParsing_Only-8        | 1.75ms  | 0.71MB  | 19.3K  |
| BenchmarkMJMLComponentCreation-8   | 0.17ms  | 0.38MB  |  4.6K  |
| BenchmarkMJMLFullPipeline-8        | 5.50ms  | 9.75MB  | 76.5K  |
| BenchmarkMJMLTemplateGeneration-8  | 0.12ms  | 0.59MB  |  0.1K  |

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

### Email Client Compatibility

Generated HTML works across all major email clients:

- **Microsoft Outlook** (2007, 2010, 2013, 2016, 365)
- **Gmail** (Web, iOS, Android)
- **Apple Mail** (macOS, iOS)
- **Thunderbird**
- **Yahoo Mail**
- **Outlook.com / Hotmail**

### Email-Specific Features

- **MSO Conditional Comments**: Outlook-specific styling and layout fixes
- **CSS Inlining Ready**: Structure compatible with CSS inlining tools
- **Mobile Responsive**: Automatic mobile breakpoints and media queries
- **Web Font Support**: Google Fonts integration with fallbacks

## 🔗 Related Projects

- **[MJML](https://mjml.io/)** - Original JavaScript implementation and framework specification
- **[MRML](https://github.com/jdrouet/mrml)** - Rust implementation used as reference for testing
- **[MJML Documentation](https://documentation.mjml.io/)** - Official MJML component specification
- **[MJML Try It Live](https://mjml.io/try-it-live)** - Online MJML editor and tester
