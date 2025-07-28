package components

import (
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

const (
	defaultVerticalAlign = "top"
)

// generateDecimalCSSClass creates precise CSS class names like mj-column-per-14-285714285714286
func generateDecimalCSSClass(percentage float64) string {
	integerPart := int(percentage)
	decimalPart := percentage - float64(integerPart)

	if decimalPart == 0 {
		// No decimal part (e.g., 50.0% -> mj-column-per-50)
		return fmt.Sprintf("mj-column-per-%d", integerPart)
	}

	// With decimal part (e.g., 14.285714285714286% -> mj-column-per-14-285714285714286)
	decimalString := fmt.Sprintf("%.15f", decimalPart)[2:] // Remove "0."
	decimalString = strings.TrimRight(decimalString, "0")  // Remove trailing zeros
	return fmt.Sprintf("mj-column-per-%d-%s", integerPart, decimalString)
}

// MJGroupComponent represents mj-group - horizontal grouping of columns
type MJGroupComponent struct {
	*BaseComponent
}

// NewMJGroupComponent creates a new mj-group component
func NewMJGroupComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJGroupComponent {
	return &MJGroupComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJGroupComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "direction":
		return "ltr"
	case "vertical-align":
		return defaultVerticalAlign
	case "width":
		return "100%"
	default:
		return ""
	}
}

func (c *MJGroupComponent) getAttribute(name string) string {
	return c.GetAttributeWithDefault(c, name)
}

func (c *MJGroupComponent) Render() (string, error) {
	var output strings.Builder

	direction := c.getAttribute("direction")
	verticalAlign := c.getAttribute("vertical-align")
	backgroundColor := c.getAttribute("background-color")

	// Count mj-column children to calculate percentage per column
	columnCount := 0
	for _, child := range c.Children {
		if _, ok := child.(*MJColumnComponent); ok {
			columnCount++
		}
	}

	// Calculate precise percentage per column (default to 100% if no columns)
	percentagePerColumn := 100.0
	if columnCount > 0 {
		percentagePerColumn = 100.0 / float64(columnCount)
	}

	// Group always takes full width of its container
	cssClass := "mj-column-per-100"

	// Root div wrapper (following MRML set_style_root_div)
	// Note: Class order should be "mj-column-per-100 mj-outlook-group-fix" to match MRML
	rootDiv := html.NewHTMLTag("div")
	c.AddDebugAttribute(rootDiv, "group")
	rootDiv.AddAttribute("class", fmt.Sprintf("%s mj-outlook-group-fix", cssClass)).
		AddStyle("font-size", "0"). // Note: "0" not "0px" to match MRML
		AddStyle("line-height", "0").
		AddStyle("text-align", "left").
		AddStyle("display", "inline-block").
		AddStyle("width", "100%").
		AddStyle("direction", direction)

	// Only add vertical-align if it's not the default value
	if verticalAlign != defaultVerticalAlign {
		rootDiv.AddStyle("vertical-align", verticalAlign)
	}

	if backgroundColor != "" {
		rootDiv.AddStyle("background-color", backgroundColor)
	}

	output.WriteString(rootDiv.RenderOpen())

	// MSO conditional table structure
	output.WriteString(html.RenderMSOConditional(
		"<table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr>"))

	// Render each column in the group
	for _, child := range c.Children {
		if columnComp, ok := child.(*MJColumnComponent); ok {
			// Set the column width to the calculated percentage if no explicit width
			if columnComp.GetAttribute("width") == nil {
				// Override the column's calculated width for group context with precise decimal
				percentageWidth := fmt.Sprintf("%.15f%%", percentagePerColumn)
				// Remove trailing zeros for cleaner output
				percentageWidth = strings.TrimRight(percentageWidth, "0")
				percentageWidth = strings.TrimRight(percentageWidth, ".")
				if !strings.HasSuffix(percentageWidth, "%") {
					percentageWidth += "%"
				}
				columnComp.Attrs["width"] = percentageWidth
			}

			// Set mobile-width signal for MRML compatibility (like group/render.rs:93)
			columnComp.Attrs["mobile-width"] = "mobile-width"

			// MSO conditional TD for each column (following MRML render_children pattern)
			output.WriteString(html.RenderMSOConditional(
				fmt.Sprintf("<td style=\"vertical-align:%s;width:%s;\">", verticalAlign, columnComp.GetWidthAsPixel())))

			// Set group context for child rendering
			childOpts := *c.RenderOpts // Copy the options
			childOpts.InsideGroup = true
			columnComp.RenderOpts = &childOpts

			// Render column content with padding support table wrapper
			childHTML, err := child.Render()
			if err != nil {
				return "", err
			}

			// Render column directly without extra table wrapper (MRML structure)
			output.WriteString(childHTML)

			// Close MSO conditional TD
			output.WriteString(html.RenderMSOConditional("</td>"))
		}
	}

	// Close MSO conditional table
	output.WriteString(html.RenderMSOConditional("</tr></table>"))

	// Close root div
	output.WriteString(rootDiv.RenderClose())

	return output.String(), nil
}

func (c *MJGroupComponent) GetTagName() string {
	return "mj-group"
}
