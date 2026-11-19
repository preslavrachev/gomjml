package components

import (
	"strings"
	"unicode/utf8"
)

func tagSet(names string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, name := range strings.Fields(names) {
		set[name] = struct{}{}
	}
	return set
}

// Tag tables of html-minifier 4.0.0.
var (
	minifierInlineTags            = tagSet("a abbr acronym b bdi bdo big button cite code del dfn em font i ins kbd label mark math nobr object q rp rt rtc ruby s samp select small span strike strong sub sup svg textarea time tt u var")
	minifierInlineTextTags        = tagSet("a abbr acronym b big del em font i ins kbd mark nobr rp s samp small span strike strong sub sup time tt u var")
	minifierSelfClosingInlineTags = tagSet("comment img input wbr")
	minifierVoidTags              = tagSet("area base basefont br col embed frame hr img input isindex keygen link meta param source track wbr")
)

type fragmentPartKind uint8

const (
	partText fragmentPartKind = iota
	partRawText
	partStartTag
	partEndTag
	partOther // comments, CDATA sections and declarations
)

type fragmentPart struct {
	kind  fragmentPartKind
	name  string // tag name as written
	s     string
	unary bool
}

// collapseHTMLWhitespace collapses the whitespace of an HTML fragment the way
// html-minifier 4.0.0 does with collapseWhitespace, the option MJML's CLI
// minifies with, treating the fragment as the whole content of a block
// element. Tags are kept as written. Like the minifier with caseSensitive,
// only the tag before a text is looked up lower-cased.
func collapseHTMLWhitespace(fragment string) string {
	if !strings.ContainsRune(fragment, '<') {
		return collapseWhitespace(fragment, true, true, true)
	}
	parts := splitHTMLFragment(fragment)
	c := whitespaceCollapser{buf: make([]fragmentPart, 0, len(parts)+2)}
	c.start(fragmentPart{kind: partStartTag, name: "div"})
	prevTag := "div"
	for i, part := range parts {
		switch part.kind {
		case partStartTag:
			c.start(part)
			prevTag = strings.ToLower(part.name)
		case partEndTag:
			c.end(part)
			prevTag = "/" + strings.ToLower(part.name)
		case partOther:
			c.buf = append(c.buf, part)
			prevTag = ""
		case partRawText:
			c.rawChars(part.s)
			prevTag = ""
		case partText:
			nextTag := "/div"
			if i+1 < len(parts) {
				switch next := parts[i+1]; next.kind {
				case partStartTag:
					nextTag = next.name
				case partEndTag:
					nextTag = "/" + next.name
				default:
					nextTag = ""
				}
			}
			c.chars(part.s, prevTag, nextTag)
			prevTag = ""
		}
	}
	c.end(fragmentPart{kind: partEndTag, name: "div"})

	var out strings.Builder
	out.Grow(len(fragment))
	for _, part := range c.buf[1 : len(c.buf)-1] {
		out.WriteString(part.s)
	}
	return out.String()
}

// splitHTMLFragment splits a fragment into tags, text, comments and the raw
// text of script and style elements.
func splitHTMLFragment(s string) []fragmentPart {
	var parts []fragmentPart
	text := 0
	flush := func(end int) {
		if end > text {
			parts = append(parts, fragmentPart{kind: partText, s: s[text:end]})
		}
	}
	for i := 0; i < len(s); {
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			break
		}
		i += lt
		if end := skipHTMLDeclaration(s, i); end > i {
			flush(i)
			parts = append(parts, fragmentPart{kind: partOther, s: s[i:end]})
			i, text = end, end
			continue
		}
		name, closing := htmlTagName(s, i)
		end, selfClosing := htmlTagEnd(s, i)
		if name == "" || end < 0 {
			i++
			continue
		}
		flush(i)
		if closing {
			parts = append(parts, fragmentPart{kind: partEndTag, name: name, s: s[i:end]})
			i, text = end, end
			continue
		}
		_, void := minifierVoidTags[strings.ToLower(name)]
		parts = append(parts, fragmentPart{kind: partStartTag, name: name, s: s[i:end], unary: void || selfClosing})
		i, text = end, end
		if (strings.EqualFold(name, "script") || strings.EqualFold(name, "style")) && !selfClosing {
			closeAt := indexFold(s[i:], "</"+name)
			if closeAt < 0 {
				closeAt = len(s) - i
			}
			parts = append(parts, fragmentPart{kind: partRawText, s: s[i : i+closeAt]})
			i += closeAt
			text = i
		}
	}
	flush(len(s))
	return parts
}

func skipHTMLDeclaration(s string, i int) int {
	rest := s[i:]
	var open, close string
	switch {
	case strings.HasPrefix(rest, "<!--"):
		open, close = "<!--", "-->"
	case strings.HasPrefix(rest, "<!["):
		open, close = "<![", "]>"
	case strings.HasPrefix(rest, "<!"):
		open, close = "<!", ">"
	default:
		return i
	}
	if j := strings.Index(rest[len(open):], close); j >= 0 {
		return i + len(open) + j + len(close)
	}
	return i
}

func htmlTagName(s string, i int) (name string, closing bool) {
	j := i + 1
	if j < len(s) && s[j] == '/' {
		closing = true
		j++
	}
	k := j
	for k < len(s) && (('a' <= s[k] && s[k] <= 'z') || ('A' <= s[k] && s[k] <= 'Z') ||
		(k > j && (('0' <= s[k] && s[k] <= '9') || s[k] == '-' || s[k] == '_' || s[k] == ':' || s[k] == '.'))) {
		k++
	}
	return s[j:k], closing
}

// htmlTagEnd returns the index just past the '>' ending the tag at s[i] and
// whether it closes itself. A quote opens a value only right after '='.
func htmlTagEnd(s string, i int) (end int, selfClosing bool) {
	afterEquals := false
	for j := i + 1; j < len(s); j++ {
		switch c := s[j]; c {
		case '"', '\'':
			if afterEquals {
				k := strings.IndexByte(s[j+1:], c)
				if k < 0 {
					return -1, false
				}
				j += k + 1
			}
			afterEquals = false
		case '=':
			afterEquals = true
		case ' ', '\t', '\r', '\n', '\f':
		case '>':
			return j + 1, s[j-1] == '/'
		default:
			afterEquals = false
		}
	}
	return -1, false
}

func indexFold(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if strings.EqualFold(s[i:i+len(substr)], substr) {
			return i
		}
	}
	return -1
}

// whitespaceCollapser ports the collapseWhitespace handling of
// html-minifier's start, end and chars callbacks. Its buffer holds the output;
// lastChar is the last rune of the minifier's currentChars, or 0 while it is
// empty.
type whitespaceCollapser struct {
	buf        []fragmentPart
	lastChar   rune
	currentTag string
	hasChars   bool
	noTrim     []string
	noCollapse []string
}

func (c *whitespaceCollapser) start(tag fragmentPart) {
	c.currentTag = tag.name
	if _, ok := minifierInlineTextTags[tag.name]; !ok {
		c.lastChar = 0
	}
	c.hasChars = false
	if len(c.noTrim) == 0 {
		c.squashTrailingWhitespace(tag.name)
	}
	if !tag.unary {
		if tag.name == "pre" || tag.name == "textarea" || len(c.noTrim) > 0 {
			c.noTrim = append(c.noTrim, tag.name)
		}
		if tag.name == "pre" || tag.name == "textarea" || tag.name == "script" || tag.name == "style" || len(c.noCollapse) > 0 {
			c.noCollapse = append(c.noCollapse, tag.name)
		}
	}
	c.buf = append(c.buf, tag)
}

func (c *whitespaceCollapser) end(tag fragmentPart) {
	if n := len(c.noTrim); n > 0 {
		if c.noTrim[n-1] == tag.name {
			c.noTrim = c.noTrim[:n-1]
		}
	} else {
		c.squashTrailingWhitespace("/" + tag.name)
	}
	if n := len(c.noCollapse); n > 0 && c.noCollapse[n-1] == tag.name {
		c.noCollapse = c.noCollapse[:n-1]
	}
	empty := false
	if tag.name == c.currentTag {
		c.currentTag = ""
		empty = !c.hasChars
	}
	c.buf = append(c.buf, tag)
	if _, ok := minifierInlineTags[tag.name]; !ok {
		c.lastChar = 0
	} else if empty {
		c.lastChar = '|'
	}
}

func (c *whitespaceCollapser) chars(text, prevTag, nextTag string) {
	if prevTag == "" {
		prevTag = "comment"
	}
	if nextTag == "" {
		nextTag = "comment"
	}
	if len(c.noTrim) == 0 {
		if prevTag == "comment" && len(c.buf) > 1 && c.lastChar == ' ' {
			// Whitespace before a comment moves after it.
			before := &c.buf[len(c.buf)-2]
			trimmed := strings.TrimRightFunc(before.s, isJSSpace)
			text = before.s[len(trimmed):] + text
			before.s = trimmed
		}
		if prevTag == "/nobr" || prevTag == "wbr" {
			if r, _ := utf8.DecodeRuneInString(text); isJSSpace(r) {
				kind, name := partStartTag, prevTag
				if prevTag == "/nobr" {
					kind, name = partEndTag, "nobr"
				}
				i := len(c.buf) - 1
				for i > 0 && (c.buf[i].kind != kind || c.buf[i].name != name) {
					i--
				}
				c.trimTrailingWhitespace(i-1, "br")
			}
		} else if _, ok := minifierInlineTextTags[strings.TrimPrefix(prevTag, "/")]; ok {
			text = collapseWhitespace(text, c.lastChar == 0 || isJSSpace(c.lastChar), false, false)
		}
		text = collapseWhitespaceSmart(text, prevTag, nextTag)
		if text == "" && isJSSpace(c.lastChar) && prevTag[0] == '/' {
			c.trimTrailingWhitespace(len(c.buf)-1, nextTag)
		}
	}
	c.addChars(text)
}

// rawChars handles the content of script and style, which html-minifier's
// parser passes without neighbouring tags.
func (c *whitespaceCollapser) rawChars(text string) {
	if len(c.noTrim) == 0 {
		text = collapseWhitespace(text, true, true, false)
	}
	c.addChars(text)
}

func (c *whitespaceCollapser) addChars(text string) {
	if r, _ := utf8.DecodeLastRuneInString(text); text != "" {
		c.lastChar = r
		c.hasChars = true
	}
	c.buf = append(c.buf, fragmentPart{kind: partText, s: text})
}

func (c *whitespaceCollapser) squashTrailingWhitespace(nextTag string) {
	i := len(c.buf) - 1
	if len(c.buf) > 1 {
		if last := c.buf[i]; last.kind == partOther || (last.kind == partText && last.s == "") {
			i--
		}
	}
	c.trimTrailingWhitespace(i, nextTag)
}

// trimTrailingWhitespace trims the text before a tag, looking back past end
// tags, as html-minifier's function of the same name does.
func (c *whitespaceCollapser) trimTrailingWhitespace(i int, nextTag string) {
	for endTag := ""; i >= 0 && endTag != "pre" && endTag != "textarea"; i-- {
		part := &c.buf[i]
		if part.kind == partEndTag {
			endTag = part.name
			continue
		}
		if (part.kind != partText && part.kind != partRawText) || strings.HasSuffix(part.s, ">") {
			break
		}
		if part.s = collapseWhitespaceSmart(part.s, "", nextTag); part.s != "" {
			break
		}
	}
}

func collapseWhitespaceSmart(s, prevTag, nextTag string) string {
	trimLeft := false
	if _, inline := minifierSelfClosingInlineTags[prevTag]; prevTag != "" && !inline {
		if name, ok := strings.CutPrefix(prevTag, "/"); ok {
			_, inline = minifierInlineTags[name]
		} else {
			_, inline = minifierInlineTextTags[prevTag]
		}
		trimLeft = !inline
	}
	trimRight := false
	if _, inline := minifierSelfClosingInlineTags[nextTag]; nextTag != "" && !inline {
		if name, ok := strings.CutPrefix(nextTag, "/"); ok {
			_, inline = minifierInlineTextTags[name]
		} else {
			_, inline = minifierInlineTags[nextTag]
		}
		trimRight = !inline
	}
	return collapseWhitespace(s, trimLeft, trimRight, prevTag != "" && nextTag != "")
}

func isMinifierSpace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == '\f' || r == '\u00a0'
}

// isJSSpace reports whether r matches \s in a JavaScript regular expression.
func isJSSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\v', '\f', '\r', '\u00a0', '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000', '\ufeff':
		return true
	}
	return '\u2000' <= r && r <= '\u200a'
}

// collapseWhitespace is html-minifier's function of the same name with
// conservativeCollapse and preserveLineBreaks off. A non-breaking space is
// kept, while each run of other whitespace becomes one space, or nothing at a
// trimmed edge.
func collapseWhitespace(s string, trimLeft, trimRight, collapseAll bool) string {
	var out strings.Builder
	changed := false
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if !isMinifierSpace(r) {
			if changed {
				out.WriteString(s[i : i+size])
			}
			i += size
			continue
		}
		j := i
		for j < len(s) {
			r, size := utf8.DecodeRuneInString(s[j:])
			if !isMinifierSpace(r) {
				break
			}
			j += size
		}
		atStart, atEnd := i == 0, j == len(s)
		var run string
		switch {
		case (atStart && trimLeft) || (atEnd && trimRight) || collapseAll:
			run = collapseSpaceRun(s[i:j], atStart && trimLeft, atEnd && trimRight, collapseAll)
		default:
			run = s[i:j]
		}
		if !changed && run != s[i:j] {
			changed = true
			out.Grow(len(s))
			out.WriteString(s[:i])
		}
		if changed {
			out.WriteString(run)
		}
		i = j
	}
	if !changed {
		return s
	}
	return out.String()
}

// collapseSpaceRun rewrites one run of whitespace. Its non-breaking spaces
// are kept and every stretch of other whitespace becomes one space, except
// that a stretch at a trimmed edge is dropped. When only trimming, a run
// without non-breaking spaces is left alone away from the trimmed edge.
func collapseSpaceRun(run string, trimLeft, trimRight, collapseAll bool) string {
	if collapseAll && run == "\t" && !trimLeft && !trimRight {
		return run
	}
	var out strings.Builder
	first := true
	for i := 0; i < len(run); {
		nbsp := strings.HasPrefix(run[i:], "\u00a0")
		j := i
		for j < len(run) && strings.HasPrefix(run[j:], "\u00a0") == nbsp {
			if nbsp {
				j += len("\u00a0")
			} else {
				j++
			}
		}
		last := j == len(run)
		switch {
		case nbsp:
			out.WriteString(run[i:j])
		case first && trimLeft, last && trimRight:
		default:
			out.WriteByte(' ')
		}
		first = false
		i = j
	}
	return out.String()
}
