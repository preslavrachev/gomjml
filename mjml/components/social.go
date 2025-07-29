package components

import (
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

func (c *MJSocialComponent) RenderString() (string, error) {
	var output strings.Builder
	err := c.Render(&output)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// Render implements optimized Writer-based rendering for MJSocialComponent
func (c *MJSocialComponent) Render(w io.Writer) error {
	padding := c.getAttribute("padding")

	// Outer table cell
	td := html.NewHTMLTag("td").
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("word-break", "break-word")

	if _, err := w.Write([]byte(td.RenderOpen())); err != nil {
		return err
	}

	// MSO conditional opening
	if _, err := w.Write([]byte("<!--[if mso | IE]><table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\" align=\"center\"><tr><![endif]-->")); err != nil {
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
		if name := c.Node.GetAttribute("name"); name != "" {
			return name
		}
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
		case "instagram":
			return "https://www.mailjet.com/images/theme/v1/icons/ico-social/instagram.png"
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
	// Handle special case for href - if no href is provided, generate platform-specific sharing URL
	if name == "href" {
		if value := c.Node.GetAttribute(name); value != "" {
			socialName := c.Node.GetAttribute("name")
			baseURL := value

			// Generate platform-specific sharing URLs
			switch socialName {
			case "facebook":
				return "https://www.facebook.com/sharer/sharer.php?u=" + baseURL
			case "twitter":
				return "https://twitter.com/home?status=" + baseURL
			case "linkedin":
				return "https://www.linkedin.com/shareArticle?mini=true&url=" + baseURL + "&title=&summary=&source="
			case "google":
				return "https://plus.google.com/share?url=" + baseURL
			default:
				return value
			}
		}
		return ""
	}

	// First check if element has the attribute explicitly set
	if value := c.Node.GetAttribute(name); value != "" {
		return value
	}

	// For non-inheritable attributes like padding, go directly to defaults
	if name == "padding" {
		return c.GetDefaultAttribute(name)
	}

	// Check if parent has the attribute for inheritable attributes
	inheritableAttrs := map[string]bool{
		"icon-size": true, "font-size": true, "border-radius": true,
	}

	if c.parentSocial != nil && inheritableAttrs[name] {
		if parentValue := c.parentSocial.Node.GetAttribute(name); parentValue != "" {
			return parentValue
		}
	}

	// Fall back to default attributes
	return c.GetDefaultAttribute(name)
}

// InheritFromParent sets the parent reference for attribute inheritance
func (c *MJSocialElementComponent) InheritFromParent(parent *MJSocialComponent) {
	c.parentSocial = parent
}

func (c *MJSocialElementComponent) RenderString() (string, error) {
	var output strings.Builder
	err := c.Render(&output)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// Render implements optimized Writer-based rendering for MJSocialElementComponent
func (c *MJSocialElementComponent) Render(w io.Writer) error {
	padding := c.getAttribute("padding")
	iconSize := c.getAttribute("icon-size")
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

	// Outer table (inline-table display)
	outerTable := html.NewHTMLTag("table")
	c.AddDebugAttribute(outerTable, "social-element")
	outerTable.
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center").
		AddStyle("float", "none").
		AddStyle("display", "inline-table")

	if _, err := w.Write([]byte(outerTable.RenderOpen())); err != nil {
		return err
	}
	if _, err := w.Write([]byte("<tbody><tr>")); err != nil {
		return err
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
		AddStyle("height", iconSize).
		AddStyle("vertical-align", "middle").
		AddStyle("width", iconSize)

	if _, err := w.Write([]byte(iconTd.RenderOpen())); err != nil {
		return err
	}

	// Image with optional link - remove "px" suffix from dimensions for HTML attributes
	heightAttr := strings.TrimSuffix(iconSize, "px")
	widthAttr := strings.TrimSuffix(iconSize, "px")

	img := html.NewHTMLTag("img").
		AddAttribute("height", heightAttr).
		AddAttribute("src", src).
		AddAttribute("width", widthAttr).
		AddStyle("border-radius", borderRadius).
		AddStyle("display", "block")

	if alt != "" {
		img.AddAttribute("alt", alt)
	}

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
