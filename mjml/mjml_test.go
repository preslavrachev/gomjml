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

// A template that emits markup before <mjml> must fail rather than render
// only the stray element (issue #36).
func TestRenderRejectsElementBeforeRoot(t *testing.T) {
	input := `
<mj-text font-size="18px">I SHOULD NOT BE HERE</mj-text>

<mjml>
  <mj-body>
    <mj-section>
      <mj-column>
        <mj-text>Hello Ed</mj-text>
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
`
	html, err := Render(input)
	if err == nil {
		t.Fatalf("Render() error = nil, output %q", html)
	}
	if want := "expected <mjml> root, found <mj-text> at line 2"; !strings.Contains(err.Error(), want) {
		t.Errorf("Render() error = %q, want it to contain %q", err, want)
	}
	if html != "" {
		t.Errorf("Render() output = %q, want none", html)
	}
}

func TestRenderRejectsFragmentRoot(t *testing.T) {
	html, err := Render(`<mj-body><mj-section><mj-column><mj-text>Hi</mj-text></mj-column></mj-section></mj-body>`)
	if err == nil {
		t.Fatalf("Render() error = nil, output %q", html)
	}
	if want := "expected <mjml> root, found <mj-body>"; !strings.Contains(err.Error(), want) {
		t.Errorf("Render() error = %q, want it to contain %q", err, want)
	}
}

func TestRenderFromASTAcceptsMJMLPrefixedFragment(t *testing.T) {
	ast, err := ParseMJML(`<mj-raw><mjml-logo/></mj-raw>`)
	if err != nil {
		t.Fatalf("ParseMJML() error = %v", err)
	}
	html, err := RenderFromAST(ast)
	if err != nil {
		t.Fatalf("RenderFromAST() error = %v", err)
	}
	if !strings.Contains(html, "<mjml-logo") {
		t.Errorf("RenderFromAST() = %q, want the raw <mjml-logo> element", html)
	}
}
