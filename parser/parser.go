// Package parser provides MJML XML parsing functionality.
// It converts MJML markup into an Abstract Syntax Tree (AST) representation
// that can be used by the mjml package for component creation and rendering.
package parser

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/debug"
)

// Compiled regex for robust text splitting in mixed content
var mixedContentSplitRegex = regexp.MustCompile(`\s*[\r\n]+\s*`)

// MJMLNode represents a node in the MJML AST
type MJMLNode struct {
	XMLName  xml.Name
	Text     string
	Attrs    []xml.Attr
	Children []*MJMLNode
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
		XMLName:  start.Name,
		Attrs:    start.Attr,
		Children: make([]*MJMLNode, 0),
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
		raw, err := parseRawContent(decoder, node.XMLName)
		if err != nil {
			return nil, err
		}
		node.Text = raw
		return node, nil
	}

	var textBuilder strings.Builder

	for {
		tok, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			child, err := parseNode(decoder, t)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, child)

		case xml.EndElement:
			if t.Name == node.XMLName {
				node.Text = textBuilder.String()
				return node, nil
			}
			return nil, fmt.Errorf("unexpected end element: %s", t.Name.Local)

		case xml.CharData:
			textBuilder.Write(t)
		case xml.Comment:
			// Preserve comments as part of text content
			textBuilder.WriteString("<!--")
			textBuilder.WriteString(string(t))
			textBuilder.WriteString("-->")
		}
	}
}

// parseRawContent reads tokens until the matching end tag and returns the raw HTML content
func parseRawContent(decoder *xml.Decoder, name xml.Name) (string, error) {
	var builder strings.Builder
	depth := 1
	voidStack := make([]bool, 0)
	isVoid := func(tag string) bool {
		switch tag {
		case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr":
			return true
		}
		return false
	}
	for depth > 0 {
		tok, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			void := isVoid(t.Name.Local)
			voidStack = append(voidStack, void)
			builder.WriteString("<")
			builder.WriteString(t.Name.Local)
			for _, attr := range t.Attr {
				builder.WriteString(" ")
				builder.WriteString(attr.Name.Local)
				builder.WriteString("=\"")
				builder.WriteString(attr.Value)
				builder.WriteString("\"")
			}
			if void {
				builder.WriteString(" />")
			} else {
				builder.WriteString(">")
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				break
			}
			if len(voidStack) > 0 {
				void := voidStack[len(voidStack)-1]
				voidStack = voidStack[:len(voidStack)-1]
				if !void {
					builder.WriteString("</")
					builder.WriteString(t.Name.Local)
					builder.WriteString(">")
				}
			}
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

	if len(n.Children) == 0 {
		result := strings.TrimSpace(n.Text)
		debug.DebugLogWithData("parser", "text-only", "Returning plain text content", map[string]interface{}{
			"content": result,
		})
		return result
	}

	var result strings.Builder

	// For mixed content, we need to reconstruct the original structure
	// The parser splits text at child elements, so we need to interleave them
	textParts := mixedContentSplitRegex.Split(n.Text, -1)
	cleanTextParts := make([]string, 0, len(textParts))

	// Clean up text parts (remove excessive whitespace but preserve structure)
	for _, part := range textParts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			cleanTextParts = append(cleanTextParts, trimmed)
		}
	}

	// Write content with children interspersed
	// Note: This assumes children are inline elements within the text flow
	if len(cleanTextParts) > 0 {
		result.WriteString(cleanTextParts[0])
	}

	// Add child elements with remaining text parts
	for i, child := range n.Children {
		// Render child element as HTML
		result.WriteString("<")
		result.WriteString(child.XMLName.Local)

		// Add attributes if any
		for _, attr := range child.Attrs {
			result.WriteString(" ")
			result.WriteString(attr.Name.Local)
			result.WriteString("=\"")
			result.WriteString(attr.Value)
			result.WriteString("\"")
		}
		result.WriteString(">")

		// Add child's content (recursively handle mixed content)
		result.WriteString(child.GetMixedContent())

		// Close tag
		result.WriteString("</")
		result.WriteString(child.XMLName.Local)
		result.WriteString(">")

		// Add remaining text part if available
		if i+1 < len(cleanTextParts) {
			result.WriteString(cleanTextParts[i+1])
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
