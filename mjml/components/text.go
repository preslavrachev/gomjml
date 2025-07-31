package components

import (
	"io"
	"strings"

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
func (c *MJTextComponent) Render(w io.Writer) error {
	debug.DebugLog("mj-text", "render-start", "Starting text component rendering")

	// Get raw inner HTML content (preserve HTML tags and formatting)
	textContent := c.getRawInnerHTML()
	debug.DebugLogWithData("mj-text", "content", "Processing text content", map[string]interface{}{
		"content_length":  len(textContent),
		"container_width": c.GetContainerWidth(),
	})

	// Helper function to get attribute with default
	getAttr := func(name string) string {
		if attr := c.GetAttribute(name); attr != nil {
			return *attr
		}
		return c.GetDefaultAttribute(name)
	}

	// Get attributes
	align := getAttr("align")
	padding := getAttr("padding")

	// Create TR element
	if _, err := w.Write([]byte("<tr>")); err != nil {
		return err
	}

	// Create TD with alignment and base styles
	tdTag := html.NewHTMLTag("td").
		AddAttribute("align", align).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding)

	// Add css-class if present
	if cssClass := c.BuildClassAttribute(); cssClass != "" {
		tdTag.AddAttribute("class", cssClass)
	}

	// Add specific padding overrides if they exist (following MRML/section pattern)
	if paddingTopAttr := c.GetAttribute("padding-top"); paddingTopAttr != nil {
		tdTag.AddStyle("padding-top", *paddingTopAttr)
	}
	if paddingBottomAttr := c.GetAttribute("padding-bottom"); paddingBottomAttr != nil {
		tdTag.AddStyle("padding-bottom", *paddingBottomAttr)
	}
	if paddingLeftAttr := c.GetAttribute("padding-left"); paddingLeftAttr != nil {
		tdTag.AddStyle("padding-left", *paddingLeftAttr)
	}
	if paddingRightAttr := c.GetAttribute("padding-right"); paddingRightAttr != nil {
		tdTag.AddStyle("padding-right", *paddingRightAttr)
	}

	tdTag.AddStyle("word-break", "break-word")

	if _, err := w.Write([]byte(tdTag.RenderOpen())); err != nil {
		return err
	}

	// Create inner div with font styling
	divTag := html.NewHTMLTag("div")
	c.AddDebugAttribute(divTag, "text")

	// Apply font styles using the proper interface method
	fontFamily := c.GetAttributeWithDefault(c, "font-family")
	fontSize := c.GetAttributeWithDefault(c, "font-size")
	fontWeight := c.GetAttributeWithDefault(c, "font-weight")
	fontStyle := c.GetAttributeWithDefault(c, "font-style")
	color := c.GetAttributeWithDefault(c, "color")
	lineHeight := c.GetAttributeWithDefault(c, "line-height")
	textAlign := c.GetAttributeWithDefault(c, "align")
	textDecoration := c.GetAttributeWithDefault(c, "text-decoration")
	textTransform := c.GetAttributeWithDefault(c, "text-transform")
	letterSpacing := c.GetAttributeWithDefault(c, "letter-spacing")

	// Apply styles in the order expected by MRML
	if fontFamily != "" {
		divTag.AddStyle("font-family", fontFamily)
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
		divTag.AddStyle("color", color)
	}
	if fontStyle != "" {
		divTag.AddStyle("font-style", fontStyle)
	}
	if textDecoration != "" {
		divTag.AddStyle("text-decoration", textDecoration)
	}

	if _, err := w.Write([]byte(divTag.RenderOpen())); err != nil {
		return err
	}
	if _, err := w.Write([]byte(textContent)); err != nil {
		return err
	}
	if _, err := w.Write([]byte(divTag.RenderClose())); err != nil {
		return err
	}
	if _, err := w.Write([]byte(tdTag.RenderClose())); err != nil {
		return err
	}
	if _, err := w.Write([]byte("</tr>")); err != nil {
		return err
	}

	return nil
}

func (c *MJTextComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "font-size":
		return "13px"
	case "color":
		return "#000000"
	case "align":
		return "left"
	case "font-family":
		return fonts.DefaultFontStack
	case "line-height":
		return "1"
	case "padding":
		return "10px 25px"
	default:
		return ""
	}
}

// getRawInnerHTML reconstructs the original inner HTML content of the mj-text element
// This is needed because our parser splits content, but mj-text needs to preserve HTML
func (c *MJTextComponent) getRawInnerHTML() string {
	// If we have children (HTML elements), we need to reconstruct the original HTML
	if len(c.Node.Children) > 0 {
		var html strings.Builder

		// Add any text content before children (trimmed)
		if c.Node.Text != "" {
			trimmedText := strings.TrimSpace(c.Node.Text)
			if trimmedText != "" {
				html.WriteString(c.restoreHTMLEntities(trimmedText))
			}
		}

		// Add children as HTML elements
		for _, child := range c.Node.Children {
			html.WriteString(c.reconstructHTMLElement(child))
		}

		return html.String()
	}

	// If no children, return the text content with HTML entities restored and whitespace trimmed
	return c.restoreHTMLEntities(strings.TrimSpace(c.Node.Text))
}

// restoreHTMLEntities converts Unicode characters back to HTML entities for proper output
func (c *MJTextComponent) restoreHTMLEntities(text string) string {
	// Convert Unicode non-breaking space back to HTML entity
	result := strings.ReplaceAll(text, "\u00A0", "&#xA0;")
	return result
}

// reconstructHTMLElement reconstructs an HTML element from a parsed node
func (c *MJTextComponent) reconstructHTMLElement(node *parser.MJMLNode) string {
	var html strings.Builder

	tagName := node.XMLName.Local

	// Check if this is a void element (self-closing)
	isVoidElement := isVoidHTMLElement(tagName)

	// Opening tag
	html.WriteString("<")
	html.WriteString(tagName)

	// Attributes
	for _, attr := range node.Attrs {
		html.WriteString(" ")
		html.WriteString(attr.Name.Local)
		html.WriteString(`="`)
		html.WriteString(attr.Value)
		html.WriteString(`"`)
	}

	if isVoidElement {
		// Self-closing tag
		html.WriteString(" />")
		return html.String()
	}

	html.WriteString(">")

	// Content (text + children)
	if node.Text != "" {
		html.WriteString(c.restoreHTMLEntities(node.Text))
	}

	for _, child := range node.Children {
		html.WriteString(c.reconstructHTMLElement(child))
	}

	// Closing tag
	html.WriteString("</")
	html.WriteString(tagName)
	html.WriteString(">")

	return html.String()
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
