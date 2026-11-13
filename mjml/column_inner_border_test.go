package mjml

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

// The styles are MJML 4.16.1's for the same columns: with padding, the inner
// table takes the inner-* attributes; without, the column's own.
func TestColumnInnerTableStyle(t *testing.T) {
	columns := []struct{ attrs, want string }{
		{
			`padding="10px" inner-border="1px solid #336699" inner-border-radius="4px" border="2px solid red" border-radius="8px" inner-background-color="#eee"`,
			"background-color:#eeeeee;border:1px solid #336699;border-radius:4px;",
		},
		{
			`inner-border="1px solid #336699" inner-border-radius="1px solid #336699" border-radius="3px" inner-background-color="#eee"`,
			"border-radius:3px;vertical-align:top;",
		},
		{
			`padding-left="5px" inner-border-left="2px dashed red" inner-border-top="1px solid blue" border="none"`,
			"border-left:2px dashed red;border-top:1px solid blue;",
		},
		{`padding="4px" inner-border-right="1px solid #336699"`, "border-right:1px solid #336699;"},
		{`padding="4px"`, ""},
		{`border="none"`, "border:none;vertical-align:top;"},
	}
	innerTable := regexp.MustCompile(`<table border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"(?: style="([^"]*)")?><tbody><tr><td align`)
	for _, c := range columns {
		html, err := Render(`<mjml><mj-body><mj-section><mj-column ` + c.attrs + `><mj-text>x</mj-text></mj-column></mj-section></mj-body></mjml>`)
		if err != nil {
			t.Fatal(err)
		}
		m := innerTable.FindStringSubmatch(html)
		if m == nil {
			t.Fatalf("<mj-column %s>: no inner table in\n%s", c.attrs, html)
		}
		if got := styleDeclarations(m[1]); got != styleDeclarations(c.want) {
			t.Errorf("<mj-column %s>: inner table style %q, want %q", c.attrs, m[1], c.want)
		}
	}
}

// styleDeclarations sorts a style attribute's declarations, since their order does not matter.
func styleDeclarations(style string) string {
	var out []string
	for _, d := range strings.Split(style, ";") {
		if d = strings.TrimSpace(d); d != "" {
			out = append(out, d)
		}
	}
	slices.Sort(out)
	return strings.Join(out, ";")
}
