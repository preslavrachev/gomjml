// Package parser provides MJML XML parsing functionality.
// It converts MJML markup into an Abstract Syntax Tree (AST) representation
// that can be used by the mjml package for component creation and rendering.
package parser

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// MJMLNode represents a node in the MJML AST
type MJMLNode struct {
	XMLName  xml.Name
	Text     string
	Attrs    []xml.Attr
	Children []*MJMLNode
}

// ParseMJML parses an MJML string into an AST
func ParseMJML(mjmlContent string) (*MJMLNode, error) {
	decoder := xml.NewDecoder(strings.NewReader(mjmlContent))
	root, err := parseNode(decoder, xml.StartElement{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse MJML: %w", err)
	}
	return root, nil
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
		}
	}
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
