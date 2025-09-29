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
	if _, err := w.WriteString(`<!--[if mso | IE]><table border="0" cellpadding="0" cellspacing="0" role="presentation" align="`); err != nil {
		return err
	}
	if _, err := w.WriteString(align); err != nil {
		return err
	}
	if _, err := w.WriteString(`" width="`); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(msoWidth)); err != nil {
		return err
	}
	if _, err := w.WriteString(`px" style="border-top:`); err != nil {
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
	if _, err := w.WriteString(`px;"><tr><td style="height:0;line-height:0;">&nbsp;</td></tr></table><![endif]-->`); err != nil {
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
	if len(padding) == 0 {
		return 0, 0
	}

	// Fast path: single value like "20px"
	if len(padding) > 2 && padding[len(padding)-2:] == "px" {
		// Parse without allocating substring
		if value, err := strconv.Atoi(padding[:len(padding)-2]); err == nil {
			return value, value // same value for all sides
		}
	}

	// Handle "10px 25px" format - find space separator without Fields allocation
	spaceIdx := -1
	for i := 0; i < len(padding); i++ {
		if padding[i] == ' ' {
			spaceIdx = i
			break
		}
	}

	if spaceIdx > 0 {
		// Find second token (skip spaces)
		secondStart := spaceIdx + 1
		for secondStart < len(padding) && padding[secondStart] == ' ' {
			secondStart++
		}

		if secondStart < len(padding) && len(padding) > secondStart+2 && padding[len(padding)-2:] == "px" {
			// Parse second value (left/right padding) without substring allocation
			if value, err := strconv.Atoi(padding[secondStart : len(padding)-2]); err == nil {
				return value, value
			}
		}
	}

	return 0, 0
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
