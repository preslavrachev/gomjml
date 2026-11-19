package mjml

import (
	"regexp"
	"strings"
	"testing"
)

// imgWidth returns the <img> tag of src in html when its width attribute is
// want, failing the test otherwise.
func imgWidth(t *testing.T, html, src, want string) string {
	t.Helper()
	tag := regexp.MustCompile(`<img[^>]*src="` + regexp.QuoteMeta(src) + `"[^>]*>`).FindString(html)
	if got := regexp.MustCompile(` width="([^"]*)"`).FindStringSubmatch(tag); got == nil || got[1] != want {
		t.Fatalf("%s: want width=%q in %s", src, want, tag)
	}
	return tag
}

// The widths are MJML 4.16.1's for the same documents.
func TestBoxWidths(t *testing.T) {
	tests := []struct {
		name, bodyAttrs, body string
		want                  []string
	}{
		{
			name: "longhand padding wins over the shorthand",
			body: `<mj-section><mj-column><mj-image padding-left="0px" padding-right="0px" padding="0px 25px" width="780px" src="a.png" /></mj-column></mj-section>`,
			want: []string{`img a.png 600`},
		},
		{
			name: "percent padding counts as its number",
			body: `<mj-section padding="5%"><mj-column><mj-text>x</mj-text></mj-column></mj-section>`,
			want: []string{`vertical-align:top;width:590px;`},
		},
		{
			name: "section and column borders",
			body: `<mj-section border-left="2px solid red" padding-right="19px"><mj-column border="3px solid red"><mj-image src="b.png" border-left="4px solid" /></mj-column><mj-column><mj-text>y</mj-text></mj-column></mj-section>`,
			want: []string{`vertical-align:top;width:289.5px;`, `width:229px;`, `img b.png 229`, `vertical-align:top;width:289.5px;`},
		},
		{
			name: "group padding and column widths",
			body: `<mj-section><mj-group mj-class="p7"><mj-column width="40%"><mj-image src="c.png" /></mj-column><mj-column width="60%"><mj-text>z</mj-text></mj-column></mj-group></mj-section>`,
			want: []string{`vertical-align:top;width:234px;`, `img c.png 184`, `vertical-align:top;width:351px;`},
		},
		{
			name: "hero padding",
			body: `<mj-hero padding-left="12px" padding-right="12px"><mj-image src="d.png" /></mj-hero>`,
			want: []string{`img d.png 526`},
		},
		{
			name: "image wider than its column",
			body: `<mj-section padding="0 10px"><mj-column><mj-image src="e.png" width="700px" padding="0" /></mj-column></mj-section>`,
			want: []string{`width:580px;`, `img e.png 580`},
		},
		{
			name:      "a zero body width is a width",
			bodyAttrs: ` width="0px"`,
			body:      `<mj-section><mj-column><mj-image src="z.png" padding="0" /></mj-column></mj-section>`,
			want:      []string{`vertical-align:top;width:0px;`, `img z.png 0`},
		},
		{
			name:      "a zero body width reaches a hero",
			bodyAttrs: ` width="0px"`,
			body:      `<mj-hero><mj-image src="h.png" padding="0 10px" /></mj-hero>`,
			want:      []string{`img h.png -20`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `<mjml><mj-head><mj-attributes><mj-class name="p7" padding="7px" /></mj-attributes></mj-head><mj-body` + tt.bodyAttrs + `>` + tt.body + `</mj-body></mjml>`
			html, err := Render(input)
			if err != nil {
				t.Fatal(err)
			}
			rest := html
			for _, want := range tt.want {
				if fields := strings.Fields(want); fields[0] == "img" {
					want = imgWidth(t, html, fields[1], fields[2])
				}
				i := strings.Index(rest, want)
				if i < 0 {
					t.Fatalf("output lacks %q (in order) in\n%s", want, html)
				}
				rest = rest[i+len(want):]
			}
		})
	}
}
