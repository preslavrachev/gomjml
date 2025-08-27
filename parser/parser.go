// Package parser provides MJML XML parsing functionality.
// It converts MJML markup into an Abstract Syntax Tree (AST) representation
// that can be used by the mjml package for component creation and rendering.
package parser

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/debug"
)

// htmlVoidElements contains HTML elements that do not require closing tags.
//
// Reference: https://html.spec.whatwg.org/\#void-elements
var htmlVoidElements = map[string]struct{}{
	"area":   {},
	"base":   {},
	"br":     {},
	"col":    {},
	"embed":  {},
	"hr":     {},
	"img":    {},
	"input":  {},
	"link":   {},
	"meta":   {},
	"param":  {},
	"source": {},
	"track":  {},
	"wbr":    {},
}

// isVoidHTMLElement reports whether the provided tag name is an HTML void element.
func isVoidHTMLElement(tag string) bool {
	_, ok := htmlVoidElements[strings.ToLower(tag)]
	return ok
}

// MJMLNode represents a node in the MJML AST
type MJMLNode struct {
	XMLName  xml.Name
	Text     string
	Attrs    []xml.Attr
	Children []*MJMLNode
	// MixedContent preserves the interleaving order of text nodes and child elements
	// as they originally appeared in the MJML source. Each entry contains either
	// a text segment or a pointer to a child node.
	MixedContent []MixedContentPart
}

// MixedContentPart represents either a piece of text or a child node in the
// mixed content sequence of an MJML node.
type MixedContentPart struct {
	Text string
	Node *MJMLNode
}

// ParseMJML parses an MJML string into an AST
func ParseMJML(mjmlContent string) (*MJMLNode, error) {
	// Pre-process HTML entities that XML parser doesn't handle
	processedContent := preprocessHTMLEntities(mjmlContent)

	decoder := xml.NewDecoder(strings.NewReader(processedContent))
	root, err := parseNode(decoder, xml.StartElement{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse MJML: %w", err)
	}
	return root, nil
}

// preprocessHTMLEntities replaces common HTML entities with Unicode characters
func preprocessHTMLEntities(content string) string {
	// Replace the most common HTML entities with Unicode characters
	result := content
	result = strings.ReplaceAll(result, "&copy;", "©")
	result = strings.ReplaceAll(result, "&reg;", "®")
	result = strings.ReplaceAll(result, "&trade;", "™")
	result = strings.ReplaceAll(result, "&amp;", "&")
	result = strings.ReplaceAll(result, "&lt;", "<")
	result = strings.ReplaceAll(result, "&gt;", ">")
	result = strings.ReplaceAll(result, "&quot;", `"`)
	result = strings.ReplaceAll(result, "&apos;", "'")
	result = strings.ReplaceAll(result, "&nbsp;", "\u00A0") // Unicode non-breaking space
	result = strings.ReplaceAll(result, "&#xA0;", "\u00A0") // Numeric character reference for non-breaking space
	result = strings.ReplaceAll(result, "&#160;", "\u00A0") // Decimal numeric reference for non-breaking space
	result = strings.ReplaceAll(result, "&ndash;", "–")
	result = strings.ReplaceAll(result, "&mdash;", "—")
	result = strings.ReplaceAll(result, "&hellip;", "…")

	return result
}

// parseNode recursively parses XML nodes
func parseNode(decoder *xml.Decoder, start xml.StartElement) (*MJMLNode, error) {
	node := &MJMLNode{
		XMLName:      start.Name,
		Attrs:        start.Attr,
		Children:     make([]*MJMLNode, 0),
		MixedContent: make([]MixedContentPart, 0),
	}

	// If this is called with empty start element, get the first element
	if start.Name.Local == "" {
		tok, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		if se, ok := tok.(xml.StartElement); ok {
			node.XMLName = se.Name
			node.Attrs = se.Attr
		} else {
			return nil, fmt.Errorf("expected start element")
		}
	}

	// Special handling for mj-raw: capture original inner content including comments
	if node.XMLName.Local == "mj-raw" {
		raw, err := parseRawContent(decoder)
		if err != nil {
			return nil, err
		}
		node.Text = raw
		node.MixedContent = []MixedContentPart{{Text: raw}}
		return node, nil
	}

	var textBuilder strings.Builder
	var segmentBuilder strings.Builder

	flushSegment := func() {
		if segmentBuilder.Len() > 0 {
			text := segmentBuilder.String()
			node.MixedContent = append(node.MixedContent, MixedContentPart{Text: text})
			segmentBuilder.Reset()
		}
	}

	for {
		tok, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			flushSegment()
			child, err := parseNode(decoder, t)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, child)
			node.MixedContent = append(node.MixedContent, MixedContentPart{Node: child})

		case xml.EndElement:
			if t.Name == node.XMLName {
				flushSegment()
				node.Text = textBuilder.String()
				return node, nil
			}
			return nil, fmt.Errorf("unexpected end element: %s", t.Name.Local)

		case xml.CharData:
			textBuilder.Write(t)
			segmentBuilder.Write(t)
		case xml.Comment:
			// Preserve comments as part of text content
			textBuilder.WriteString("<!--")
			textBuilder.WriteString(string(t))
			textBuilder.WriteString("-->")
			segmentBuilder.WriteString("<!--")
			segmentBuilder.WriteString(string(t))
			segmentBuilder.WriteString("-->")
		}
	}
}

// parseRawContent reads tokens until the matching end tag and returns the raw HTML content
func parseRawContent(decoder *xml.Decoder) (string, error) {
	origStrict := decoder.Strict
	decoder.Strict = false
	defer func() { decoder.Strict = origStrict }()

	var builder strings.Builder
	depth := 1
	tagStack := make([]string, 0)
	for depth > 0 {
		tok, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			tagName := t.Name.Local
			builder.WriteString("<")
			builder.WriteString(tagName)
			for _, attr := range t.Attr {
				builder.WriteString(" ")
				builder.WriteString(attr.Name.Local)
				builder.WriteString("=\"")
				builder.WriteString(attr.Value)
				builder.WriteString("\"")
			}
			builder.WriteString(">")

			// Track depth for all start elements
			depth++

			if !isVoidHTMLElement(tagName) {
				// Only non-void elements participate in stack tracking
				tagStack = append(tagStack, tagName)
			}

		case xml.EndElement:
			tagName := t.Name.Local

			if isVoidHTMLElement(tagName) {
				// Ignore end tags for void elements
				depth--
				if depth == 0 {
					break
				}
				continue
			}

			depth--
			if len(tagStack) > 0 {
				lastTag := tagStack[len(tagStack)-1]
				if lastTag == tagName {
					tagStack = tagStack[:len(tagStack)-1]
				}
			}

			if depth == 0 {
				break
			}

			builder.WriteString("</")
			builder.WriteString(tagName)
			builder.WriteString(">")
		case xml.CharData:
			builder.WriteString(string(t))
		case xml.Comment:
			builder.WriteString("<!--")
			builder.WriteString(string(t))
			builder.WriteString("-->")
		case xml.Directive:
			builder.WriteString("<!")
			builder.WriteString(string(t))
			builder.WriteString(">")
		case xml.ProcInst:
			builder.WriteString("<")
			builder.WriteString("?")
			builder.WriteString(t.Target)
			if len(t.Inst) > 0 {
				builder.WriteString(" ")
				builder.Write(t.Inst)
			}
			builder.WriteString("?>")
		}
	}
	return builder.String(), nil
}

// GetAttribute retrieves an attribute value by name
func (n *MJMLNode) GetAttribute(name string) string {
	for _, attr := range n.Attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

// GetTagName returns the local name of the XML tag
func (n *MJMLNode) GetTagName() string {
	return n.XMLName.Local
}

// GetTextContent returns the trimmed text content
func (n *MJMLNode) GetTextContent() string {
	return strings.TrimSpace(n.Text)
}

// GetMixedContent returns the full mixed content including HTML child elements
// This reconstructs the original content like "Share <b>test</b> hi" from the AST
func (n *MJMLNode) GetMixedContent() string {
	debug.DebugLogWithData("parser", "mixed-content", "Processing mixed content", map[string]interface{}{
		"tag_name":       n.XMLName.Local,
		"plain_text":     n.Text,
		"children_count": len(n.Children),
		"has_children":   len(n.Children) > 0,
	})

	if len(n.MixedContent) == 0 {
		result := strings.TrimSpace(n.Text)
		debug.DebugLogWithData("parser", "text-only", "Returning plain text content", map[string]interface{}{
			"content": result,
		})
		return result
	}

	var result strings.Builder
	for i, part := range n.MixedContent {
		if part.Node != nil {
			tag := part.Node.XMLName.Local
			result.WriteString("<")
			result.WriteString(tag)
			for _, attr := range part.Node.Attrs {
				result.WriteString(" ")
				result.WriteString(attr.Name.Local)
				result.WriteString("=\"")
				result.WriteString(attr.Value)
				result.WriteString("\"")
			}
			if isVoidHTMLElement(tag) {
				result.WriteString(" />")
				continue
			}
			result.WriteString(">")
			result.WriteString(part.Node.GetMixedContent())
			result.WriteString("</")
			result.WriteString(tag)
			result.WriteString(">")
		} else {
			text := part.Text
			if i == 0 {
				text = strings.TrimLeft(text, " \n\r\t")
			}
			if i == len(n.MixedContent)-1 {
				text = strings.TrimRight(text, " \n\r\t")
			}
			result.WriteString(text)
		}
	}

	finalResult := strings.TrimSpace(result.String())
	debug.DebugLogWithData("parser", "mixed-complete", "Mixed content reconstructed", map[string]interface{}{
		"original_text":     n.Text,
		"final_content":     finalResult,
		"children_rendered": len(n.Children),
	})
	return finalResult
}

// FindFirstChild finds the first child with the given tag name
func (n *MJMLNode) FindFirstChild(tagName string) *MJMLNode {
	for _, child := range n.Children {
		if child.GetTagName() == tagName {
			return child
		}
	}
	return nil
}

// FindAllChildren finds all children with the given tag name
func (n *MJMLNode) FindAllChildren(tagName string) []*MJMLNode {
	var result []*MJMLNode
	for _, child := range n.Children {
		if child.GetTagName() == tagName {
			result = append(result, child)
		}
	}
	return result
}
