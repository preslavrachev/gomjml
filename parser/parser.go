// Package parser provides MJML XML parsing functionality.
// It converts MJML markup into an Abstract Syntax Tree (AST) representation
// that can be used by the mjml package for component creation and rendering.
package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"regexp"
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

	// Wrap mj-text inner content in CDATA to preserve raw HTML
	processedContent = wrapMJTextContent(processedContent)

	decoder := xml.NewDecoder(strings.NewReader(processedContent))
	root, err := parseNode(decoder, xml.StartElement{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse MJML: %w", err)
	}
	return root, nil
}

// preprocessHTMLEntities replaces common HTML entities with Unicode characters
// and properly escapes ampersands in attribute values
func preprocessHTMLEntities(content string) string {
	// First, escape raw ampersands in attribute values that aren't part of valid entities
	result := escapeAttributeAmpersands(content)

	// Replace the most common HTML entities with Unicode characters
	// NOTE: We skip &amp; replacement here because we just escaped raw ampersands
	// to &amp; in the previous step. The XML parser will handle &amp; correctly.
	result = strings.ReplaceAll(result, "&copy;", "©")
	result = strings.ReplaceAll(result, "&reg;", "®")
	result = strings.ReplaceAll(result, "&trade;", "™")
	// result = strings.ReplaceAll(result, "&amp;", "&")  // SKIP THIS - let XML parser handle it
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

// escapeAttributeAmpersands escapes raw ampersands in XML attribute values
// that aren't part of valid HTML entities. This prevents XML parsing errors
// when URLs contain query parameters like "?param1=value1&param2=value2"
func escapeAttributeAmpersands(content string) string {
	// attrPattern is a regular expression that matches HTML/XML attribute assignments.
	// It captures both double-quoted and single-quoted attribute values, including escaped quotes within the value.
	// The pattern extracts the attribute name, the equals sign (with optional whitespace), and the quoted value.
	attrPattern := regexp.MustCompile(`(\w+)(\s*=\s*)"((?:[^"\\]|\\.)*)"|(\w+)(\s*=\s*)'((?:[^'\\]|\\.)*)'`)

	return attrPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the parts of the match
		parts := attrPattern.FindStringSubmatch(match)
		if len(parts) < 7 {
			return match // Return unchanged if parsing failed
		}

		var attrName, attrValue, quote, spacing string

		// Check which pattern matched (double quotes or single quotes)
		if parts[1] != "" && parts[2] != "" && parts[3] != "" {
			// Double quote pattern matched
			attrName = parts[1]
			spacing = parts[2]
			attrValue = parts[3]
			quote = `"`
		} else if parts[4] != "" && parts[5] != "" && parts[6] != "" {
			// Single quote pattern matched
			attrName = parts[4]
			spacing = parts[5]
			attrValue = parts[6]
			quote = `'`
		} else {
			return match // No valid pattern matched
		}

		// Escape ampersands in the attribute value that aren't part of valid entities
		escapedValue := escapeAmperands(attrValue)

		return attrName + spacing + quote + escapedValue + quote
	})
}

// escapeAmperands escapes ampersands that aren't part of valid HTML entities
func escapeAmperands(value string) string {
	// List of known HTML entities (without the & and ;)
	validEntities := []string{
		"amp", "lt", "gt", "quot", "apos", "nbsp", "copy", "reg", "trade",
		"ndash", "mdash", "hellip", "laquo", "raquo", "ldquo", "rdquo",
		"lsquo", "rsquo", "times", "divide",
	}

	// Also handle numeric character references like &#160; and &#xA0;
	numericEntityPattern := regexp.MustCompile(`&(#(?:\d+|x[0-9A-Fa-f]+));`)

	result := value
	i := 0

	for i < len(result) {
		if result[i] == '&' {
			// Check if this is a valid HTML entity
			isValidEntity := false

			// Check for numeric entities first
			matches := numericEntityPattern.FindStringSubmatch(result[i:])
			if len(matches) > 0 && strings.HasPrefix(result[i:], matches[0]) {
				i += len(matches[0])
				isValidEntity = true
				continue
			}

			// Check for named entities
			for _, entity := range validEntities {
				entityWithMarkers := "&" + entity + ";"
				if i+len(entityWithMarkers) <= len(result) &&
					result[i:i+len(entityWithMarkers)] == entityWithMarkers {
					i += len(entityWithMarkers)
					isValidEntity = true
					break
				}
			}

			if !isValidEntity {
				// This is a raw ampersand that needs escaping
				result = result[:i] + "&amp;" + result[i+1:]
				i += 5 // Move past "&amp;"
			}
		} else {
			i++
		}
	}

	return result
}

const (
	openNeedle   = "<mj-text"
	closeNeedle  = "</mj-text>"
	cdataStart   = "<![CDATA["
	cdataEnd     = "]]>"
	cdataEndSafe = "]]]]><![CDATA[>"
)

// wrapMJTextContent wraps the inner content of every <mj-text>...</mj-text>
// in a CDATA section and normalizes void tags inside. It is case-insensitive
// on tag names, handles attributes with quotes, and supports self-closing tags.
func wrapMJTextContent(content string) string {
	if content == "" {
		return ""
	}

	b := []byte(content)
	var out strings.Builder
	out.Grow(len(content) + 64)

	pos := 0
	for {
		idx := indexCI(b, []byte(openNeedle), pos)
		if idx < 0 {
			out.Write(b[pos:])
			break
		}

		out.Write(b[pos:idx])

		endStart, selfClosing := findTagEnd(b, idx)
		if endStart < 0 {
			out.Write(b[idx:])
			break
		}

		out.Write(b[idx:endStart])

		if selfClosing {
			pos = endStart
			continue
		}

		closeIdx := indexCI(b, []byte(closeNeedle), endStart)
		if closeIdx < 0 {
			out.Write(b[endStart:])
			break
		}

		inner := b[endStart:closeIdx]

		alreadyCDATA := bytes.HasPrefix(bytes.TrimLeft(inner, " \t\r\n"), []byte(cdataStart))

		inner = normalizeSelfClosingVoidTags(inner)

		if alreadyCDATA {
			out.Write(inner)
		} else {
			if bytes.Contains(inner, []byte(cdataEnd)) {
				inner = bytes.ReplaceAll(inner, []byte(cdataEnd), []byte(cdataEndSafe))
			}
			out.WriteString(cdataStart)
			out.Write(inner)
			out.WriteString(cdataEnd)
		}

		out.WriteString(closeNeedle)

		pos = closeIdx + len(closeNeedle)
	}

	return out.String()
}

// findTagEnd returns the index *after* the '>' of the start tag at 'start'
// and whether it was self-closing (<.../>). Respects single/double quotes.
func findTagEnd(b []byte, start int) (end int, selfClosing bool) {
	i := start
	inQuote := byte(0)
	for i < len(b) {
		c := b[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			i++
			continue
		}
		switch c {
		case '"', '\'':
			inQuote = c
			i++
		case '>':
			selfClosing = i > start && previousNonSpace(b, i-1) == '/'
			return i + 1, selfClosing
		default:
			i++
		}
	}
	return -1, false
}

// previousNonSpace returns the previous non-space byte at or before idx; 0 if none.
func previousNonSpace(b []byte, idx int) byte {
	for i := idx; i >= 0; i-- {
		if b[i] != ' ' && b[i] != '\t' && b[i] != '\n' && b[i] != '\r' {
			return b[i]
		}
	}
	return 0
}

// indexCI finds needle in haystack starting at 'from', ASCII case-insensitive.
// Avoids allocating by not lowercasing the whole string.
func indexCI(haystack, needle []byte, from int) int {
	if from < 0 {
		from = 0
	}
	n := len(needle)
	if n == 0 {
		return from
	}
	h := haystack
	max := len(h) - n
	for i := from; i <= max; i++ {
		if equalFoldASCII(h[i:i+n], needle) {
			return i
		}
	}
	return -1
}

func equalFoldASCII(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ai := a[i]
		bi := b[i]
		if ai == bi {
			continue
		}
		if 'A' <= ai && ai <= 'Z' {
			ai += 'a' - 'A'
		}
		if 'A' <= bi && bi <= 'Z' {
			bi += 'a' - 'A'
		}
		if ai != bi {
			return false
		}
	}
	return true
}

// normalizeSelfClosingVoidTags ensures that void HTML elements use a space before the
// closing slash (e.g., <br/> becomes <br />). This matches how XML parsers normalize
// self-closing tags.
var voidSelfClosingRe = regexp.MustCompile(`(?i)<(?:area|base|br|col|embed|hr|img|input|link|meta|param|source|track|wbr)([^>]*?)/>`)

func normalizeSelfClosingVoidTags(b []byte) []byte {
	return voidSelfClosingRe.ReplaceAllFunc(b, func(m []byte) []byte {
		base := bytes.TrimRight(m[:len(m)-2], " ")
		res := make([]byte, len(base)+3)
		copy(res, base)
		copy(res[len(base):], []byte(" />"))
		return res
	})
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
