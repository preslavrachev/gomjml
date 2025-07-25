package components

import (
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/parser"
)

// MJImageComponent represents mj-image
type MJImageComponent struct {
	*BaseComponent
}

// NewMJImageComponent creates a new mj-image component
func NewMJImageComponent(node *parser.MJMLNode) *MJImageComponent {
	return &MJImageComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJImageComponent) Render() (string, error) {
	var output strings.Builder

	// Helper function to get attribute with default
	getAttr := func(name string) string {
		if attr := c.GetAttribute(name); attr != nil {
			return *attr
		}
		return c.GetDefaultAttribute(name)
	}

	// Get attributes with defaults
	align := getAttr("align")
	alt := getAttr("alt")
	border := getAttr("border")
	borderRadius := getAttr("border-radius")
	height := getAttr("height")
	href := getAttr("href")
	padding := getAttr("padding")
	rel := getAttr("rel")
	src := getAttr("src")
	target := getAttr("target")
	title := getAttr("title")
	width := getAttr("width")

	if src == "" {
		return "", fmt.Errorf("mj-image requires src attribute")
	}

	// Parse width to remove 'px' suffix for img width attribute
	imgWidth := width
	if strings.HasSuffix(width, "px") {
		imgWidth = strings.TrimSuffix(width, "px")
	}

	// Create TR element
	output.WriteString("<tr>")

	// Create TD container with alignment and base styles
	tdTag := html.NewHTMLTag("td").
		AddAttribute("align", align).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("word-break", "break-word")

	output.WriteString(tdTag.RenderOpen())

	// Image table
	tableTag := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddStyle("border-collapse", "collapse").
		AddStyle("border-spacing", "0px")

	output.WriteString(tableTag.RenderOpen())
	output.WriteString("<tbody><tr>")

	// Image cell with width constraint
	imageTdTag := html.NewHTMLTag("td")
	if width != "" {
		imageTdTag.AddStyle("width", width)
	}

	output.WriteString(imageTdTag.RenderOpen())

	// Optional link wrapper
	if href != "" {
		linkTag := html.NewHTMLTag("a").
			AddAttribute("href", href)

		if rel != "" {
			linkTag.AddAttribute("rel", rel)
		}
		if target != "" {
			linkTag.AddAttribute("target", target)
		}

		output.WriteString(linkTag.RenderOpen())
	}

	// Image element with styles
	imgTag := html.NewHTMLTag("img")
	c.AddDebugAttribute(imgTag, "image")

	// Set image attributes - always include alt for accessibility
	imgTag.AddAttribute("alt", alt)
	if height != "" {
		imgTag.AddAttribute("height", height)
	}
	imgTag.AddAttribute("src", src)
	if title != "" {
		imgTag.AddAttribute("title", title)
	}
	if imgWidth != "" {
		imgTag.AddAttribute("width", imgWidth)
	}

	// Apply image styles
	imgTag.AddStyle("border", border).
		AddStyle("display", "block").
		AddStyle("outline", "none").
		AddStyle("text-decoration", "none").
		AddStyle("height", height).
		AddStyle("width", "100%").
		AddStyle("font-size", "13px")

	if borderRadius != "" {
		imgTag.AddStyle("border-radius", borderRadius)
	}

	output.WriteString(imgTag.RenderSelfClosing())

	// Close optional link wrapper
	if href != "" {
		output.WriteString("</a>")
	}

	output.WriteString(imageTdTag.RenderClose())
	output.WriteString("</tr></tbody>")
	output.WriteString(tableTag.RenderClose())
	output.WriteString(tdTag.RenderClose())
	output.WriteString("</tr>")

	return output.String(), nil
}

func (c *MJImageComponent) GetTagName() string {
	return "mj-image"
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
		return ""
	default:
		return ""
	}
}
