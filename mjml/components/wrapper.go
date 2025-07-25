package components

import (
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/parser"
)

// MJWrapperComponent represents mj-wrapper
type MJWrapperComponent struct {
	*BaseComponent
}

// NewMJWrapperComponent creates a new mj-wrapper component
func NewMJWrapperComponent(node *parser.MJMLNode) *MJWrapperComponent {
	return &MJWrapperComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJWrapperComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "background-position":
		return "top center"
	case "background-repeat":
		return "repeat"
	case "background-size":
		return "auto"
	case "direction":
		return "ltr"
	case "padding":
		return "20px 0"
	case "text-align":
		return "center"
	case "text-padding":
		return "4px 4px 4px 0"
	default:
		return ""
	}
}

func (c *MJWrapperComponent) getAttribute(name string) string {
	return c.GetAttributeWithDefault(c, name)
}

// getBorderWidth calculates the total border width from border attribute
func (c *MJWrapperComponent) getBorderWidth() int {
	border := c.getAttribute("border")
	if border == "" {
		return 0
	}

	// Parse border width (e.g., "2px solid #333" -> 2)
	// Simple parsing for the common case
	if strings.Contains(border, "1px") {
		return 2 // 1px on each side
	}
	if strings.Contains(border, "2px") {
		return 4 // 2px on each side
	}
	if strings.Contains(border, "3px") {
		return 6 // 3px on each side
	}
	return 0
}

// getEffectiveWidth calculates width minus border width
func (c *MJWrapperComponent) getEffectiveWidth() int {
	baseWidth := GetDefaultBodyWidthPixels()
	borderWidth := c.getBorderWidth()
	return baseWidth - borderWidth
}

func (c *MJWrapperComponent) isFullWidth() bool {
	// Full width if explicitly set
	if c.getAttribute("full-width") == "full-width" {
		return true
	}

	// Full width if has background attributes (like background-color)
	if c.getAttribute("background-color") != "" {
		return true
	}
	if c.getAttribute("background-image") != "" {
		return true
	}

	return false
}

func (c *MJWrapperComponent) renderFullWidth() (string, error) {
	var output strings.Builder

	// Get wrapper attributes
	padding := c.getAttribute("padding")
	textAlign := c.getAttribute("text-align")
	direction := c.getAttribute("direction")
	cssClass := c.getAttribute("css-class")

	// Full width wrapper div with background styles
	wrapperDiv := html.NewHTMLTag("div")
	if cssClass != "" {
		wrapperDiv.AddAttribute("class", cssClass)
	}

	// Apply background and border styles using helper methods
	c.ApplyBackgroundStyles(wrapperDiv)
	c.ApplyBorderStyles(wrapperDiv)

	output.WriteString(wrapperDiv.RenderOpen())

	// MSO conditional for inner container
	msoTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("width", "100%")

	// Add bgcolor to MSO table if background-color is set
	if bgColor := c.getAttribute("background-color"); bgColor != "" {
		msoTable.AddAttribute("bgcolor", bgColor).
			AddAttribute("align", "center").
			AddAttribute("width", fmt.Sprintf("%d", GetDefaultBodyWidthPixels())).
			AddStyle("width", GetDefaultBodyWidth())
	}

	msoTd := html.NewHTMLTag("td").
		AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("text-align", textAlign)

	output.WriteString(html.RenderMSOConditional(
		msoTable.RenderOpen() + "<tr>" + msoTd.RenderOpen()))

	// Inner container div
	effectiveWidth := c.getEffectiveWidth()
	innerDiv := html.NewHTMLTag("div").
		AddStyle("margin", "0px auto").
		AddStyle("max-width", fmt.Sprintf("%dpx", effectiveWidth))

	output.WriteString(innerDiv.RenderOpen())

	// Inner table
	innerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddStyle("width", "100%")

	// Apply border styles to inner table as well
	c.ApplyBorderStyles(innerTable)

	output.WriteString(innerTable.RenderOpen())
	output.WriteString("<tbody><tr>")

	// Inner TD
	innerTd := html.NewHTMLTag("td").
		AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("text-align", textAlign)

	// Apply border styles to the inner TD which should contain the actual border
	c.ApplyBorderStyles(innerTd)

	output.WriteString(innerTd.RenderOpen())

	// Render children - pass the effective width (600px - border width)
	effectiveWidth = c.getEffectiveWidth()
	for _, child := range c.Children {
		child.SetContainerWidth(effectiveWidth)
		childHTML, err := child.Render()
		if err != nil {
			return "", err
		}
		output.WriteString(childHTML)
	}

	output.WriteString(innerTd.RenderClose())
	output.WriteString("</tr></tbody>")
	output.WriteString(innerTable.RenderClose())
	output.WriteString(innerDiv.RenderClose())

	// Close MSO conditional
	output.WriteString(html.RenderMSOConditional(msoTd.RenderClose() + "</tr>" + msoTable.RenderClose()))
	output.WriteString(wrapperDiv.RenderClose())

	return output.String(), nil
}

func (c *MJWrapperComponent) renderSimple() (string, error) {
	var output strings.Builder

	// Get wrapper attributes
	padding := c.getAttribute("padding")
	textAlign := c.getAttribute("text-align")
	direction := c.getAttribute("direction")
	effectiveWidth := c.getEffectiveWidth()

	// MSO conditional table wrapper (should use full default body width, not effective width)
	msoTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center").
		AddAttribute("width", fmt.Sprintf("%d", GetDefaultBodyWidthPixels())).
		AddStyle("width", GetDefaultBodyWidth())

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	output.WriteString(html.RenderMSOConditional(
		msoTable.RenderOpen() + "<tr>" + msoTd.RenderOpen()))

	// Main wrapper div (match MRML property order: margin, border-radius, max-width)
	wrapperDiv := html.NewHTMLTag("div").
		AddStyle("margin", "0px auto")

	// Add border-radius before max-width to match MRML order
	if borderRadius := c.getAttribute("border-radius"); borderRadius != "" {
		wrapperDiv.AddStyle("border-radius", borderRadius)
	}

	wrapperDiv.AddStyle("max-width", GetDefaultBodyWidth())

	output.WriteString(wrapperDiv.RenderOpen())

	// Inner table (match MRML order: width, border-radius)
	innerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center").
		AddStyle("width", "100%")

	// Add border-radius after width to match MRML order
	if borderRadius := c.getAttribute("border-radius"); borderRadius != "" {
		innerTable.AddStyle("border-radius", borderRadius)
	}

	output.WriteString(innerTable.RenderOpen())
	output.WriteString("<tbody><tr>")

	// Main TD with wrapper styles (match MRML order: border first, then other properties)
	mainTd := html.NewHTMLTag("td")

	// Add border first to match MRML order
	if border := c.getAttribute("border"); border != "" {
		mainTd.AddStyle("border", border)
	}

	mainTd.AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("text-align", textAlign)

	output.WriteString(mainTd.RenderOpen())

	// For basic wrapper, we need a specific MSO conditional pattern
	// that matches MRML's output more closely - use original body width for wrapper MSO
	output.WriteString(html.RenderMSOConditional(
		fmt.Sprintf("<table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr><td width=\"%dpx\">", GetDefaultBodyWidthPixels())))

	// Render children - pass the effective width (600px - border width)
	for _, child := range c.Children {
		child.SetContainerWidth(effectiveWidth)
		childHTML, err := child.Render()
		if err != nil {
			return "", err
		}
		output.WriteString(childHTML)
	}

	output.WriteString(html.RenderMSOConditional("</td></tr></table>"))

	output.WriteString(mainTd.RenderClose())
	output.WriteString("</tr></tbody>")
	output.WriteString(innerTable.RenderClose())
	output.WriteString(wrapperDiv.RenderClose())

	// Close MSO conditional
	output.WriteString(html.RenderMSOConditional(msoTd.RenderClose() + "</tr>" + msoTable.RenderClose()))

	return output.String(), nil
}

func (c *MJWrapperComponent) Render() (string, error) {
	if c.isFullWidth() {
		return c.renderFullWidth()
	}
	return c.renderSimple()
}

func (c *MJWrapperComponent) GetTagName() string {
	return "mj-wrapper"
}
