package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

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
	case "align":
		return "center"
	case "border-radius":
		return "3px"
	case "color":
		return "#333333"
	case "font-family":
		return "Ubuntu, Helvetica, Arial, sans-serif"
	case "font-size":
		return "13px"
	case "icon-size":
		return "20px"
	case "inner-padding":
		return "4px"
	case "line-height":
		return "22px"
	case "mode":
		return "horizontal"
	case "padding":
		return "10px 25px"
	case "table-layout":
		return "auto"
	case "text-decoration":
		return "none"
	default:
		return ""
	}
}

func (c *MJSocialComponent) getAttribute(name string) string {
	return c.GetAttributeWithDefault(c, name)
}

// Render implements optimized Writer-based rendering for MJSocialComponent
func (c *MJSocialComponent) Render(w io.Writer) error {
	padding := c.getAttribute("padding")
	align := c.getAttribute("align")

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
	cssClass := c.Node.GetAttribute("css-class")
	if cssClass != "" {
		td.AddAttribute("class", cssClass)
	}

	td.AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("word-break", "break-word")

	if _, err := w.Write([]byte(td.RenderOpen())); err != nil {
		return err
	}

	// MSO conditional opening - use parent's align attribute
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
			// Pass parent attributes to child if not explicitly set
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
}

// NewMJSocialElementComponent creates a new mj-social-element component
func NewMJSocialElementComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJSocialElementComponent {
	return &MJSocialElementComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJSocialElementComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "align":
		return "left"
	case "alt":
		return ""
	case "border-radius":
		return "3px"
	case "color":
		return "#000"
	case "font-family":
		return "Ubuntu, Helvetica, Arial, sans-serif"
	case "font-size":
		return "13px"
	case "font-style":
		return "normal"
	case "font-weight":
		return "normal"
	case "href":
		return ""
	case "icon-size":
		return "20px"
	case "icon-height":
		return "" // No default, falls back to icon-size
	case "line-height":
		return "1"
	case "name":
		return ""
	case "padding":
		return "4px"
	case "src":
		// Default social icons from MJML standard locations
		// Get name directly from node to avoid circular dependency
		nameAttr := c.Node.GetAttribute("name")
		switch nameAttr {
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
		default:
			return ""
		}
	case "target":
		return "_blank"
	case "text-decoration":
		return "none"
	case "text-padding":
		return "4px 4px 4px 0"
	case "vertical-align":
		return "middle"
	default:
		return ""
	}
}

func (c *MJSocialElementComponent) getAttribute(name string) string {
	// 1. Check explicit element attribute first
	if value := c.Node.GetAttribute(name); value != "" {
		return value
	}

	// 2. Check parent mj-social for inheritable attributes
	if c.parentSocial != nil {
		inheritableAttrs := []string{
			"color", "font-family", "font-size", "line-height",
			"text-decoration", "border-radius", "icon-size",
		}
		for _, attr := range inheritableAttrs {
			if attr == name {
				// First check parent's explicit attribute
				if parentValue := c.parentSocial.Node.GetAttribute(name); parentValue != "" {
					return parentValue
				}
				// Then check parent's default attribute
				if parentDefault := c.parentSocial.GetDefaultAttribute(name); parentDefault !=
					"" {
					return parentDefault
				}
			}
		}
	}

	// 3. Check platform-specific defaults (for background-color)
	if name == "background-color" {
		socialName := c.Node.GetAttribute("name")
		platformDefaults := map[string]string{
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
		}
		if bgColor, exists := platformDefaults[socialName]; exists {
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
	target := c.getAttribute("target")
	backgroundColor := c.getAttribute("background-color")
	borderRadius := c.getAttribute("border-radius")

	// Skip rendering if no src provided
	if src == "" {
		return nil
	}

	// MSO conditional for individual social element
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
	paddingTd := html.NewHTMLTag("td").
		AddStyle("padding", padding).
		AddStyle("vertical-align", "middle")

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
	iconTd := html.NewHTMLTag("td").
		AddStyle("font-size", "0").
		AddStyle("height", iconHeight).
		AddStyle("vertical-align", "middle").
		AddStyle("width", iconSize)

	if _, err := w.Write([]byte(iconTd.RenderOpen())); err != nil {
		return err
	}

	// Image with optional link - remove "px" suffix from dimensions for HTML attributes
	heightAttr := strings.TrimSuffix(iconHeight, "px")
	widthAttr := strings.TrimSuffix(iconSize, "px")

	img := html.NewHTMLTag("img")

	// Add alt first to match MRML attribute order
	if alt != "" {
		img.AddAttribute("alt", alt)
	}

	img.AddAttribute("height", heightAttr).
		AddAttribute("src", src).
		AddAttribute("width", widthAttr).
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
	// Preserve original whitespace like MRML does (don't trim)
	textContent := c.Node.Text
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

		// Add font-style in correct position if explicitly set
		if fontStyle := c.Node.GetAttribute("font-style"); fontStyle != "" {
			textElement.AddStyle("font-style", fontStyle)
		}

		// Add font-weight in correct position if explicitly set
		if fontWeight := c.Node.GetAttribute("font-weight"); fontWeight != "" {
			textElement.AddStyle("font-weight", fontWeight)
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
