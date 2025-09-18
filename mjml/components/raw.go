package components

import (
	"io"
	"regexp"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJRawComponent represents an mj-raw component
// It outputs its inner content exactly as provided without any additional wrappers.
type MJRawComponent struct {
	*BaseComponent
	// Content stores the original inner HTML captured during parsing
	Content string
}

var (
	conditionalCommentGapAfter  = regexp.MustCompile(`(-->)\s+(<)`)
	conditionalCommentGapBefore = regexp.MustCompile(`(>)\s+(<!--)`)
	interTagWhitespace          = regexp.MustCompile(`>(\s+)<`)
)

// NewMJRawComponent creates a new mj-raw component
func NewMJRawComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJRawComponent {
	return &MJRawComponent{
		BaseComponent: NewBaseComponent(node, opts),
		Content:       node.Text,
	}
}

// GetTagName returns the component's tag name
func (c *MJRawComponent) GetTagName() string { return "mj-raw" }

// IsRawElement indicates this component should be treated as a raw element
func (c *MJRawComponent) IsRawElement() bool { return true }

// GetDefaultAttribute returns default values for the component's attributes
func (c *MJRawComponent) GetDefaultAttribute(name string) string { return "" }

// Render writes the original content trimmed of leading/trailing whitespace
func (c *MJRawComponent) Render(w io.StringWriter) error {
	content := strings.TrimSpace(c.Content)
	if strings.Contains(content, "<!--") {
		content = conditionalCommentGapAfter.ReplaceAllString(content, "${1}${2}")
		content = conditionalCommentGapBefore.ReplaceAllString(content, "${1}${2}")
	}

	content = interTagWhitespace.ReplaceAllString(content, "><")

	if _, err := w.WriteString(content); err != nil {
		return err
	}
	return nil
}
