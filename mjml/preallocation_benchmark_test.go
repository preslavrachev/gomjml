package mjml

import (
	"os"
	"strings"
	"testing"
)

// PreallocationStrategy represents different buffer sizing strategies
type PreallocationStrategy struct {
	Name        string
	Calculator  func(mjmlSize int, componentCount int, nestingDepth int) int
	Description string
}

// Helper function to count components in MJML content
func countComponents(mjmlContent string) int {
	return strings.Count(mjmlContent, "<mj-")
}

// Helper function to estimate nesting depth
func estimateNestingDepth(mjmlContent string) int {
	maxDepth := 0
	currentDepth := 0

	lines := strings.Split(mjmlContent, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<mj-") && !strings.Contains(trimmed, "/>") && !strings.HasPrefix(trimmed, "</") {
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		} else if strings.HasPrefix(trimmed, "</mj-") {
			currentDepth--
		}
	}
	return maxDepth
}

// loadRealWorldTemplate loads the Austin layout template from testdata
func loadRealWorldTemplate() string {
	content, err := os.ReadFile("testdata/austin-layout-from-mjml-io.mjml")
	if err != nil {
		panic("Failed to load real-world template: " + err.Error())
	}
	return string(content)
}

// Pre-allocation strategies to test
var strategies = []PreallocationStrategy{
	{
		Name: "Current_4x",
		Calculator: func(mjmlSize, componentCount, nestingDepth int) int {
			return mjmlSize * 4
		},
		Description: "Current strategy: 4x MJML input size",
	},
	{
		Name: "Size_2x_Plus_ComponentFactor",
		Calculator: func(mjmlSize, componentCount, nestingDepth int) int {
			return mjmlSize*2 + componentCount*150 // ~150 bytes avg per component
		},
		Description: "2x input size + 150 bytes per component",
	},
	{
		Name: "Size_3x_Plus_ComponentFactor",
		Calculator: func(mjmlSize, componentCount, nestingDepth int) int {
			return mjmlSize*3 + componentCount*100 // ~100 bytes avg per component
		},
		Description: "3x input size + 100 bytes per component",
	},
	{
		Name: "ComponentBased_Heavy",
		Calculator: func(mjmlSize, componentCount, nestingDepth int) int {
			base := mjmlSize * 2
			componentFactor := componentCount * 200 // Heavy component weight
			nestingFactor := nestingDepth * 50      // Nesting adds wrapper overhead
			return base + componentFactor + nestingFactor
		},
		Description: "2x input + 200 bytes per component + 50 bytes per nesting level",
	},
	{
		Name: "ComponentBased_Light",
		Calculator: func(mjmlSize, componentCount, nestingDepth int) int {
			base := mjmlSize * 2
			componentFactor := componentCount * 120 // Light component weight
			nestingFactor := nestingDepth * 30      // Less nesting overhead
			return base + componentFactor + nestingFactor
		},
		Description: "2x input + 120 bytes per component + 30 bytes per nesting level",
	},
	{
		Name: "Adaptive_MSO",
		Calculator: func(mjmlSize, componentCount, nestingDepth int) int {
			base := mjmlSize * 3
			msoOverhead := mjmlSize / 3                      // ~30% MSO overhead
			componentFactor := componentCount * 130          // Medium component weight
			nestingBonus := nestingDepth * nestingDepth * 10 // Quadratic nesting penalty
			return base + msoOverhead + componentFactor + nestingBonus
		},
		Description: "3x input + 30% MSO overhead + 130 bytes per component + quadratic nesting",
	},
	{
		Name: "Conservative_5x",
		Calculator: func(mjmlSize, componentCount, nestingDepth int) int {
			return mjmlSize * 5
		},
		Description: "Conservative 5x MJML input size",
	},
	{
		Name: "Minimal_2x",
		Calculator: func(mjmlSize, componentCount, nestingDepth int) int {
			return mjmlSize * 2
		},
		Description: "Minimal 2x MJML input size",
	},
	{
		Name: "Dynamic_ScaledByComplexity",
		Calculator: func(mjmlSize, componentCount, nestingDepth int) int {
			// Scale multiplier based on complexity
			complexity := float64(componentCount) / float64(mjmlSize) * 1000 // components per 1000 chars
			if complexity > 10 {
				// Very dense template
				return mjmlSize*5 + componentCount*180
			} else if complexity > 5 {
				// Medium density
				return mjmlSize*4 + componentCount*140
			} else {
				// Light template
				return mjmlSize*3 + componentCount*100
			}
		},
		Description: "Dynamic scaling based on component density",
	},
}

// BenchmarkPreallocationStrategies_Simple tests different strategies on simple templates
func BenchmarkPreallocationStrategies_Simple(b *testing.B) {
	template := loadRealWorldTemplate() // Real Austin layout template
	mjmlSize := len(template)
	componentCount := countComponents(template)
	nestingDepth := estimateNestingDepth(template)

	for _, strategy := range strategies {
		b.Run(strategy.Name, func(b *testing.B) {
			bufferSize := strategy.Calculator(mjmlSize, componentCount, nestingDepth)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ast, err := ParseMJML(template)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
				component, err := NewFromAST(ast)
				if err != nil {
					b.Fatalf("Component creation failed: %v", err)
				}

				var buf strings.Builder
				buf.Grow(bufferSize) // Pre-allocate with strategy
				err = component.RenderHTML(&buf)
				if err != nil {
					b.Fatalf("RenderHTML failed: %v", err)
				}
				_ = buf.String() // Force evaluation
			}
		})
	}
}

// BenchmarkPreallocationStrategies_Medium tests different strategies on medium templates
func BenchmarkPreallocationStrategies_Medium(b *testing.B) {
	template := loadRealWorldTemplate() // Real Austin layout template
	mjmlSize := len(template)
	componentCount := countComponents(template)
	nestingDepth := estimateNestingDepth(template)

	for _, strategy := range strategies {
		b.Run(strategy.Name, func(b *testing.B) {
			bufferSize := strategy.Calculator(mjmlSize, componentCount, nestingDepth)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ast, err := ParseMJML(template)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
				component, err := NewFromAST(ast)
				if err != nil {
					b.Fatalf("Component creation failed: %v", err)
				}

				var buf strings.Builder
				buf.Grow(bufferSize) // Pre-allocate with strategy
				err = component.RenderHTML(&buf)
				if err != nil {
					b.Fatalf("RenderHTML failed: %v", err)
				}
				_ = buf.String() // Force evaluation
			}
		})
	}
}

// BenchmarkPreallocationStrategies_Complex tests different strategies on complex templates
func BenchmarkPreallocationStrategies_Complex(b *testing.B) {
	template := loadRealWorldTemplate() // Real Austin layout template
	mjmlSize := len(template)
	componentCount := countComponents(template)
	nestingDepth := estimateNestingDepth(template)

	for _, strategy := range strategies {
		b.Run(strategy.Name, func(b *testing.B) {
			bufferSize := strategy.Calculator(mjmlSize, componentCount, nestingDepth)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ast, err := ParseMJML(template)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
				component, err := NewFromAST(ast)
				if err != nil {
					b.Fatalf("Component creation failed: %v", err)
				}

				var buf strings.Builder
				buf.Grow(bufferSize) // Pre-allocate with strategy
				err = component.RenderHTML(&buf)
				if err != nil {
					b.Fatalf("RenderHTML failed: %v", err)
				}
				_ = buf.String() // Force evaluation
			}
		})
	}
}

// BenchmarkPreallocationStrategies_RealWorld tests strategies on real-world test files
func BenchmarkPreallocationStrategies_RealWorld(b *testing.B) {
	template := loadRealWorldTemplate() // Real Austin layout template

	mjmlSize := len(template)
	componentCount := countComponents(template)
	nestingDepth := estimateNestingDepth(template)

	for _, strategy := range strategies {
		b.Run(strategy.Name, func(b *testing.B) {
			bufferSize := strategy.Calculator(mjmlSize, componentCount, nestingDepth)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ast, err := ParseMJML(template)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
				component, err := NewFromAST(ast)
				if err != nil {
					b.Fatalf("Component creation failed: %v", err)
				}

				var buf strings.Builder
				buf.Grow(bufferSize) // Pre-allocate with strategy
				err = component.RenderHTML(&buf)
				if err != nil {
					b.Fatalf("RenderHTML failed: %v", err)
				}
				_ = buf.String() // Force evaluation
			}
		})
	}
}

// BenchmarkBufferGrowth_Analysis analyzes how much buffers actually grow during rendering
func BenchmarkBufferGrowth_Analysis(b *testing.B) {
	testCases := []struct {
		name     string
		template string
	}{
		{"RealWorld", loadRealWorldTemplate()},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			mjmlSize := len(tc.template)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				ast, err := ParseMJML(tc.template)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
				component, err := NewFromAST(ast)
				if err != nil {
					b.Fatalf("Component creation failed: %v", err)
				}

				// No pre-allocation to see natural growth
				var buf strings.Builder
				err = component.RenderHTML(&buf)
				if err != nil {
					b.Fatalf("RenderHTML failed: %v", err)
				}

				result := buf.String()
				expansionRatio := float64(len(result)) / float64(mjmlSize)
				_ = expansionRatio // Use the ratio (in real usage, you'd log this)
			}
		})
	}
}

// Helper function to test expansion ratios
func TestExpansionRatios(t *testing.T) {
	testCases := []struct {
		name     string
		template string
	}{
		{"RealWorld", loadRealWorldTemplate()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := ParseMJML(tc.template)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			component, err := NewFromAST(ast)
			if err != nil {
				t.Fatalf("Component creation failed: %v", err)
			}

			var buf strings.Builder
			err = component.RenderHTML(&buf)
			if err != nil {
				t.Fatalf("RenderHTML failed: %v", err)
			}

			result := buf.String()
			mjmlSize := len(tc.template)
			htmlSize := len(result)
			componentCount := countComponents(tc.template)
			nestingDepth := estimateNestingDepth(tc.template)
			expansionRatio := float64(htmlSize) / float64(mjmlSize)

			t.Logf("Template: %s", tc.name)
			t.Logf("  MJML size: %d bytes", mjmlSize)
			t.Logf("  HTML size: %d bytes", htmlSize)
			t.Logf("  Components: %d", componentCount)
			t.Logf("  Nesting depth: %d", nestingDepth)
			t.Logf("  Expansion ratio: %.2fx", expansionRatio)
			t.Logf("  Component density: %.2f components/1000 chars", float64(componentCount)/float64(mjmlSize)*1000)
		})
	}
}
