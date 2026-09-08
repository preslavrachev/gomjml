package components

import (
	"testing"

	"github.com/preslavrachev/gomjml/mjml/options"
)

// TestRenderMetadataFollowsContainerChildSelection verifies that
// CollectRenderMetadata only reports fonts reachable through the exact
// children each container actually renders: the last title/text under an
// accordion element (mirroring its duplicate-child overwrite behavior), and
// only the child types a container's renderable-children filter accepts
// (raw elements and mj-column under mj-group), never a font merely present
// somewhere in the source tree.
func TestRenderMetadataFollowsContainerChildSelection(t *testing.T) {
	tests := []struct {
		name        string
		buildRoot   func(t *testing.T) Component
		wantFonts   []string
		unwantFonts []string
	}{
		{
			name: "duplicate accordion title and text keep only the last of each",
			buildRoot: func(t *testing.T) Component {
				opts := &options.RenderOpts{FontTracker: options.NewFontTracker()}
				element := NewMJAccordionElementComponent(nodeWithAttrs("mj-accordion-element", nil), opts)
				firstTitle := NewMJAccordionTitleComponent(nodeWithAttrs("mj-accordion-title", map[string]string{"font-family": "Lato, Arial, sans-serif"}), opts)
				secondTitle := NewMJAccordionTitleComponent(nodeWithAttrs("mj-accordion-title", map[string]string{"font-family": "Roboto, Arial, sans-serif"}), opts)
				firstText := NewMJAccordionTextComponent(nodeWithAttrs("mj-accordion-text", map[string]string{"font-family": "Montserrat, Arial, sans-serif"}), opts)
				secondText := NewMJAccordionTextComponent(nodeWithAttrs("mj-accordion-text", map[string]string{"font-family": "Open Sans, Arial, sans-serif"}), opts)
				element.Children = []Component{firstTitle, secondTitle, firstText, secondText}
				return element
			},
			// Render only ever sees the last title and last text among an
			// accordion-element's children (see renderedTitleAndText).
			wantFonts:   []string{"Roboto, Arial, sans-serif", "Open Sans, Arial, sans-serif"},
			unwantFonts: []string{"Lato, Arial, sans-serif", "Montserrat, Arial, sans-serif"},
		},
		{
			name: "an unsupported child directly under mj-group is never reached",
			buildRoot: func(t *testing.T) Component {
				opts := &options.RenderOpts{FontTracker: options.NewFontTracker()}
				group := NewMJGroupComponent(nodeWithAttrs("mj-group", nil), opts)

				column := NewMJColumnComponent(nodeWithAttrs("mj-column", nil), opts)
				column.Children = []Component{
					NewMJTextComponent(nodeWithAttrs("mj-text", map[string]string{"font-family": "Roboto, Arial, sans-serif"}), opts),
				}

				// mj-group only ever renders raw elements and mj-column children
				// (see MJGroupComponent.renderableChildren); a stray mj-text
				// child placed directly under it, as could happen from a
				// malformed tree, must never be reached.
				strayText := NewMJTextComponent(nodeWithAttrs("mj-text", map[string]string{"font-family": "Lato, Arial, sans-serif"}), opts)

				group.Children = []Component{column, strayText}
				return group
			},
			wantFonts:   []string{"Roboto, Arial, sans-serif"},
			unwantFonts: []string{"Lato, Arial, sans-serif"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//GIVEN: a component tree containing both renderable and non-renderable font sources.
			root := tt.buildRoot(t)

			//WHEN: CollectRenderMetadata walks the tree.
			meta := CollectRenderMetadata(root)
			got := make(map[string]bool)
			for _, family := range meta.FontFamilies() {
				got[family] = true
			}

			//THEN: only fonts reachable through the actual render path are present.
			for _, want := range tt.wantFonts {
				if !got[want] {
					t.Errorf("expected font %q to be collected, got %v", want, meta.FontFamilies())
				}
			}
			for _, unwant := range tt.unwantFonts {
				if got[unwant] {
					t.Errorf("did not expect unreachable font %q to be collected, got %v", unwant, meta.FontFamilies())
				}
			}
		})
	}
}
