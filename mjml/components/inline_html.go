package components

import (
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/options"
)

type inlineHTMLAttr struct {
	Prefix   string
	Name     string
	Value    string
	Quote    byte
	HasValue bool
}

// ApplyInlineStylesToHTMLContent processes an HTML fragment and inlines CSS declarations
// collected from mj-style inline rules onto elements with matching class attributes.
func (bc *BaseComponent) ApplyInlineStylesToHTMLContent(html string) string {
	if html == "" || bc == nil || bc.RenderOpts == nil || len(bc.RenderOpts.InlineClassStyles) == 0 {
		return html
	}
	return applyInlineStylesToHTML(html, bc.RenderOpts.InlineClassStyles, bc)
}

func applyInlineStylesToHTML(html string, styles map[string][]options.InlineStyle, bc *BaseComponent) string {
	if html == "" {
		return html
	}

	var builder strings.Builder
	builder.Grow(len(html))

	i := 0
	for i < len(html) {
		lt := strings.IndexByte(html[i:], '<')
		if lt == -1 {
			builder.WriteString(html[i:])
			break
		}
		lt += i
		// write preceding text
		if lt > i {
			builder.WriteString(html[i:lt])
		}
		if lt+1 >= len(html) {
			builder.WriteString(html[lt:])
			break
		}

		next := html[lt+1]
		if next == '/' || next == '!' || next == '?' {
			end := findTagEnd(html, lt+1)
			if end == -1 {
				builder.WriteString(html[lt:])
				break
			}
			builder.WriteString(html[lt : end+1])
			i = end + 1
			continue
		}

		end := findTagEnd(html, lt+1)
		if end == -1 {
			builder.WriteString(html[lt:])
			break
		}

		original := html[lt : end+1]
		rebuilt := inlineStylesInTag(original, styles, bc)
		builder.WriteString(rebuilt)
		i = end + 1
	}

	return builder.String()
}

func findTagEnd(value string, start int) int {
	inQuote := byte(0)
	for i := start; i < len(value); i++ {
		c := value[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			inQuote = c
		case '>':
			return i
		}
	}
	return -1
}

func inlineStylesInTag(tag string, styles map[string][]options.InlineStyle, bc *BaseComponent) string {
	if len(styles) == 0 {
		return tag
	}

	tagName, attrs, selfClosing, closingSuffix := parseTag(tag)
	if tagName == "" || len(attrs) == 0 {
		return tag
	}

	var classValue string
	classIndex := -1
	styleIndex := -1

	for idx, attr := range attrs {
		if strings.EqualFold(attr.Name, constants.AttrClass) {
			classValue = attr.Value
			classIndex = idx
		} else if strings.EqualFold(attr.Name, constants.AttrStyle) {
			styleIndex = idx
		}
	}

	if classIndex == -1 {
		return tag
	}

	inlineStyle := bc.BuildInlineStyleString(classValue)
	if inlineStyle == "" {
		return tag
	}

	if styleIndex >= 0 {
		attrs[styleIndex].Value = mergeInlineStyleValues(attrs[styleIndex].Value, inlineStyle)
	} else {
		attrs = append(attrs, inlineHTMLAttr{
			Prefix:   " ",
			Name:     constants.AttrStyle,
			Value:    inlineStyle,
			Quote:    '"',
			HasValue: true,
		})
	}

	var builder strings.Builder
	builder.Grow(len(tag) + len(inlineStyle))
	builder.WriteByte('<')
	builder.WriteString(tagName)
	for _, attr := range attrs {
		builder.WriteString(attr.Prefix)
		builder.WriteString(attr.Name)
		if attr.HasValue {
			quote := attr.Quote
			if quote == 0 {
				quote = '"'
			}
			builder.WriteByte('=')
			builder.WriteByte(quote)
			builder.WriteString(attr.Value)
			builder.WriteByte(quote)
		}
	}
	if selfClosing {
		builder.WriteString(closingSuffix)
	}
	builder.WriteByte('>')
	return builder.String()
}

func parseTag(tag string) (string, []inlineHTMLAttr, bool, string) {
	if len(tag) < 2 || tag[0] != '<' {
		return "", nil, false, ""
	}

	i := 1
	for i < len(tag) && isSpace(tag[i]) {
		i++
	}
	start := i
	for i < len(tag) && !isSpace(tag[i]) && tag[i] != '>' && tag[i] != '/' {
		i++
	}
	if start == i {
		return "", nil, false, ""
	}
	tagName := tag[start:i]
	attrs := make([]inlineHTMLAttr, 0, 4)

	for i < len(tag) {
		prefixStart := i
		for i < len(tag) && isSpace(tag[i]) {
			i++
		}
		prefix := tag[prefixStart:i]
		if i >= len(tag) {
			break
		}
		if tag[i] == '>' {
			break
		}
		if tag[i] == '/' {
			// Skip over trailing slash and any spaces before closing
			i++
			for i < len(tag) && isSpace(tag[i]) {
				i++
			}
			closingSuffix := ""
			if prefix == "" {
				closingSuffix = "/"
			} else {
				closingSuffix = prefix + "/"
			}
			if i < len(tag) && tag[i] == '>' {
				return tagName, attrs, true, closingSuffix
			}
			return tagName, attrs, true, closingSuffix
		}

		nameStart := i
		for i < len(tag) && !isSpace(tag[i]) && tag[i] != '=' && tag[i] != '>' && tag[i] != '/' {
			i++
		}
		name := tag[nameStart:i]
		if name == "" {
			break
		}

		for i < len(tag) && isSpace(tag[i]) {
			i++
		}

		hasValue := false
		value := ""
		quote := byte(0)
		if i < len(tag) && tag[i] == '=' {
			hasValue = true
			i++
			for i < len(tag) && isSpace(tag[i]) {
				i++
			}
			if i < len(tag) && (tag[i] == '"' || tag[i] == '\'') {
				quote = tag[i]
				i++
				valueStart := i
				for i < len(tag) && tag[i] != quote {
					i++
				}
				value = tag[valueStart:i]
				if i < len(tag) {
					i++
				}
			} else {
				valueStart := i
				for i < len(tag) && !isSpace(tag[i]) && tag[i] != '>' && tag[i] != '/' {
					i++
				}
				value = tag[valueStart:i]
			}
		}

		attrs = append(attrs, inlineHTMLAttr{
			Prefix:   prefix,
			Name:     name,
			Value:    value,
			Quote:    quote,
			HasValue: hasValue,
		})
	}

	closingSuffix := ""
	trimmed := strings.TrimSpace(tag)
	if strings.HasSuffix(trimmed, "/>") {
		if strings.HasSuffix(tag, " />") {
			closingSuffix = " />"
		} else {
			closingSuffix = "/"
		}
		return tagName, attrs, true, closingSuffix
	}
	return tagName, attrs, false, ""
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t'
}
