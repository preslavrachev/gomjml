package components

import (
	"io"
	"strconv"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
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
func (c *MJButtonComponent) Render(w io.StringWriter) error {
	// Get text content
	textContent := c.Node.Text
	if textContent == "" {
		textContent = "Button"
	}

	// Get attributes with proper default and global attribute resolution
	align := c.GetAttributeWithDefault(c, constants.MJMLAlign)
	backgroundColor := c.GetAttributeWithDefault(c, constants.MJMLBackgroundColor)
	border := c.GetAttributeWithDefault(c, constants.MJMLBorder)
	borderRadius := c.GetAttributeWithDefault(c, constants.MJMLBorderRadius)
	innerPadding := c.GetAttributeWithDefault(c, constants.MJMLInnerPadding)
	padding := c.GetAttributeWithDefault(c, constants.MJMLPadding)
	target := c.GetAttributeWithDefault(c, constants.MJMLTarget)
	verticalAlign := c.GetAttributeWithDefault(c, constants.MJMLVerticalAlign)
	href := c.GetAttributeWithDefault(c, constants.MJMLHref)
	width := c.GetAttributeWithDefault(c, constants.MJMLWidth)
	containerBackground := c.GetAttributeWithDefault(c, constants.MJMLContainerBackgroundColor)
	borderTop := c.GetAttributeWithDefault(c, constants.MJMLBorderTop)
	borderRight := c.GetAttributeWithDefault(c, constants.MJMLBorderRight)
	borderBottom := c.GetAttributeWithDefault(c, constants.MJMLBorderBottom)
	borderLeft := c.GetAttributeWithDefault(c, constants.MJMLBorderLeft)
	paddingTop := c.GetAttributeWithDefault(c, constants.MJMLPaddingTop)
	paddingRight := c.GetAttributeWithDefault(c, constants.MJMLPaddingRight)
	paddingBottom := c.GetAttributeWithDefault(c, constants.MJMLPaddingBottom)
	paddingLeft := c.GetAttributeWithDefault(c, constants.MJMLPaddingLeft)

	// Font-related attributes (needed by both td and content elements)
	fontFamily := c.GetAttributeWithDefault(c, constants.MJMLFontFamily)
	fontSize := c.GetAttributeWithDefault(c, constants.MJMLFontSize)
	fontWeight := c.GetAttributeWithDefault(c, constants.MJMLFontWeight)
	fontStyle := c.GetAttributeWithDefault(c, constants.MJMLFontStyle)
	color := c.GetAttributeWithDefault(c, constants.MJMLColor)
	lineHeight := c.GetAttributeWithDefault(c, constants.MJMLLineHeight)
	textDecoration := c.GetAttributeWithDefault(c, constants.MJMLTextDecoration)
	textTransform := c.GetAttributeWithDefault(c, "text-transform")

	// Determine if we use <a> or <p> tag
	tagName := "p"
	if href != "" {
		tagName = "a"
	}

	// Create TR element
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}

	// Create TD with alignment and base styles
	tdTag := html.NewHTMLTag("td").
		AddAttribute(constants.AttrAlign, align)

	// Only add vertical-align attribute if not inside an mj-hero
	// In mj-hero context, MRML doesn't include this attribute
	if !c.RenderOpts.InsideHero {
		tdTag.AddAttribute(constants.AttrVerticalAlign, verticalAlign)
	}

	if containerBackground != "" {
		tdTag.AddStyle(constants.CSSBackground, containerBackground)
	}

	tdTag.AddStyle(constants.CSSFontSize, "0px").
		AddStyle(constants.CSSPadding, padding)

	if paddingTop != "" {
		tdTag.AddStyle(constants.CSSPaddingTop, paddingTop)
	}
	if paddingRight != "" {
		tdTag.AddStyle(constants.CSSPaddingRight, paddingRight)
	}
	if paddingBottom != "" {
		tdTag.AddStyle(constants.CSSPaddingBottom, paddingBottom)
	}
	if paddingLeft != "" {
		tdTag.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}

	tdTag.AddStyle(constants.CSSWordBreak, "break-word")

	// Add css-class if present
	if cssClass := c.BuildClassAttribute(); cssClass != "" {
		tdTag.AddAttribute(constants.AttrClass, cssClass)
	}

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// Button table structure
	tableTag := html.NewHTMLTag("table")
	c.AddDebugAttribute(tableTag, "button")
	tableTag.
		AddAttribute(constants.AttrBorder, "0").
		AddAttribute(constants.AttrCellPadding, "0").
		AddAttribute(constants.AttrCellSpacing, "0").
		AddAttribute(constants.AttrRole, "presentation").
		AddStyle(constants.CSSBorderCollapse, constants.BorderCollapseSeparate)

	// Add width to table if specified
	if width != "" {
		tableTag.AddStyle(constants.CSSWidth, width)
	}

	tableTag.AddStyle(constants.CSSLineHeight, "100%")

	if err := tableTag.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody><tr>"); err != nil {
		return err
	}

	// Button cell with background and border styles
	buttonTdTag := html.NewHTMLTag("td").
		AddAttribute(constants.AttrAlign, constants.AlignCenter).
		AddAttribute(constants.AttrBgcolor, backgroundColor).
		AddAttribute(constants.AttrRole, "presentation").
		AddAttribute(constants.AttrValign, verticalAlign).
		AddStyle(constants.CSSBorder, border)

	if borderBottom != "" {
		buttonTdTag.AddStyle(constants.CSSBorderBottom, borderBottom)
	}
	if borderLeft != "" {
		buttonTdTag.AddStyle(constants.CSSBorderLeft, borderLeft)
	}

	buttonTdTag.AddStyle(constants.CSSBorderRadius, borderRadius)

	if borderRight != "" {
		buttonTdTag.AddStyle(constants.CSSBorderRight, borderRight)
	}
	if borderTop != "" {
		buttonTdTag.AddStyle(constants.CSSBorderTop, borderTop)
	}

	buttonTdTag.AddStyle("cursor", "auto")
	if fontStyle != "" {
		buttonTdTag.AddStyle(constants.CSSFontStyle, fontStyle)
	}
	buttonTdTag.AddStyle("mso-padding-alt", innerPadding).
		AddStyle(constants.CSSBackground, backgroundColor)

	if err := buttonTdTag.RenderOpen(w); err != nil {
		return err
	}

	// Button content (a or p tag)
	contentTag := html.NewHTMLTag(tagName)
	if href != "" {
		contentTag.AddAttribute(constants.AttrHref, href)
		if target != "" {
			contentTag.AddAttribute(constants.AttrTarget, target)
		}
		// Add rel attribute if specified
		rel := c.GetAttributeWithDefault(c, constants.AttrRel)
		if rel != "" {
			contentTag.AddAttribute(constants.AttrRel, rel)
		}
	}

	// Calculate inner width for anchor tag
	innerWidth := c.calculateInnerWidth(width, innerPadding)

	// Apply button content styles in MRML order
	contentTag.AddStyle(constants.CSSDisplay, constants.DisplayInlineBlock)

	// Add width if calculated
	if innerWidth != "" {
		contentTag.AddStyle(constants.CSSWidth, innerWidth)
	}

	contentTag.AddStyle(constants.CSSBackground, backgroundColor).
		AddStyle(constants.CSSColor, color).
		AddStyle(constants.CSSFontFamily, fontFamily).
		AddStyle(constants.CSSFontSize, fontSize)

	if fontStyle != "" {
		contentTag.AddStyle(constants.CSSFontStyle, fontStyle)
	}

	contentTag.AddStyle(constants.CSSFontWeight, fontWeight).
		AddStyle(constants.CSSLineHeight, lineHeight).
		AddStyle(constants.CSSMargin, "0").
		AddStyle(constants.CSSTextDecoration, textDecoration).
		AddStyle(constants.CSSTextTransform, textTransform).
		AddStyle(constants.CSSPadding, innerPadding).
		AddStyle("mso-padding-alt", "0px").
		AddStyle(constants.CSSBorderRadius, borderRadius)

	if err := contentTag.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString(textContent); err != nil {
		return err
	}
	if err := contentTag.RenderClose(w); err != nil {
		return err
	}
	if err := buttonTdTag.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr></tbody>"); err != nil {
		return err
	}
	if err := tableTag.RenderClose(w); err != nil {
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
