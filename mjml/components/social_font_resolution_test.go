package components

import (
	"testing"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/globals"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// TestMJSocialElementResolveAttribute_FontPrecedenceWithoutTracking verifies
// mj-social-element's parent-aware font-family precedence (element, mj-class,
// parent explicit/resolved, global, default) resolves without tracking, and
// that tracking happens exactly once via the getAttribute wrapper.
func TestMJSocialElementResolveAttribute_FontPrecedenceWithoutTracking(t *testing.T) {
	tests := []struct {
		name          string
		buildElement  func(t *testing.T) *MJSocialElementComponent
		expectedValue string
	}{
		{
			name: "element attribute",
			buildElement: func(t *testing.T) *MJSocialElementComponent {
				doc := parseSocialFixture(t, `<mj-social><mj-social-element name="twitter" font-family="Georgia" src="i.png">Hi</mj-social-element></mj-social>`)
				return newSocialElementFromFixture(t, doc, &options.RenderOpts{FontTracker: options.NewFontTracker()})
			},
			expectedValue: "Georgia",
		},
		{
			name: "mj-class attribute",
			buildElement: func(t *testing.T) *MJSocialElementComponent {
				doc := parseSocialFixture(t, `<mj-social><mj-social-element name="twitter" mj-class="branded" src="i.png">Hi</mj-social-element></mj-social>`)
				opts := &options.RenderOpts{
					FontTracker:      options.NewFontTracker(),
					GlobalAttributes: globalsWithClass("branded", "font-family", "Verdana"),
				}
				return newSocialElementFromFixture(t, doc, opts)
			},
			expectedValue: "Verdana",
		},
		{
			name: "parent explicit attribute",
			buildElement: func(t *testing.T) *MJSocialElementComponent {
				doc := parseSocialFixture(t, `<mj-social font-family="Impact"><mj-social-element name="twitter" src="i.png">Hi</mj-social-element></mj-social>`)
				return newSocialElementFromFixture(t, doc, &options.RenderOpts{FontTracker: options.NewFontTracker()})
			},
			expectedValue: "Impact",
		},
		{
			// Uses a component-specific default scoped to "mj-social" (not
			// "mj-all"), which cannot resolve directly on the child element.
			// This is what actually exercises the parent-resolved branch, as
			// opposed to a global mj-all default the child would also match
			// on its own.
			name: "parent resolved via mj-social component-specific global",
			buildElement: func(t *testing.T) *MJSocialElementComponent {
				doc := parseSocialFixture(t, `<mj-social><mj-social-element name="twitter" src="i.png">Hi</mj-social-element></mj-social>`)
				opts := &options.RenderOpts{
					FontTracker:      options.NewFontTracker(),
					GlobalAttributes: globalsWithComponentDefault("mj-social", "font-family", "Courier New"),
				}
				return newSocialElementFromFixture(t, doc, opts)
			},
			expectedValue: "Courier New",
		},
		{
			name: "parent resolved via parent's own mj-class",
			buildElement: func(t *testing.T) *MJSocialElementComponent {
				doc := parseSocialFixture(t, `<mj-social mj-class="branded"><mj-social-element name="twitter" src="i.png">Hi</mj-social-element></mj-social>`)
				opts := &options.RenderOpts{
					FontTracker:      options.NewFontTracker(),
					GlobalAttributes: globalsWithClass("branded", "font-family", "Georgia"),
				}
				return newSocialElementFromFixture(t, doc, opts)
			},
			expectedValue: "Georgia",
		},
		{
			name: "child explicit overrides a different parent value",
			buildElement: func(t *testing.T) *MJSocialElementComponent {
				doc := parseSocialFixture(t, `<mj-social font-family="Impact"><mj-social-element name="twitter" font-family="Georgia" src="i.png">Hi</mj-social-element></mj-social>`)
				return newSocialElementFromFixture(t, doc, &options.RenderOpts{FontTracker: options.NewFontTracker()})
			},
			expectedValue: "Georgia",
		},
		{
			name: "component default",
			buildElement: func(t *testing.T) *MJSocialElementComponent {
				doc := parseSocialFixture(t, `<mj-social><mj-social-element name="twitter" src="i.png">Hi</mj-social-element></mj-social>`)
				return newSocialElementFromFixture(t, doc, &options.RenderOpts{FontTracker: options.NewFontTracker()})
			},
			expectedValue: "Ubuntu, Helvetica, Arial, sans-serif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//GIVEN: a social element wired to its parent, resolving font-family via a specific branch.
			elem := tt.buildElement(t)
			tracker := elem.RenderOpts.FontTracker

			//WHEN: resolveAttribute is called directly, bypassing the tracking wrapper.
			got := elem.resolveAttribute(constants.MJMLFontFamily)

			//THEN: the resolved value matches expectations and the tracker stays empty.
			if got != tt.expectedValue {
				t.Errorf("resolveAttribute() = %q, want %q", got, tt.expectedValue)
			}
			if fonts := tracker.GetFonts(); len(fonts) != 0 {
				t.Errorf("resolveAttribute() must not track fonts, but tracker has %v", fonts)
			}

			//WHEN: getAttribute (the tracking wrapper) resolves the same attribute.
			got2 := elem.getAttribute(constants.MJMLFontFamily)

			//THEN: it returns the identical value and tracks it exactly once.
			if got2 != tt.expectedValue {
				t.Errorf("getAttribute() = %q, want %q", got2, tt.expectedValue)
			}
			fonts := tracker.GetFonts()
			if len(fonts) != 1 || fonts[0] != tt.expectedValue {
				t.Errorf("getAttribute() tracked %v, want [%q]", fonts, tt.expectedValue)
			}
		})
	}
}

// parseSocialFixture parses a snippet containing a single mj-social element
// and returns the root document node.
func parseSocialFixture(t *testing.T, mjmlSnippet string) *parser.MJMLNode {
	t.Helper()
	doc, err := parser.ParseMJML(`<mjml><mj-body>` + mjmlSnippet + `</mj-body></mjml>`)
	if err != nil {
		t.Fatalf("failed to parse fixture: %v", err)
	}
	return doc
}

// newSocialElementFromFixture builds a wired-up MJSocialElementComponent
// (parent inheritance included) from a parsed fixture document.
func newSocialElementFromFixture(t *testing.T, doc *parser.MJMLNode, opts *options.RenderOpts) *MJSocialElementComponent {
	t.Helper()
	if opts.GlobalAttributes == nil {
		opts.GlobalAttributes = globals.NewGlobalAttributes()
	}

	socialNode := findMJMLElement(doc, "mj-social")
	if socialNode == nil {
		t.Fatal("mj-social not found in fixture")
	}
	elementNode := findMJMLElement(socialNode, "mj-social-element")
	if elementNode == nil {
		t.Fatal("mj-social-element not found in fixture")
	}

	social := NewMJSocialComponent(socialNode, opts)
	element := NewMJSocialElementComponent(elementNode, opts)
	element.InheritFromParent(social)
	return element
}
