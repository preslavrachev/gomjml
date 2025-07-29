package mjml

import (
	"fmt"
	"strings"
	"testing"
)

// generateMJMLTemplate creates a dynamic MJML template with specified number of sections
func generateMJMLTemplate(sections int) string {
	var builder strings.Builder

	// Start with basic MJML structure
	builder.WriteString(`<mjml>
  <mj-head>
    <mj-title>Benchmark Test Email</mj-title>
    <mj-font name="Roboto" href="https://fonts.googleapis.com/css?family=Roboto:300,400,500,700" />
    <mj-style>
      .custom-class { color: #e74c3c; }
      .highlight { background-color: #f39c12; }
    </mj-style>
  </mj-head>
  <mj-body background-color="#f4f4f4">
`)

	// Generate dynamic sections with columns and content
	for i := 0; i < sections; i++ {
		builder.WriteString(fmt.Sprintf(`    <mj-section background-color="#ffffff" padding="20px">
      <mj-column width="50%%">
        <mj-text font-size="16px" color="#333333" font-family="Roboto, Arial, sans-serif">
          <h2>Section %d - Column 1</h2>
          <p>This is dynamically generated content for section %d. It includes various MJML components to simulate a real email template with rich content and styling.</p>
        </mj-text>
        <mj-button background-color="#e74c3c" color="white" href="https://example.com/section-%d">
          Learn More %d
        </mj-button>
      </mj-column>
      <mj-column width="50%%">
        <mj-image src="https://via.placeholder.com/300x200?text=Image+%d" alt="Section %d Image" />
        <mj-text font-size="14px" color="#666666" align="center">
          <p class="custom-class">Featured content for section %d with custom styling and multiple components.</p>
        </mj-text>
      </mj-column>
    </mj-section>
`, i+1, i+1, i+1, i+1, i+1, i+1, i+1))

		// Add a divider between sections (except last one)
		if i < sections-1 {
			builder.WriteString(`    <mj-section>
      <mj-column>
        <mj-divider border-color="#e0e0e0" border-width="1px" />
      </mj-column>
    </mj-section>
`)
		}
	}

	// Add footer section
	builder.WriteString(`    <mj-section background-color="#34495e" padding="20px">
      <mj-column>
        <mj-text color="white" align="center" font-size="14px">
          <p>© 2024 Benchmark Test. This email was generated with <span class="highlight">gomjml</span>.</p>
          <p>Total sections: ` + fmt.Sprintf("%d", sections) + `</p>
        </mj-text>
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>`)

	return builder.String()
}

// BenchmarkMJMLRender_Small benchmarks rendering with 10 sections
func BenchmarkMJMLRender_Small(b *testing.B) {
	template := generateMJMLTemplate(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Render(template)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkMJMLRender_Medium benchmarks rendering with 100 sections
func BenchmarkMJMLRender_Medium(b *testing.B) {
	template := generateMJMLTemplate(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Render(template)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkMJMLRender_Large benchmarks rendering with 1000 sections
func BenchmarkMJMLRender_Large(b *testing.B) {
	template := generateMJMLTemplate(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Render(template)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkMJMLRender_SmallMemory benchmarks memory allocations with 10 sections
func BenchmarkMJMLRender_SmallMemory(b *testing.B) {
	template := generateMJMLTemplate(10)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Render(template)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkMJMLRender_MediumMemory benchmarks memory allocations with 100 sections
func BenchmarkMJMLRender_MediumMemory(b *testing.B) {
	template := generateMJMLTemplate(100)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Render(template)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkMJMLRender_LargeMemory benchmarks memory allocations with 1000 sections
func BenchmarkMJMLRender_LargeMemory(b *testing.B) {
	template := generateMJMLTemplate(1000)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Render(template)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkMJMLParsing_Only benchmarks just the parsing phase
func BenchmarkMJMLParsing_Only(b *testing.B) {
	template := generateMJMLTemplate(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseMJML(template)
		if err != nil {
			b.Fatalf("ParseMJML failed: %v", err)
		}
	}
}

// BenchmarkMJMLComponentCreation benchmarks AST to component tree conversion
func BenchmarkMJMLComponentCreation(b *testing.B) {
	template := generateMJMLTemplate(100)
	ast, err := ParseMJML(template)
	if err != nil {
		b.Fatalf("ParseMJML failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewFromAST(ast)
		if err != nil {
			b.Fatalf("NewFromAST failed: %v", err)
		}
	}
}

// BenchmarkMJMLFullPipeline benchmarks the complete rendering pipeline step by step
func BenchmarkMJMLFullPipeline(b *testing.B) {
	template := generateMJMLTemplate(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Parse MJML to AST
		ast, err := ParseMJML(template)
		if err != nil {
			b.Fatalf("ParseMJML failed: %v", err)
		}

		// Convert AST to component tree
		component, err := NewFromAST(ast)
		if err != nil {
			b.Fatalf("NewFromAST failed: %v", err)
		}

		// Render to HTML
		_, err = component.Render()
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkMJMLTemplateGeneration benchmarks the template generation itself
func BenchmarkMJMLTemplateGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateMJMLTemplate(100)
	}
}
