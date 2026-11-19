package parser

import (
	"strings"
	"testing"
)

// MJML's htmlparser2 ends a tag at `</name` followed by '>' or whitespace, so
// Prettier's `</mj-text\n>` closes the text.
func TestEndingTagCloseWithWhitespace(t *testing.T) {
	closes := map[string]string{
		"space":      "</mj-text >",
		"newline":    "</mj-text\n          >",
		"tab":        "</mj-text\t>",
		"upper-case": "</MJ-TEXT >",
		"no-space":   "</mj-text>",
	}
	for name, close := range closes {
		t.Run(name, func(t *testing.T) {
			input := `<mjml><mj-body><mj-section><mj-column>` +
				`<mj-text>One <b>bold</b>` + close +
				`<mj-button href="#">Go</mj-button>` +
				`<mj-text>Two</mj-text>` +
				`</mj-column></mj-section></mj-body></mjml>`
			root, err := ParseMJML(input)
			if err != nil {
				t.Fatalf("ParseMJML: %v", err)
			}
			column := root.FindFirstChild("mj-body").FindFirstChild("mj-section").FindFirstChild("mj-column")
			var tags []string
			for _, child := range column.Children {
				tags = append(tags, child.GetTagName())
			}
			if got, want := strings.Join(tags, ","), "mj-text,mj-button,mj-text"; got != want {
				t.Fatalf("column children = %s, want %s", got, want)
			}
			if got := column.Children[0].Text; strings.Contains(got, "mj-button") || !strings.Contains(got, "<b>bold</b>") {
				t.Errorf("first mj-text content = %q", got)
			}
		})
	}
}

// A name that only starts with mj-text is not its close tag.
func TestEndingTagCloseNeedsNameBoundary(t *testing.T) {
	got := wrapEndingTagContent(`<mj-text>a</mj-textx>b</mj-text>`)
	want := `<mj-text><![CDATA[a</mj-textx>b]]></mj-text>`
	if got != want {
		t.Errorf("wrapEndingTagContent = %q, want %q", got, want)
	}
}
