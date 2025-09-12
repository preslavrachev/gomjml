// Package parser provides MJML XML parsing functionality.
// It converts MJML markup into an Abstract Syntax Tree (AST) representation
// that can be used by the mjml package for component creation and rendering.
package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"regexp"
	"sort"
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

// buildVoidElementsRegexPattern creates a regex pattern from the htmlVoidElements map
func buildVoidElementsRegexPattern() string {
	elements := make([]string, 0, len(htmlVoidElements))
	for element := range htmlVoidElements {
		elements = append(elements, element)
	}
	sort.Strings(elements) // Ensure deterministic order
	return `(?i)<(?:` + strings.Join(elements, "|") + `)([^>]*?)/>`
}

// namedHTMLEntities lists HTML entities that should remain unescaped.
var namedHTMLEntities = map[string]struct{}{
	"amp":    {},
	"lt":     {},
	"gt":     {},
	"quot":   {},
	"apos":   {},
	"nbsp":   {},
	"copy":   {},
	"reg":    {},
	"trade":  {},
	"ndash":  {},
	"mdash":  {},
	"hellip": {},
	"laquo":  {},
	"raquo":  {},
	"ldquo":  {},
	"rdquo":  {},
	"lsquo":  {},
	"rsquo":  {},
	"times":  {},
	"divide": {},
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

// AIDEV-NOTE: mjml-spec-structure; MJML document structure per official spec
// Minimal valid MJML structure:
// <mjml>
//   <mj-body> (required)
//     <!-- at least one component -->
//   </mj-body>
// </mjml>
// The <mj-head> section is OPTIONAL and can be omitted entirely.

// ParseMJML parses an MJML string into an AST
func ParseMJML(mjmlContent string) (*MJMLNode, error) {
	// AIDEV-NOTE: comment-preservation; Preserve all XML comments for MRML compatibility
	// MRML preserves regular XML comments and wraps them with MSO conditionals
	processedContent := stripNonMSOComments(mjmlContent)

	// Pre-process HTML entities that XML parser doesn't handle
	processedContent = preprocessHTMLEntities(processedContent)

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
// and properly escapes ampersands in attribute values. Raw ampersands are first
// escaped to &amp; for XML safety, then most entities are replaced with Unicode.
// The &amp; entities are left for the XML parser to handle, preventing re-introduction
// of invalid raw ampersands that would break XML parsing.
func preprocessHTMLEntities(content string) string {
	// First, escape raw ampersands in attribute values that aren't part of valid entities
	result := escapeAttributeAmpersands(content)

	// Replace the most common HTML entities with Unicode characters
	// NOTE: &amp; entities are intentionally preserved - the XML parser will convert
	// them to raw ampersands safely after parsing, maintaining XML validity.
	result = strings.ReplaceAll(result, "&copy;", "©")
	result = strings.ReplaceAll(result, "&reg;", "®")
	result = strings.ReplaceAll(result, "&trade;", "™")
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
// when URLs contain query parameters like "?param1=value1&param2=value2".
func escapeAttributeAmpersands(content string) string {
	var out strings.Builder
	out.Grow(len(content))

	inTag := false
	var quote byte

	for i := 0; i < len(content); i++ {
		c := content[i]
		if quote != 0 {
			if c == quote {
				out.WriteByte(c)
				quote = 0
				continue
			}
			if c == '&' {
				j := i + 1
				for j < len(content) && content[j] != quote && !isEntityTerminator(content[j]) {
					j++
				}
				if j < len(content) && content[j] == ';' && isValidEntity(content[i+1:j]) {
					out.WriteString(content[i : j+1])
					i = j
				} else {
					out.WriteString("&amp;")
				}
				continue
			}
			out.WriteByte(c)
			continue
		}

		switch c {
		case '<':
			inTag = true
		case '>':
			inTag = false
		case '\'', '"':
			if inTag {
				quote = c
			}
		}
		out.WriteByte(c)
	}

	return out.String()
}

// escapeAmperands escapes ampersands that aren't part of valid HTML entities.
func escapeAmperands(value string) string {
	if strings.IndexByte(value, '&') == -1 {
		return value
	}

	var b strings.Builder
	b.Grow(len(value))

	for i := 0; i < len(value); i++ {
		c := value[i]
		if c != '&' {
			b.WriteByte(c)
			continue
		}

		j := i + 1
		for j < len(value) && !isEntityTerminator(value[j]) {
			j++
		}
		if j < len(value) && value[j] == ';' && isValidEntity(value[i+1:j]) {
			b.WriteString(value[i : j+1])
			i = j
		} else {
			b.WriteString("&amp;")
		}
	}

	return b.String()
}

func isValidEntity(s string) bool {
	if len(s) == 0 {
		return false
	}
	if s[0] == '#' {
		if len(s) == 1 {
			return false
		}
		if s[1] == 'x' || s[1] == 'X' {
			if len(s) == 2 {
				return false
			}
			for i := 2; i < len(s); i++ {
				if !isHexDigit(s[i]) {
					return false
				}
			}
			return true
		}
		for i := 1; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return false
			}
		}
		return true
	}
	_, ok := namedHTMLEntities[s]
	return ok
}

func isHexDigit(b byte) bool {
	return ('0' <= b && b <= '9') || ('a' <= b && b <= 'f') || ('A' <= b && b <= 'F')
}

// isEntityTerminator returns true if the byte is a character that terminates
// an HTML entity sequence (anything that would end an entity name).
func isEntityTerminator(c byte) bool {
	return c == ';' || c == '&' || c == ' ' || c == '\n' || c == '\t' ||
		c == '"' || c == '\'' || c == '<' || c == '>'
}

// stripNonMSOComments removes comments that appear before the <mjml> root node.
// Comments inside the document are preserved so they can be rendered in the
// output HTML. This mirrors the behaviour of the MRML reference implementation
// which keeps user comments intact while ensuring the XML decoder starts at the
// root element.
func stripNonMSOComments(content string) string {
	// Find <mjml case-insensitively without creating a copy of the entire content
	idx := findMjmlTagIndex(content)
	if idx == -1 {
		return content
	}

	prefix := content[:idx]

	// Fast check: if there are no comments in the prefix, just trim whitespace and return
	// Use byte-level search to avoid string allocation
	hasComments := false
	for i := 0; i <= len(prefix)-4; i++ {
		if prefix[i] == '<' && prefix[i+1] == '!' && prefix[i+2] == '-' && prefix[i+3] == '-' {
			hasComments = true
			break
		}
	}
	if !hasComments {
		prefix = trimLeftInPlace(prefix)
		return prefix + content[idx:]
	}

	// Use strings.Builder for efficient string building instead of concatenation
	var result strings.Builder
	result.Grow(len(content)) // Pre-allocate to avoid repeated allocations

	// Strip all HTML comments from the prefix to avoid parse errors when the
	// XML decoder expects a start element.
	writePos := 0
	for {
		start := strings.Index(prefix[writePos:], "<!--")
		if start == -1 {
			// No more comments, write remaining prefix
			result.WriteString(prefix[writePos:])
			break
		}
		start += writePos // Make relative to prefix start

		// Write content before comment
		result.WriteString(prefix[writePos:start])

		// Find comment end
		end := strings.Index(prefix[start+4:], "-->")
		if end == -1 {
			// Malformed comment; drop everything from start
			break
		}
		end += start + 4 + 3 // Point after "-->"

		// Skip the comment by updating write position
		writePos = end
	}

	// Trim any leftover whitespace before the root element efficiently
	// Write directly to avoid creating intermediate string
	resultStr := result.String()
	trimmed := trimLeftInPlace(resultStr)

	return trimmed + content[idx:]
}

// trimLeftInPlace efficiently trims leading whitespace without allocating new strings
func trimLeftInPlace(s string) string {
	start := 0
	for start < len(s) {
		c := s[start]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			break
		}
		start++
	}
	return s[start:]
}

// findMjmlTagIndex finds the index of "<mjml" case-insensitively without allocating
func findMjmlTagIndex(content string) int {
	needle := "<mjml"
	for i := 0; i <= len(content)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			c := content[i+j]
			n := needle[j]
			// Convert to lowercase for comparison
			if c >= 'A' && c <= 'Z' {
				c = c + 'a' - 'A'
			}
			if c != n {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// writeNonWhitespace writes content to builder, trimming leading and trailing whitespace
func writeNonWhitespace(builder *strings.Builder, content string) {
	// Trim leading and trailing whitespace efficiently
	start := 0
	for start < len(content) {
		c := content[start]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			break
		}
		start++
	}

	end := len(content)
	for end > start {
		c := content[end-1]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			break
		}
		end--
	}

	if start < end {
		builder.WriteString(content[start:end])
	}
}

// isMSOConditionalComment checks if a comment is an MSO conditional comment
// that should be preserved for email client compatibility
func isMSOConditionalComment(comment string) bool {
	// MSO conditional comments patterns
	msoPatterns := []string{
		"<!--[if mso",
		"<!--[if !mso",
		"<!--[if lte mso",
		"<!--[if gte mso",
		"<!--[if lt mso",
		"<!--[if gt mso",
		"<![endif]-->",
	}

	commentLower := strings.ToLower(comment)
	for _, pattern := range msoPatterns {
		if strings.Contains(commentLower, pattern) {
			return true
		}
	}

	return false
}

const (
	openNeedle  = "<mj-text"
	closeNeedle = "</mj-text>"
	cdataStart  = "<![CDATA["
	cdataEnd    = "]]>"
	// cdataEndSafe is used to escape CDATA end sequences within CDATA sections.
	// When "]]>" appears in content that will be wrapped in CDATA, it's replaced
	// with "]]]]><![CDATA[>" which effectively closes the current CDATA section,
	// outputs "]]>", then starts a new CDATA section. This prevents XML parsing
	// errors that would occur if "]]>" appeared within a CDATA block.
	// See: https://www.w3.org/TR/xml/#sec-cdata-sect (W3C XML 1.0, section 2.7 CDATA Sections)
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
var voidSelfClosingRe = regexp.MustCompile(buildVoidElementsRegexPattern())

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
