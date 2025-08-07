package components

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/fonts"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// Global counter for unique navbar checkbox IDs
var navbarIDCounter int64

// ResetNavbarIDCounter resets the global counter for deterministic testing
func ResetNavbarIDCounter() {
	atomic.StoreInt64(&navbarIDCounter, 0)
}

// MJNavbarComponent represents the mj-navbar component
type MJNavbarComponent struct {
	*BaseComponent
	Children []Component
}

func NewMJNavbarComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJNavbarComponent {
	return &MJNavbarComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJNavbarComponent) Render(w io.StringWriter) error {
	align := c.getAttribute(constants.MJMLAlign)
	baseURL := c.getAttribute("base-url")
	hamburger := c.getAttribute("hamburger")

	// Generate unique ID for checkbox
	checkboxID := c.generateCheckboxID()

	// Start table cell wrapper
	if err := c.renderCellOpen(w, align); err != nil {
		return err
	}

	// Render hamburger checkbox and trigger (mobile only)
	if hamburger != "" {
		if err := c.renderHamburgerToggle(w, checkboxID); err != nil {
			return err
		}
	}

	// Render inline links container
	if err := c.renderInlineLinks(w, baseURL); err != nil {
		return err
	}

	// Close table cell and row wrapper
	if _, err := w.WriteString("</td></tr>"); err != nil {
		return err
	}

	return nil
}

func (c *MJNavbarComponent) renderCellOpen(w io.StringWriter, align string) error {
	// Start table row first
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}

	// Create table cell with alignment and CSS class
	cellTag := html.NewHTMLTag("td").
		AddAttribute(constants.AttrAlign, align).
		AddStyle(constants.CSSFontSize, "0px").
		AddStyle("word-break", "break-word")

	// Only add CSS class if it's not empty
	if cssClass := c.getAttribute(constants.MJMLCSSClass); cssClass != "" {
		cellTag.AddAttribute(constants.AttrClass, cssClass)
	}

	return cellTag.RenderOpen(w)
}

func (c *MJNavbarComponent) renderHamburgerToggle(w io.StringWriter, checkboxID string) error {
	// Render checkbox input (hidden)
	if _, err := w.WriteString("<!--[if !mso]><!-->"); err != nil {
		return err
	}

	inputTag := html.NewHTMLTag("input").
		AddAttribute("id", checkboxID).
		AddAttribute(constants.AttrType, "checkbox").
		AddAttribute(constants.AttrClass, "mj-menu-checkbox").
		AddStyle(constants.CSSDisplay, "none !important").
		AddStyle("max-height", "0").
		AddStyle("visibility", "hidden")

	if err := inputTag.RenderSelfClosing(w); err != nil {
		return err
	}

	if _, err := w.WriteString("<!--<![endif]-->"); err != nil {
		return err
	}

	// Render hamburger trigger/label
	triggerDiv := html.NewHTMLTag("div").
		AddAttribute(constants.AttrClass, "mj-menu-trigger").
		AddStyle(constants.CSSDisplay, constants.DisplayNone).
		AddStyle("max-height", "0px").
		AddStyle("max-width", "0px").
		AddStyle(constants.CSSFontSize, "0px").
		AddStyle("overflow", "hidden")

	if err := triggerDiv.RenderOpen(w); err != nil {
		return err
	}

	// Render label with hamburger icons
	if err := c.renderHamburgerLabel(w, checkboxID); err != nil {
		return err
	}

	if _, err := w.WriteString("</div>"); err != nil {
		return err
	}

	return nil
}

func (c *MJNavbarComponent) renderHamburgerLabel(w io.StringWriter, checkboxID string) error {
	icoAlign := c.getAttribute("ico-align")
	icoColor := c.getAttribute("ico-color")
	icoFontSize := c.getAttribute("ico-font-size")
	icoFontFamily := c.getAttribute("ico-font-family")
	icoTextTransform := c.getAttribute("ico-text-transform")
	icoTextDecoration := c.getAttribute("ico-text-decoration")
	icoLineHeight := c.getAttribute("ico-line-height")
	icoPadding := c.getAttribute("ico-padding")
	icoOpen := c.getAttribute("ico-open")
	icoClose := c.getAttribute("ico-close")

	// Handle individual padding properties
	icoPaddingTop := c.getAttribute("ico-padding-top")
	icoPaddingRight := c.getAttribute("ico-padding-right")
	icoPaddingBottom := c.getAttribute("ico-padding-bottom")
	icoPaddingLeft := c.getAttribute("ico-padding-left")

	labelTag := html.NewHTMLTag("label").
		AddAttribute(constants.AttrAlign, icoAlign).
		AddAttribute("for", checkboxID).
		AddAttribute(constants.AttrClass, "mj-menu-label").
		AddStyle(constants.CSSDisplay, constants.DisplayBlock).
		AddStyle("cursor", "pointer").
		AddStyle("mso-hide", "all").
		AddStyle("-moz-user-select", "none").
		AddStyle("user-select", "none").
		AddStyle(constants.CSSColor, icoColor).
		AddStyle(constants.CSSFontSize, icoFontSize).
		AddStyle(constants.CSSFontFamily, icoFontFamily).
		AddStyle(constants.CSSTextTransform, icoTextTransform).
		AddStyle(constants.CSSTextDecoration, icoTextDecoration).
		AddStyle(constants.CSSLineHeight, icoLineHeight).
		AddStyle(constants.CSSPadding, icoPadding)

	// Only add individual padding properties if they're not empty
	if icoPaddingTop != "" {
		labelTag.AddStyle(constants.CSSPaddingTop, icoPaddingTop)
	}
	if icoPaddingRight != "" {
		labelTag.AddStyle(constants.CSSPaddingRight, icoPaddingRight)
	}
	if icoPaddingBottom != "" {
		labelTag.AddStyle(constants.CSSPaddingBottom, icoPaddingBottom)
	}
	if icoPaddingLeft != "" {
		labelTag.AddStyle(constants.CSSPaddingLeft, icoPaddingLeft)
	}

	if err := labelTag.RenderOpen(w); err != nil {
		return err
	}

	// Render open icon (hamburger)
	openSpan := html.NewHTMLTag("span").
		AddAttribute(constants.AttrClass, "mj-menu-icon-open").
		AddStyle("mso-hide", "all")

	if err := openSpan.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString(icoOpen); err != nil {
		return err
	}
	if _, err := w.WriteString("</span>"); err != nil {
		return err
	}

	// Render close icon (X)
	closeSpan := html.NewHTMLTag("span").
		AddAttribute(constants.AttrClass, "mj-menu-icon-close").
		AddStyle(constants.CSSDisplay, constants.DisplayNone).
		AddStyle("mso-hide", "all")

	if err := closeSpan.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString(icoClose); err != nil {
		return err
	}
	if _, err := w.WriteString("</span>"); err != nil {
		return err
	}

	if _, err := w.WriteString("</label>"); err != nil {
		return err
	}

	return nil
}

func (c *MJNavbarComponent) renderInlineLinks(w io.StringWriter, baseURL string) error {
	// Start inline links container
	linksDiv := html.NewHTMLTag("div").
		AddAttribute(constants.AttrClass, "mj-inline-links")

	if err := linksDiv.RenderOpen(w); err != nil {
		return err
	}

	// MSO table for Outlook compatibility with correct alignment
	align := c.getAttribute(constants.MJMLAlign)
	if _, err := w.WriteString("<!--[if mso | IE]><table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\" align=\"" + align + "\"><tr><![endif]-->"); err != nil {
		return err
	}

	// Render navbar links
	for _, child := range c.Children {
		if navbarLink, ok := child.(*MJNavbarLinkComponent); ok {
			// MSO table cell with CSS class and padding
			originalClass := navbarLink.getAttribute(constants.MJMLCSSClass)
			if _, err := w.WriteString("<!--[if mso | IE]><td"); err != nil {
				return err
			}
			if originalClass != "" {
				msoClass := originalClass + "-outlook"
				if _, err := w.WriteString(" class=\"" + msoClass + "\""); err != nil {
					return err
				}
			}
			if _, err := w.WriteString(" style=\"padding:" + navbarLink.getAttribute(constants.MJMLPadding) + ";\"><![endif]-->"); err != nil {
				return err
			}

			if err := navbarLink.RenderWithBaseURL(w, baseURL); err != nil {
				return err
			}

			if _, err := w.WriteString("<!--[if mso | IE]></td><![endif]-->"); err != nil {
				return err
			}
		}
	}

	if _, err := w.WriteString("<!--[if mso | IE]></tr></table><![endif]-->"); err != nil {
		return err
	}

	if _, err := w.WriteString("</div>"); err != nil {
		return err
	}

	return nil
}

func (c *MJNavbarComponent) generateCheckboxID() string {
	// Generate unique ID using atomic counter: 00000000, 00000001, 00000002, etc.
	// This maintains compatibility with existing tests while fixing multi-navbar conflicts
	counter := atomic.LoadInt64(&navbarIDCounter)
	id := fmt.Sprintf("%08d", counter) // Start from 00000000
	atomic.AddInt64(&navbarIDCounter, 1)
	return id
}

func (c *MJNavbarComponent) getAttribute(name string) string {
	value := c.GetAttributeWithDefault(c, name)
	if name == "ico-font-family" && value != "" {
		c.TrackFontFamily(value)
	}
	return value
}

func (c *MJNavbarComponent) GetTagName() string {
	return "mj-navbar"
}

func (c *MJNavbarComponent) GetDefaultAttribute(name string) string {
	switch name {
	case constants.MJMLAlign:
		return constants.AlignCenter
	case "ico-align":
		return constants.AlignCenter
	case "ico-close":
		return "&#8855;"
	case "ico-color":
		return "#000000"
	case "ico-font-family":
		return fonts.DefaultFontStack
	case "ico-font-size":
		return "30px"
	case "ico-line-height":
		return "30px"
	case "ico-open":
		return "&#9776;"
	case "ico-padding":
		return "10px"
	case "ico-text-decoration":
		return constants.TextDecorationNone
	case "ico-text-transform":
		return "uppercase"
	default:
		return ""
	}
}

// MJNavbarLinkComponent represents the mj-navbar-link component
type MJNavbarLinkComponent struct {
	*BaseComponent
}

func NewMJNavbarLinkComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJNavbarLinkComponent {
	return &MJNavbarLinkComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJNavbarLinkComponent) Render(w io.StringWriter) error {
	return c.RenderWithBaseURL(w, "")
}

func (c *MJNavbarLinkComponent) RenderWithBaseURL(w io.StringWriter, baseURL string) error {
	href := c.getAttribute(constants.MJMLHref)
	target := c.getAttribute(constants.MJMLTarget)
	color := c.getAttribute(constants.CSSColor)
	fontFamily := c.getAttribute(constants.MJMLFontFamily)
	fontSize := c.getAttribute(constants.MJMLFontSize)
	fontStyle := c.getAttribute(constants.MJMLFontStyle)
	fontWeight := c.getAttribute(constants.MJMLFontWeight)
	lineHeight := c.getAttribute(constants.MJMLLineHeight)
	textDecoration := c.getAttribute(constants.MJMLTextDecoration)
	textTransform := c.getAttribute("text-transform")
	padding := c.getAttribute(constants.MJMLPadding)

	// Handle individual padding properties
	paddingTop := c.getAttribute(constants.MJMLPaddingTop)
	paddingRight := c.getAttribute(constants.MJMLPaddingRight)
	paddingBottom := c.getAttribute(constants.MJMLPaddingBottom)
	paddingLeft := c.getAttribute(constants.MJMLPaddingLeft)

	// Build full URL by combining base URL and href
	fullHref := href
	if baseURL != "" && href != "" {
		if strings.HasSuffix(baseURL, "/") && strings.HasPrefix(href, "/") {
			fullHref = baseURL + href[1:] // Remove duplicate slash
		} else if !strings.HasSuffix(baseURL, "/") && !strings.HasPrefix(href, "/") {
			fullHref = baseURL + "/" + href // Add missing slash
		} else {
			fullHref = baseURL + href
		}
	}

	// Build CSS class attribute
	cssClass := "mj-link"
	if additionalClass := c.getAttribute(constants.MJMLCSSClass); additionalClass != "" {
		cssClass = cssClass + " " + additionalClass
	}

	linkTag := html.NewHTMLTag("a").
		AddAttribute(constants.AttrHref, fullHref).
		AddAttribute(constants.AttrTarget, target).
		AddAttribute(constants.AttrClass, cssClass).
		AddStyle(constants.CSSDisplay, constants.DisplayInlineBlock).
		AddStyle(constants.CSSColor, color).
		AddStyle(constants.CSSFontFamily, fontFamily).
		AddStyle(constants.CSSFontSize, fontSize).
		AddStyle(constants.CSSFontWeight, fontWeight).
		AddStyle(constants.CSSLineHeight, lineHeight).
		AddStyle(constants.CSSTextDecoration, textDecoration).
		AddStyle("text-transform", textTransform).
		AddStyle(constants.CSSPadding, padding)

	// Only add rel attribute if it's not empty
	if rel := c.getAttribute("rel"); rel != "" {
		linkTag.AddAttribute(constants.AttrRel, rel)
	}

	// Only add individual padding properties if they're not empty
	if paddingTop != "" {
		linkTag.AddStyle(constants.CSSPaddingTop, paddingTop)
	}
	if paddingRight != "" {
		linkTag.AddStyle(constants.CSSPaddingRight, paddingRight)
	}
	if paddingBottom != "" {
		linkTag.AddStyle(constants.CSSPaddingBottom, paddingBottom)
	}
	if paddingLeft != "" {
		linkTag.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}

	// Only add font-style if it's not empty
	if fontStyle != "" {
		linkTag.AddStyle(constants.CSSFontStyle, fontStyle)
	}

	if err := linkTag.RenderOpen(w); err != nil {
		return err
	}

	// Render link content (text)
	content := strings.TrimSpace(c.Node.Text)
	if _, err := w.WriteString(content); err != nil {
		return err
	}

	if _, err := w.WriteString("</a>"); err != nil {
		return err
	}

	return nil
}

func (c *MJNavbarLinkComponent) getAttribute(name string) string {
	value := c.GetAttributeWithDefault(c, name)
	if name == constants.MJMLFontFamily && value != "" {
		c.TrackFontFamily(value)
	}
	return value
}

func (c *MJNavbarLinkComponent) GetTagName() string {
	return "mj-navbar-link"
}

func (c *MJNavbarLinkComponent) GetDefaultAttribute(name string) string {
	switch name {
	case constants.CSSColor:
		return "#000000"
	case constants.MJMLFontFamily:
		return fonts.DefaultFontStack
	case constants.MJMLFontSize:
		return "13px"
	case constants.MJMLFontWeight:
		return constants.FontWeightNormal
	case constants.MJMLLineHeight:
		return "22px"
	case constants.MJMLPadding:
		return "15px 10px"
	case constants.MJMLTarget:
		return constants.TargetBlank
	case constants.MJMLTextDecoration:
		return constants.TextDecorationNone
	case "text-transform":
		return "uppercase"
	default:
		return ""
	}
}
