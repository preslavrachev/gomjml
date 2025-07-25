package components

import (
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/parser"
)

// MJSocialComponent represents mj-social
type MJSocialComponent struct {
	*BaseComponent
}

// NewMJSocialComponent creates a new mj-social component
func NewMJSocialComponent(node *parser.MJMLNode) *MJSocialComponent {
	return &MJSocialComponent{
		BaseComponent: NewBaseComponent(node),
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
	align := c.getAttribute("align")

	// Outer table cell
	td := html.NewHTMLTag("td").
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("word-break", "break-word")

	output.WriteString(td.RenderOpen())

	// Inner table for social elements
	table := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddStyle("float", "none").
		AddStyle("display", "inline-table")

	output.WriteString(fmt.Sprintf(`<div align="%s">`, align))
	output.WriteString(table.RenderOpen())
	output.WriteString("<tr>")

	// Render social elements
	for _, child := range c.Children {
		if socialElement, ok := child.(*MJSocialElementComponent); ok {
			socialElement.SetContainerWidth(c.GetContainerWidth())
			childHTML, err := socialElement.Render()
			if err != nil {
				return "", err
			}
			output.WriteString(childHTML)
		}
	}

	output.WriteString("</tr>")
	output.WriteString(table.RenderClose())
	output.WriteString("</div>")
	output.WriteString(td.RenderClose())

	return output.String(), nil
}

func (c *MJSocialComponent) GetTagName() string {
	return "mj-social"
}

// MJSocialElementComponent represents mj-social-element
type MJSocialElementComponent struct {
	*BaseComponent
}

// NewMJSocialElementComponent creates a new mj-social-element component
func NewMJSocialElementComponent(node *parser.MJMLNode) *MJSocialElementComponent {
	return &MJSocialElementComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJSocialElementComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "align":
		return "center"
	case "alt":
		if name := c.getAttribute("name"); name != "" {
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
		// Default social icons - simplified
		switch c.getAttribute("name") {
		case "facebook":
			return "https://www.mailjet.com/images/social/facebook.png"
		case "twitter":
			return "https://www.mailjet.com/images/social/twitter.png"
		case "linkedin":
			return "https://www.mailjet.com/images/social/linkedin.png"
		case "instagram":
			return "https://www.mailjet.com/images/social/instagram.png"
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
	return c.GetAttributeWithDefault(c, name)
}

func (c *MJSocialElementComponent) Render() (string, error) {
	padding := c.getAttribute("padding")
	iconSize := c.getAttribute("icon-size")
	src := c.getAttribute("src")
	href := c.getAttribute("href")
	alt := c.getAttribute("alt")
	target := c.getAttribute("target")

	// Table cell
	td := html.NewHTMLTag("td").
		AddStyle("padding", padding).
		AddStyle("vertical-align", "middle")

	content := ""
	if src != "" {
		img := html.NewHTMLTag("img").
			AddAttribute("height", iconSize).
			AddAttribute("src", src).
			AddAttribute("style", fmt.Sprintf("border-radius:%s;display:block;", c.getAttribute("border-radius"))).
			AddAttribute("width", iconSize)

		if alt != "" {
			img.AddAttribute("alt", alt)
		}

		if href != "" {
			link := html.NewHTMLTag("a").
				AddAttribute("href", href).
				AddAttribute("target", target)
			content = link.RenderOpen() + img.RenderSelfClosing() + link.RenderClose()
		} else {
			content = img.RenderSelfClosing()
		}
	}

	return td.RenderOpen() + content + td.RenderClose(), nil
}

func (c *MJSocialElementComponent) GetTagName() string {
	return "mj-social-element"
}