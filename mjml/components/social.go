package components

import (
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

func (c *MJSocialComponent) Render() (string, error) {
	var output strings.Builder

	padding := c.getAttribute("padding")

	// Outer table cell
	td := html.NewHTMLTag("td").
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("word-break", "break-word")

	output.WriteString(td.RenderOpen())

	// MSO conditional opening
	output.WriteString("<!--[if mso | IE]><table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\" align=\"center\"><tr><![endif]-->")

	// Render social elements
	for _, child := range c.Children {
		if socialElement, ok := child.(*MJSocialElementComponent); ok {
			socialElement.SetContainerWidth(c.GetContainerWidth())
			// Pass parent attributes to child if not explicitly set
			socialElement.InheritFromParent(c)
			childHTML, err := socialElement.Render()
			if err != nil {
				return "", err
			}
			output.WriteString(childHTML)
		}
	}

	// MSO conditional closing
	output.WriteString("<!--[if mso | IE]></tr></table><![endif]-->")
	output.WriteString(td.RenderClose())

	return output.String(), nil
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
		return "center"
	case "alt":
		if name := c.Node.GetAttribute("name"); name != "" {
			return name
		}
		return ""
	case "border-radius":
		return "3px"
	case "color":
		return "#333333"
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
		return "22px"
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
	case "vertical-align":
		return "middle"
	default:
		return ""
	}
}

func (c *MJSocialElementComponent) getAttribute(name string) string {
	// First check if element has the attribute explicitly set
	if value := c.Node.GetAttribute(name); value != "" {
		return value
	}

	// Check if parent has the attribute for inheritable attributes
	inheritableAttrs := map[string]bool{
		"icon-size": true, "font-size": true, "padding": true, "border-radius": true,
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

func (c *MJSocialElementComponent) Render() (string, error) {
	var output strings.Builder

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
		return "", nil
	}

	// MSO conditional for individual social element
	output.WriteString("<!--[if mso | IE]><td><![endif]-->")

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

	output.WriteString(outerTable.RenderOpen())
	output.WriteString("<tbody><tr>")

	// Padding cell
	paddingTd := html.NewHTMLTag("td").
		AddStyle("padding", padding).
		AddStyle("vertical-align", "middle")

	output.WriteString(paddingTd.RenderOpen())

	// Inner table with background color
	innerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddStyle("background", backgroundColor).
		AddStyle("border-radius", borderRadius).
		AddStyle("width", iconSize)

	output.WriteString(innerTable.RenderOpen())
	output.WriteString("<tbody><tr>")

	// Icon cell
	iconTd := html.NewHTMLTag("td").
		AddStyle("font-size", "0").
		AddStyle("height", iconSize).
		AddStyle("vertical-align", "middle").
		AddStyle("width", iconSize)

	output.WriteString(iconTd.RenderOpen())

	// Image with optional link
	img := html.NewHTMLTag("img").
		AddAttribute("height", iconSize).
		AddAttribute("src", src).
		AddAttribute("width", iconSize).
		AddStyle("border-radius", borderRadius).
		AddStyle("display", "block")

	if alt != "" {
		img.AddAttribute("alt", alt)
	}

	if href != "" {
		link := html.NewHTMLTag("a").
			AddAttribute("href", href).
			AddAttribute("target", target)
		output.WriteString(link.RenderOpen())
		output.WriteString(img.RenderSelfClosing())
		output.WriteString(link.RenderClose())
	} else {
		output.WriteString(img.RenderSelfClosing())
	}

	output.WriteString(iconTd.RenderClose())
	output.WriteString("</tr></tbody>")
	output.WriteString(innerTable.RenderClose())
	output.WriteString(paddingTd.RenderClose())
	output.WriteString("</tr></tbody>")
	output.WriteString(outerTable.RenderClose())

	// Close MSO conditional
	output.WriteString("<!--[if mso | IE]></td><![endif]-->")

	return output.String(), nil
}

func (c *MJSocialElementComponent) GetTagName() string {
	return "mj-social-element"
}
