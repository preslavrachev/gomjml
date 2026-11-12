package mjml

import (
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/preslavrachev/gomjml/mjml/fonts"
)

// helper
func containsUbuntuBlock(html string) bool {
	return strings.Contains(html, "https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700") &&
		strings.Contains(html, "@import url(https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700);") &&
		strings.Contains(html, "<!--[if !mso]><!-->") &&
		strings.Contains(html, "<!--<![endif]-->")
}

// (If you have a RenderOptions/RenderOpts, swap in accordingly.)
func renderMJML(t *testing.T, mjml string) string {
	t.Helper()
	out, err := Render(mjml)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	return out
}

// TestFontImportTriggerMatrix verifies that the MJML renderer correctly includes or omits
// the Ubuntu font import block in the generated HTML <head> based on the presence and type
// of MJML components in the input markup. It tests various scenarios, including default
// text components, buttons, social elements, navbar links, accordion elements, hero headings,
// tables, images, and font-family overrides, ensuring that the font import logic matches
// MJML's upstream behavior and respects explicit font-family settings.
func TestFontImportTriggerMatrix(t *testing.T) {
	tests := []struct {
		name         string
		mjml         string
		expectUbuntu bool
		reason       string
	}{
		{
			name:         "Text component triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-text>Hello</mj-text></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-text default font stack includes Ubuntu",
		},
		{
			name:         "Button only triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-button>Click</mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-button default stack includes Ubuntu",
		},
		{
			name:         "Social triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-social><mj-social-element name="twitter">Tw</mj-social-element></mj-social></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-social-element text uses default stack",
		},
		{
			name:         "Navbar triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-navbar><mj-navbar-link href="https://x">X</mj-navbar-link></mj-navbar></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-navbar-link default stack includes Ubuntu",
		},
		{
			name:         "Accordion triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-accordion><mj-accordion-element><mj-accordion-title>Q</mj-accordion-title><mj-accordion-text>A</mj-accordion-text></mj-accordion-element></mj-accordion></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "accordion title/text default stack includes Ubuntu",
		},
		{
			name:         "Hero heading triggers (text inside hero)",
			mjml:         `<mjml><mj-body><mj-hero mode="fixed-height" height="200px"><mj-text>Hero</mj-text></mj-hero></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-text nested in hero still a text component",
		},
		{
			name:         "Table triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-table><tr><td>Cell</td></tr></mj-table></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-table default font-family includes Ubuntu",
		},
		{
			name:         "Image only does NOT trigger",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-image src="x.png" /></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: false,
			reason:       "No text-like component; upstream MJML omits font block",
		},
		{
			name:         "System font override removes Ubuntu",
			mjml:         `<mjml><mj-head><mj-attributes><mj-all font-family="Arial, Helvetica, sans-serif" /></mj-attributes></mj-head><mj-body><mj-section><mj-column><mj-text>Override</mj-text></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: false,
			reason:       "Override removes Google font token from first position",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			html := renderMJML(t, tc.mjml)
			got := containsUbuntuBlock(html)
			if got != tc.expectUbuntu {
				t.Errorf("Ubuntu block presence=%v want %v (%s)\nHTML(head snippet)=%s",
					got, tc.expectUbuntu, tc.reason, headSnippet(html, 500))
			}
		})
	}
}

// TestMJFontImports verifies the stylesheets imported for mj-font declarations, in order, as
// MJML 4.18 imports them for the same templates: an mj-font named like a GoogleFontsMapping
// entry replaces that entry's URL at the entry's position, other mj-font names follow the
// built-in fonts, and a name declared several times imports its last href.
func TestMJFontImports(t *testing.T) {
	const (
		googleLato   = "https://fonts.googleapis.com/css?family=Lato:300,400,500,700"
		googleRoboto = "https://fonts.googleapis.com/css?family=Roboto:300,400,500,700"
		customLato   = "https://fonts.example.com/lato.css"
		customUbuntu = "https://fonts.example.com/ubuntu.css"
		raleway      = "https://fonts.example.com/raleway.css"
		pacifico     = "https://fonts.example.com/pacifico.css"
	)

	tests := []struct {
		name string
		head string
		body string
		want []string
	}{
		{
			name: "Built-in font replaced",
			head: `<mj-font name="Lato" href="` + customLato + `" />`,
			body: `<mj-text font-family="Lato, Arial">Hi</mj-text>`,
			want: []string{customLato},
		},
		{
			name: "Replaced font keeps the built-in font's position",
			head: `<mj-font name="Lato" href="` + customLato + `" />`,
			body: `<mj-text font-family="Roboto">Hi</mj-text><mj-text font-family="Lato">Hi</mj-text>`,
			want: []string{customLato, googleRoboto},
		},
		{
			name: "Replaced font comes before other mj-font names",
			head: `<mj-font name="Pacifico" href="` + pacifico + `" /><mj-font name="Lato" href="` + customLato + `" />`,
			body: `<mj-text font-family="Pacifico">Hi</mj-text><mj-text font-family="Lato">Hi</mj-text>`,
			want: []string{customLato, pacifico},
		},
		{
			name: "Other mj-font names come after built-in fonts",
			head: `<mj-font name="Raleway" href="` + raleway + `" />`,
			body: `<mj-text font-family="Raleway">Hi</mj-text><mj-text font-family="Lato">Hi</mj-text>`,
			want: []string{googleLato, raleway},
		},
		{
			name: "Font used through mj-all",
			head: `<mj-font name="Lato" href="` + customLato + `" /><mj-attributes><mj-all font-family="Lato" /></mj-attributes>`,
			body: `<mj-text>Hi</mj-text>`,
			want: []string{customLato},
		},
		{
			name: "Font used only in inline HTML styles",
			head: `<mj-font name="Lato" href="` + customLato + `" />`,
			body: `<mj-text font-family="Arial"><span style="font-family:Lato">Hi</span></mj-text>`,
			want: []string{customLato},
		},
		{
			name: "Default font of text components replaced",
			head: `<mj-font name="Ubuntu" href="` + customUbuntu + `" />`,
			body: `<mj-text>Hi</mj-text>`,
			want: []string{customUbuntu},
		},
		{
			name: "Default font of social and button components replaced",
			head: `<mj-font name="Ubuntu" href="` + customUbuntu + `" />`,
			body: `<mj-social><mj-social-element name="twitter">Tw</mj-social-element></mj-social><mj-button>Go</mj-button>`,
			want: []string{customUbuntu},
		},
		{
			name: "Replaced default font keeps the built-in font's position",
			head: `<mj-font name="Ubuntu" href="` + customUbuntu + `" />`,
			body: `<mj-text font-family="Roboto">Hi</mj-text><mj-text font-family="Ubuntu">Hi</mj-text>`,
			want: []string{googleRoboto, customUbuntu},
		},
		{
			name: "Name differing in case does not replace",
			head: `<mj-font name="lato" href="` + customLato + `" />`,
			body: `<mj-text font-family="Lato, Arial">Hi</mj-text>`,
			want: []string{googleLato, customLato},
		},
		{
			name: "Redeclared name imports its last href",
			head: `<mj-font name="Raleway" href="https://fonts.example.com/raleway-old.css" /><mj-font name="Raleway" href="` + raleway + `" />`,
			body: `<mj-text font-family="Raleway">Hi</mj-text>`,
			want: []string{raleway},
		},
		{
			name: "Redeclared name keeps its first position",
			head: `<mj-font name="Raleway" href="https://fonts.example.com/raleway-old.css" /><mj-font name="Pacifico" href="` + pacifico + `" /><mj-font name="Raleway" href="` + raleway + `" />`,
			body: `<mj-text font-family="Raleway">Hi</mj-text><mj-text font-family="Pacifico">Hi</mj-text>`,
			want: []string{raleway, pacifico},
		},
		{
			name: "Redeclared built-in font",
			head: `<mj-font name="Lato" href="https://fonts.example.com/lato-old.css" /><mj-font name="Lato" href="` + customLato + `" />`,
			body: `<mj-text font-family="Lato">Hi</mj-text>`,
			want: []string{customLato},
		},
		{
			name: "Built-in URL declared last",
			head: `<mj-font name="Lato" href="` + customLato + `" /><mj-font name="Lato" href="` + googleLato + `" />`,
			body: `<mj-text font-family="Lato">Hi</mj-text>`,
			want: []string{googleLato},
		},
		{
			name: "Built-in URL declared first",
			head: `<mj-font name="Lato" href="` + googleLato + `" /><mj-font name="Lato" href="` + customLato + `" />`,
			body: `<mj-text font-family="Lato">Hi</mj-text>`,
			want: []string{customLato},
		},
		{
			name: "Built-in URL declared once",
			head: `<mj-font name="Lato" href="` + googleLato + `" />`,
			body: `<mj-text font-family="Lato, Arial">Hi</mj-text>`,
			want: []string{googleLato},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			html := renderMJML(t,
				`<mjml><mj-head>`+tc.head+`</mj-head><mj-body><mj-section><mj-column>`+tc.body+`</mj-column></mj-section></mj-body></mjml>`)

			links, imports := fontImports(html)
			if !slices.Equal(links, tc.want) {
				t.Errorf("font <link> hrefs = %q, want %q", links, tc.want)
			}
			if !slices.Equal(imports, tc.want) {
				t.Errorf("font @import urls = %q, want %q", imports, tc.want)
			}
		})
	}
}

var (
	fontLinkPattern   = regexp.MustCompile(`<link href="([^"]+)" rel="stylesheet" type="text/css">`)
	fontImportPattern = regexp.MustCompile(`@import url\(([^)]+)\);`)
)

// fontImports returns, in document order, the hrefs of the font <link> tags and the URLs of
// the font @import rules of a rendered document.
func fontImports(html string) (links, imports []string) {
	for _, m := range fontLinkPattern.FindAllStringSubmatch(html, -1) {
		links = append(links, m[1])
	}
	for _, m := range fontImportPattern.FindAllStringSubmatch(html, -1) {
		imports = append(imports, m[1])
	}
	return links, imports
}

// TestMJFontReplacementIsPerRender verifies that an mj-font replacing a built-in URL only
// affects the template declaring it: fonts.GoogleFontsMapping is left untouched, and
// templates with and without the mj-font rendered concurrently each get their own font.
// Run with -race to also catch an implementation that writes to the shared mapping.
func TestMJFontReplacementIsPerRender(t *testing.T) {
	const (
		googleLato  = "https://fonts.googleapis.com/css?family=Lato:300,400,500,700"
		customLato  = "https://fonts.example.com/lato.css"
		withFont    = `<mjml><mj-head><mj-font name="Lato" href="` + customLato + `" /></mj-head><mj-body><mj-section><mj-column><mj-text font-family="Lato, Arial">Hi</mj-text></mj-column></mj-section></mj-body></mjml>`
		withoutFont = `<mjml><mj-body><mj-section><mj-column><mj-text font-family="Lato, Arial">Hi</mj-text></mj-column></mj-section></mj-body></mjml>`
		renders     = 50
	)

	var wg sync.WaitGroup
	errs := make(chan string, 2*renders)
	for range renders {
		wg.Add(2)
		go func() {
			defer wg.Done()
			html, err := Render(withFont)
			if err != nil {
				errs <- err.Error()
			} else if strings.Contains(html, googleLato) || !strings.Contains(html, customLato) {
				errs <- "template with mj-font: expected only " + customLato
			}
		}()
		go func() {
			defer wg.Done()
			html, err := Render(withoutFont)
			if err != nil {
				errs <- err.Error()
			} else if strings.Contains(html, customLato) || !strings.Contains(html, googleLato) {
				errs <- "template without mj-font: expected only " + googleLato
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}

	if got := fonts.GoogleFontsMapping["Lato"]; got != googleLato {
		t.Errorf("GoogleFontsMapping[\"Lato\"] = %q, want %q", got, googleLato)
	}
}

// headSnippet extracts first n chars from <head> for debugging
func headSnippet(html string, n int) string {
	if len(html) < n {
		return html
	}
	return html[:n]
}

// TestFontsBlockOrderingRelativeToMediaQueries verifies that the conditional fonts block
// (specifically for Ubuntu fonts) is present in the rendered HTML and that it appears
// before the first responsive @media rule. This ensures correct ordering of font and
// media query blocks for proper email client rendering.
func TestFontsBlockOrderingRelativeToMediaQueries(t *testing.T) {
	html := renderMJML(t,
		`<mjml><mj-body><mj-section><mj-column><mj-text>Hello</mj-text></mj-column></mj-section></mj-body></mjml>`)

	// Basic presence
	if !containsUbuntuBlock(html) {
		t.Fatalf("expected Ubuntu fonts block")
	}

	fontBlockIdx := strings.Index(html, "<!--[if !mso]><!-->")
	mediaIdx := strings.Index(html, "@media only screen")
	if mediaIdx < 0 {
		t.Fatalf("expected a responsive @media style")
	}
	if !(fontBlockIdx >= 0 && fontBlockIdx < mediaIdx) {
		t.Errorf("expected fonts conditional block to precede first @media rule: fontBlockIdx=%d mediaIdx=%d", fontBlockIdx, mediaIdx)
	}
}
