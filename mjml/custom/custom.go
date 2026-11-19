// Package custom defines MJML tags of your own. A custom tag is a composite:
// it expands into built-in MJML before the document is rendered, so its
// output is exactly what the equivalent hand-written MJML would produce.
//
// Pass a Registry to a render with mjml.WithComponents. There is no global
// registry, so renders with different registries can run concurrently.
package custom

import (
	"encoding/xml"
	"fmt"
	"html"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/preslavrachev/gomjml/parser"
)

// Node is an MJML AST node.
type Node = parser.MJMLNode

// Call describes one occurrence of a custom tag.
type Call struct {
	// Node is the tag as written. It may be shared with other renders through
	// the AST cache, so neither it nor anything under it may be modified,
	// though its nodes may be placed in the result. Its Children are MJML to
	// place in the result, for a container tag; for a tag that holds HTML,
	// Node.GetMixedContent returns the inner markup, entities decoded.
	Node *Node
	// Parent is the tag name of the element the custom tag appears in.
	Parent string
	// Attrs holds the tag's attributes resolved in MJML's order: the
	// element's own, then mj-class, then mj-attributes for the tag, then
	// Def.Defaults. mj-all is left to the built-in tags of the expansion.
	Attrs map[string]string
}

// Expander returns the nodes that replace a custom tag. The result may hold
// further custom tags. It is copied before use, so an Expander may return
// the same parsed nodes on every call.
type Expander func(Call) ([]*Node, error)

// Def defines a custom tag.
type Def struct {
	Expand Expander
	// Attributes lists the attributes the tag accepts, besides those in
	// Defaults. Any other is reported as invalid, as on a built-in tag. Nil
	// accepts every attribute.
	Attributes []string
	// Defaults are attribute values used when neither the element, mj-class
	// nor mj-attributes sets them.
	Defaults map[string]string
	// HeadStyle is CSS added once to the document head when the tag is used,
	// ahead of the document's own mj-style.
	HeadStyle string
}

// Registry holds custom tag definitions. It is safe for concurrent use,
// including Register during renders.
type Registry struct {
	mu   sync.RWMutex
	defs map[string]Def
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{defs: make(map[string]Def)}
}

// reserved holds every MJML tag, including those gomjml does not implement
// yet, so that implementing one later cannot silently displace a custom tag.
var reserved = map[string]bool{
	"mjml": true, "mj-head": true, "mj-body": true, "mj-attributes": true, "mj-all": true,
	"mj-class": true, "mj-breakpoint": true, "mj-font": true, "mj-html-attributes": true,
	"mj-html-attribute": true, "mj-selector": true, "mj-preview": true, "mj-style": true,
	"mj-title": true, "mj-include": true, "mj-accordion": true, "mj-accordion-element": true,
	"mj-accordion-title": true, "mj-accordion-text": true, "mj-button": true, "mj-carousel": true,
	"mj-carousel-image": true, "mj-column": true, "mj-divider": true, "mj-group": true,
	"mj-hero": true, "mj-image": true, "mj-navbar": true, "mj-navbar-link": true, "mj-raw": true,
	"mj-section": true, "mj-social": true, "mj-social-element": true, "mj-spacer": true,
	"mj-table": true, "mj-text": true, "mj-wrapper": true,
}

// Register defines tag, which must not be an MJML tag.
func (r *Registry) Register(tag string, def Def) error {
	if tag == "" {
		return fmt.Errorf("custom component: empty tag name")
	}
	if reserved[tag] {
		return fmt.Errorf("custom component <%s>: an MJML tag cannot be redefined", tag)
	}
	if def.Expand == nil {
		return fmt.Errorf("custom component <%s>: Expand is nil", tag)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.defs[tag]; dup {
		return fmt.Errorf("custom component <%s>: already registered", tag)
	}
	r.defs[tag] = def
	return nil
}

// Lookup returns the definition of tag.
func (r *Registry) Lookup(tag string) (Def, bool) {
	if r == nil {
		return Def{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	def, ok := r.defs[tag]
	return def, ok
}

// Len reports the number of registered tags.
func (r *Registry) Len() int {
	if r == nil {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.defs)
}

// Parse parses a fragment of trusted MJML markup, of any number of sibling
// elements, for an Expander to return. Never build src from data: the
// parser decodes entities such as &lt; before parsing, so escaping a value
// does not stop it becoming markup. Build nodes that carry data with Element
// and Text. Line numbers are cleared so that an error inside the expansion is
// reported at the custom tag's line.
func Parse(src string) ([]*Node, error) {
	root, err := parser.ParseMJML("<mj-fragment>" + src + "</mj-fragment>")
	if err != nil {
		return nil, err
	}
	clearLines(root)
	return root.Children, nil
}

func clearLines(n *Node) {
	n.LineNumber = 0
	for _, c := range n.Children {
		clearLines(c)
	}
}

// Element builds a node without parsing, so the values it holds can never
// become markup. Attribute values are used as given, as the parser leaves
// them; gomjml writes attribute values into the HTML unescaped, so keep
// untrusted data in Text. content is Text and Element nodes.
func Element(tag string, attrs map[string]string, content ...*Node) *Node {
	n := &Node{XMLName: xml.Name{Local: tag}}
	for _, name := range slices.Sorted(maps.Keys(attrs)) {
		n.Attrs = append(n.Attrs, xml.Attr{Name: xml.Name{Local: name}, Value: attrs[name]})
	}
	var text strings.Builder
	for _, c := range content {
		switch {
		case c == nil:
		case c.XMLName.Local == "":
			text.WriteString(c.Text)
			n.MixedContent = append(n.MixedContent, parser.MixedContentPart{Text: c.Text})
		default:
			n.Children = append(n.Children, c)
			n.MixedContent = append(n.MixedContent, parser.MixedContentPart{Node: c})
		}
	}
	n.Text = text.String()
	return n
}

// Text is literal text for an Element. It is escaped, so it renders as
// written and never as markup.
func Text(s string) *Node {
	return &Node{Text: html.EscapeString(s)}
}
