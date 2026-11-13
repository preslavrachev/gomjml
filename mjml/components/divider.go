package components

import (
	"io"
	"strconv"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/mjml/styles"
	"github.com/preslavrachev/gomjml/parser"
)

// MJDividerComponent represents mj-divider
type MJDividerComponent struct {
	*BaseComponent
}

// NewMJDividerComponent creates a new mj-divider component
func NewMJDividerComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJDividerComponent {
	return &MJDividerComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJDividerComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "align":
		return "center"
	case "border-color":
		return "#000000"
	case "border-style":
		return "solid"
	case "border-width":
		return "4px"
	case "container-background-color":
		return "transparent"
	case "padding":
		return "10px 25px"
	case "width":
		return "100%"
	default:
		return ""
	}
}

func (c *MJDividerComponent) getAttribute(name string) string {
	return c.GetAttributeWithDefault(c, name)
}

// Render implements optimized Writer-based rendering for MJDividerComponent
func (c *MJDividerComponent) Render(w io.StringWriter) error {
	padding := c.getAttribute(constants.MJMLPadding)
	borderColor := c.getAttribute("border-color")
	borderStyle := c.getAttribute("border-style")
	borderWidth := c.getAttribute("border-width")
	align := strings.ToLower(strings.TrimSpace(c.getAttribute(constants.MJMLAlign)))
	if align == "" {
		align = constants.AlignCenter
	}

	margin := c.marginForAlign(align)

	// Create TR element
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}

	// Table cell with padding and alignment
	td := html.NewHTMLTag("td")
	if align != "" {
		td.AddAttribute(constants.AttrAlign, align)
	}

	// Add css-class if present
	c.SetClassAttribute(td)

	td.AddStyle(constants.CSSFontSize, "0px").
		AddStyle(constants.CSSPadding, padding).
		AddStyle(constants.CSSWordBreak, "break-word")

	// Handle container background color
	containerBgAttr := c.Node.GetAttribute(constants.MJMLContainerBackgroundColor)
	containerBg := c.GetAttributeFast(c, constants.MJMLContainerBackgroundColor)
	if containerBgAttr != "" || containerBg != c.GetDefaultAttribute(constants.MJMLContainerBackgroundColor) {
		td.AddStyle(constants.CSSBackground, containerBg)
	}

	// Handle individual padding properties
	if paddingTop := c.GetAttributeFast(c, constants.MJMLPaddingTop); paddingTop != "" {
		td.AddStyle(constants.CSSPaddingTop, paddingTop)
	}
	if paddingRight := c.GetAttributeFast(c, constants.MJMLPaddingRight); paddingRight != "" {
		td.AddStyle(constants.CSSPaddingRight, paddingRight)
	}
	if paddingBottom := c.GetAttributeFast(c, constants.MJMLPaddingBottom); paddingBottom != "" {
		td.AddStyle(constants.CSSPaddingBottom, paddingBottom)
	}
	if paddingLeft := c.GetAttributeFast(c, constants.MJMLPaddingLeft); paddingLeft != "" {
		td.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}

	if err := td.RenderOpen(w); err != nil {
		return err
	}

	// Create paragraph with border styles matching MRML exact order
	p := html.NewHTMLTag("p")
	c.AddDebugAttribute(p, "divider")
	p.
		AddStyle("border-top", borderStyle+" "+borderWidth+" "+borderColor).
		AddStyle("font-size", "1px").
		AddStyle("margin", margin)

	// Add width (MRML includes default width of 100%)
	width := c.getAttribute(constants.MJMLWidth)
	p = p.AddStyle("width", width)

	// Render paragraph - must be empty, not self-closing to match MRML
	if err := p.RenderOpen(w); err != nil {
		return err
	}
	if err := p.RenderClose(w); err != nil {
		return err
	}

	msoWidth := c.outlookWidth(width)

	// Build MSO table directly to writer to avoid fmt.Sprintf allocation
	if _, err := w.WriteString(`<!--[if mso | IE]><table align="`); err != nil {
		return err
	}
	if _, err := w.WriteString(align); err != nil {
		return err
	}
	if _, err := w.WriteString(`" border="0" cellpadding="0" cellspacing="0" style="border-top:`); err != nil {
		return err
	}
	if _, err := w.WriteString(borderStyle); err != nil {
		return err
	}
	if _, err := w.WriteString(" "); err != nil {
		return err
	}
	if _, err := w.WriteString(borderWidth); err != nil {
		return err
	}
	if _, err := w.WriteString(" "); err != nil {
		return err
	}
	if _, err := w.WriteString(borderColor); err != nil {
		return err
	}
	if _, err := w.WriteString(`;font-size:1px;margin:`); err != nil {
		return err
	}
	if _, err := w.WriteString(margin); err != nil {
		return err
	}
	if _, err := w.WriteString(`;width:`); err != nil {
		return err
	}
	if _, err := w.WriteString(msoWidth); err != nil {
		return err
	}
	if _, err := w.WriteString(`;" role="presentation" width="`); err != nil {
		return err
	}
	if _, err := w.WriteString(msoWidth); err != nil {
		return err
	}
	if _, err := w.WriteString(`" ><tr><td style="height:0;line-height:0;"> &nbsp; </td></tr></table><![endif]-->`); err != nil {
		return err
	}

	if err := td.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr>"); err != nil {
		return err
	}

	return nil
}

func (c *MJDividerComponent) GetTagName() string {
	return "mj-divider"
}

// outlookWidth is the width of the divider's Outlook table, as MJML's
// getOutlookWidth computes it: a px width as given, a percentage of what the
// horizontal paddings leave of the container, or all of it.
func (c *MJDividerComponent) outlookWidth(width string) string {
	containerWidth := c.GetEffectiveWidth()
	paddingSize := c.ShorthandAttrValue(c, "padding", "left") + c.ShorthandAttrValue(c, "padding", "right")
	parsedWidth, _ := styles.ParseLeadingInt(width)
	switch styles.WidthUnit(width) {
	case "%":
		return styles.FormatJSNumber(float64(containerWidth-paddingSize)*(float64(parsedWidth)/100)) + "px"
	case "px":
		return width
	default:
		return strconv.Itoa(containerWidth-paddingSize) + "px"
	}
}

func (c *MJDividerComponent) marginForAlign(align string) string {
	switch align {
	case constants.AlignLeft:
		return "0px"
	case constants.AlignRight:
		return "0px 0px 0px auto"
	default:
		return "0px auto"
	}
}
