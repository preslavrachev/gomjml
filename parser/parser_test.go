package parser

import (
	"testing"
)

func TestParseMJML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantTag string
		wantErr bool
	}{
		{
			name:    "basic mjml",
			input:   `<mjml><mj-body><mj-text>Hello</mj-text></mj-body></mjml>`,
			wantTag: "mjml",
			wantErr: false,
		},
		{
			name:    "with attributes",
			input:   `<mjml version="4.0"><mj-head><mj-title>Test</mj-title></mj-head></mjml>`,
			wantTag: "mjml",
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantTag: "",
			wantErr: true,
		},
		{
			name:    "invalid xml",
			input:   `<mjml><mj-body><mj-text>Hello</mj-body></mjml>`,
			wantTag: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseMJML(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseMJML() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseMJML() error = %v", err)
				return
			}

			if node.GetTagName() != tt.wantTag {
				t.Errorf("ParseMJML() tag = %v, want %v", node.GetTagName(), tt.wantTag)
			}
		})
	}
}

func TestMJMLNode_GetAttribute(t *testing.T) {
	input := `<mjml version="4.0" lang="en"><mj-head></mj-head></mjml>`
	node, err := ParseMJML(input)
	if err != nil {
		t.Fatalf("ParseMJML() error = %v", err)
	}

	tests := []struct {
		name string
		attr string
		want string
	}{
		{"existing attribute", "version", "4.0"},
		{"another existing attribute", "lang", "en"},
		{"non-existing attribute", "nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := node.GetAttribute(tt.attr)
			if got != tt.want {
				t.Errorf("GetAttribute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMJMLNode_FindFirstChild(t *testing.T) {
	input := `<mjml><mj-head><mj-title>Test</mj-title></mj-head><mj-body><mj-text>Hello</mj-text></mj-body></mjml>`
	node, err := ParseMJML(input)
	if err != nil {
		t.Fatalf("ParseMJML() error = %v", err)
	}

	tests := []struct {
		name    string
		tagName string
		wantTag string
		wantNil bool
	}{
		{"find head", "mj-head", "mj-head", false},
		{"find body", "mj-body", "mj-body", false},
		{"find non-existing", "mj-section", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := node.FindFirstChild(tt.tagName)

			if tt.wantNil {
				if child != nil {
					t.Errorf("FindFirstChild() expected nil but got %v", child.GetTagName())
				}
				return
			}

			if child == nil {
				t.Errorf("FindFirstChild() expected child but got nil")
				return
			}

			if child.GetTagName() != tt.wantTag {
				t.Errorf("FindFirstChild() tag = %v, want %v", child.GetTagName(), tt.wantTag)
			}
		})
	}
}

func TestMJMLNode_GetTextContent(t *testing.T) {
	input := `<mjml><mj-body><mj-text>  Hello World  </mj-text></mj-body></mjml>`
	node, err := ParseMJML(input)
	if err != nil {
		t.Fatalf("ParseMJML() error = %v", err)
	}

	textNode := node.FindFirstChild("mj-body").FindFirstChild("mj-text")
	if textNode == nil {
		t.Fatal("Could not find mj-text node")
	}

	got := textNode.GetTextContent()
	want := "Hello World"

	if got != want {
		t.Errorf("GetTextContent() = %q, want %q", got, want)
	}
}
