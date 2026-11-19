package components

import (
	"strings"

	"github.com/preslavrachev/gomjml/parser"
)

// endingTagHTML returns the content of an ending tag (mj-button, mj-table, ...)
// the way MJML places it in its output: as written, trimmed, with <br/> written
// as <br> as in mj-text, and whitespace collapsed as the reference goldens'
// html-minifier pass collapses it.
func endingTagHTML(node *parser.MJMLNode) string {
	content := node.GetMixedContent()
	if strings.Contains(content, "/>") {
		content = normalizeVoidHTMLTags(content)
	}
	return collapseHTMLWhitespace(content)
}
