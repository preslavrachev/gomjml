# gomjml - Native Go MJML Compiler

A native Go implementation of the MJML email framework, providing fast compilation of MJML markup to responsive HTML. This implementation has been inspired by and tested against [MRML](https://github.com/jdrouet/mrml), the Rust implementation of MJML.

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

## 📋 Supported Components

### Core Components ✅
- **`mjml`** - Root document container with DOCTYPE and HTML structure
- **`mj-head`** - Document metadata container  
- **`mj-body`** - Email body container with responsive layout
- **`mj-section`** - Layout sections with background color support
- **`mj-column`** - Responsive columns with automatic width calculation
- **`mj-text`** - Text content with full styling support (fonts, colors, alignment)

### Interactive Components ✅  
- **`mj-button`** - Email-safe buttons with customizable styling and links
- **`mj-image`** - Responsive images with link wrapping and alt text

### Head Components ✅
- **`mj-title`** - Document title for email clients
- **`mj-font`** - Custom font imports with Google Fonts support

### Advanced Components (Future Phases)
- **`mj-divider`** - Visual separators and spacing elements (planned)
- **`mj-spacer`** - Layout spacing control (planned)
- **`mj-navbar`** - Navigation components (planned)
- **`mj-hero`** - Header/banner sections (planned)

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
