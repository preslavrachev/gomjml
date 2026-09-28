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
	XMLName    xml.Name
	Text       string
	Attrs      []xml.Attr
	Children   []*MJMLNode
	LineNumber int
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

type lineLookup struct {
	lineOffsets []int
	lastOffset  int64
	lastIndex   int
}

func newLineLookup(content []byte) *lineLookup {
	offsets := make([]int, 0, 64)
	offsets = append(offsets, 0)
	for i, b := range content {
		if b == '\n' {
			offsets = append(offsets, i+1)
		}
	}
	return &lineLookup{lineOffsets: offsets, lastOffset: -1}
}

func (ll *lineLookup) Line(offset int64) int {
	if ll == nil || len(ll.lineOffsets) == 0 {
		return 1
	}
	if offset >= ll.lastOffset {
		idx := ll.lastIndex
		for idx+1 < len(ll.lineOffsets) && ll.lineOffsets[idx+1] <= int(offset) {
			idx++
		}
		ll.lastIndex = idx
		ll.lastOffset = offset
		return idx + 1
	}

	idx := max(sort.Search(len(ll.lineOffsets), func(i int) bool {
		return ll.lineOffsets[i] > int(offset)
	})-1, 0)
	ll.lastIndex = idx
	ll.lastOffset = offset
	return idx + 1
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

	// Keep ending-tag content as written, as MJML does
	processedContent = wrapEndingTagContent(processedContent)

	contentBytes := []byte(processedContent)
	lookup := newLineLookup(contentBytes)

	decoder := xml.NewDecoder(bytes.NewReader(contentBytes))
	root, err := parseNode(decoder, xml.StartElement{}, lookup, 0, contentBytes)
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
	cdataStart = "<![CDATA["
	cdataEnd   = "]]>"
	// cdataEndSafe is used to escape CDATA end sequences within CDATA sections.
	// When "]]>" appears in content that will be wrapped in CDATA, it's replaced
	// with "]]]]><![CDATA[>" which effectively closes the current CDATA section,
	// outputs "]]>", then starts a new CDATA section. This prevents XML parsing
	// errors that would occur if "]]>" appeared within a CDATA block.
	// See: https://www.w3.org/TR/xml/#sec-cdata-sect (W3C XML 1.0, section 2.7 CDATA Sections)
	cdataEndSafe = "]]]]><![CDATA[>"
)

// endingTags are the components MJML declares with endingTag = true: it keeps
// their content as written instead of parsing it as MJML.
var endingTags = map[string]struct{}{
	"mj-accordion-text":  {},
	"mj-accordion-title": {},
	"mj-breakpoint":      {},
	"mj-button":          {},
	"mj-carousel-image":  {},
	"mj-navbar-link":     {},
	"mj-preview":         {},
	"mj-raw":             {},
	"mj-social-element":  {},
	"mj-style":           {},
	"mj-table":           {},
	"mj-text":            {},
	"mj-title":           {},
}

// wrapEndingTagContent wraps the content of every ending tag in a CDATA
// section, so that the XML decoder returns it as written, and normalizes void
// tags inside mj-text. Tag names are matched case-insensitively.
func wrapEndingTagContent(content string) string {
	var out strings.Builder
	pos := 0
	for i := 0; i < len(content); {
		lt := strings.IndexByte(content[i:], '<')
		if lt < 0 {
			break
		}
		i += lt
		if next := skipMarkupDeclaration(content, i); next > i {
			i = next
			continue
		}
		name, closing := scanTagName(content, i)
		if name == "" {
			i++
			continue
		}
		end, selfClosing := findTagEnd(content, i)
		if end < 0 {
			break
		}
		i = end
		if _, ok := endingTags[name]; !ok || closing || selfClosing {
			continue
		}
		closeStart, closeEnd := findEndingTagClose(content, end)
		if closeStart < 0 {
			break // the XML decoder reports the unclosed tag
		}
		if pos == 0 {
			out.Grow(len(content) + 64)
		}
		out.WriteString(content[pos:end])
		writeEndingTagContent(&out, name, content[end:closeStart])
		// htmlparser2 lower-cases the name and skips to '>', so "</MJ-TEXT\n>"
		// closes <mj-text>; the XML decoder needs it written exactly.
		closeName, _ := scanTagName(content, closeStart)
		out.WriteString("</" + closeName + ">")
		pos, i = closeEnd, closeEnd
	}
	if pos == 0 {
		return content
	}
	out.WriteString(content[pos:])
	return out.String()
}

// writeEndingTagContent writes inner as CDATA. A CDATA section the author put
// first is left as it is, so the decoder unwraps it, except in mj-raw, which
// passes everything through.
func writeEndingTagContent(out *strings.Builder, name, inner string) {
	if name == "mj-text" {
		inner = normalizeSelfClosingVoidTags(inner)
	}
	if name != "mj-raw" && strings.HasPrefix(strings.TrimLeft(inner, " \t\r\n"), cdataStart) {
		out.WriteString(inner)
		return
	}
	out.WriteString(cdataStart)
	out.WriteString(strings.ReplaceAll(inner, cdataEnd, cdataEndSafe))
	out.WriteString(cdataEnd)
}

// findEndingTagClose returns the bounds of the end tag that closes an ending
// tag whose content starts at from. Like MJML, it counts every ending tag
// nested in the content, whatever its name.
func findEndingTagClose(s string, from int) (start, end int) {
	depth := 1
	for i := from; i < len(s); {
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			break
		}
		i += lt
		if next := skipMarkupDeclaration(s, i); next > i {
			i = next
			continue
		}
		name, closing := scanTagName(s, i)
		if name == "" {
			i++
			continue
		}
		tagEnd, selfClosing := findTagEnd(s, i)
		if tagEnd < 0 {
			break
		}
		if _, ok := endingTags[name]; ok {
			if closing {
				if depth--; depth == 0 {
					return i, tagEnd
				}
			} else if !selfClosing {
				depth++
			}
		}
		i = tagEnd
	}
	return -1, -1
}

// skipMarkupDeclaration returns the index just past the comment, CDATA
// section, declaration or processing instruction at s[i], or i if there is
// none there.
func skipMarkupDeclaration(s string, i int) int {
	rest := s[i:]
	var open, close string
	switch {
	case strings.HasPrefix(rest, "<!--"):
		open, close = "<!--", "-->"
	case strings.HasPrefix(rest, cdataStart):
		open, close = cdataStart, cdataEnd
	case strings.HasPrefix(rest, "<!"):
		open, close = "<!", ">"
	case strings.HasPrefix(rest, "<?"):
		open, close = "<?", "?>"
	default:
		return i
	}
	if j := strings.Index(rest[len(open):], close); j >= 0 {
		return i + len(open) + j + len(close)
	}
	return len(s)
}

// scanTagName returns the lower-cased name of the start or end tag at s[i],
// or "" if s[i] does not begin a tag.
func scanTagName(s string, i int) (name string, closing bool) {
	j := i + 1
	if j < len(s) && s[j] == '/' {
		closing = true
		j++
	}
	if j >= len(s) || !isASCIILetter(s[j]) {
		return "", false
	}
	k := j + 1
	for k < len(s) && (isASCIILetter(s[k]) || ('0' <= s[k] && s[k] <= '9') || strings.IndexByte("-_:.", s[k]) >= 0) {
		k++
	}
	name = s[j:k]
	for n := 0; n < len(name); n++ {
		if 'A' <= name[n] && name[n] <= 'Z' {
			return strings.ToLower(name), closing
		}
	}
	return name, closing
}

func isASCIILetter(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

// findTagEnd returns the index just past the '>' that ends the tag at s[start]
// and whether the tag closes itself. As in HTML, a quote only opens an
// attribute value right after '=', so unquoted values may contain quotes.
func findTagEnd(s string, start int) (end int, selfClosing bool) {
	afterEquals := false
	for i := start + 1; i < len(s); i++ {
		switch c := s[i]; c {
		case '"', '\'':
			if afterEquals {
				j := strings.IndexByte(s[i+1:], c)
				if j < 0 {
					return -1, false
				}
				i += j + 1
			}
			afterEquals = false
		case '=':
			afterEquals = true
		case ' ', '\t', '\r', '\n':
		case '>':
			return i + 1, previousNonSpace(s, i-1) == '/'
		default:
			afterEquals = false
		}
	}
	return -1, false
}

// previousNonSpace returns the previous non-space byte at or before idx; 0 if none.
func previousNonSpace(s string, idx int) byte {
	for i := idx; i >= 0; i-- {
		if s[i] != ' ' && s[i] != '\t' && s[i] != '\n' && s[i] != '\r' {
			return s[i]
		}
	}
	return 0
}

// normalizeSelfClosingVoidTags ensures that void HTML elements use a space before the
// closing slash (e.g., <br/> becomes <br />). This matches how XML parsers normalize
// self-closing tags.
var voidSelfClosingRe = regexp.MustCompile(buildVoidElementsRegexPattern())

func normalizeSelfClosingVoidTags(s string) string {
	if !strings.Contains(s, "/>") {
		return s
	}
	return voidSelfClosingRe.ReplaceAllStringFunc(s, func(m string) string {
		return strings.TrimRight(m[:len(m)-2], " ") + " />"
	})
}

// parseNode recursively parses XML nodes
func parseNode(decoder *xml.Decoder, start xml.StartElement, lookup *lineLookup, startOffset int64, content []byte) (*MJMLNode, error) {
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
			startOffset = decoder.InputOffset()
		} else {
			return nil, fmt.Errorf("expected start element")
		}
	}

	if lookup != nil && len(node.Attrs) > 0 {
		node.LineNumber = lookup.Line(startOffset)
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
			childOffset := decoder.InputOffset()
			child, err := parseNode(decoder, t, lookup, childOffset, content)
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

// GetLineNumber returns the starting line number for this node in the processed MJML source.
func (n *MJMLNode) GetLineNumber() int {
	if n == nil || n.LineNumber <= 0 {
		return 1
	}
	return n.LineNumber
}

// GetTextContent returns the trimmed text content
func (n *MJMLNode) GetTextContent() string {
	return strings.TrimSpace(n.Text)
}

// GetMixedContent returns the full mixed content including HTML child elements
// This reconstructs the original content like "Share <b>test</b> hi" from the AST
func (n *MJMLNode) GetMixedContent() string {
	if debug.Enabled() {
		debug.DebugLogWithData("parser", "mixed-content", "Processing mixed content", map[string]any{
			"tag_name":       n.XMLName.Local,
			"plain_text":     n.Text,
			"children_count": len(n.Children),
			"has_children":   len(n.Children) > 0,
		})
	}

	if len(n.MixedContent) == 0 {
		result := strings.TrimSpace(n.Text)
		if debug.Enabled() {
			debug.DebugLogWithData("parser", "text-only", "Returning plain text content", map[string]any{
				"content": result,
			})
		}
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
	if debug.Enabled() {
		debug.DebugLogWithData("parser", "mixed-complete", "Mixed content reconstructed", map[string]any{
			"original_text":     n.Text,
			"final_content":     finalResult,
			"children_rendered": len(n.Children),
		})
	}
	return finalResult
}

// HasRenderableMixedContent reports whether GetMixedContent would return a
// non-empty string, without paying the cost of building it. A child node
// part always renders at least its own tag markup, so its presence alone is
// enough; text parts must contain non-whitespace to count.
func (n *MJMLNode) HasRenderableMixedContent() bool {
	if len(n.MixedContent) == 0 {
		return strings.TrimSpace(n.Text) != ""
	}
	for _, part := range n.MixedContent {
		if part.Node != nil {
			return true
		}
		if strings.TrimSpace(part.Text) != "" {
			return true
		}
	}
	return false
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
