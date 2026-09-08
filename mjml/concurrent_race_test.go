package mjml

import (
	"fmt"
	"sync"
	"testing"
)

// tplWithColor renders an mj-text with the given color via mj-attributes.
// Two goroutines using different colors will corrupt each other's output
// when the global attributes singleton is unprotected.
func tplWithColor(color string) string {
	return fmt.Sprintf(`<mjml>
  <mj-head>
    <mj-attributes>
      <mj-text color="%s" />
    </mj-attributes>
  </mj-head>
  <mj-body>
    <mj-section><mj-column>
      <mj-text>Hello</mj-text>
    </mj-column></mj-section>
  </mj-body>
</mjml>`, color)
}

// TestConcurrentRender_GlobalAttributesRace runs concurrent renders with
// different mj-attributes and verifies each output contains the correct color.
// Run with -race to surface the data race; run without -race to expose wrong
// output caused by the shared global singleton.
func TestConcurrentRender_GlobalAttributesRace(t *testing.T) {
	colors := []string{"red", "blue", "green", "purple", "orange", "teal", "navy", "maroon"}
	const iterations = 20

	for iter := range iterations {
		var wg sync.WaitGroup
		type result struct {
			color string
			html  string
			err   error
		}
		results := make([]result, len(colors))

		for i, color := range colors {
			wg.Add(1)
			go func(idx int, c string) {
				defer wg.Done()
				html, err := Render(tplWithColor(c))
				results[idx] = result{color: c, html: html, err: err}
			}(i, color)
		}
		wg.Wait()

		for _, r := range results {
			if r.err != nil {
				t.Errorf("iter %d: Render(%s) error: %v", iter, r.color, r.err)
				continue
			}
			// Each render must contain its own color, not a neighbor's
			needle := fmt.Sprintf(`color:%s`, r.color)
			if !contains(r.html, needle) {
				t.Errorf("iter %d: expected output for color=%q to contain %q\ngot (first 500 chars):\n%s",
					iter, r.color, needle, truncate(r.html, 500))
			}
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
