package mjml

import (
	"errors"
	"strings"
	"testing"
)

// TestRenderSameComponentTwiceProducesIdenticalOutput verifies that rendering
// one parsed component instance more than once (as callers reusing a cached
// AST do) produces byte-identical HTML both times, with no font or style-tag
// state leaking from the first render into the second.
func TestRenderSameComponentTwiceProducesIdenticalOutput(t *testing.T) {
	//GIVEN: a component built once via NewFromAST from a template with text and a Google font.
	mjmlContent := `<mjml><mj-body><mj-section><mj-column><mj-text font-family="Roboto, Arial, sans-serif">Hello</mj-text></mj-column></mj-section></mj-body></mjml>`
	ast, err := ParseMJML(mjmlContent)
	if err != nil {
		t.Fatalf("ParseMJML failed: %v", err)
	}
	component, err := NewFromAST(ast)
	if err != nil {
		t.Fatalf("NewFromAST failed: %v", err)
	}

	//WHEN: the same component instance is rendered twice.
	var first, second strings.Builder
	if err := component.Render(&first); err != nil {
		t.Fatalf("first Render failed: %v", err)
	}
	if err := component.Render(&second); err != nil {
		t.Fatalf("second Render failed: %v", err)
	}

	//THEN: both renders produce identical HTML.
	if first.String() != second.String() {
		t.Fatalf("expected identical output across repeated renders\nfirst:\n%s\nsecond:\n%s", first.String(), second.String())
	}
}

// TestRenderEmptyBodyPreservesWrapperAndWordSpacing verifies that an empty
// mj-body still renders its exact standard article wrapper, with the body
// tag keeping the word-spacing:normal style MJML emits whenever a body is
// present.
func TestRenderEmptyBodyPreservesWrapperAndWordSpacing(t *testing.T) {
	//GIVEN: an MJML document with an empty mj-body.
	mjmlContent := `<mjml><mj-body></mj-body></mjml>`

	//WHEN: it is rendered.
	html, err := Render(mjmlContent)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	//THEN: the document ends with the exact body tag, empty article wrapper div, and closing tags.
	const wantSuffix = `<body style="word-spacing:normal;"><div aria-roledescription="email" role="article" lang="und" dir="auto"></div></body></html>`
	if !strings.HasSuffix(html, wantSuffix) {
		t.Errorf("expected output to end with %q, got: %s", wantSuffix, html)
	}
}

// TestRenderGoogleFontViaMJClassAppearsInHead verifies that a Google font
// resolved only through an mj-class definition (not an element attribute)
// still produces a <link> and @import entry in the document head.
func TestRenderGoogleFontViaMJClassAppearsInHead(t *testing.T) {
	//GIVEN: mj-text referencing an mj-class that sets a Google font-family.
	mjmlContent := `<mjml>
		<mj-head>
			<mj-attributes>
				<mj-class name="branded" font-family="Lato, Arial, sans-serif" />
			</mj-attributes>
		</mj-head>
		<mj-body><mj-section><mj-column><mj-text mj-class="branded">Hello</mj-text></mj-column></mj-section></mj-body>
	</mjml>`

	//WHEN: it is rendered.
	html, err := Render(mjmlContent)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	//THEN: the Lato Google Font link and @import both appear in the head.
	latoURL := "https://fonts.googleapis.com/css?family=Lato:300,400,500,700"
	if !strings.Contains(html, `<link href="`+latoURL+`"`) {
		t.Errorf("expected Lato font link in head, got: %s", html)
	}
	if !strings.Contains(html, "@import url("+latoURL+");") {
		t.Errorf("expected Lato font @import in head, got: %s", html)
	}
}

// TestRenderGoogleFontViaComponentGlobalAttributeAppearsInHead verifies that
// a font declared through a component-specific global default inside
// mj-attributes (rather than mj-all or an element attribute) is detected
// independently of the existing mj-all coverage.
func TestRenderGoogleFontViaComponentGlobalAttributeAppearsInHead(t *testing.T) {
	//GIVEN: mj-attributes sets font-family on mj-text specifically, with no element-level override.
	mjmlContent := `<mjml>
		<mj-head>
			<mj-attributes>
				<mj-text font-family="Roboto, Arial, sans-serif" />
			</mj-attributes>
		</mj-head>
		<mj-body><mj-section><mj-column><mj-text>Hello</mj-text></mj-column></mj-section></mj-body>
	</mjml>`

	//WHEN: it is rendered.
	html, err := Render(mjmlContent)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	//THEN: the Roboto Google Font link and @import both appear in the head.
	robotoURL := "https://fonts.googleapis.com/css?family=Roboto:300,400,500,700"
	if !strings.Contains(html, `<link href="`+robotoURL+`"`) {
		t.Errorf("expected Roboto font link in head, got: %s", html)
	}
	if !strings.Contains(html, "@import url("+robotoURL+");") {
		t.Errorf("expected Roboto font @import in head, got: %s", html)
	}
}

// TestRenderSocialFontInheritanceAndOverrideMatchRenderedStyle verifies that
// the fonts imported in the head match what social elements actually render:
// one element inherits the parent's Google font, another overrides it with
// its own, and both fonts must be present in the head.
func TestRenderSocialFontInheritanceAndOverrideMatchRenderedStyle(t *testing.T) {
	//GIVEN: mj-social sets a Google font that one child inherits, while a sibling overrides it with a different Google font.
	mjmlContent := `<mjml><mj-body><mj-section><mj-column>
		<mj-social font-family="Lato, Arial, sans-serif">
			<mj-social-element name="twitter">Inherits</mj-social-element>
			<mj-social-element name="facebook" font-family="Roboto, Arial, sans-serif">Overrides</mj-social-element>
		</mj-social>
	</mj-column></mj-section></mj-body></mjml>`

	//WHEN: it is rendered.
	html, err := Render(mjmlContent)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	//THEN: both the inherited (Lato) and overriding (Roboto) fonts appear in the head,
	// matching the font-family styles actually rendered on each element.
	latoURL := "https://fonts.googleapis.com/css?family=Lato:300,400,500,700"
	robotoURL := "https://fonts.googleapis.com/css?family=Roboto:300,400,500,700"
	if !strings.Contains(html, `<link href="`+latoURL+`"`) {
		t.Errorf("expected inherited Lato font link in head, got: %s", html)
	}
	if !strings.Contains(html, `<link href="`+robotoURL+`"`) {
		t.Errorf("expected overriding Roboto font link in head, got: %s", html)
	}
	if !strings.Contains(html, "font-family:Lato, Arial, sans-serif") {
		t.Errorf("expected inherited Lato font-family to be rendered on the twitter element, got: %s", html)
	}
	if !strings.Contains(html, "font-family:Roboto, Arial, sans-serif") {
		t.Errorf("expected overriding Roboto font-family to be rendered on the facebook element, got: %s", html)
	}
}

// TestRenderDeduplicatesFontsAndPreservesRegistryOrder verifies that a font
// used by multiple components is imported only once, and that two distinct
// Google fonts appear in the head in the registry's fixed order (Lato before
// Roboto), independent of the order components using them appear in the body.
func TestRenderDeduplicatesFontsAndPreservesRegistryOrder(t *testing.T) {
	//GIVEN: a document where a Roboto-using button appears before a Lato-using text component
	// in body order, and Roboto is used twice.
	mjmlContent := `<mjml><mj-body><mj-section><mj-column>
		<mj-button font-family="Roboto, Arial, sans-serif">Click</mj-button>
		<mj-text font-family="Lato, Arial, sans-serif">Hello</mj-text>
		<mj-table font-family="Roboto, Arial, sans-serif"><tr><td>Cell</td></tr></mj-table>
	</mj-column></mj-section></mj-body></mjml>`

	//WHEN: it is rendered.
	html, err := Render(mjmlContent)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	//THEN: each font's link and @import appear exactly once, and Lato (registry order)
	// precedes Roboto even though Roboto appears first in the body.
	robotoURL := "https://fonts.googleapis.com/css?family=Roboto:300,400,500,700"
	latoURL := "https://fonts.googleapis.com/css?family=Lato:300,400,500,700"

	if count := strings.Count(html, `<link href="`+robotoURL+`"`); count != 1 {
		t.Errorf("expected exactly 1 Roboto link tag, got %d", count)
	}
	if count := strings.Count(html, "@import url("+robotoURL+");"); count != 1 {
		t.Errorf("expected exactly 1 Roboto @import, got %d", count)
	}

	robotoIdx := strings.Index(html, robotoURL)
	latoIdx := strings.Index(html, latoURL)
	if robotoIdx == -1 || latoIdx == -1 {
		t.Fatalf("expected both fonts present in head, got: %s", html)
	}
	if !(latoIdx < robotoIdx) {
		t.Errorf("expected Lato (registry order) to precede Roboto regardless of body order, latoIdx=%d robotoIdx=%d", latoIdx, robotoIdx)
	}
}

// TestRenderStreamingBodyFailurePropagatesSentinelError verifies that when
// the destination writer starts failing right after the body opening tag is
// written, Render returns that exact error unchanged instead of swallowing
// or wrapping it.
func TestRenderStreamingBodyFailurePropagatesSentinelError(t *testing.T) {
	//GIVEN: a component with body content and a writer that fails immediately after the body tag opens.
	mjmlContent := `<mjml><mj-body><mj-section><mj-column><mj-text>Hello</mj-text></mj-column></mj-section></mj-body></mjml>`
	ast, err := ParseMJML(mjmlContent)
	if err != nil {
		t.Fatalf("ParseMJML failed: %v", err)
	}
	component, err := NewFromAST(ast)
	if err != nil {
		t.Fatalf("NewFromAST failed: %v", err)
	}
	w := &failingAfterBodyOpenWriter{}

	//WHEN: the component is rendered to that writer.
	renderErr := component.Render(w)

	//THEN: the sentinel error is returned unchanged.
	if !errors.Is(renderErr, sentinelWriterError) {
		t.Fatalf("expected sentinel error to propagate unchanged, got: %v", renderErr)
	}
}

// sentinelWriterError is returned by failingAfterBodyOpenWriter once its
// failure trigger has been written, to verify streaming errors are returned
// unchanged rather than swallowed or replaced.
var sentinelWriterError = errors.New("sentinel write failure")

// failingAfterBodyOpenWriter is an io.StringWriter that succeeds normally
// until it observes the opening <body tag, after which every subsequent
// WriteString call fails with sentinelWriterError.
type failingAfterBodyOpenWriter struct {
	strings.Builder
	bodyOpened bool
	failing    bool
}

func (w *failingAfterBodyOpenWriter) WriteString(s string) (int, error) {
	if w.failing {
		return 0, sentinelWriterError
	}
	if strings.Contains(s, "<body") {
		w.bodyOpened = true
	}
	n, err := w.Builder.WriteString(s)
	if w.bodyOpened {
		w.failing = true
	}
	return n, err
}
