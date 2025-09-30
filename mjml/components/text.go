package components

import (
	"io"
	"regexp"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/debug"
	"github.com/preslavrachev/gomjml/mjml/fonts"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJTextComponent represents mj-text
type MJTextComponent struct {
	*BaseComponent
}

var selfClosingVoidTagPattern = regexp.MustCompile(`(?i)<(area|base|br|col|embed|hr|img|input|link|meta|param|source|track|wbr)\b([^>]*)/>`)

// NewMJTextComponent creates a new mj-text component
func NewMJTextComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJTextComponent {
	return &MJTextComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJTextComponent) GetTagName() string {
	return "mj-text"
}

// Render implements optimized Writer-based rendering for MJTextComponent
func (c *MJTextComponent) Render(w io.StringWriter) error {
	if debug.Enabled() {
		debug.DebugLog("mj-text", "render-start", "Starting text component rendering")
		debug.DebugLogWithData("mj-text", "content", "Processing text content", map[string]any{
			"container_width": c.GetContainerWidth(),
		})
	}

	// Get attributes using full resolution order (element > mj-class > global > default)
	align := c.GetAttributeFast(c, constants.MJMLAlign)
	padding := c.GetAttributeFast(c, constants.MJMLPadding)
	containerBg := c.GetAttributeFast(c, constants.MJMLContainerBackgroundColor)

	// Create TR element
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}

	// Create TD with alignment and base styles
	tdTag := html.NewHTMLTag("td").
		AddAttribute(constants.AttrAlign, align)

	if containerBg != "" {
		tdTag.AddStyle(constants.CSSBackground, containerBg)
	}

	tdTag.AddStyle(constants.CSSFontSize, "0px").
		AddStyle(constants.CSSPadding, padding)

		// Add css-class if present
	c.SetClassAttribute(tdTag)

	// Add specific padding overrides if they exist (following MRML/section pattern)
	if paddingTop := c.GetAttributeFast(c, constants.MJMLPaddingTop); paddingTop != "" {
		tdTag.AddStyle(constants.CSSPaddingTop, paddingTop)
	}
	if paddingBottom := c.GetAttributeFast(c, constants.MJMLPaddingBottom); paddingBottom != "" {
		tdTag.AddStyle(constants.CSSPaddingBottom, paddingBottom)
	}
	if paddingLeft := c.GetAttributeFast(c, constants.MJMLPaddingLeft); paddingLeft != "" {
		tdTag.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}
	if paddingRight := c.GetAttributeFast(c, constants.MJMLPaddingRight); paddingRight != "" {
		tdTag.AddStyle(constants.CSSPaddingRight, paddingRight)
	}

	tdTag.AddStyle("word-break", "break-word")

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// Create inner div with font styling
	divTag := html.NewHTMLTag("div")
	c.AddDebugAttribute(divTag, "text")

	// Apply font styles using the proper interface method
	fontFamily := c.GetAttributeWithDefault(c, constants.MJMLFontFamily)
	fontSize := c.GetAttributeWithDefault(c, "font-size")
	fontWeight := c.GetAttributeWithDefault(c, "font-weight")
	fontStyle := c.GetAttributeWithDefault(c, "font-style")
	color := c.GetAttributeWithDefault(c, constants.MJMLColor)
	lineHeight := c.GetAttributeWithDefault(c, "line-height")
	textAlign := c.GetAttributeWithDefault(c, constants.MJMLAlign)
	textDecoration := c.GetAttributeWithDefault(c, "text-decoration")
	textTransform := c.GetAttributeWithDefault(c, "text-transform")
	letterSpacing := c.GetAttributeWithDefault(c, "letter-spacing")

	// Apply styles in the order expected by MRML
	if fontFamily != "" {
		divTag.AddStyle(constants.CSSFontFamily, fontFamily)
	}
	if fontSize != "" {
		divTag.AddStyle("font-size", fontSize)
	}
	if fontWeight != "" {
		divTag.AddStyle("font-weight", fontWeight)
	}
	if letterSpacing != "" {
		divTag.AddStyle("letter-spacing", letterSpacing)
	}
	if lineHeight != "" {
		divTag.AddStyle("line-height", lineHeight)
	}
	if textAlign != "" {
		divTag.AddStyle("text-align", textAlign)
	}
	if textTransform != "" {
		divTag.AddStyle("text-transform", textTransform)
	}
	if color != "" {
		divTag.AddStyle(constants.CSSColor, color)
	}
	if fontStyle != "" {
		divTag.AddStyle("font-style", fontStyle)
	}
	if textDecoration != "" {
		divTag.AddStyle("text-decoration", textDecoration)
	}

	if err := divTag.RenderOpen(w); err != nil {
		return err
	}
	innerHTML, err := c.buildRawInnerHTML()
	if err != nil {
		return err
	}
	if innerHTML != "" {
		normalized := ensureVoidHTMLTagsSelfClosed(innerHTML)
		if _, err := w.WriteString(normalized); err != nil {
			return err
		}
	}
	if err := divTag.RenderClose(w); err != nil {
		return err
	}
	if err := tdTag.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr>"); err != nil {
		return err
	}

	return nil
}

func (c *MJTextComponent) GetDefaultAttribute(name string) string {
	switch name {
	case constants.MJMLFontSize:
		return "13px"
	case constants.MJMLColor:
		return "#000000"
	case constants.MJMLAlign:
		return constants.AlignLeft
	case constants.MJMLFontFamily:
		return fonts.DefaultFontStack
	case constants.MJMLLineHeight:
		return "1"
	case constants.MJMLPadding:
		return "10px 25px"
	default:
		return ""
	}
}

// writeRawInnerHTML writes the original inner HTML content of the mj-text element to the writer
// This is needed because our parser splits content, but mj-text needs to preserve HTML
func (c *MJTextComponent) writeRawInnerHTML(w io.StringWriter) error {
	innerHTML, err := c.buildRawInnerHTML()
	if err != nil {
		return err
	}
	if innerHTML == "" {
		return nil
	}
	_, err = w.WriteString(innerHTML)
	return err
}

func (c *MJTextComponent) buildRawInnerHTML() (string, error) {
	// If we have mixed content, reconstruct it preserving original order
	if len(c.Node.MixedContent) > 0 {
		var builder strings.Builder
		for i, part := range c.Node.MixedContent {
			if part.Node != nil {
				if err := c.reconstructHTMLElement(part.Node, &builder); err != nil {
					return "", err
				}
				continue
			}

			text := part.Text
			if i == 0 {
				text = strings.TrimLeft(text, " \n\r\t")
			}
			if i == len(c.Node.MixedContent)-1 {
				text = strings.TrimRight(text, " \n\r\t")
			}
			if text != "" {
				builder.WriteString(c.restoreHTMLEntities(text))
			}
		}
		return builder.String(), nil
	}

	// Fallback: no mixed content, use trimmed text content
	return c.restoreHTMLEntities(strings.TrimSpace(c.Node.Text)), nil
}

// restoreHTMLEntities converts Unicode characters back to HTML entities for proper output
func (c *MJTextComponent) restoreHTMLEntities(text string) string {
	// Convert Unicode non-breaking space back to HTML entity
	result := strings.ReplaceAll(text, "\u00A0", "&#xA0;")
	return result
}

func ensureVoidHTMLTagsSelfClosed(html string) string {
	if html == "" {
		return html
	}

	return selfClosingVoidTagPattern.ReplaceAllStringFunc(html, func(tag string) string {
		base := strings.TrimRight(tag[:len(tag)-2], " \n\r\t")
		return base + " />"
	})
}

// reconstructHTMLElement reconstructs an HTML element from a parsed node
func (c *MJTextComponent) reconstructHTMLElement(node *parser.MJMLNode, w io.StringWriter) error {
	tagName := node.XMLName.Local

	// Check if this is a void element (self-closing)
	isVoidElement := isVoidHTMLElement(tagName)

	// Opening tag
	if _, err := w.WriteString("<"); err != nil {
		return err
	}
	if _, err := w.WriteString(tagName); err != nil {
		return err
	}

	// Attributes
	for _, attr := range node.Attrs {
		if _, err := w.WriteString(" "); err != nil {
			return err
		}
		if _, err := w.WriteString(attr.Name.Local); err != nil {
			return err
		}
		if _, err := w.WriteString(`="`); err != nil {
			return err
		}
		if _, err := w.WriteString(attr.Value); err != nil {
			return err
		}
		if _, err := w.WriteString(`"`); err != nil {
			return err
		}
	}

	if isVoidElement {
		// Self-closing tag
		_, err := w.WriteString(" />")
		return err
	}

	if _, err := w.WriteString(">"); err != nil {
		return err
	}

	// Content (text + children) preserving original order
	if len(node.MixedContent) > 0 {
		for _, part := range node.MixedContent {
			if part.Node != nil {
				if err := c.reconstructHTMLElement(part.Node, w); err != nil {
					return err
				}
				continue
			}
			if _, err := w.WriteString(c.restoreHTMLEntities(part.Text)); err != nil {
				return err
			}
		}
	} else {
		if node.Text != "" {
			if _, err := w.WriteString(c.restoreHTMLEntities(node.Text)); err != nil {
				return err
			}
		}
		for _, child := range node.Children {
			if err := c.reconstructHTMLElement(child, w); err != nil {
				return err
			}
		}
	}

	// Closing tag
	if _, err := w.WriteString("</"); err != nil {
		return err
	}
	if _, err := w.WriteString(tagName); err != nil {
		return err
	}
	_, err := w.WriteString(">")
	return err
}

// isVoidHTMLElement checks if an HTML element is a void element (self-closing)
func isVoidHTMLElement(tagName string) bool {
	voidElements := map[string]bool{
		"area":   true,
		"base":   true,
		"br":     true,
		"col":    true,
		"embed":  true,
		"hr":     true,
		"img":    true,
		"input":  true,
		"link":   true,
		"meta":   true,
		"param":  true,
		"source": true,
		"track":  true,
		"wbr":    true,
	}
	return voidElements[tagName]
}
