package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
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

func (c *MJGroupComponent) GetTagName() string {
	return "mj-group"
}

// Render implements optimized Writer-based rendering for MJGroupComponent
func (c *MJGroupComponent) Render(w io.StringWriter) error {
	direction := c.getAttribute("direction")
	verticalAlign := c.getAttribute("vertical-align")
	backgroundColor := c.getAttribute("background-color")
	groupWidth := c.getAttribute("width")

	// Count mj-column children to calculate width per column
	columnCount := 0
	for _, child := range c.Children {
		if _, ok := child.(*MJColumnComponent); ok {
			columnCount++
		}
	}

	// Determine group width based on attribute and container width
	var widthClass string
	var groupWidthPx int
	var childWidthPx int

	containerWidth := c.GetEffectiveWidth()

	if strings.HasSuffix(groupWidth, "px") {
		// Pixel width provided explicitly
		fmt.Sscanf(groupWidth, "%dpx", &groupWidthPx)
		widthClass = fmt.Sprintf("mj-column-px-%d", groupWidthPx)
	} else if strings.HasSuffix(groupWidth, "%") {
		// Percentage width – compute relative to container width
		var percent float64
		fmt.Sscanf(groupWidth, "%f%%", &percent)
		groupWidthPx = int(float64(containerWidth) * percent / 100.0)
		widthClass = generateDecimalCSSClass(percent)
	} else {
		// Fallback to 100% of container width
		groupWidthPx = containerWidth
		widthClass = "mj-column-per-100"
	}

	if columnCount > 0 {
		childWidthPx = groupWidthPx / columnCount
	}

	// Root div wrapper (following MRML set_style_root_div)
	// Note: Class order should match MRML output
	rootDiv := html.NewHTMLTag("div")
	c.AddDebugAttribute(rootDiv, "group")
	c.SetClassAttribute(rootDiv, widthClass, "mj-outlook-group-fix")

	rootDiv.AddStyle("font-size", "0"). // Note: "0" not "0px" to match MRML
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

	if err := rootDiv.RenderOpen(w); err != nil {
		return err
	}

	// Determine Outlook-specific class name (css-class + "-outlook")
	outlookClass := ""
	if cssClass := c.GetCSSClass(); cssClass != "" {
		outlookClass = cssClass + "-outlook"
	}

	// Only include vertical-align style in MSO wrapper when explicitly set
	msoVerticalAlign := ""
	if verticalAlign != defaultVerticalAlign {
		msoVerticalAlign = verticalAlign
	}

	// MSO conditional table structure with dynamic bgcolor and wrapper metadata
	if err := html.RenderMSOGroupTableOpen(w, groupWidthPx, backgroundColor, outlookClass, msoVerticalAlign); err != nil {
		return err
	}

	// Render each column in the group
	renderedColumns := 0
	for _, child := range c.Children {
		if child.IsRawElement() {
			if err := child.Render(w); err != nil {
				return err
			}
			continue
		}
		if columnComp, ok := child.(*MJColumnComponent); ok {
			isFirstColumn := renderedColumns == 0
			isLastColumn := renderedColumns == columnCount-1

			// Set the column width based on group's width distribution
			if columnComp.GetAttribute("width") == nil {
				if strings.HasSuffix(groupWidth, "px") {
					// For pixel-based groups, set pixel width for each column
					columnComp.Attrs["width"] = getPixelWidthString(childWidthPx)
				} else {
					// For percentage-based groups, calculate percentage per column
					percentagePerColumn := 100.0 / float64(columnCount)
					percentageWidth := fmt.Sprintf("%.15f%%", percentagePerColumn)
					// Remove trailing zeros for cleaner output
					percentageWidth = strings.TrimRight(percentageWidth, "0")
					percentageWidth = strings.TrimRight(percentageWidth, ".")
					if !strings.HasSuffix(percentageWidth, "%") {
						percentageWidth += "%"
					}
					columnComp.Attrs["width"] = percentageWidth
				}
			}

			// Set mobile-width signal for MRML compatibility (like group/render.rs:93)
			columnComp.Attrs["mobile-width"] = "mobile-width"

			// Ensure child columns receive the group's full width for internal calculations
			columnComp.SetContainerWidth(groupWidthPx)

			// MSO conditional TD for each column with correct width and vertical alignment
			msoWidth := getPixelWidthString(childWidthPx)
			colVAlign := columnComp.GetAttributeWithDefault(columnComp, constants.MJMLVerticalAlign)

			if err := html.RenderMSOGroupTDOpen(w, "", colVAlign, msoWidth, backgroundColor, isFirstColumn); err != nil {
				return err
			}

			// Set group context for child rendering
			childOpts := *c.RenderOpts // Copy the options
			childOpts.InsideGroup = true
			childOpts.GroupColumnCount = columnCount
			columnComp.RenderOpts = &childOpts

			// Render column content with padding support table wrapper
			if err := child.Render(w); err != nil {
				return err
			}

			// Close MSO conditional TD
			if err := html.RenderMSOGroupTDClose(w, isLastColumn); err != nil {
				return err
			}
			renderedColumns++
		}
	}

	// Close MSO conditional table
	if err := html.RenderMSOGroupTableClose(w); err != nil {
		return err
	}

	// Close root div
	if err := rootDiv.RenderClose(w); err != nil {
		return err
	}

	return nil
}
