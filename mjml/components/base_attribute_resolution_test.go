package components

import (
	"encoding/xml"
	"testing"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/globals"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// TestResolveAttribute_FontFamilyDoesNotTrack verifies that ResolveAttribute
// resolves font-family through every precedence branch (element, mj-class,
// global, default) without recording anything in the FontTracker. A future
// render-metadata pass depends on this pure-resolution contract.
func TestResolveAttribute_FontFamilyDoesNotTrack(t *testing.T) {
	tests := []struct {
		name          string
		node          *parser.MJMLNode
		globals       *globals.GlobalAttributes
		expectedValue string
	}{
		{
			name:          "element attribute",
			node:          nodeWithAttrs("mj-text", map[string]string{"font-family": "Georgia"}),
			expectedValue: "Georgia",
		},
		{
			name:          "mj-class attribute",
			node:          nodeWithAttrs("mj-text", map[string]string{"mj-class": "branded"}),
			globals:       globalsWithClass("branded", "font-family", "Verdana"),
			expectedValue: "Verdana",
		},
		{
			name:          "global mj-all attribute",
			node:          nodeWithAttrs("mj-text", nil),
			globals:       globalsWithAll("font-family", "Courier New"),
			expectedValue: "Courier New",
		},
		{
			name:          "component default",
			node:          nodeWithAttrs("mj-text", nil),
			expectedValue: "Ubuntu, Helvetica, Arial, sans-serif", // MJTextComponent's own default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//GIVEN: a component whose font-family resolves via a specific precedence branch.
			tracker := options.NewFontTracker()
			opts := &options.RenderOpts{FontTracker: tracker, GlobalAttributes: tt.globals}
			comp := NewMJTextComponent(tt.node, opts)

			//WHEN: ResolveAttribute is called directly, bypassing the tracking wrapper.
			got := comp.ResolveAttribute(comp, constants.MJMLFontFamily)

			//THEN: the resolved value matches expectations and the tracker stays empty.
			if got != tt.expectedValue {
				t.Errorf("ResolveAttribute() = %q, want %q", got, tt.expectedValue)
			}
			if fonts := tracker.GetFonts(); len(fonts) != 0 {
				t.Errorf("ResolveAttribute() must not track fonts, but tracker has %v", fonts)
			}
		})
	}
}

// TestGetAttributeWithDefault_FontFamilyTracksResolvedValue verifies that
// GetAttributeWithDefault records exactly the value it resolved into the
// FontTracker, across every precedence branch. This is the tracking half of
// the pure/tracking split introduced alongside ResolveAttribute.
func TestGetAttributeWithDefault_FontFamilyTracksResolvedValue(t *testing.T) {
	tests := []struct {
		name          string
		node          *parser.MJMLNode
		globals       *globals.GlobalAttributes
		expectedValue string
	}{
		{
			name:          "element attribute",
			node:          nodeWithAttrs("mj-text", map[string]string{"font-family": "Georgia"}),
			expectedValue: "Georgia",
		},
		{
			name:          "mj-class attribute",
			node:          nodeWithAttrs("mj-text", map[string]string{"mj-class": "branded"}),
			globals:       globalsWithClass("branded", "font-family", "Verdana"),
			expectedValue: "Verdana",
		},
		{
			name:          "global mj-all attribute",
			node:          nodeWithAttrs("mj-text", nil),
			globals:       globalsWithAll("font-family", "Courier New"),
			expectedValue: "Courier New",
		},
		{
			name:          "component default",
			node:          nodeWithAttrs("mj-text", nil),
			expectedValue: "Ubuntu, Helvetica, Arial, sans-serif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//GIVEN: a component whose font-family resolves via a specific precedence branch.
			tracker := options.NewFontTracker()
			opts := &options.RenderOpts{FontTracker: tracker, GlobalAttributes: tt.globals}
			comp := NewMJTextComponent(tt.node, opts)

			//WHEN: GetAttributeWithDefault resolves the font-family attribute.
			got := comp.GetAttributeWithDefault(comp, constants.MJMLFontFamily)

			//THEN: the returned value is tracked, and only that value is tracked.
			if got != tt.expectedValue {
				t.Errorf("GetAttributeWithDefault() = %q, want %q", got, tt.expectedValue)
			}
			fonts := tracker.GetFonts()
			if len(fonts) != 1 || fonts[0] != tt.expectedValue {
				t.Errorf("GetAttributeWithDefault() tracked %v, want [%q]", fonts, tt.expectedValue)
			}
		})
	}
}

// nodeWithAttrs builds a minimal MJMLNode of the given tag with the provided
// attributes, for tests that exercise attribute resolution directly.
func nodeWithAttrs(tag string, attrs map[string]string) *parser.MJMLNode {
	node := &parser.MJMLNode{XMLName: xml.Name{Local: tag}}
	for k, v := range attrs {
		node.Attrs = append(node.Attrs, xml.Attr{Name: xml.Name{Local: k}, Value: v})
	}
	return node
}

// globalsWithAll builds a GlobalAttributes store with a single mj-all entry.
func globalsWithAll(name, value string) *globals.GlobalAttributes {
	head := &parser.MJMLNode{
		XMLName: xml.Name{Local: "mj-head"},
		Children: []*parser.MJMLNode{
			{
				XMLName: xml.Name{Local: "mj-attributes"},
				Children: []*parser.MJMLNode{
					{
						XMLName: xml.Name{Local: "mj-all"},
						Attrs:   []xml.Attr{{Name: xml.Name{Local: name}, Value: value}},
					},
				},
			},
		},
	}
	ga := globals.NewGlobalAttributes()
	ga.ProcessAttributesFromHead(head)
	return ga
}

// globalsWithClass builds a GlobalAttributes store with a single named
// mj-class definition.
func globalsWithClass(className, attrName, attrValue string) *globals.GlobalAttributes {
	head := &parser.MJMLNode{
		XMLName: xml.Name{Local: "mj-head"},
		Children: []*parser.MJMLNode{
			{
				XMLName: xml.Name{Local: "mj-attributes"},
				Children: []*parser.MJMLNode{
					{
						XMLName: xml.Name{Local: "mj-class"},
						Attrs: []xml.Attr{
							{Name: xml.Name{Local: "name"}, Value: className},
							{Name: xml.Name{Local: attrName}, Value: attrValue},
						},
					},
				},
			},
		},
	}
	ga := globals.NewGlobalAttributes()
	ga.ProcessAttributesFromHead(head)
	return ga
}

// globalsWithComponentDefault builds a GlobalAttributes store with a single
// component-specific default (e.g. <mj-social font-family="...">), which
// resolves only for components whose tag name matches componentTag.
func globalsWithComponentDefault(componentTag, attrName, attrValue string) *globals.GlobalAttributes {
	head := &parser.MJMLNode{
		XMLName: xml.Name{Local: "mj-head"},
		Children: []*parser.MJMLNode{
			{
				XMLName: xml.Name{Local: "mj-attributes"},
				Children: []*parser.MJMLNode{
					{
						XMLName: xml.Name{Local: componentTag},
						Attrs:   []xml.Attr{{Name: xml.Name{Local: attrName}, Value: attrValue}},
					},
				},
			},
		},
	}
	ga := globals.NewGlobalAttributes()
	ga.ProcessAttributesFromHead(head)
	return ga
}
