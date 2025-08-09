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

// MJHeroComponent represents the mj-hero component
type MJHeroComponent struct {
	*BaseComponent
}

func NewMJHeroComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJHeroComponent {
	return &MJHeroComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJHeroComponent) Render(w io.StringWriter) error {
	// Helper function to get attribute with default
	getAttr := func(name string) string {
		if attr := c.GetAttribute(name); attr != nil {
			return *attr
		}
		return c.GetDefaultAttribute(name)
	}

	// Get attributes
	backgroundColor := getAttr(constants.MJMLBackgroundColor)
	backgroundPosition := getAttr(constants.MJMLBackgroundPosition)
	backgroundRepeat := constants.BackgroundRepeatNoRepeat
	backgroundUrl := getAttr(constants.MJMLBackgroundUrl)
	backgroundHeight := getAttr(constants.MJMLBackgroundHeight)
	height := getAttr(constants.MJMLHeight)
	padding := getAttr(constants.MJMLPadding)
	verticalAlign := getAttr(constants.MJMLVerticalAlign)

	// Calculate effective height by subtracting padding
	effectiveHeight := height
	if height != "" && height != "0px" {
		effectiveHeight = c.calculateEffectiveHeight(height, padding)
	}

	// Calculate container width - use parent width or default 600px
	containerWidth := c.GetContainerWidth()
	if containerWidth == 0 {
		containerWidth = 600
	}
	containerWidthPx := fmt.Sprintf("%dpx", containerWidth)

	// MSO conditional comment for Outlook support
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}

	// MSO table structure
	msoTable := fmt.Sprintf(`<table border="0" cellpadding="0" cellspacing="0" role="presentation" align="center" width="%d" style="width:%s;"><tr><td style="line-height:0;font-size:0;mso-line-height-rule:exactly;">`, containerWidth, containerWidthPx)
	if _, err := w.WriteString(msoTable); err != nil {
		return err
	}

	// Add VML image for Outlook if background-url is provided
	if backgroundUrl != "" {
		vmlStyle := fmt.Sprintf("border:0;mso-position-horizontal:center;position:absolute;top:0;width:%s;z-index:-3;", containerWidthPx)
		if backgroundHeight != "" && backgroundHeight != "0px" {
			vmlStyle = fmt.Sprintf("border:0;height:%s;mso-position-horizontal:center;position:absolute;top:0;width:%s;z-index:-3;", backgroundHeight, containerWidthPx)
		}
		vmlImage := fmt.Sprintf(`<v:image src="%s" xmlns:v="urn:schemas-microsoft-com:vml" style="%s" />`, backgroundUrl, vmlStyle)
		if _, err := w.WriteString(vmlImage); err != nil {
			return err
		}
	} else {
		vmlImage := fmt.Sprintf(`<v:image xmlns:v="urn:schemas-microsoft-com:vml" style="border:0;mso-position-horizontal:center;position:absolute;top:0;width:%s;z-index:-3;" />`, containerWidthPx)
		if _, err := w.WriteString(vmlImage); err != nil {
			return err
		}
	}

	if _, err := w.WriteString("<![endif]-->"); err != nil {
		return err
	}

	// Main responsive container using HTMLTag builder
	divTag := html.NewHTMLTag("div").
		AddStyle(constants.CSSMargin, "0 auto").
		AddStyle(constants.CSSMaxWidth, containerWidthPx)

	// Add css-class if present
	if cssClass := c.BuildClassAttribute(); cssClass != "" {
		divTag.AddAttribute(constants.AttrClass, cssClass)
	}

	if err := divTag.RenderOpen(w); err != nil {
		return err
	}

	// Main table using HTMLTag builder
	mainTableTag := html.NewHTMLTag("table").
		AddAttribute(constants.AttrBorder, "0").
		AddAttribute(constants.AttrCellPadding, "0").
		AddAttribute(constants.AttrCellSpacing, "0").
		AddAttribute(constants.AttrRole, "presentation").
		AddStyle(constants.CSSWidth, "100%")

	if err := mainTableTag.RenderOpen(w); err != nil {
		return err
	}

	// tbody and tr
	tbodyTag := html.NewHTMLTag("tbody")
	if err := tbodyTag.RenderOpen(w); err != nil {
		return err
	}

	trTag := html.NewHTMLTag("tr").
		AddStyle(constants.CSSVerticalAlign, constants.VAlignTop)
	if err := trTag.RenderOpen(w); err != nil {
		return err
	}

	// Main TD with background and height using HTMLTag builder
	tdTag := html.NewHTMLTag("td").
		AddAttribute(constants.AttrHeight, strings.TrimSuffix(effectiveHeight, "px")).
		AddStyle(constants.CSSBackground, backgroundColor).
		AddStyle(constants.CSSBackgroundPosition, backgroundPosition).
		AddStyle(constants.CSSBackgroundRepeat, backgroundRepeat).
		AddStyle(constants.CSSPadding, padding).
		AddStyle(constants.CSSVerticalAlign, verticalAlign).
		AddStyle(constants.CSSHeight, effectiveHeight)

	// Add background image if provided
	if backgroundUrl != "" {
		tdTag.AddAttribute(constants.AttrBackground, backgroundUrl)
		// Add CSS shorthand background for modern email clients
		shorthandBg := fmt.Sprintf("%s url('%s') %s %s / cover", backgroundColor, backgroundUrl, backgroundRepeat, backgroundPosition)
		tdTag.AddStyle(constants.CSSBackground, shorthandBg)
	}

	// Add individual padding overrides (similar to other components)
	if paddingTopAttr := c.GetAttribute(constants.MJMLPaddingTop); paddingTopAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingTop, *paddingTopAttr)
	}
	if paddingRightAttr := c.GetAttribute(constants.MJMLPaddingRight); paddingRightAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingRight, *paddingRightAttr)
	}
	if paddingBottomAttr := c.GetAttribute(constants.MJMLPaddingBottom); paddingBottomAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingBottom, *paddingBottomAttr)
	}
	if paddingLeftAttr := c.GetAttribute(constants.MJMLPaddingLeft); paddingLeftAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingLeft, *paddingLeftAttr)
	}

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// MSO inner table
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}

	msoInnerTable := fmt.Sprintf(`<table border="0" cellpadding="0" cellspacing="0" width="%d" style="width:%s;"><tr><td><![endif]-->`, containerWidth, containerWidthPx)
	if _, err := w.WriteString(msoInnerTable); err != nil {
		return err
	}

	// Hero content container using HTMLTag builder
	heroContentTag := html.NewHTMLTag("div").
		AddAttribute(constants.AttrClass, "mj-hero-content").
		AddStyle(constants.CSSMargin, "0px auto")
	if err := heroContentTag.RenderOpen(w); err != nil {
		return err
	}

	// Content table structure using HTMLTag builder
	contentTableTag := html.NewHTMLTag("table").
		AddAttribute(constants.AttrBorder, "0").
		AddAttribute(constants.AttrCellPadding, "0").
		AddAttribute(constants.AttrCellSpacing, "0").
		AddAttribute(constants.AttrRole, "presentation").
		AddStyle(constants.CSSWidth, "100%").
		AddStyle(constants.CSSMargin, "0px")
	if err := contentTableTag.RenderOpen(w); err != nil {
		return err
	}

	// tbody, tr, td
	contentTbodyTag := html.NewHTMLTag("tbody")
	if err := contentTbodyTag.RenderOpen(w); err != nil {
		return err
	}

	contentTrTag := html.NewHTMLTag("tr")
	if err := contentTrTag.RenderOpen(w); err != nil {
		return err
	}

	contentTdTag := html.NewHTMLTag("td")
	if err := contentTdTag.RenderOpen(w); err != nil {
		return err
	}

	// Inner content table for children using HTMLTag builder
	innerTableTag := html.NewHTMLTag("table").
		AddAttribute(constants.AttrBorder, "0").
		AddAttribute(constants.AttrCellPadding, "0").
		AddAttribute(constants.AttrCellSpacing, "0").
		AddAttribute(constants.AttrRole, "presentation").
		AddStyle(constants.CSSWidth, "100%").
		AddStyle(constants.CSSMargin, "0px")
	if err := innerTableTag.RenderOpen(w); err != nil {
		return err
	}

	innerTbodyTag := html.NewHTMLTag("tbody")
	if err := innerTbodyTag.RenderOpen(w); err != nil {
		return err
	}

	// Render child components
	for _, child := range c.Children {
		// Set container width for children
		child.SetContainerWidth(containerWidth)

		// Set hero context for child rendering
		childOpts := *c.RenderOpts // Copy the options
		childOpts.InsideHero = true

		// Set render options based on component type
		switch comp := child.(type) {
		case *MJTextComponent:
			comp.RenderOpts = &childOpts
		case *MJButtonComponent:
			comp.RenderOpts = &childOpts
		case *MJImageComponent:
			comp.RenderOpts = &childOpts
			// Add more component types as needed
		}

		if err := child.Render(w); err != nil {
			return err
		}
	}

	// Close inner content table
	if _, err := w.WriteString("</tbody></table>"); err != nil {
		return err
	}

	// Close content structure
	if _, err := w.WriteString("</td></tr></tbody></table></div>"); err != nil {
		return err
	}

	// Close MSO inner table
	if _, err := w.WriteString("<!--[if mso | IE]></td></tr></table><![endif]-->"); err != nil {
		return err
	}

	// Close main TD, tr, tbody, table, and div
	if _, err := w.WriteString("</td></tr></tbody></table></div>"); err != nil {
		return err
	}

	// Close MSO main table
	if _, err := w.WriteString("<!--[if mso | IE]></td></tr></table><![endif]-->"); err != nil {
		return err
	}

	return nil
}

func (c *MJHeroComponent) GetTagName() string {
	return "mj-hero"
}

// calculateEffectiveHeight calculates the effective height by subtracting top and bottom padding
func (c *MJHeroComponent) calculateEffectiveHeight(height, padding string) string {
	if height == "" || height == "0px" {
		return height
	}

	// Parse height value
	heightPx := strings.TrimSuffix(height, "px")
	heightVal := 0
	if _, err := fmt.Sscanf(heightPx, "%d", &heightVal); err != nil {
		return height // Return original if parsing fails
	}

	// Parse padding value
	paddingPx := strings.TrimSuffix(padding, "px")
	paddingVal := 0
	if paddingPx != "" && paddingPx != "0" {
		if _, err := fmt.Sscanf(paddingPx, "%d", &paddingVal); err != nil {
			paddingVal = 0 // Default to 0 if parsing fails
		}
	}

	// Get individual padding overrides
	topPadding := paddingVal
	bottomPadding := paddingVal

	if paddingTopAttr := c.GetAttribute(constants.MJMLPaddingTop); paddingTopAttr != nil {
		topPaddingPx := strings.TrimSuffix(*paddingTopAttr, "px")
		if _, err := fmt.Sscanf(topPaddingPx, "%d", &topPadding); err != nil {
			topPadding = paddingVal // Fallback to general padding if parsing fails
		}
	}

	if paddingBottomAttr := c.GetAttribute(constants.MJMLPaddingBottom); paddingBottomAttr != nil {
		bottomPaddingPx := strings.TrimSuffix(*paddingBottomAttr, "px")
		if _, err := fmt.Sscanf(bottomPaddingPx, "%d", &bottomPadding); err != nil {
			bottomPadding = paddingVal // Fallback to general padding if parsing fails
		}
	}

	// Calculate effective height = original height - top padding - bottom padding
	effectiveHeightVal := heightVal - topPadding - bottomPadding
	if effectiveHeightVal < 0 {
		effectiveHeightVal = 0
	}

	return fmt.Sprintf("%dpx", effectiveHeightVal)
}

func (c *MJHeroComponent) GetDefaultAttribute(name string) string {
	switch name {
	case constants.MJMLBackgroundColor:
		return "#ffffff"
	case constants.MJMLBackgroundPosition:
		return "center center"
	case constants.MJMLHeight:
		return "0px"
	case constants.MJMLMode:
		return "fixed-height"
	case constants.MJMLPadding:
		return "0px"
	case constants.MJMLVerticalAlign:
		return constants.VAlignTop
	default:
		return ""
	}
}
