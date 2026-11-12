package mjml

import (
	"strings"
	"testing"
)

// mjml 4.x renders these placements with a warning, so their children must reach the
// output. gomjml used to drop them without any error.
func TestMisplacedComponentsKeepTheirChildren(t *testing.T) {
	tests := []struct {
		name   string
		mjml   string
		marker string
	}{
		{
			name: "section inside column",
			mjml: `<mjml><mj-body><mj-section><mj-column>
<mj-section><mj-column><mj-text>NESTED-SECTION-TEXT</mj-text></mj-column></mj-section>
</mj-column></mj-section></mj-body></mjml>`,
			marker: "NESTED-SECTION-TEXT",
		},
		{
			name:   "group directly in body",
			mjml:   `<mjml><mj-body><mj-group><mj-column><mj-text>GROUP-IN-BODY-TEXT</mj-text></mj-column></mj-group></mj-body></mjml>`,
			marker: "GROUP-IN-BODY-TEXT",
		},
		{
			name:   "group directly in wrapper",
			mjml:   `<mjml><mj-body><mj-wrapper><mj-group><mj-column><mj-text>GROUP-IN-WRAPPER-TEXT</mj-text></mj-column></mj-group></mj-wrapper></mj-body></mjml>`,
			marker: "GROUP-IN-WRAPPER-TEXT",
		},
		{
			name:   "column directly in body",
			mjml:   `<mjml><mj-body><mj-column><mj-text>COLUMN-IN-BODY-TEXT</mj-text></mj-column></mj-body></mjml>`,
			marker: "COLUMN-IN-BODY-TEXT",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := Render(tt.mjml)
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if !strings.Contains(html, tt.marker) {
				t.Errorf("%s is missing from the output", tt.marker)
			}
		})
	}
}
