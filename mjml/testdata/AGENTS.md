# MJML Test Data Reference

## Purpose

This directory contains reference test cases for validating the gomjml compiler against the official MJML specification. Each test case consists of:

- **Input**: `.mjml` file containing MJML markup
- **Expected Output**: `.html` file containing the reference HTML output

The HTML files were generated using the reference MJML compiler (MRML - Rust implementation) and serve as the ground truth for verifying that our Go implementation produces spec-compliant output.

## File Structure

```
testdata/
├── basic.mjml          # Input MJML template
├── basic.html          # Expected HTML output
├── mj-button.mjml      # Button component test
├── mj-button.html      # Expected button HTML
└── ...                 # Additional test cases
```

## Integration with Tests

These test files are consumed by `mjml/integration_test.go` in the `TestMJMLAgainstExpected` function. The test:

1. Reads each `.mjml` file from the test case list
2. Compiles it using our Go implementation (`mjml.Render()`)
3. Compares the output against the corresponding `.html` file
4. Reports comprehensive differences if outputs don't match

## Test Case Categories

### Core Components
- `mjml.mjml/html` - Root document structure
- `mj-body.mjml/html` - Body container with attributes
- `mj-section.mjml/html` - Section layouts and styling
- `mj-column.mjml/html` - Column layouts and responsive behavior
- `mj-wrapper.mjml/html` - Wrapper components for grouping

### Content Components
- `mj-text.mjml/html` - Text rendering and styling
- `mj-button.mjml/html` - Button components with various attributes
- `mj-image.mjml/html` - Image handling and responsive behavior
- `mj-divider.mjml/html` - Divider styling and positioning
- `mj-social.mjml/html` - Social media icons and links

### Advanced Components
- `mj-group.mjml/html` - Column grouping and alignment
- `mj-hero.mjml/html` - Hero sections with background images
- `mj-navbar.mjml/html` - Navigation components
- `mj-accordion.mjml/html` - Collapsible content sections
- `mj-table.mjml/html` - Email-safe table rendering

### Layout Tests
- `complex-layout.mjml/html` - Multi-section complex layouts
- `austin-*.mjml/html` - Real-world template examples
- `wrapper-*.mjml/html` - Various wrapper configurations

## Debugging Failed Tests

When tests fail, the integration test provides detailed analysis:

### 1. Use htmlcompare Utility
```bash
# From project root
./bin/htmlcompare basic --testdata-dir mjml/testdata

# From testdata directory
../../bin/htmlcompare basic
```

### 2. Check Debug Files
Failed tests write temporary files for manual inspection:
- `/tmp/expected_{testname}.html` - Reference output
- `/tmp/actual_{testname}.html` - Our implementation output

### 3. Common Difference Types
- **DOM Structure**: Element count or nesting differences
- **CSS Properties**: Style attribute ordering or values
- **MSO Conditionals**: Outlook-specific conditional comments
- **VML Attributes**: Vector Markup Language attributes for email clients
- **HTML Entities**: Encoding differences (&amp; vs &)

## Adding New Test Cases

1. **Create MJML file**: Add `{testname}.mjml` with your test case
2. **Generate reference**: Use MRML to create `{testname}.html`
3. **Add to test list**: Include in `testCases` array in `integration_test.go`
4. **Verify**: Run `go test -v ./mjml -run TestMJMLAgainstExpected`

## Reference Implementation

The expected HTML outputs were generated using MRML (Rust MJML implementation), which serves as the canonical reference for MJML specification compliance. This ensures our Go implementation maintains compatibility with the broader MJML ecosystem.

## Performance Testing

Some test cases are also used for benchmarking:
- `basic.mjml` - Simple performance baseline
- `complex-layout.mjml` - Complex template performance
- `austin-*.mjml` - Real-world template performance

Run benchmarks with: `./bench.sh`