package components

import (
	"io"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/fonts"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJAccordionComponent represents the mj-accordion component
type MJAccordionComponent struct {
	*BaseComponent
}

func NewMJAccordionComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJAccordionComponent {
	return &MJAccordionComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJAccordionComponent) Render(w io.StringWriter) error {
	border := c.GetAttributeWithDefault(c, constants.MJMLBorder)
	fontFamily := c.GetAttributeWithDefault(c, constants.MJMLFontFamily)
	padding := c.GetAttributeWithDefault(c, constants.MJMLPadding)
	paddingRight := c.GetAttributeWithDefault(c, constants.MJMLPaddingRight)
	containerBackgroundColor := c.GetAttributeWithDefault(c, constants.MJMLContainerBackgroundColor)
	cssClass := c.GetAttributeWithDefault(c, constants.MJMLCSSClass)

	// Track font family if specified
	if fontFamily != "" {
		c.TrackFontFamily(fontFamily)
	}

	// Create TR element (required wrapper for content components)
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}

	// Create TD with alignment and base styles using HTMLTag
	tdTag := html.NewHTMLTag("td")

	// Add CSS class if specified
	if cssClass != "" {
		tdTag.AddAttribute(constants.AttrClass, cssClass)
	}

	// Add container background color first if specified (MRML order)
	if containerBackgroundColor != "" {
		tdTag.AddStyle(constants.CSSBackground, containerBackgroundColor)
	}

	// Add base styles in MRML order
	tdTag.AddStyle(constants.CSSFontSize, "0px").
		AddStyle(constants.CSSPadding, padding)

	// Add individual padding if specified
	if paddingRight != "" {
		tdTag.AddStyle(constants.CSSPaddingRight, paddingRight)
	}

	// Add word-break last
	tdTag.AddStyle("word-break", "break-word")

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// Start the accordion table wrapper
	tableTag := html.NewHTMLTag("table").
		AddAttribute(constants.AttrCellSpacing, "0").
		AddAttribute(constants.AttrCellPadding, "0").
		AddAttribute(constants.AttrClass, "mj-accordion").
		AddStyle(constants.CSSWidth, "100%").
		AddStyle(constants.CSSBorderCollapse, constants.BorderCollapseCollapse).
		AddStyle(constants.CSSBorder, border).
		AddStyle(constants.CSSBorderBottom, "none").
		AddStyle(constants.CSSFontFamily, fontFamily)

	if err := tableTag.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody>"); err != nil {
		return err
	}

	// Render accordion elements
	for _, child := range c.Children {
		if accordionElement, ok := child.(*MJAccordionElementComponent); ok {
			accordionElement.SetContainerWidth(c.GetContainerWidth())
			accordionElement.inheritFromParent(c)
			if err := accordionElement.Render(w); err != nil {
				return err
			}
		}
	}

	// Close the accordion table and wrapper elements
	if _, err := w.WriteString("</tbody></table></td></tr>"); err != nil {
		return err
	}

	return nil
}

func (c *MJAccordionComponent) GetTagName() string {
	return "mj-accordion"
}

func (c *MJAccordionComponent) GetDefaultAttribute(name string) string {
	switch name {
	case constants.MJMLBorder:
		return "2px solid black"
	case constants.MJMLFontFamily:
		return fonts.DefaultFontStack
	case "icon-align":
		return constants.VAlignMiddle
	case "icon-height":
		return "32px"
	case "icon-position":
		return constants.AlignRight
	case "icon-unwrapped-alt":
		return "-"
	case "icon-unwrapped-url":
		return "https://i.imgur.com/w4uTygT.png"
	case "icon-width":
		return "32px"
	case "icon-wrapped-alt":
		return "+"
	case "icon-wrapped-url":
		return "https://i.imgur.com/bIXv1bk.png"
	case constants.MJMLPadding:
		return "10px 25px"
	default:
		return ""
	}
}

// MJAccordionTextComponent represents the mj-accordion-text component
type MJAccordionTextComponent struct {
	*BaseComponent
}

func NewMJAccordionTextComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJAccordionTextComponent {
	return &MJAccordionTextComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJAccordionTextComponent) Render(w io.StringWriter) error {
	// Render the raw content inside the accordion text
	content := strings.TrimSpace(c.Node.Text)
	if content != "" {
		if _, err := w.WriteString(content); err != nil {
			return err
		}
	}

	// Also render any child HTML elements
	for _, child := range c.Node.Children {
		if err := c.renderHTMLChild(w, child); err != nil {
			return err
		}
	}

	return nil
}

func (c *MJAccordionTextComponent) renderHTMLChild(w io.StringWriter, node *parser.MJMLNode) error {
	// Simple HTML rendering for child elements like <span>
	tagName := node.XMLName.Local
	if tagName != "" {
		if _, err := w.WriteString("<" + tagName + ">"); err != nil {
			return err
		}

		// Render text content
		if content := strings.TrimSpace(node.Text); content != "" {
			if _, err := w.WriteString(content); err != nil {
				return err
			}
		}

		// Render children recursively
		for _, child := range node.Children {
			if err := c.renderHTMLChild(w, child); err != nil {
				return err
			}
		}

		if _, err := w.WriteString("</" + tagName + ">"); err != nil {
			return err
		}
	} else {
		// Text node
		if content := strings.TrimSpace(node.Text); content != "" {
			if _, err := w.WriteString(content); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *MJAccordionTextComponent) GetTagName() string {
	return "mj-accordion-text"
}

func (c *MJAccordionTextComponent) GetDefaultAttribute(name string) string {
	switch name {
	case constants.MJMLFontSize:
		return "13px"
	case constants.MJMLLineHeight:
		return "1"
	case constants.MJMLPadding:
		return "16px"
	case constants.MJMLFontFamily:
		return fonts.DefaultFontStack
	default:
		return ""
	}
}

// MJAccordionTitleComponent represents the mj-accordion-title component
type MJAccordionTitleComponent struct {
	*BaseComponent
}

func NewMJAccordionTitleComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJAccordionTitleComponent {
	return &MJAccordionTitleComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJAccordionTitleComponent) Render(w io.StringWriter) error {
	// Render the raw content inside the accordion title
	content := strings.TrimSpace(c.Node.Text)
	if content != "" {
		if _, err := w.WriteString(content); err != nil {
			return err
		}
	}
	return nil
}

func (c *MJAccordionTitleComponent) GetTagName() string {
	return "mj-accordion-title"
}

func (c *MJAccordionTitleComponent) GetDefaultAttribute(name string) string {
	switch name {
	case constants.MJMLFontSize:
		return "13px"
	case constants.MJMLPadding:
		return "16px"
	case constants.MJMLFontFamily:
		return fonts.DefaultFontStack
	default:
		return ""
	}
}

// MJAccordionElementComponent represents the mj-accordion-element component
type MJAccordionElementComponent struct {
	*BaseComponent
	parentAccordion *MJAccordionComponent // Reference to parent for attribute inheritance
}

func NewMJAccordionElementComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJAccordionElementComponent {
	return &MJAccordionElementComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJAccordionElementComponent) Render(w io.StringWriter) error {
	fontSize := c.getAttribute("font-size")
	fontFamily := c.getAttribute("font-family")
	backgroundColor := c.getAttribute("background-color")
	cssClass := c.getAttribute("css-class")
	iconAlign := c.getAttribute("icon-align")
	iconHeight := c.getAttribute("icon-height")
	iconWidth := c.getAttribute("icon-width")
	iconWrappedUrl := c.getAttribute("icon-wrapped-url")
	iconUnwrappedUrl := c.getAttribute("icon-unwrapped-url")
	iconWrappedAlt := c.getAttribute("icon-wrapped-alt")
	iconUnwrappedAlt := c.getAttribute("icon-unwrapped-alt")

	// Start accordion element row
	if _, err := w.WriteString("<tr"); err != nil {
		return err
	}

	// Add CSS class to row if specified
	if cssClass != "" {
		if _, err := w.WriteString(" class=\"" + cssClass + "\""); err != nil {
			return err
		}
	}

	if _, err := w.WriteString(">"); err != nil {
		return err
	}

	// Create TD with background color
	tdTag := html.NewHTMLTag("td").
		AddStyle(constants.CSSPadding, "0px")

	if backgroundColor != "" {
		tdTag.AddStyle(constants.CSSBackgroundColor, backgroundColor)
	}

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// Create label element
	labelTag := html.NewHTMLTag("label").
		AddAttribute(constants.AttrClass, "mj-accordion-element").
		AddStyle(constants.CSSFontSize, fontSize)

	if fontFamily != "" {
		labelTag.AddStyle(constants.CSSFontFamily, fontFamily)
	}

	if err := labelTag.RenderOpen(w); err != nil {
		return err
	}

	// Add checkbox input (hidden for accessibility)
	if _, err := w.WriteString("<!--[if !mso | IE]><!--><input type=\"checkbox\" class=\"mj-accordion-checkbox\" style=\"display:none;\" /><!--<![endif]-->"); err != nil {
		return err
	}

	// Start accordion wrapper div
	if _, err := w.WriteString("<div>"); err != nil {
		return err
	}

	// Find title and content components
	var titleComponent *MJAccordionTitleComponent
	var textComponent *MJAccordionTextComponent

	for _, child := range c.Children {
		if title, ok := child.(*MJAccordionTitleComponent); ok {
			titleComponent = title
		} else if text, ok := child.(*MJAccordionTextComponent); ok {
			textComponent = text
		}
	}

	// Render title section
	if titleComponent != nil {
		if err := c.renderTitle(w, titleComponent, iconAlign, iconHeight, iconWidth, iconWrappedUrl, iconUnwrappedUrl, iconWrappedAlt, iconUnwrappedAlt); err != nil {
			return err
		}
	}

	// Render content section
	if textComponent != nil {
		if err := c.renderContent(w, textComponent); err != nil {
			return err
		}
	}

	// Close accordion wrapper div, label, cell, and row
	if _, err := w.WriteString("</div></label></td></tr>"); err != nil {
		return err
	}

	return nil
}

func (c *MJAccordionElementComponent) renderTitle(w io.StringWriter, titleComponent *MJAccordionTitleComponent, iconAlign, iconHeight, iconWidth, iconWrappedUrl, iconUnwrappedUrl, iconWrappedAlt, iconUnwrappedAlt string) error {
	border := c.parentAccordion.GetAttributeWithDefault(c.parentAccordion, constants.MJMLBorder)
	fontSize := titleComponent.GetAttributeWithDefault(titleComponent, constants.MJMLFontSize)
	// Only get font-family if explicitly set on title element
	fontFamily := ""
	if value := titleComponent.Node.GetAttribute(constants.MJMLFontFamily); value != "" {
		fontFamily = value
		titleComponent.TrackFontFamily(value)
	}
	padding := titleComponent.GetAttributeWithDefault(titleComponent, constants.MJMLPadding)
	paddingTop := titleComponent.GetAttributeWithDefault(titleComponent, constants.MJMLPaddingTop)
	paddingBottom := titleComponent.GetAttributeWithDefault(titleComponent, constants.MJMLPaddingBottom)
	paddingLeft := titleComponent.GetAttributeWithDefault(titleComponent, constants.MJMLPaddingLeft)
	paddingRight := titleComponent.GetAttributeWithDefault(titleComponent, constants.MJMLPaddingRight)

	// Get title-specific attributes
	backgroundColor := titleComponent.Node.GetAttribute(constants.MJMLBackgroundColor)
	color := titleComponent.Node.GetAttribute(constants.MJMLColor)
	cssClass := titleComponent.Node.GetAttribute(constants.MJMLCSSClass)

	// Get icon position to determine order
	iconPosition := c.getAttribute("icon-position")

	// Start title section
	divTag := html.NewHTMLTag("div").AddAttribute(constants.AttrClass, "mj-accordion-title")
	if err := divTag.RenderOpen(w); err != nil {
		return err
	}

	tableTag := html.NewHTMLTag("table").
		AddAttribute(constants.AttrCellSpacing, "0").
		AddAttribute(constants.AttrCellPadding, "0").
		AddStyle(constants.CSSWidth, "100%").
		AddStyle(constants.CSSBorderBottom, border)

	if err := tableTag.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody><tr>"); err != nil {
		return err
	}

	// Render icon on left if icon-position="left"
	if iconPosition == constants.AlignLeft {
		if err := c.renderIconCell(w, iconAlign, iconHeight, iconWidth, iconWrappedUrl, iconUnwrappedUrl, iconWrappedAlt, iconUnwrappedAlt, titleComponent, backgroundColor); err != nil {
			return err
		}
	}

	// Create title cell
	tdTag := html.NewHTMLTag("td").
		AddStyle(constants.CSSWidth, "100%")

	// Add CSS class if specified
	if cssClass != "" {
		tdTag.AddAttribute(constants.AttrClass, cssClass)
	}

	// Add background color first if specified (MRML order)
	if backgroundColor != "" {
		tdTag.AddStyle(constants.CSSBackgroundColor, backgroundColor)
	}

	// Add color if specified
	if color != "" {
		tdTag.AddStyle(constants.CSSColor, color)
	}

	// Add font size
	tdTag.AddStyle(constants.CSSFontSize, fontSize)

	// Add font-family if specified
	if fontFamily != "" {
		tdTag.AddStyle(constants.CSSFontFamily, fontFamily)
	}

	// Handle individual padding properties if specified, otherwise use general padding
	if paddingTop != "" {
		tdTag.AddStyle(constants.CSSPaddingTop, paddingTop)
	}
	if paddingBottom != "" {
		tdTag.AddStyle(constants.CSSPaddingBottom, paddingBottom)
	}
	if paddingLeft != "" {
		tdTag.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}
	if paddingRight != "" {
		tdTag.AddStyle(constants.CSSPaddingRight, paddingRight)
	}
	// Always add base padding
	tdTag.AddStyle(constants.CSSPadding, padding)

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// Render title content
	if err := titleComponent.Render(w); err != nil {
		return err
	}

	// Close title cell
	if _, err := w.WriteString("</td>"); err != nil {
		return err
	}

	// Render icon on right if icon-position="right" (default)
	if iconPosition != constants.AlignLeft {
		if err := c.renderIconCell(w, iconAlign, iconHeight, iconWidth, iconWrappedUrl, iconUnwrappedUrl, iconWrappedAlt, iconUnwrappedAlt, titleComponent, backgroundColor); err != nil {
			return err
		}
	}

	// Close table
	if _, err := w.WriteString("</tr></tbody></table></div>"); err != nil {
		return err
	}

	return nil
}

func (c *MJAccordionElementComponent) renderIconCell(w io.StringWriter, iconAlign, iconHeight, iconWidth, iconWrappedUrl, iconUnwrappedUrl, iconWrappedAlt, iconUnwrappedAlt string, titleComponent *MJAccordionTitleComponent, titleBackgroundColor string) error {
	// Add icon cell using MSO conditional comments
	if _, err := w.WriteString("<!--[if !mso | IE]><!-->"); err != nil {
		return err
	}

	// Use titleComponent's default padding (16px) for icon, not accordion padding
	iconPadding := titleComponent.GetDefaultAttribute(constants.MJMLPadding)
	iconCellTag := html.NewHTMLTag("td").
		AddAttribute(constants.AttrClass, "mj-accordion-ico").
		AddStyle(constants.CSSPadding, iconPadding)

	// Add background color if specified on title (MRML order - before vertical-align)
	if titleBackgroundColor != "" {
		iconCellTag.AddStyle(constants.CSSBackground, titleBackgroundColor)
	}

	// Add vertical align last
	iconCellTag.AddStyle(constants.CSSVerticalAlign, iconAlign)

	if err := iconCellTag.RenderOpen(w); err != nil {
		return err
	}

	// Wrapped state icon (plus)
	wrappedImg := html.NewHTMLTag("img").
		AddAttribute(constants.AttrSrc, iconWrappedUrl).
		AddAttribute(constants.AttrAlt, iconWrappedAlt).
		AddAttribute(constants.AttrClass, "mj-accordion-more").
		AddStyle(constants.CSSDisplay, constants.DisplayNone).
		AddStyle(constants.CSSWidth, iconWidth).
		AddStyle(constants.CSSHeight, iconHeight)

	if err := wrappedImg.RenderSelfClosing(w); err != nil {
		return err
	}

	// Unwrapped state icon (minus)
	unwrappedImg := html.NewHTMLTag("img").
		AddAttribute(constants.AttrSrc, iconUnwrappedUrl).
		AddAttribute(constants.AttrAlt, iconUnwrappedAlt).
		AddAttribute(constants.AttrClass, "mj-accordion-less").
		AddStyle(constants.CSSDisplay, constants.DisplayNone).
		AddStyle(constants.CSSWidth, iconWidth).
		AddStyle(constants.CSSHeight, iconHeight)

	if err := unwrappedImg.RenderSelfClosing(w); err != nil {
		return err
	}

	// Close icon cell
	if _, err := w.WriteString("</td><!--<![endif]-->"); err != nil {
		return err
	}

	return nil
}

func (c *MJAccordionElementComponent) renderContent(w io.StringWriter, textComponent *MJAccordionTextComponent) error {
	border := c.parentAccordion.GetAttributeWithDefault(c.parentAccordion, constants.MJMLBorder)
	fontSize := textComponent.GetAttributeWithDefault(textComponent, constants.MJMLFontSize)
	// Only get font-family if explicitly set on text element
	fontFamily := ""
	if value := textComponent.Node.GetAttribute(constants.MJMLFontFamily); value != "" {
		fontFamily = value
		textComponent.TrackFontFamily(value)
	}
	lineHeight := textComponent.GetAttributeWithDefault(textComponent, constants.MJMLLineHeight)
	padding := textComponent.GetAttributeWithDefault(textComponent, constants.MJMLPadding)
	paddingTop := textComponent.GetAttributeWithDefault(textComponent, constants.MJMLPaddingTop)
	paddingBottom := textComponent.GetAttributeWithDefault(textComponent, constants.MJMLPaddingBottom)
	paddingLeft := textComponent.GetAttributeWithDefault(textComponent, constants.MJMLPaddingLeft)
	paddingRight := textComponent.GetAttributeWithDefault(textComponent, constants.MJMLPaddingRight)

	// Get text-specific attributes
	backgroundColor := textComponent.Node.GetAttribute(constants.MJMLBackgroundColor)
	color := textComponent.Node.GetAttribute(constants.MJMLColor)
	cssClass := textComponent.Node.GetAttribute(constants.MJMLCSSClass)

	// Start content section
	divTag := html.NewHTMLTag("div").AddAttribute(constants.AttrClass, "mj-accordion-content")
	if err := divTag.RenderOpen(w); err != nil {
		return err
	}

	tableTag := html.NewHTMLTag("table").
		AddAttribute(constants.AttrCellSpacing, "0").
		AddAttribute(constants.AttrCellPadding, "0").
		AddStyle(constants.CSSWidth, "100%").
		AddStyle(constants.CSSBorderBottom, border)

	if err := tableTag.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody><tr>"); err != nil {
		return err
	}

	// Create content cell
	tdTag := html.NewHTMLTag("td")

	// Add CSS class if specified
	if cssClass != "" {
		tdTag.AddAttribute(constants.AttrClass, cssClass)
	}

	// Add background color first if specified (MRML order)
	if backgroundColor != "" {
		tdTag.AddStyle(constants.CSSBackground, backgroundColor)
	}

	// Add font size
	tdTag.AddStyle(constants.CSSFontSize, fontSize)

	// Add line height after font size
	tdTag.AddStyle(constants.CSSLineHeight, lineHeight)

	// Add color if specified
	if color != "" {
		tdTag.AddStyle(constants.CSSColor, color)
	}

	// Add font-family if specified
	if fontFamily != "" {
		tdTag.AddStyle(constants.CSSFontFamily, fontFamily)
	}

	// Handle individual padding properties if specified, otherwise use general padding
	if paddingTop != "" {
		tdTag.AddStyle(constants.CSSPaddingTop, paddingTop)
	}
	if paddingBottom != "" {
		tdTag.AddStyle(constants.CSSPaddingBottom, paddingBottom)
	}
	if paddingLeft != "" {
		tdTag.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}
	if paddingRight != "" {
		tdTag.AddStyle(constants.CSSPaddingRight, paddingRight)
	}
	// Always add base padding
	tdTag.AddStyle(constants.CSSPadding, padding)

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// Render text content
	if err := textComponent.Render(w); err != nil {
		return err
	}

	// Close content section
	if _, err := w.WriteString("</td></tr></tbody></table></div>"); err != nil {
		return err
	}

	return nil
}

func (c *MJAccordionElementComponent) GetTagName() string {
	return "mj-accordion-element"
}

func (c *MJAccordionElementComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "icon-align":
		return constants.VAlignMiddle
	case "icon-height":
		return "32px"
	case "icon-position":
		return constants.AlignRight
	case "icon-unwrapped-alt":
		return "-"
	case "icon-unwrapped-url":
		return "https://i.imgur.com/w4uTygT.png"
	case "icon-width":
		return "32px"
	case "icon-wrapped-alt":
		return "+"
	case "icon-wrapped-url":
		return "https://i.imgur.com/bIXv1bk.png"
	case constants.MJMLFontSize:
		return "13px"
	case constants.MJMLPadding:
		return "16px"
	default:
		return ""
	}
}

func (c *MJAccordionElementComponent) getAttribute(name string) string {
	// 1. Check explicit element attribute first
	if value := c.Node.GetAttribute(name); value != "" {
		return value
	}

	// 2. Check parent accordion attributes (but not for font-family or css-class)
	if c.parentAccordion != nil && name != constants.MJMLFontFamily && name != constants.MJMLCSSClass {
		if value := c.parentAccordion.Node.GetAttribute(name); value != "" {
			return value
		}
	}

	// 3. Fall back to component defaults
	return c.GetDefaultAttribute(name)
}

// inheritFromParent sets the parent reference for attribute inheritance
func (c *MJAccordionElementComponent) inheritFromParent(parent *MJAccordionComponent) {
	c.parentAccordion = parent
}
