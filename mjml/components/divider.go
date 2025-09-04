package components

import (
	"io"
	"strconv"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
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
	padding := c.getAttribute("padding")
	borderColor := c.getAttribute("border-color")
	borderStyle := c.getAttribute("border-style")
	borderWidth := c.getAttribute("border-width")
	align := c.getAttribute("align")

	// Calculate margin based on alignment (matching MRML logic)
	var margin string
	switch align {
	case "left":
		margin = "0px"
	case "right":
		margin = "0px 0px 0px auto"
	default:
		margin = "0px auto"
	}

	// Create TR element
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}

	// Table cell with padding and center alignment
	td := html.NewHTMLTag("td").
		AddAttribute("align", "center").
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("word-break", "break-word")

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
	width := c.getAttribute("width")
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
	msoWidth := containerWidth - leftPadding - rightPadding

	// Build MSO table directly to writer to avoid fmt.Sprintf allocation
	if _, err := w.WriteString(`<!--[if mso | IE]><table border="0" cellpadding="0" cellspacing="0" role="presentation" align="center" width="`); err != nil {
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
	if _, err := w.WriteString(`;font-size:1px;margin:0px auto;width:`); err != nil {
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
