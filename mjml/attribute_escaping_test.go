package mjml

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// wrapImage embeds a single mj-image snippet in a minimal valid MJML document.
func wrapImage(image string) string {
	return `<mjml><mj-body><mj-section><mj-column>` +
		image +
		`</mj-column></mj-section></mj-body></mjml>`
}

// renderImgAlt renders the MJML document, parses the resulting HTML the way a
// browser would, and returns the decoded (browser-visible) value of the alt
// attribute of the first <img> element together with the raw rendered HTML.
func renderImgAlt(t *testing.T, mjmlDoc string) (alt, rawHTML string) {
	t.Helper()
	rawHTML, err := Render(mjmlDoc)
	if err != nil {
		t.Fatalf("Render returned an error: %v", err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		t.Fatalf("failed to parse rendered HTML: %v", err)
	}
	img := doc.Find("img").First()
	if img.Length() == 0 {
		t.Fatalf("no <img> element found in rendered HTML:\n%s", rawHTML)
	}
	alt, _ = img.Attr("alt")
	return alt, rawHTML
}

// TestAttributeValueEscaping covers the four attribute-escaping cases from the
// bug report. In every case the author's intended alt text is `27" monitor` (or,
// for the angle-bracket case, `tag <b> here`), and the browser-visible alt value
// must read back exactly that text.
//
// Before the fix: case A and case C failed the XML parse outright, and case B
// rendered invalid markup (alt="27" monitor") whose embedded quote truncated the
// attribute. Case D (the double-escaped-ampersand workaround) was the only input
// that produced correct output; it must keep working.
func TestAttributeValueEscaping(t *testing.T) {
	const wantQuote = `27" monitor`
	const wantAngle = `tag <b> here`

	cases := []struct {
		name  string
		image string
		want  string
	}{
		{
			name:  "A_named_quot_entity",
			image: `<mj-image src="https://e.test/a.png" alt="27&quot; monitor"/>`,
			want:  wantQuote,
		},
		{
			name:  "B_numeric_quote_entity",
			image: `<mj-image src="https://e.test/a.png" alt="27&#34; monitor"/>`,
			want:  wantQuote,
		},
		{
			name:  "C_named_angle_bracket_entities",
			image: `<mj-image src="https://e.test/a.png" alt="tag &lt;b&gt; here"/>`,
			want:  wantAngle,
		},
		{
			name:  "D_double_escaped_ampersand_workaround",
			image: `<mj-image src="https://e.test/a.png" alt="27&amp;quot; monitor"/>`,
			want:  wantQuote,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, rawHTML := renderImgAlt(t, wrapImage(tc.image))
			if got != tc.want {
				t.Fatalf("browser-visible alt = %q, want %q\nrendered HTML:\n%s", got, tc.want, rawHTML)
			}
		})
	}
}

// TestAttributeValueEscapingSerialization asserts that the natural, XML-escaped
// input alt="27&quot; monitor" serializes with the quote kept encoded, so it can
// never terminate the attribute, rather than emitting a raw double quote.
func TestAttributeValueEscapingSerialization(t *testing.T) {
	_, rawHTML := renderImgAlt(t, wrapImage(`<mj-image src="https://e.test/a.png" alt="27&quot; monitor"/>`))

	if !strings.Contains(rawHTML, `alt="27&quot; monitor"`) {
		t.Fatalf("expected the quote to be emitted encoded as alt=\"27&quot; monitor\", got:\n%s", rawHTML)
	}
	if strings.Contains(rawHTML, `alt="27" monitor"`) {
		t.Fatalf("rendered a raw double quote inside the alt attribute (invalid markup):\n%s", rawHTML)
	}
}

// TestMjTextEntityDecodingUnaffected guards the defect-1 caveat: mj-text content
// is CDATA-wrapped after the entity pre-pass, and encoding/xml does not decode
// entities inside CDATA, so the pre-pass remains the only decoder for mj-text.
// Scoping the pre-pass to non-attribute context (rather than removing the
// decodes) must keep mj-text rendering its entities exactly as before.
func TestMjTextEntityDecodingUnaffected(t *testing.T) {
	doc := `<mjml><mj-body><mj-section><mj-column>` +
		`<mj-text>&copy; 2026 &amp; &lt;b&gt;bold&lt;/b&gt; &quot;q&quot;</mj-text>` +
		`</mj-column></mj-section></mj-body></mjml>`

	out, err := Render(doc)
	if err != nil {
		t.Fatalf("Render returned an error: %v", err)
	}

	// &lt;b&gt; must be decoded to a real <b> tag, not emitted literally.
	if !strings.Contains(out, "<b>bold</b>") {
		t.Fatalf("expected mj-text to decode &lt;b&gt; to a real <b> tag, got:\n%s", out)
	}
	if strings.Contains(out, "&lt;b&gt;") {
		t.Fatalf("mj-text emitted a literal &lt;b&gt; instead of decoding it:\n%s", out)
	}
	// &amp; stays encoded (so a browser reads a single ampersand).
	if !strings.Contains(out, "&amp;") {
		t.Fatalf("expected mj-text to preserve &amp;, got:\n%s", out)
	}
	// The HTML-only entity &copy; is still decoded to its glyph.
	if !strings.Contains(out, "© 2026") {
		t.Fatalf("expected mj-text to decode &copy; to the copyright glyph, got:\n%s", out)
	}
}
