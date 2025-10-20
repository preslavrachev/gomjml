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

	// MSO conditional comment for Outlook compatibility - calculate width based on container width minus padding
	// Container width minus divider padding (25px left + 25px right = 50px total from default "10px 25px")
	// AIDEV-NOTE: width-flow-divider; divider gets containerWidth from column and subtracts its own padding for MSO table width
	containerWidth := c.GetContainerWidth()
	if containerWidth <= 0 {
		containerWidth = 600 // fallback
	}

	// Parse divider padding to get accurate left + right values
	leftPadding, rightPadding := c.parseDividerPaddingLeftRight(padding)

	// Override with individual padding attributes if present
	if pl := c.GetAttributeFast(c, constants.MJMLPaddingLeft); pl != "" {
		if px, err := styles.ParsePixel(pl); err == nil && px != nil {
			leftPadding = int(px.Value)
		}
	}
	if pr := c.GetAttributeFast(c, constants.MJMLPaddingRight); pr != "" {
		if px, err := styles.ParsePixel(pr); err == nil && px != nil {
			rightPadding = int(px.Value)
		}
	}

	availableWidth := containerWidth - leftPadding - rightPadding
	msoWidth := availableWidth

	// Apply width attribute (supports percentages)
	if size, err := styles.ParseSize(width); err == nil {
		if size.IsPercent() {
			msoWidth = int(float64(availableWidth) * size.Value() / 100.0)
		} else {
			msoWidth = int(size.Value())
		}
	}

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
	if _, err := w.WriteString(strconv.Itoa(msoWidth)); err != nil {
		return err
	}
	if _, err := w.WriteString(`px;" role="presentation" width="`); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(msoWidth)); err != nil {
		return err
	}
	if _, err := w.WriteString(`px" ><tr><td style="height:0;line-height:0;"> &nbsp; </td></tr></table><![endif]-->`); err != nil {
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

// parseDividerPaddingLeftRight parses CSS padding shorthand to get left and right padding values in pixels
// Optimized to avoid allocations by parsing in place
func (c *MJDividerComponent) parseDividerPaddingLeftRight(padding string) (left, right int) {
	if padding == "" {
		return 0, 0
	}

	parts := strings.Fields(padding)
	parse := func(part string) int {
		px, err := styles.ParsePixel(part)
		if err != nil || px == nil {
			return 0
		}
		return int(px.Value)
	}

	switch len(parts) {
	case 1:
		v := parse(parts[0])
		return v, v
	case 2:
		v := parse(parts[1])
		return v, v
	case 3:
		v := parse(parts[1])
		return v, v
	case 4:
		return parse(parts[3]), parse(parts[1])
	default:
		return 0, 0
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
