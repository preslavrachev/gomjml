package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/mjml/styles"
	"github.com/preslavrachev/gomjml/parser"
)

// MJImageComponent represents mj-image
type MJImageComponent struct {
	*BaseComponent
}

// NewMJImageComponent creates a new mj-image component
func NewMJImageComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJImageComponent {
	return &MJImageComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJImageComponent) GetTagName() string {
	return "mj-image"
}

// Render implements optimized Writer-based rendering for MJImageComponent
func (c *MJImageComponent) Render(w io.StringWriter) error {
	// Get attributes with proper resolution order (element -> class -> global -> default)
	align := c.GetAttributeWithDefault(c, constants.MJMLAlign)
	border := c.GetAttributeWithDefault(c, constants.MJMLBorder)
	borderRadius := c.GetAttributeWithDefault(c, constants.MJMLBorderRadius)
	height := c.GetAttributeWithDefault(c, constants.MJMLHeight)
	href := c.GetAttributeWithDefault(c, constants.MJMLHref)
	padding := c.GetAttributeWithDefault(c, constants.MJMLPadding)
	rel := c.GetAttributeWithDefault(c, "rel")
	src := c.GetAttributeWithDefault(c, constants.MJMLSrc)
	target := c.GetAttributeWithDefault(c, constants.MJMLTarget)
	title := c.GetAttributeWithDefault(c, constants.MJMLTitle)

	widthAttr := c.GetAttribute("width")
	width := ""
	if widthAttr != nil && *widthAttr != "" {
		width = *widthAttr
	} else {
		width = c.calculateDefaultWidth()
	}
	containerBackground := c.GetAttributeWithDefault(c, constants.MJMLContainerBackgroundColor)
	fluidOnMobile := c.GetAttributeWithDefault(c, "fluid-on-mobile")
	paddingTop := c.GetAttributeWithDefault(c, constants.MJMLPaddingTop)
	paddingRight := c.GetAttributeWithDefault(c, constants.MJMLPaddingRight)
	paddingBottom := c.GetAttributeWithDefault(c, constants.MJMLPaddingBottom)
	paddingLeft := c.GetAttributeWithDefault(c, constants.MJMLPaddingLeft)

	// MJML always emits an alt attribute (falling back to the empty string when
	// not provided) to preserve accessibility defaults. Use the attribute
	// resolution pipeline so global attributes and mj-attributes blocks are
	// honoured.
	alt := c.GetAttributeWithDefault(c, constants.MJMLAlt)

	if src == "" {
		return fmt.Errorf("mj-image requires src attribute")
	}

	// Parse width to remove 'px' suffix for img width attribute
	imgWidth := width
	if strings.HasSuffix(width, "px") {
		imgWidth = strings.TrimSuffix(width, "px")
	}

	// Parse height to remove 'px' suffix for img height attribute
	imgHeight := height
	if strings.HasSuffix(height, "px") {
		imgHeight = strings.TrimSuffix(height, "px")
	}

	// Create TR element
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}

	// Create TD container with alignment and base styles
	tdTag := html.NewHTMLTag("td").
		AddAttribute(constants.AttrAlign, align).
		AddStyle(constants.CSSFontSize, "0px").
		AddStyle(constants.CSSPadding, padding).
		AddStyle("word-break", "break-word")

	if containerBackground != "" {
		tdTag.AddStyle(constants.CSSBackground, containerBackground)
	}
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

	// Add css-class if present
	if cssClass := c.BuildClassAttribute(); cssClass != "" {
		tdTag.AddAttribute("class", cssClass)
	}

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// Image table
	tableTag := html.NewHTMLTag("table").
		AddAttribute(constants.AttrBorder, "0").
		AddAttribute(constants.AttrCellPadding, "0").
		AddAttribute(constants.AttrCellSpacing, "0").
		AddAttribute(constants.AttrRole, "presentation").
		AddStyle(constants.CSSBorderCollapse, "collapse").
		AddStyle("border-spacing", "0px")

	if fluidOnMobile == "true" {
		tableTag.AddAttribute(constants.AttrClass, "mj-full-width-mobile")
	}

	if err := tableTag.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody><tr>"); err != nil {
		return err
	}

	// Image cell with width constraint
	imageTdTag := html.NewHTMLTag("td")
	if width != "" {
		imageTdTag.AddStyle(constants.CSSWidth, width)
	}
	if fluidOnMobile == "true" {
		imageTdTag.AddAttribute(constants.AttrClass, "mj-full-width-mobile")
	}

	if err := imageTdTag.RenderOpen(w); err != nil {
		return err
	}

	// Optional link wrapper
	if href != "" {
		linkTag := html.NewHTMLTag("a").
			AddAttribute(constants.AttrHref, href)

		if rel != "" {
			linkTag.AddAttribute(constants.AttrRel, rel)
		}
		if target != "" {
			linkTag.AddAttribute(constants.AttrTarget, target)
		}

		if err := linkTag.RenderOpen(w); err != nil {
			return err
		}
	}

	// Image element with styles
	imgTag := html.NewHTMLTag("img")
	c.AddDebugAttribute(imgTag, "image")

	// Set image attributes following MJML ordering.
	imgTag.AddAttribute("alt", alt)
	if imgHeight != "" {
		imgTag.AddAttribute(constants.AttrHeight, imgHeight)
	}
	imgTag.AddAttribute(constants.AttrSrc, src)
	if title != "" {
		imgTag.AddAttribute(constants.AttrTitle, title)
	}
	if imgWidth != "" {
		imgTag.AddAttribute(constants.AttrWidth, imgWidth)
	}

	// Apply image styles
	imgTag.AddStyle(constants.CSSBorder, border).
		AddStyle("display", "block").
		AddStyle("outline", "none").
		AddStyle("text-decoration", "none").
		AddStyle(constants.CSSHeight, height).
		AddStyle(constants.CSSWidth, "100%").
		AddStyle(constants.CSSFontSize, "13px")

	if borderRadius != "" {
		imgTag.AddStyle("border-radius", borderRadius)
	}

	if err := imgTag.RenderOpen(w); err != nil {
		return err
	}

	// Close optional link wrapper
	if href != "" {
		if _, err := w.WriteString("</a>"); err != nil {
			return err
		}
	}

	if err := imageTdTag.RenderClose(w); err != nil {
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

func (c *MJImageComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "align":
		return "center"
	case "alt":
		return ""
	case "border":
		return "0"
	case "border-radius":
		return ""
	case "font-size":
		return "13px"
	case "height":
		return "auto"
	case "href":
		return ""
	case "padding":
		return "10px 25px"
	case "rel":
		return ""
	case "src":
		return ""
	case "target":
		return "_blank"
	case "title":
		return ""
	case "width":
		return c.calculateDefaultWidth()
	case "fluid-on-mobile":
		return "false"
	case constants.MJMLContainerBackgroundColor:
		return ""
	default:
		return ""
	}
}

// calculateDefaultWidth calculates the default width for the image
// based on the container width minus horizontal padding
func (c *MJImageComponent) calculateDefaultWidth() string {
	containerWidth := c.GetEffectiveWidth()

	// Determine horizontal padding using shorthand and individual overrides
	padding := c.GetAttributeWithDefault(c, constants.MJMLPadding)
	left, right := 0.0, 0.0
	if spacing, err := styles.ParseSpacing(padding); err == nil && spacing != nil {
		left, right = spacing.Left, spacing.Right
	} else if parts := strings.Fields(padding); len(parts) == 3 {
		if lr, err := styles.ParsePixel(parts[1]); err == nil && lr != nil {
			left, right = lr.Value, lr.Value
		}
	}
	if pl := c.GetAttributeWithDefault(c, constants.MJMLPaddingLeft); pl != "" {
		if px, err := styles.ParsePixel(pl); err == nil && px != nil {
			left = px.Value
		}
	}
	if pr := c.GetAttributeWithDefault(c, constants.MJMLPaddingRight); pr != "" {
		if px, err := styles.ParsePixel(pr); err == nil && px != nil {
			right = px.Value
		}
	}
	horizontalPadding := int(left + right)

	// Subtract border widths if present
	if border := c.GetAttributeWithDefault(c, constants.MJMLBorder); border != "" && border != "none" {
		bw := styles.ParseBorderWidth(border)
		horizontalPadding += bw * 2
	}

	// Calculate available width
	availableWidth := containerWidth - horizontalPadding
	if availableWidth <= 0 {
		availableWidth = containerWidth
	}

	return getPixelWidthString(availableWidth)
}
