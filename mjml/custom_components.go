package mjml

import (
	"encoding/xml"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/components"
	"github.com/preslavrachev/gomjml/mjml/custom"
	"github.com/preslavrachev/gomjml/parser"
)

// WithComponents renders with the custom tags defined in reg. Custom tags
// expand in mj-body before the component tree is built.
func WithComponents(reg *custom.Registry) RenderOption {
	return func(opts *RenderOpts) {
		opts.Components = reg
	}
}

// maxExpansionDepth bounds custom tags expanding into custom tags, the only
// way a registry can recurse forever.
const maxExpansionDepth = 32

// opaqueTags hold HTML rather than MJML, so custom tags are never looked for
// inside them. mj-head is opaque because mj-attributes is read before
// expansion.
var opaqueTags = map[string]bool{
	"mj-head": true, "mj-text": true, "mj-button": true, "mj-table": true, "mj-raw": true,
	"mj-navbar-link": true, "mj-accordion-title": true, "mj-accordion-text": true,
	"mj-social-element": true,
}

// createExpandedComponent is CreateComponent after custom tags have expanded.
func createExpandedComponent(ast *MJMLNode, opts *RenderOpts) (Component, error) {
	expanded, err := expandCustomComponents(ast, opts)
	if err != nil {
		return nil, err
	}
	return CreateComponent(expanded, opts)
}

// expandCustomComponents returns root with every custom tag replaced by its
// expansion. root is never modified, since the AST cache may share it.
func expandCustomComponents(root *MJMLNode, opts *RenderOpts) (*MJMLNode, error) {
	if opts.Components.Len() == 0 {
		return root, nil
	}
	x := &expander{opts: opts, allowed: map[string]map[string]struct{}{}}
	out, err := x.children(root, 0)
	if err != nil {
		return nil, err
	}
	if len(x.headStyles) > 0 && out.GetTagName() == "mjml" {
		out = withHeadStyle(out, strings.Join(x.headStyles, "\n"))
	}
	return out, nil
}

type expander struct {
	opts       *RenderOpts
	allowed    map[string]map[string]struct{}
	headStyles []string
}

// children expands n's children, copying n only if one of them changed.
func (x *expander) children(n *MJMLNode, depth int) (*MJMLNode, error) {
	var replaced map[*MJMLNode][]*MJMLNode
	for _, c := range n.Children {
		out, err := x.expand(c, n.GetTagName(), depth)
		if err != nil {
			return nil, err
		}
		if len(out) != 1 || out[0] != c {
			if replaced == nil {
				replaced = map[*MJMLNode][]*MJMLNode{}
			}
			replaced[c] = out
		}
	}
	if replaced == nil {
		return n, nil
	}

	cp := *n
	cp.Children = make([]*MJMLNode, 0, len(n.Children))
	for _, c := range n.Children {
		if out, ok := replaced[c]; ok {
			cp.Children = append(cp.Children, out...)
		} else {
			cp.Children = append(cp.Children, c)
		}
	}
	cp.MixedContent = make([]parser.MixedContentPart, 0, len(n.MixedContent))
	for _, part := range n.MixedContent {
		out, ok := replaced[part.Node]
		if !ok {
			cp.MixedContent = append(cp.MixedContent, part)
			continue
		}
		for _, o := range out {
			cp.MixedContent = append(cp.MixedContent, parser.MixedContentPart{Node: o})
		}
	}
	return &cp, nil
}

func (x *expander) expand(n *MJMLNode, parentTag string, depth int) ([]*MJMLNode, error) {
	tag := n.GetTagName()
	def, ok := x.opts.Components.Lookup(tag)
	if !ok {
		if opaqueTags[tag] {
			return []*MJMLNode{n}, nil
		}
		out, err := x.children(n, depth)
		return []*MJMLNode{out}, err
	}

	line := n.LineNumber
	if depth >= maxExpansionDepth {
		return nil, fmt.Errorf("%s expands deeper than %d levels", describe(n), maxExpansionDepth)
	}
	x.validate(n, def)

	// Clipped, so that an append in the Expander cannot write into the
	// backing arrays of an AST the cache shares.
	call := *n
	call.Children = slices.Clip(n.Children)
	call.MixedContent = slices.Clip(n.MixedContent)
	call.Attrs = slices.Clip(n.Attrs)
	nodes, err := def.Expand(custom.Call{Node: &call, Parent: parentTag, Attrs: x.resolve(n, def)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", describe(n), err)
	}
	if def.HeadStyle != "" && !slices.Contains(x.headStyles, def.HeadStyle) {
		x.headStyles = append(x.headStyles, def.HeadStyle)
	}

	var out []*MJMLNode
	for _, node := range nodes {
		if node == nil || node.GetTagName() == "" {
			return nil, fmt.Errorf("%s: expanded to a nil node or to text outside an element", describe(n))
		}
		expanded, err := x.expand(copyExpansion(node, line, n), parentTag, depth+1)
		if err != nil {
			return nil, err
		}
		out = append(out, expanded...)
	}
	return out, nil
}

// describe names a custom tag for an error. The parser records a line only
// for elements with attributes.
func describe(n *MJMLNode) string {
	if line := n.LineNumber; line > 0 {
		return fmt.Sprintf("line %d: <%s>", line, n.GetTagName())
	}
	return fmt.Sprintf("<%s>", n.GetTagName())
}

// resolve follows BaseComponent.ResolveAttribute's order, leaving out mj-all.
func (x *expander) resolve(n *MJMLNode, def custom.Def) map[string]string {
	attrs := maps.Clone(def.Defaults)
	if attrs == nil {
		attrs = make(map[string]string, len(n.Attrs))
	}
	if ga := x.opts.GlobalAttributes; ga != nil {
		maps.Copy(attrs, ga.GetComponentAttributes(n.GetTagName()))
		var cssClasses []string
		for _, class := range strings.Fields(n.GetAttribute("mj-class")) {
			for name, value := range ga.GetClassAttributes(class) {
				if name == "css-class" {
					cssClasses = append(cssClasses, value)
				} else {
					attrs[name] = value
				}
			}
		}
		if len(cssClasses) > 0 {
			attrs["css-class"] = strings.Join(cssClasses, " ")
		}
	}
	for _, a := range n.Attrs {
		if a.Name.Local != "mj-class" {
			attrs[a.Name.Local] = a.Value
		}
	}
	return attrs
}

func (x *expander) validate(n *MJMLNode, def custom.Def) {
	if def.Attributes == nil {
		return
	}
	tag := n.GetTagName()
	allowed, ok := x.allowed[tag]
	if !ok {
		allowed = make(map[string]struct{}, len(def.Attributes)+len(def.Defaults))
		for _, name := range def.Attributes {
			allowed[name] = struct{}{}
		}
		for name := range def.Defaults {
			allowed[name] = struct{}{}
		}
		x.allowed[tag] = allowed
	}
	components.ValidateAttributes(n, allowed, x.opts)
}

// copyExpansion copies a node an Expander returned, so that it may return
// shared nodes, and gives nodes without a line the custom tag's line. The
// custom tag's own children are placed as they are.
func copyExpansion(n *MJMLNode, line int, tag *MJMLNode) *MJMLNode {
	if slices.Contains(tag.Children, n) {
		return n
	}
	cp := *n
	if cp.LineNumber == 0 {
		cp.LineNumber = line
	}
	copied := make(map[*MJMLNode]*MJMLNode, len(n.Children))
	cp.Children = make([]*MJMLNode, len(n.Children))
	for i, c := range n.Children {
		cp.Children[i] = copyExpansion(c, line, tag)
		copied[c] = cp.Children[i]
	}
	cp.MixedContent = slices.Clone(n.MixedContent)
	for i, part := range cp.MixedContent {
		if c, ok := copied[part.Node]; ok {
			cp.MixedContent[i].Node = c
		}
	}
	return &cp
}

// withHeadStyle puts css ahead of the document's own mj-style, where MJML
// puts component head styles.
func withHeadStyle(root *MJMLNode, css string) *MJMLNode {
	style := &MJMLNode{XMLName: xml.Name{Local: "mj-style"}, Text: css}
	style.MixedContent = []parser.MixedContentPart{{Text: css}}

	head := &MJMLNode{XMLName: xml.Name{Local: "mj-head"}}
	existing := root.FindFirstChild("mj-head")
	if existing != nil {
		*head = *existing
	}
	head.Children = append([]*MJMLNode{style}, head.Children...)
	head.MixedContent = append([]parser.MixedContentPart{{Node: style}}, head.MixedContent...)

	cp := *root
	if existing == nil {
		cp.Children = append([]*MJMLNode{head}, root.Children...)
		cp.MixedContent = append([]parser.MixedContentPart{{Node: head}}, root.MixedContent...)
		return &cp
	}
	cp.Children = slices.Clone(root.Children)
	cp.MixedContent = slices.Clone(root.MixedContent)
	for i, c := range cp.Children {
		if c == existing {
			cp.Children[i] = head
		}
	}
	for i, part := range cp.MixedContent {
		if part.Node == existing {
			cp.MixedContent[i].Node = head
		}
	}
	return &cp
}
