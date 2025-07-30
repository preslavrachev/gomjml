package mjml

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		contains []string
	}{
		{
			name: "basic mjml",
			input: `<mjml>
				<mj-body>
					<mj-section>
						<mj-column>
							<mj-text>Hello World</mj-text>
						</mj-column>
					</mj-section>
				</mj-body>
			</mjml>`,
			wantErr: false,
			contains: []string{
				"<!doctype html>",
				"Hello World",
				"mj-column-per-100",
			},
		},
		{
			name: "with head",
			input: `<mjml>
				<mj-head>
					<mj-title>Test Email</mj-title>
				</mj-head>
				<mj-body>
					<mj-section>
						<mj-column>
							<mj-text>Hello</mj-text>
						</mj-column>
					</mj-section>
				</mj-body>
			</mjml>`,
			wantErr: false,
			contains: []string{
				"<title>Test Email</title>",
				"Hello",
			},
		},
		{
			name:     "invalid xml",
			input:    `<mjml><mj-body><mj-text>Hello</mj-body></mjml>`,
			wantErr:  true,
			contains: nil,
		},
		{
			name:     "empty input",
			input:    "",
			wantErr:  true,
			contains: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := Render(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Render() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Render() error = %v", err)
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(html, want) {
					t.Errorf("Render() output should contain %q", want)
				}
			}
		})
	}
}

func TestCreateComponent(t *testing.T) {
	// Test basic component creation
	ast, err := ParseMJML(`<mjml><mj-body><mj-text>Hello</mj-text></mj-body></mjml>`)
	if err != nil {
		t.Fatalf("ParseMJML() error = %v", err)
	}

	comp, err := CreateComponent(ast, nil)
	if err != nil {
		t.Fatalf("CreateComponent() error = %v", err)
	}

	if comp.GetTagName() != "mjml" {
		t.Errorf("CreateComponent() tag = %v, want mjml", comp.GetTagName())
	}

	// Test rendering
	html, err := RenderComponentString(comp)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.Contains(html, "Hello") {
		t.Errorf("Render() output should contain 'Hello'")
	}
}
