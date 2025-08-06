package mjml

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

// TestMJRawWithGoTemplate tests the integration of mj-raw components with Go's html/template
// This test demonstrates how MJML templates can be used with Go templating for dynamic email generation
func TestMJRawWithGoTemplate(t *testing.T) {
	// Define the MJML template with Go template syntax inside mj-raw components
	mjmlTemplate := `<mjml>
  <mj-body>
    <mj-raw>{{range .Names}}</mj-raw>
    <mj-section>
      <mj-column>
        <mj-text>Hello, dear {{.Name}}</mj-text>
      </mj-column>
    </mj-section>
    <mj-raw>{{end}}</mj-raw>
  </mj-body>
</mjml>`

	// Define test data structure
	data := struct {
		Names []struct {
			Name string
		}
	}{
		Names: []struct {
			Name string
		}{
			{Name: "John"},
			{Name: "Jane"},
			{Name: "Bob"},
		},
	}

	// Parse the template
	tmpl, err := template.New("mjml").Parse(mjmlTemplate)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Execute the template with test data
	var templatedMJML bytes.Buffer
	err = tmpl.Execute(&templatedMJML, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Convert templated MJML to HTML using gomjml
	finalHTML, err := RenderHTML(templatedMJML.String())
	if err != nil {
		t.Fatalf("Failed to render MJML: %v", err)
	}

	// Verify that the final HTML contains the expected personalized content
	expectedContent := []string{
		"Hello, dear John",
		"Hello, dear Jane",
		"Hello, dear Bob",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(finalHTML, expected) {
			t.Errorf("Final HTML should contain %q", expected)
		}
	}

	// Verify basic MJML structure is maintained
	if !strings.Contains(finalHTML, "<!doctype html>") {
		t.Error("Final HTML should contain DOCTYPE declaration")
	}

	// Verify that we have the correct number of sections (one per name)
	// Count the occurrences of each personalized message to ensure all were generated
	for _, name := range data.Names {
		expectedMessage := "Hello, dear " + name.Name
		count := strings.Count(finalHTML, expectedMessage)
		if count != 1 {
			t.Errorf("Expected exactly 1 occurrence of %q, got %d", expectedMessage, count)
		}
	}

	// Optional: Print the final HTML for manual inspection during development
	t.Logf("Generated HTML:\n%s", finalHTML)
}
