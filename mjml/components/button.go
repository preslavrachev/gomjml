package components

import (
	"io"
	"strconv"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/fonts"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJButtonComponent represents mj-button
type MJButtonComponent struct {
	*BaseComponent
}

// NewMJButtonComponent creates a new mj-button component
func NewMJButtonComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJButtonComponent {
	return &MJButtonComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

// calculateInnerWidth calculates the inner width of the button by subtracting horizontal padding
func (c *MJButtonComponent) calculateInnerWidth(width, innerPadding string) string {
	if width == "" {
		return ""
	}

	// Parse width (remove "px" suffix)
	widthStr := strings.TrimSuffix(width, "px")
	widthVal, err := strconv.Atoi(widthStr)
	if err != nil {
		return ""
	}

	// Parse inner-padding (format: "10px 25px" or "10px")
	parts := strings.Fields(innerPadding)
	if len(parts) == 0 {
		return width
	}

	// Get horizontal padding (right padding)
	horizontalPadding := parts[0]
	if len(parts) >= 2 {
		horizontalPadding = parts[1]
	}

	// Parse horizontal padding
	paddingStr := strings.TrimSuffix(horizontalPadding, "px")
	paddingVal, err := strconv.Atoi(paddingStr)
	if err != nil {
		return width
	}

	// Calculate inner width (subtract padding from both sides)
	innerWidth := widthVal - (paddingVal * 2)
	if innerWidth <= 0 {
		return width
	}

	return strconv.Itoa(innerWidth) + "px"
}

func (c *MJButtonComponent) GetTagName() string {
	return "mj-button"
}

// Render implements optimized Writer-based rendering for MJButtonComponent
func (c *MJButtonComponent) Render(w io.Writer) error {
	// Get text content
	textContent := c.Node.Text
	if textContent == "" {
		textContent = "Button"
	}

	// Helper function to get attribute with default
	getAttr := func(name string) string {
		if attr := c.GetAttribute(name); attr != nil {
			return *attr
		}
		return c.GetDefaultAttribute(name)
	}

	// Get attributes with defaults matching MRML
	align := getAttr("align")
	backgroundColor := getAttr("background-color")
	border := getAttr("border")
	borderRadius := getAttr("border-radius")
	innerPadding := getAttr("inner-padding")
	padding := getAttr("padding")
	target := getAttr("target")
	verticalAlign := getAttr("vertical-align")
	href := getAttr("href")
	width := getAttr("width")

	// Determine if we use <a> or <p> tag
	tagName := "p"
	if href != "" {
		tagName = "a"
	}

	// Create TR element
	if _, err := w.Write([]byte("<tr>")); err != nil {
		return err
	}

	// Create TD with alignment and base styles
	tdTag := html.NewHTMLTag("td").
		AddAttribute("align", align).
		AddAttribute("vertical-align", verticalAlign).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("word-break", "break-word")

	// Add css-class if present
	if cssClass := c.BuildClassAttribute(); cssClass != "" {
		tdTag.AddAttribute("class", cssClass)
	}

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// Button table structure
	tableTag := html.NewHTMLTag("table")
	c.AddDebugAttribute(tableTag, "button")
	tableTag.
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddStyle("border-collapse", "separate")

	// Add width to table if specified
	if width != "" {
		tableTag.AddStyle("width", width)
	}

	tableTag.AddStyle("line-height", "100%")

	if err := tableTag.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.Write([]byte("<tbody><tr>")); err != nil {
		return err
	}

	// Button cell with background and border styles
	buttonTdTag := html.NewHTMLTag("td").
		AddAttribute("align", "center").
		AddAttribute("bgcolor", backgroundColor).
		AddAttribute("role", "presentation").
		AddAttribute("valign", verticalAlign).
		AddStyle("border", border).
		AddStyle("border-radius", borderRadius).
		AddStyle("cursor", "auto").
		AddStyle("mso-padding-alt", innerPadding).
		AddStyle("background", backgroundColor)

	if err := buttonTdTag.RenderOpen(w); err != nil {
		return err
	}

	// Button content (a or p tag)
	contentTag := html.NewHTMLTag(tagName)
	if href != "" {
		contentTag.AddAttribute("href", href)
		if target != "" {
			contentTag.AddAttribute("target", target)
		}
	}

	// Get font styles first
	fontFamily := c.GetAttributeWithDefault(c, "font-family")
	fontSize := c.GetAttributeWithDefault(c, "font-size")
	fontWeight := c.GetAttributeWithDefault(c, "font-weight")
	color := c.GetAttributeWithDefault(c, "color")
	lineHeight := c.GetAttributeWithDefault(c, "line-height")
	textDecoration := c.GetAttributeWithDefault(c, "text-decoration")
	textTransform := c.GetAttributeWithDefault(c, "text-transform")

	// Calculate inner width for anchor tag
	innerWidth := c.calculateInnerWidth(width, innerPadding)

	// Apply button content styles in MRML order
	contentTag.AddStyle("display", "inline-block")

	// Add width if calculated
	if innerWidth != "" {
		contentTag.AddStyle("width", innerWidth)
	}

	contentTag.AddStyle("background", backgroundColor).
		AddStyle("color", color).
		AddStyle("font-family", fontFamily).
		AddStyle("font-size", fontSize).
		AddStyle("font-weight", fontWeight).
		AddStyle("line-height", lineHeight).
		AddStyle("margin", "0").
		AddStyle("text-decoration", textDecoration).
		AddStyle("text-transform", textTransform).
		AddStyle("padding", innerPadding).
		AddStyle("mso-padding-alt", "0px").
		AddStyle("border-radius", borderRadius)

	if err := contentTag.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.Write([]byte(textContent)); err != nil {
		return err
	}
	if err := contentTag.RenderClose(w); err != nil {
		return err
	}
	if err := buttonTdTag.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.Write([]byte("</tr></tbody>")); err != nil {
		return err
	}
	if err := tableTag.RenderClose(w); err != nil {
		return err
	}
	if err := tdTag.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.Write([]byte("</tr>")); err != nil {
		return err
	}

	return nil
}

func (c *MJButtonComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "align":
		return "center"
	case "background-color":
		return "#414141"
	case "border":
		return "none"
	case "border-radius":
		return "3px"
	case "color":
		return "#ffffff"
	case "font-family":
		return fonts.DefaultFontStack
	case "font-size":
		return "13px"
	case "font-weight":
		return "normal"
	case "inner-padding":
		return "10px 25px"
	case "line-height":
		return "120%"
	case "padding":
		return "10px 25px"
	case "target":
		return "_blank"
	case "text-decoration":
		return "none"
	case "text-transform":
		return "none"
	case "vertical-align":
		return "middle"
	case "href":
		return ""
	default:
		return ""
	}
}
