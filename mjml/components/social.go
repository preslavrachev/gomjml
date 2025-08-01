package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/debug"
	"github.com/preslavrachev/gomjml/mjml/fonts"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// stripPxSuffix removes the "px" suffix from a CSS value if present
func stripPxSuffix(value string) string {
	return strings.TrimSuffix(value, "px")
}

// platformDefaults defines the default background colors for social media platforms
var platformDefaults = map[string]string{
	"youtube":    "#EB3323",
	"facebook":   "#3b5998",
	"twitter":    "#55acee",
	"google":     "#dc4e41",
	"github":     "#000000",
	"dribbble":   "#D95988",
	"instagram":  "#3f729b",
	"linkedin":   "#0077b5",
	"pinterest":  "#bd081c",
	"medium":     "#000000",
	"tumblr":     "#344356",
	"vimeo":      "#53B4E7",
	"web":        "#4BADE9",
	"snapchat":   "#FFFA54",
	"soundcloud": "#EF7F31",
	"xing":       "#296366",
}

// MJSocialComponent represents mj-social
type MJSocialComponent struct {
	*BaseComponent
}

// NewMJSocialComponent creates a new mj-social component
func NewMJSocialComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJSocialComponent {
	return &MJSocialComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJSocialComponent) GetDefaultAttribute(name string) string {
	switch name {
	case constants.MJMLAlign:
		return constants.AlignCenter
	case constants.MJMLBorderRadius:
		return "3px"
	case constants.MJMLColor:
		return "#333333"
	case constants.MJMLFontFamily:
		return fonts.DefaultFontStack
	case constants.MJMLFontSize:
		return "13px"
	case constants.MJMLIconSize:
		return "20px"
	case constants.MJMLInnerPadding:
		return "4px"
	case constants.MJMLLineHeight:
		return "22px"
	case constants.MJMLMode:
		return "horizontal"
	case constants.MJMLPadding:
		return "10px 25px"
	case constants.MJMLTableLayout:
		return "auto"
	case constants.MJMLTextDecoration:
		return constants.TextDecorationNone
	default:
		return ""
	}
}

func (c *MJSocialComponent) getAttribute(name string) string {
	value := c.GetAttributeWithDefault(c, name)
	// Ensure font families are tracked
	if name == constants.MJMLFontFamily && value != "" {
		c.TrackFontFamily(value)
	}
	return value
}

// Render implements optimized Writer-based rendering for MJSocialComponent
func (c *MJSocialComponent) Render(w io.Writer) error {
	padding := c.getAttribute(constants.MJMLPadding)
	align := c.getAttribute(constants.MJMLAlign)
	mode := c.getAttribute(constants.MJMLMode)

	// Wrap in table row (required when inside column tbody)
	if _, err := w.Write([]byte("<tr>")); err != nil {
		return err
	}

	// Outer table cell with align attribute
	td := html.NewHTMLTag("td")
	if align != "" {
		td.AddAttribute("align", align)
	}

	// Add CSS class if specified
	cssClass := c.Node.GetAttribute(constants.MJMLCSSClass)
	if cssClass != "" {
		td.AddAttribute(constants.AttrClass, cssClass)
	}

	// Add container background color if specified
	containerBg := c.Node.GetAttribute("container-background-color")
	if containerBg != "" {
		td.AddStyle(constants.CSSBackground, containerBg)
	}

	td.AddStyle(constants.CSSFontSize, "0px").
		AddStyle(constants.CSSPadding, padding)

	// Handle padding-left separately if specified
	paddingLeft := c.Node.GetAttribute(constants.MJMLPaddingLeft)
	if paddingLeft != "" {
		td.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}

	td.AddStyle("word-break", "break-word")

	if _, err := w.Write([]byte(td.RenderOpen())); err != nil {
		return err
	}

	if mode == "vertical" {
		// Vertical mode: single table with margin:0px
		table := html.NewHTMLTag("table").
			AddAttribute("border", "0").
			AddAttribute("cellpadding", "0").
			AddAttribute("cellspacing", "0").
			AddAttribute("role", "presentation").
			AddStyle("margin", "0px")

		if _, err := w.Write([]byte(table.RenderOpen())); err != nil {
			return err
		}
		if _, err := w.Write([]byte("<tbody>")); err != nil {
			return err
		}

		// Render social elements as table rows
		for _, child := range c.Children {
			if socialElement, ok := child.(*MJSocialElementComponent); ok {
				socialElement.SetContainerWidth(c.GetContainerWidth())
				socialElement.InheritFromParent(c)
				socialElement.SetVerticalMode(true)
				if err := socialElement.Render(w); err != nil {
					return err
				}
			}
		}

		if _, err := w.Write([]byte("</tbody>")); err != nil {
			return err
		}
		if _, err := w.Write([]byte(table.RenderClose())); err != nil {
			return err
		}
	} else {
		// Horizontal mode (default): MSO conditional with inline tables
		msoAlign := align
		if msoAlign == "" {
			msoAlign = "center"
		}
		msoTable := fmt.Sprintf(
			"<!--[if mso | IE]><table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\" align=\"%s\"><tr><![endif]-->",
			msoAlign,
		)
		if _, err := w.Write([]byte(msoTable)); err != nil {
			return err
		}

		// Render social elements
		for _, child := range c.Children {
			if socialElement, ok := child.(*MJSocialElementComponent); ok {
				socialElement.SetContainerWidth(c.GetContainerWidth())
				socialElement.InheritFromParent(c)
				if err := socialElement.Render(w); err != nil {
					return err
				}
			}
		}

		// MSO conditional closing
		if _, err := w.Write([]byte("<!--[if mso | IE]></tr></table><![endif]-->")); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte(td.RenderClose())); err != nil {
		return err
	}

	// Close table row
	if _, err := w.Write([]byte("</tr>")); err != nil {
		return err
	}

	return nil
}

func (c *MJSocialComponent) GetTagName() string {
	return "mj-social"
}

// MJSocialElementComponent represents mj-social-element
type MJSocialElementComponent struct {
	*BaseComponent
	parentSocial *MJSocialComponent // Reference to parent for attribute inheritance
	verticalMode bool               // Whether this element should render in vertical mode
}

// NewMJSocialElementComponent creates a new mj-social-element component
func NewMJSocialElementComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJSocialElementComponent {
	return &MJSocialElementComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJSocialElementComponent) GetDefaultAttribute(name string) string {
	switch name {
	case constants.MJMLAlign:
		return constants.AlignLeft
	case constants.MJMLAlt:
		return ""
	case constants.MJMLBorderRadius:
		return "3px"
	case constants.MJMLColor:
		return "#000"
	case constants.MJMLFontFamily:
		return fonts.DefaultFontStack
	case constants.MJMLFontSize:
		return "13px"
	case constants.MJMLFontStyle:
		return constants.FontStyleNormal
	case constants.MJMLFontWeight:
		return constants.FontWeightNormal
	case constants.MJMLHref:
		return ""
	case constants.MJMLIconSize:
		return "20px"
	case constants.MJMLIconHeight:
		return "" // No default, falls back to icon-size
	case constants.MJMLLineHeight:
		return "1"
	case constants.MJMLName:
		return ""
	case constants.MJMLPadding:
		return "4px"
	case "src":
		// Default social icons from MJML standard locations
		// Get name directly from node to avoid circular dependency
		nameAttr := c.Node.GetAttribute("name")

		// Handle variants like "facebook-noshare" by extracting base platform name
		baseName := nameAttr
		if strings.Contains(nameAttr, "-") {
			baseName = strings.Split(nameAttr, "-")[0]
		}

		switch baseName {
		case "facebook":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/facebook.png"
		case "twitter":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/twitter.png"
		case "linkedin":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/linkedin.png"
		case "google":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/google-plus.png"
		case "github":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/github.png"
		case "dribbble":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/dribbble.png"
		case "instagram":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/instagram.png"
		case "youtube":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/youtube.png"
		case "pinterest":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/pinterest.png"
		case "medium":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/medium.png"
		case "tumblr":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/tumblr.png"
		case "vimeo":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/vimeo.png"
		case "web":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/web.png"
		case "snapchat":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/snapchat.png"
		case "soundcloud":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/soundcloud.png"
		case "xing":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/xing.png"
		default:
			return ""
		}
	case constants.MJMLTarget:
		return constants.TargetBlank
	case constants.MJMLTextDecoration:
		return constants.TextDecorationNone
	case constants.MJMLTextPadding:
		return "4px 4px 4px 0"
	case constants.MJMLVerticalAlign:
		return constants.VAlignMiddle
	default:
		return ""
	}
}

func (c *MJSocialElementComponent) getAttribute(name string) string {
	// 1. Check explicit element attribute first
	if value := c.Node.GetAttribute(name); value != "" {
		// Track font families
		if name == constants.MJMLFontFamily {
			c.TrackFontFamily(value)
		}
		return value
	}

	// 2. Check parent mj-social for inheritable attributes
	if c.parentSocial != nil {
		inheritableAttrs := []string{
			"color", "font-family", "font-size", "line-height",
			"text-decoration", "border-radius", "icon-size",
			"font-weight", "font-style", "icon-height", "icon-padding",
			"inner-padding", "text-padding",
		}
		for _, attr := range inheritableAttrs {
			if attr == name {
				// First check parent's explicit attribute
				if parentValue := c.parentSocial.Node.GetAttribute(name); parentValue != "" {
					debug.DebugLogWithData(
						"social-attr",
						"parent-explicit",
						"Using parent explicit attribute",
						map[string]interface{}{
							"attr":    name,
							"value":   parentValue,
							"element": c.Node.GetAttribute("name"),
						},
					)
					// Track font families
					if name == constants.MJMLFontFamily {
						c.TrackFontFamily(parentValue)
					}
					return parentValue
				}
				// Then check parent's default attribute
				if parentDefault := c.parentSocial.GetDefaultAttribute(name); parentDefault != "" {
					debug.DebugLogWithData(
						"social-attr",
						"parent-default",
						"Using parent default attribute",
						map[string]interface{}{
							"attr":    name,
							"value":   parentDefault,
							"element": c.Node.GetAttribute("name"),
						},
					)
					// Track font families
					if name == constants.MJMLFontFamily {
						c.TrackFontFamily(parentDefault)
					}
					return parentDefault
				}
			}
		}
	}

	// 3. Check platform-specific defaults (for background-color)
	if name == constants.MJMLBackgroundColor {
		socialName := c.Node.GetAttribute("name")

		// Handle variants like "facebook-noshare" by extracting base platform name
		baseName := socialName
		if strings.Contains(socialName, "-") {
			baseName = strings.Split(socialName, "-")[0]
		}

		if bgColor, exists := platformDefaults[baseName]; exists {
			return bgColor
		}
	}

	// 4. Fall back to component defaults
	return c.GetDefaultAttribute(name)
}

// InheritFromParent sets the parent reference for attribute inheritance
func (c *MJSocialElementComponent) InheritFromParent(parent *MJSocialComponent) {
	c.parentSocial = parent
}

// SetVerticalMode sets whether this element should render in vertical mode
func (c *MJSocialElementComponent) SetVerticalMode(vertical bool) {
	c.verticalMode = vertical
}

// Render implements optimized Writer-based rendering for MJSocialElementComponent
func (c *MJSocialElementComponent) Render(w io.Writer) error {
	padding := c.getAttribute("padding")
	iconSize := c.getAttribute("icon-size")
	iconHeight := c.getAttribute("icon-height")
	if iconHeight == "" {
		iconHeight = iconSize // fallback to icon-size
	}
	src := c.getAttribute("src")
	href := c.getAttribute("href")
	alt := c.getAttribute("alt")

	// Handle special sharing URL generation for known platforms
	if href != "" && !strings.HasPrefix(href, "http") {
		nameAttr := c.Node.GetAttribute("name")
		if nameAttr == "facebook" {
			// Convert simple href to Facebook sharing URL
			href = "https://www.facebook.com/sharer/sharer.php?u=" + href
		}
	}
	target := c.getAttribute("target")
	backgroundColor := c.getAttribute("background-color")
	borderRadius := c.getAttribute("border-radius")

	// Skip rendering if no src provided
	if src == "" {
		return nil
	}

	if c.verticalMode {
		// Vertical mode: render as table row without MSO conditionals
		if _, err := w.Write([]byte("<tr>")); err != nil {
			return err
		}

		// Icon cell
		iconTd := html.NewHTMLTag("td").
			AddStyle("padding", padding).
			AddStyle("vertical-align", "middle")

		if _, err := w.Write([]byte(iconTd.RenderOpen())); err != nil {
			return err
		}

		// Inner table with background color
		innerTable := html.NewHTMLTag("table").
			AddAttribute("border", "0").
			AddAttribute("cellpadding", "0").
			AddAttribute("cellspacing", "0").
			AddAttribute("role", "presentation").
			AddStyle("background", backgroundColor).
			AddStyle("border-radius", borderRadius).
			AddStyle("width", iconSize)

		if _, err := w.Write([]byte(innerTable.RenderOpen())); err != nil {
			return err
		}
		if _, err := w.Write([]byte("<tbody><tr>")); err != nil {
			return err
		}

		// Icon cell
		iconInnerTd := html.NewHTMLTag("td").
			AddStyle("font-size", "0").
			AddStyle("height", iconHeight).
			AddStyle("vertical-align", "middle").
			AddStyle("width", iconSize)

		if _, err := w.Write([]byte(iconInnerTd.RenderOpen())); err != nil {
			return err
		}

		// Image without link in vertical mode (as per MRML output)
		heightAttr := stripPxSuffix(iconHeight)
		widthAttr := stripPxSuffix(iconSize)

		img := html.NewHTMLTag("img")
		if alt != "" {
			img.AddAttribute("alt", alt)
		}
		img.AddAttribute("height", heightAttr).
			AddAttribute("src", src).
			AddAttribute("width", widthAttr).
			AddStyle("border-radius", borderRadius).
			AddStyle("display", "block")

		if _, err := w.Write([]byte(img.RenderSelfClosing())); err != nil {
			return err
		}

		if _, err := w.Write([]byte(iconInnerTd.RenderClose())); err != nil {
			return err
		}
		if _, err := w.Write([]byte("</tr></tbody>")); err != nil {
			return err
		}
		if _, err := w.Write([]byte(innerTable.RenderClose())); err != nil {
			return err
		}
		if _, err := w.Write([]byte(iconTd.RenderClose())); err != nil {
			return err
		}

		// Text content cell
		textContent := c.Node.Text
		if textContent != "" {
			textTd := html.NewHTMLTag("td").
				AddStyle("vertical-align", "middle").
				AddStyle("padding", c.getAttribute("text-padding"))

			if _, err := w.Write([]byte(textTd.RenderOpen())); err != nil {
				return err
			}

			// Text content with span (no link in vertical mode as per MRML)
			textSpan := html.NewHTMLTag("span").
				AddStyle("color", c.getAttribute("color")).
				AddStyle("font-size", c.getAttribute("font-size")).
				AddStyle("font-family", c.getAttribute("font-family")).
				AddStyle("line-height", c.getAttribute("line-height")).
				AddStyle("text-decoration", c.getAttribute("text-decoration"))

			if _, err := w.Write([]byte(textSpan.RenderOpen())); err != nil {
				return err
			}
			if _, err := w.Write([]byte(textContent)); err != nil {
				return err
			}
			if _, err := w.Write([]byte(textSpan.RenderClose())); err != nil {
				return err
			}
			if _, err := w.Write([]byte(textTd.RenderClose())); err != nil {
				return err
			}
		}

		if _, err := w.Write([]byte("</tr>")); err != nil {
			return err
		}

		return nil
	}

	// Horizontal mode: MSO conditional for individual social element
	if _, err := w.Write([]byte("<!--[if mso | IE]><td><![endif]-->")); err != nil {
		return err
	}

	// Outer table (inline-table display) - inherit align from parent
	align := "center" // default
	if c.parentSocial != nil {
		if parentAlign := c.parentSocial.getAttribute("align"); parentAlign != "" {
			align = parentAlign
		}
	}

	outerTable := html.NewHTMLTag("table")
	c.AddDebugAttribute(outerTable, "social-element")
	outerTable.
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", align).
		AddStyle("float", "none").
		AddStyle("display", "inline-table")

	if _, err := w.Write([]byte(outerTable.RenderOpen())); err != nil {
		return err
	}

	// Add CSS class to tr if specified on individual social element
	cssClass := c.Node.GetAttribute("css-class")
	if cssClass != "" {
		trTag := fmt.Sprintf("<tbody><tr class=\"%s\">", cssClass)
		if _, err := w.Write([]byte(trTag)); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte("<tbody><tr>")); err != nil {
			return err
		}
	}

	// Padding cell
	paddingTd := html.NewHTMLTag("td")

	// Use element's own padding first, then fall back to inherited inner-padding
	iconPadding := padding
	if padding == c.GetDefaultAttribute("padding") {
		// Only use inner-padding if element doesn't have explicit padding
		if inheritedInnerPadding := c.getAttribute("inner-padding"); inheritedInnerPadding != "" {
			iconPadding = inheritedInnerPadding
		}
	}

	// Handle padding and padding-bottom specially
	paddingBottom := c.Node.GetAttribute("padding-bottom")
	if paddingBottom != "" {
		paddingTd.AddStyle("padding", iconPadding).
			AddStyle("padding-bottom", paddingBottom)
	} else {
		paddingTd.AddStyle("padding", iconPadding)
	}

	paddingTd.AddStyle("vertical-align", "middle")

	if _, err := w.Write([]byte(paddingTd.RenderOpen())); err != nil {
		return err
	}

	// Inner table with background color
	innerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddStyle("background", backgroundColor).
		AddStyle("border-radius", borderRadius).
		AddStyle("width", iconSize)

	if _, err := w.Write([]byte(innerTable.RenderOpen())); err != nil {
		return err
	}
	if _, err := w.Write([]byte("<tbody><tr>")); err != nil {
		return err
	}

	// Icon cell
	iconTd := html.NewHTMLTag("td")

	// Handle icon-padding if specified (check both element and inherited from parent)
	iconPaddingAttr := c.getAttribute("icon-padding")
	if iconPaddingAttr != "" {
		iconTd.AddStyle("padding", iconPaddingAttr)
	}

	iconTd.AddStyle("font-size", "0").
		AddStyle("height", iconHeight).
		AddStyle("vertical-align", "middle").
		AddStyle("width", iconSize)

	if _, err := w.Write([]byte(iconTd.RenderOpen())); err != nil {
		return err
	}

	// Image with optional link - remove "px" suffix from dimensions for HTML attributes
	heightAttr := stripPxSuffix(iconHeight)
	widthAttr := stripPxSuffix(iconSize)

	img := html.NewHTMLTag("img")

	// Add alt first to match MRML attribute order
	if alt != "" {
		img.AddAttribute("alt", alt)
	}

	img.AddAttribute("height", heightAttr).
		AddAttribute("src", src)

	// Add title attribute if specified
	title := c.Node.GetAttribute("title")
	if title != "" {
		img.AddAttribute("title", title)
	}

	img.AddAttribute("width", widthAttr).
		AddStyle("border-radius", borderRadius).
		AddStyle("display", "block")

	if href != "" {
		link := html.NewHTMLTag("a").
			AddAttribute("href", href).
			AddAttribute("target", target)
		if _, err := w.Write([]byte(link.RenderOpen())); err != nil {
			return err
		}
		if _, err := w.Write([]byte(img.RenderSelfClosing())); err != nil {
			return err
		}
		if _, err := w.Write([]byte(link.RenderClose())); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte(img.RenderSelfClosing())); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte(iconTd.RenderClose())); err != nil {
		return err
	}
	if _, err := w.Write([]byte("</tr></tbody>")); err != nil {
		return err
	}
	if _, err := w.Write([]byte(innerTable.RenderClose())); err != nil {
		return err
	}
	if _, err := w.Write([]byte(paddingTd.RenderClose())); err != nil {
		return err
	}

	// Render text content if present - INSIDE the same <tr>
	// Use GetMixedContent to preserve HTML tags like <b>, <i>, etc. within text
	textContent := c.Node.GetMixedContent()
	debug.DebugLogWithData(
		"social-element",
		"content-selection",
		"Selected text content source",
		map[string]interface{}{
			"element_name":   c.Node.GetAttribute("name"),
			"plain_text":     c.Node.Text,
			"mixed_content":  textContent,
			"has_children":   len(c.Node.Children) > 0,
			"content_length": len(textContent),
		},
	)
	if textContent != "" {
		// Text cell with padding and styling
		textTd := html.NewHTMLTag("td").
			AddStyle("vertical-align", c.getAttribute("vertical-align")).
			AddStyle("padding", c.getAttribute("text-padding"))

		if _, err := w.Write([]byte(textTd.RenderOpen())); err != nil {
			return err
		}

		// Text content with social styling - use <a> if href present, <span> otherwise
		var textElement *html.HTMLTag
		if href != "" {
			// Use <a> tag when there's a link
			textElement = html.NewHTMLTag("a").
				AddAttribute("href", href).
				AddAttribute("target", target)
		} else {
			// Use <span> tag when no link
			textElement = html.NewHTMLTag("span")
		}

		// Add styling - maintain MRML CSS property order
		textElement.AddStyle("color", c.getAttribute("color")).
			AddStyle("font-size", c.getAttribute("font-size"))

		// Add font-weight after font-size (inherited from parent or explicit)
		fontWeight := c.getAttribute("font-weight")
		if fontWeight != "" && fontWeight != "normal" { // Only add if not default
			textElement.AddStyle("font-weight", fontWeight)
		}

		// Add font-style after font-weight (inherited from parent or explicit)
		fontStyle := c.getAttribute("font-style")
		if fontStyle != "" && fontStyle != "normal" { // Only add if not default
			textElement.AddStyle("font-style", fontStyle)
		}

		textElement.AddStyle("font-family", c.getAttribute("font-family")).
			AddStyle("line-height", c.getAttribute("line-height")).
			AddStyle("text-decoration", c.getAttribute("text-decoration"))

		if _, err := w.Write([]byte(textElement.RenderOpen())); err != nil {
			return err
		}
		if _, err := w.Write([]byte(textContent)); err != nil {
			return err
		}
		if _, err := w.Write([]byte(textElement.RenderClose())); err != nil {
			return err
		}
		if _, err := w.Write([]byte(textTd.RenderClose())); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("</tr></tbody>")); err != nil {
		return err
	}
	if _, err := w.Write([]byte(outerTable.RenderClose())); err != nil {
		return err
	}

	// Close MSO conditional
	if _, err := w.Write([]byte("<!--[if mso | IE]></td><![endif]-->")); err != nil {
		return err
	}

	return nil
}

func (c *MJSocialElementComponent) GetTagName() string {
	return "mj-social-element"
}
