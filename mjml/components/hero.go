package components

import (
	"fmt"
	"io"
	"strings"

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
	backgroundColor := getAttr("background-color")
	backgroundPosition := getAttr("background-position")
	backgroundRepeat := "no-repeat"
	backgroundUrl := getAttr("background-url")
	backgroundHeight := getAttr("background-height")
	height := getAttr("height")
	padding := getAttr("padding")
	verticalAlign := getAttr("vertical-align")

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

	// Main responsive container
	if _, err := w.WriteString(`<div`); err != nil {
		return err
	}

	// Add css-class if present
	if cssClass := c.BuildClassAttribute(); cssClass != "" {
		if _, err := w.WriteString(` class="`); err != nil {
			return err
		}
		if _, err := w.WriteString(cssClass); err != nil {
			return err
		}
		if _, err := w.WriteString(`"`); err != nil {
			return err
		}
	}

	if _, err := w.WriteString(` style="margin:0 auto;max-width:`); err != nil {
		return err
	}
	if _, err := w.WriteString(containerWidthPx); err != nil {
		return err
	}
	if _, err := w.WriteString(`;">`); err != nil {
		return err
	}

	// Main table
	if _, err := w.WriteString(`<table border="0" cellpadding="0" cellspacing="0" role="presentation" style="width:100%;"><tbody><tr style="vertical-align:top;">`); err != nil {
		return err
	}

	// Main TD with background and height using HTMLTag builder
	tdTag := html.NewHTMLTag("td").
		AddAttribute("height", strings.TrimSuffix(effectiveHeight, "px")).
		AddStyle("background", backgroundColor).
		AddStyle("background-position", backgroundPosition).
		AddStyle("background-repeat", backgroundRepeat).
		AddStyle("padding", padding).
		AddStyle("vertical-align", verticalAlign).
		AddStyle("height", effectiveHeight)

	// Add background image if provided
	if backgroundUrl != "" {
		tdTag.AddAttribute("background", backgroundUrl)
		// Add CSS shorthand background for modern email clients
		shorthandBg := fmt.Sprintf("%s url('%s') %s %s / cover", backgroundColor, backgroundUrl, backgroundRepeat, backgroundPosition)
		tdTag.AddStyle("background", shorthandBg)
	}

	// Add individual padding overrides (similar to other components)
	if paddingTopAttr := c.GetAttribute("padding-top"); paddingTopAttr != nil {
		tdTag.AddStyle("padding-top", *paddingTopAttr)
	}
	if paddingRightAttr := c.GetAttribute("padding-right"); paddingRightAttr != nil {
		tdTag.AddStyle("padding-right", *paddingRightAttr)
	}
	if paddingBottomAttr := c.GetAttribute("padding-bottom"); paddingBottomAttr != nil {
		tdTag.AddStyle("padding-bottom", *paddingBottomAttr)
	}
	if paddingLeftAttr := c.GetAttribute("padding-left"); paddingLeftAttr != nil {
		tdTag.AddStyle("padding-left", *paddingLeftAttr)
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

	// Hero content container
	if _, err := w.WriteString(`<div class="mj-hero-content" style="margin:0px auto;">`); err != nil {
		return err
	}

	// Content table structure
	if _, err := w.WriteString(`<table border="0" cellpadding="0" cellspacing="0" role="presentation" style="width:100%;margin:0px;"><tbody><tr><td>`); err != nil {
		return err
	}

	// Inner content table for children
	if _, err := w.WriteString(`<table border="0" cellpadding="0" cellspacing="0" role="presentation" style="width:100%;margin:0px;"><tbody>`); err != nil {
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
	if _, err := w.WriteString(`</tbody></table>`); err != nil {
		return err
	}

	// Close content structure
	if _, err := w.WriteString(`</td></tr></tbody></table></div>`); err != nil {
		return err
	}

	// Close MSO inner table
	if _, err := w.WriteString("<!--[if mso | IE]></td></tr></table><![endif]-->"); err != nil {
		return err
	}

	// Close main TD and table
	if _, err := w.WriteString(`</td></tr></tbody></table></div>`); err != nil {
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
		fmt.Sscanf(paddingPx, "%d", &paddingVal)
	}

	// Get individual padding overrides
	topPadding := paddingVal
	bottomPadding := paddingVal

	if paddingTopAttr := c.GetAttribute("padding-top"); paddingTopAttr != nil {
		topPaddingPx := strings.TrimSuffix(*paddingTopAttr, "px")
		fmt.Sscanf(topPaddingPx, "%d", &topPadding)
	}

	if paddingBottomAttr := c.GetAttribute("padding-bottom"); paddingBottomAttr != nil {
		bottomPaddingPx := strings.TrimSuffix(*paddingBottomAttr, "px")
		fmt.Sscanf(bottomPaddingPx, "%d", &bottomPadding)
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
	case "background-color":
		return "#ffffff"
	case "background-position":
		return "center center"
	case "height":
		return "0px"
	case "mode":
		return "fixed-height"
	case "padding":
		return "0px"
	case "vertical-align":
		return "top"
	default:
		return ""
	}
}
