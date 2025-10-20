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

// socialNetworkDefaults describes the default MJML metadata for a social network.
type socialNetworkDefaults struct {
	backgroundColor        string
	iconURL                string
	shareURLTemplate       string
	shareURLSkipSubstrings []string
}

// baseSocialNetworkDefaults matches MJML's default network definitions.
var baseSocialNetworkDefaults = map[string]socialNetworkDefaults{
	"facebook": {
		backgroundColor:        "#3b5998",
		iconURL:                "https://www.mailjet.com/images/theme/v1/icons/ico-social/facebook.png",
		shareURLTemplate:       "https://www.facebook.com/sharer/sharer.php?u=[[URL]]",
		shareURLSkipSubstrings: []string{"facebook.com/sharer"},
	},
	"twitter": {
		backgroundColor:        "#55acee",
		iconURL:                "https://www.mailjet.com/images/theme/v1/icons/ico-social/twitter.png",
		shareURLTemplate:       "https://twitter.com/intent/tweet?url=[[URL]]",
		shareURLSkipSubstrings: []string{"twitter.com/", "x.com/"},
	},
	"x": {
		backgroundColor:        "#000000",
		iconURL:                "https://www.mailjet.com/images/theme/v1/icons/ico-social/twitter-x.png",
		shareURLTemplate:       "https://twitter.com/intent/tweet?url=[[URL]]",
		shareURLSkipSubstrings: []string{"twitter.com/", "x.com/"},
	},
	"google": {
		backgroundColor:        "#dc4e41",
		iconURL:                "https://www.mailjet.com/images/theme/v1/icons/ico-social/google-plus.png",
		shareURLTemplate:       "https://plus.google.com/share?url=[[URL]]",
		shareURLSkipSubstrings: []string{"plus.google.com/share"},
	},
	"pinterest": {
		backgroundColor:        "#bd081c",
		iconURL:                "https://www.mailjet.com/images/theme/v1/icons/ico-social/pinterest.png",
		shareURLTemplate:       "https://pinterest.com/pin/create/button/?url=[[URL]]&media=&description=",
		shareURLSkipSubstrings: []string{"pinterest.com/pin/create/button"},
	},
	"linkedin": {
		backgroundColor:        "#0077b5",
		iconURL:                "https://www.mailjet.com/images/theme/v1/icons/ico-social/linkedin.png",
		shareURLTemplate:       "https://www.linkedin.com/shareArticle?mini=true&url=[[URL]]&title=&summary=&source=",
		shareURLSkipSubstrings: []string{"linkedin.com/shareArticle"},
	},
	"instagram": {
		backgroundColor: "#3f729b",
		iconURL:         "https://www.mailjet.com/images/theme/v1/icons/ico-social/instagram.png",
	},
	"web": {
		backgroundColor: "#4BADE9",
		iconURL:         "https://www.mailjet.com/images/theme/v1/icons/ico-social/web.png",
	},
	"snapchat": {
		backgroundColor: "#FFFA54",
		iconURL:         "https://www.mailjet.com/images/theme/v1/icons/ico-social/snapchat.png",
	},
	"youtube": {
		backgroundColor: "#EB3323",
		iconURL:         "https://www.mailjet.com/images/theme/v1/icons/ico-social/youtube.png",
	},
	"tumblr": {
		backgroundColor:        "#344356",
		iconURL:                "https://www.mailjet.com/images/theme/v1/icons/ico-social/tumblr.png",
		shareURLTemplate:       "https://www.tumblr.com/widgets/share/tool?canonicalUrl=[[URL]]",
		shareURLSkipSubstrings: []string{"tumblr.com/widgets/share"},
	},
	"github": {
		backgroundColor: "#000000",
		iconURL:         "https://www.mailjet.com/images/theme/v1/icons/ico-social/github.png",
	},
	"xing": {
		backgroundColor:        "#296366",
		iconURL:                "https://www.mailjet.com/images/theme/v1/icons/ico-social/xing.png",
		shareURLTemplate:       "https://www.xing.com/app/user?op=share&url=[[URL]]",
		shareURLSkipSubstrings: []string{"xing.com/app/user?op=share"},
	},
	"vimeo": {
		backgroundColor: "#53B4E7",
		iconURL:         "https://www.mailjet.com/images/theme/v1/icons/ico-social/vimeo.png",
	},
	"medium": {
		backgroundColor: "#000000",
		iconURL:         "https://www.mailjet.com/images/theme/v1/icons/ico-social/medium.png",
	},
	"soundcloud": {
		backgroundColor: "#EF7F31",
		iconURL:         "https://www.mailjet.com/images/theme/v1/icons/ico-social/soundcloud.png",
	},
	"dribbble": {
		backgroundColor: "#D95988",
		iconURL:         "https://www.mailjet.com/images/theme/v1/icons/ico-social/dribbble.png",
	},
}

const shareURLPlaceholder = "[[URL]]"

// getSocialNetworkDefaults resolves MJML defaults for a given social element name.
func getSocialNetworkDefaults(name string) (socialNetworkDefaults, bool) {
	if name == "" {
		return socialNetworkDefaults{}, false
	}

	if defaults, exists := baseSocialNetworkDefaults[name]; exists {
		return defaults, true
	}

	if idx := strings.Index(name, "-"); idx != -1 {
		baseName := name[:idx]
		if defaults, exists := baseSocialNetworkDefaults[baseName]; exists {
			// Copy the struct to avoid mutating the base definition.
			resolved := defaults
			if strings.HasSuffix(name, "-noshare") && resolved.shareURLTemplate != "" {
				resolved.shareURLTemplate = shareURLPlaceholder
			}
			return resolved, true
		}
	}

	return socialNetworkDefaults{}, false
}

// Attributes that mj-social-element is allowed to inherit from its parent.
var socialElementInheritableAttributes = map[string]struct{}{
	"color":           {},
	"font-family":     {},
	"font-size":       {},
	"line-height":     {},
	"text-decoration": {},
	"border-radius":   {},
	"icon-size":       {},
	"font-weight":     {},
	"font-style":      {},
	"icon-height":     {},
	"icon-padding":    {},
	"inner-padding":   {},
	"text-padding":    {},
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
func (c *MJSocialComponent) Render(w io.StringWriter) error {
	hasTextContent := false
	for _, child := range c.Children {
		if elem, ok := child.(*MJSocialElementComponent); ok {
			if strings.TrimSpace(elem.Node.Text) != "" || len(elem.Node.Children) > 0 {
				hasTextContent = true
				break
			}
		}
	}

	if hasTextContent {
		c.getAttribute(constants.MJMLFontFamily)
	}

	padding := c.getAttribute(constants.MJMLPadding)
	align := c.getAttribute(constants.MJMLAlign)
	mode := c.getAttribute(constants.MJMLMode)

	// Wrap in table row (required when inside column tbody)
	if _, err := w.WriteString("<tr>"); err != nil {
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

	// Handle individual padding properties - check all sides for MRML compatibility
	if paddingTop := c.Node.GetAttribute(constants.MJMLPaddingTop); paddingTop != "" {
		td.AddStyle(constants.CSSPaddingTop, paddingTop)
	}
	if paddingRight := c.Node.GetAttribute(constants.MJMLPaddingRight); paddingRight != "" {
		td.AddStyle(constants.CSSPaddingRight, paddingRight)
	}
	if paddingBottom := c.Node.GetAttribute(constants.MJMLPaddingBottom); paddingBottom != "" {
		td.AddStyle(constants.CSSPaddingBottom, paddingBottom)
	}
	if paddingLeft := c.Node.GetAttribute(constants.MJMLPaddingLeft); paddingLeft != "" {
		td.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}

	td.AddStyle("word-break", "break-word")

	if err := td.RenderOpen(w); err != nil {
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

		if err := table.RenderOpen(w); err != nil {
			return err
		}
		if _, err := w.WriteString("<tbody>"); err != nil {
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

		if _, err := w.WriteString("</tbody>"); err != nil {
			return err
		}
		if err := table.RenderClose(w); err != nil {
			return err
		}
	} else {
		// Horizontal mode (default): MSO conditional with inline tables
		msoAlign := align
		if msoAlign == "" {
			msoAlign = "center"
		}

		// Collect social elements first to coordinate MSO conditionals
		socialElements := make([]*MJSocialElementComponent, 0, len(c.Children))
		for _, child := range c.Children {
			if socialElement, ok := child.(*MJSocialElementComponent); ok {
				socialElement.SetContainerWidth(c.GetContainerWidth())
				socialElement.InheritFromParent(c)
				socialElements = append(socialElements, socialElement)
			}
		}

		if len(socialElements) > 0 {
			msoTable := fmt.Sprintf(
				"<!--[if mso | IE]><table align=\"%s\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\" ><tr><td><![endif]-->",
				msoAlign,
			)
			if _, err := w.WriteString(msoTable); err != nil {
				return err
			}
		} else {
			msoTable := fmt.Sprintf(
				"<!--[if mso | IE]><table align=\"%s\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\" ><tr><![endif]-->",
				msoAlign,
			)
			if _, err := w.WriteString(msoTable); err != nil {
				return err
			}
		}

		// Render social elements with coordinated MSO wrappers
		for i, socialElement := range socialElements {
			previousWrap := socialElement.SetMSOConditionalWrap(false)
			if err := socialElement.Render(w); err != nil {
				socialElement.SetMSOConditionalWrap(previousWrap)
				return err
			}
			socialElement.SetMSOConditionalWrap(previousWrap)

			if i < len(socialElements)-1 {
				if _, err := w.WriteString("<!--[if mso | IE]></td><td><![endif]-->"); err != nil {
					return err
				}
			}
		}

		// MSO conditional closing
		if len(socialElements) > 0 {
			if _, err := w.WriteString("<!--[if mso | IE]></td></tr></table><![endif]-->"); err != nil {
				return err
			}
		} else {
			if _, err := w.WriteString("<!--[if mso | IE]></tr></table><![endif]-->"); err != nil {
				return err
			}
		}
	}

	if err := td.RenderClose(w); err != nil {
		return err
	}

	// Close table row
	if _, err := w.WriteString("</tr>"); err != nil {
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
	wrapMSO      bool               // Whether to wrap output with MSO conditionals
}

// NewMJSocialElementComponent creates a new mj-social-element component
func NewMJSocialElementComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJSocialElementComponent {
	return &MJSocialElementComponent{
		BaseComponent: NewBaseComponent(node, opts),
		wrapMSO:       true,
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
		if defaults, ok := getSocialNetworkDefaults(c.Node.GetAttribute("name")); ok {
			return defaults.iconURL
		}
		return ""
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

	// 2. Check mj-class definitions for the element
	if classValue := c.GetClassAttribute(name); classValue != "" {
		if name == constants.MJMLFontFamily {
			c.TrackFontFamily(classValue)
		}
		return classValue
	}

	// 3. Check parent mj-social for inheritable attributes
	if c.parentSocial != nil {
		if _, inheritable := socialElementInheritableAttributes[name]; inheritable {
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
				if name == constants.MJMLFontFamily {
					c.TrackFontFamily(parentValue)
				}
				return parentValue
			}
			// Then check parent's resolved attribute (includes global attributes)
			if parentResolved := c.parentSocial.getAttribute(name); parentResolved != "" {
				debug.DebugLogWithData(
					"social-attr",
					"parent-resolved",
					"Using parent resolved attribute",
					map[string]interface{}{
						"attr":    name,
						"value":   parentResolved,
						"element": c.Node.GetAttribute("name"),
					},
				)
				if name == constants.MJMLFontFamily {
					c.TrackFontFamily(parentResolved)
				}
				return parentResolved
			}
		}
	}

	// 4. Check global attributes and component defaults
	if resolved := c.GetAttributeWithDefault(c, name); resolved != "" {
		return resolved
	}

	// 5. Check platform-specific defaults (for background-color)
	if name == constants.MJMLBackgroundColor {
		if defaults, ok := getSocialNetworkDefaults(c.Node.GetAttribute("name")); ok {
			return defaults.backgroundColor
		}
	}

	// 6. Fall back to component defaults
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

// SetMSOConditionalWrap enables or disables MSO wrapper output, returning the previous state.
func (c *MJSocialElementComponent) SetMSOConditionalWrap(enabled bool) bool {
	previous := c.wrapMSO
	c.wrapMSO = enabled
	return previous
}

// Render implements optimized Writer-based rendering for MJSocialElementComponent
func (c *MJSocialElementComponent) Render(w io.StringWriter) error {
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
	nameAttr := c.Node.GetAttribute("name")
	if href != "" {
		if defaults, ok := getSocialNetworkDefaults(nameAttr); ok && defaults.shareURLTemplate != "" {
			hrefLower := strings.ToLower(href)
			skipShare := false
			for _, pattern := range defaults.shareURLSkipSubstrings {
				if strings.Contains(hrefLower, pattern) {
					skipShare = true
					break
				}
			}

			if !skipShare {
				template := defaults.shareURLTemplate
				if strings.Contains(template, shareURLPlaceholder) {
					href = strings.ReplaceAll(template, shareURLPlaceholder, href)
				} else {
					href = template
				}
			}
		}
	}
	// Note: Only generate default URLs when href is explicitly provided (even if empty like "#")
	// Don't add default URLs when no href attribute exists - those are text-only social elements
	target := c.getAttribute("target")
	backgroundColor := c.getAttribute("background-color")
	borderRadius := c.getAttribute("border-radius")

	// Skip rendering if no src provided
	if src == "" {
		return nil
	}

	if c.verticalMode {
		// Vertical mode: render as table row without MSO conditionals
		if _, err := w.WriteString("<tr>"); err != nil {
			return err
		}

		// Icon cell
		iconTd := html.NewHTMLTag("td").
			AddStyle("padding", padding).
			AddStyle("vertical-align", "middle")

		if err := iconTd.RenderOpen(w); err != nil {
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

		if err := innerTable.RenderOpen(w); err != nil {
			return err
		}
		if _, err := w.WriteString("<tbody><tr>"); err != nil {
			return err
		}

		// Icon cell
		iconInnerTd := html.NewHTMLTag("td").
			AddStyle("font-size", "0").
			AddStyle("height", iconHeight).
			AddStyle("vertical-align", "middle").
			AddStyle("width", iconSize)

		if err := iconInnerTd.RenderOpen(w); err != nil {
			return err
		}

		// Image without link in vertical mode (as per MRML output)
		heightAttr := stripPxSuffix(iconHeight)
		widthAttr := stripPxSuffix(iconSize)

		img := html.NewHTMLTag("img")
		img.AddAttribute("alt", alt)
		img.AddAttribute("height", heightAttr).
			AddAttribute("src", src).
			AddAttribute("width", widthAttr).
			AddStyle("border-radius", borderRadius).
			AddStyle("display", "block")

		if err := img.RenderVoid(w); err != nil {
			return err
		}

		if err := iconInnerTd.RenderClose(w); err != nil {
			return err
		}
		if _, err := w.WriteString("</tr></tbody>"); err != nil {
			return err
		}
		if err := innerTable.RenderClose(w); err != nil {
			return err
		}
		if err := iconTd.RenderClose(w); err != nil {
			return err
		}

		// Text content cell
		textContent := c.Node.Text
		if textContent != "" {
			textTd := html.NewHTMLTag("td").
				AddStyle("vertical-align", "middle").
				AddStyle("padding", c.getAttribute("text-padding"))

			if err := textTd.RenderOpen(w); err != nil {
				return err
			}

			// Text content with span (no link in vertical mode as per MRML)
			textSpan := html.NewHTMLTag("span").
				AddStyle("color", c.getAttribute("color")).
				AddStyle("font-size", c.getAttribute("font-size")).
				AddStyle("font-family", c.getAttribute("font-family")).
				AddStyle("line-height", c.getAttribute("line-height")).
				AddStyle("text-decoration", c.getAttribute("text-decoration"))

			if err := textSpan.RenderOpen(w); err != nil {
				return err
			}
			if _, err := w.WriteString(textContent); err != nil {
				return err
			}
			if err := textSpan.RenderClose(w); err != nil {
				return err
			}
			if err := textTd.RenderClose(w); err != nil {
				return err
			}
		}

		if _, err := w.WriteString("</tr>"); err != nil {
			return err
		}

		return nil
	}

	// Horizontal mode: MSO conditional for individual social element
	if c.wrapMSO {
		if _, err := w.WriteString("<!--[if mso | IE]><td><![endif]-->"); err != nil {
			return err
		}
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

	if err := outerTable.RenderOpen(w); err != nil {
		return err
	}

	// Add CSS class to tr if specified on individual social element
	cssClass := c.Node.GetAttribute("css-class")
	if cssClass != "" {
		trTag := fmt.Sprintf("<tbody><tr class=\"%s\">", cssClass)
		if _, err := w.WriteString(trTag); err != nil {
			return err
		}
	} else {
		if _, err := w.WriteString("<tbody><tr>"); err != nil {
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

	if err := paddingTd.RenderOpen(w); err != nil {
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

	if err := innerTable.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody><tr>"); err != nil {
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

	if err := iconTd.RenderOpen(w); err != nil {
		return err
	}

	// Image with optional link - remove "px" suffix from dimensions for HTML attributes
	heightAttr := stripPxSuffix(iconHeight)
	widthAttr := stripPxSuffix(iconSize)

	img := html.NewHTMLTag("img")
	img.AddAttribute("alt", alt)

	img.AddAttribute("height", heightAttr).
		AddAttribute("src", src).
		AddAttribute("width", widthAttr)

	// Add title attribute if specified
	title := c.Node.GetAttribute("title")
	if title != "" {
		img.AddAttribute("title", title)
	}

	img.AddStyle("border-radius", borderRadius).
		AddStyle("display", "block")

	if href != "" {
		link := html.NewHTMLTag("a").
			AddAttribute("href", href).
			AddAttribute("target", target)
		if err := link.RenderOpen(w); err != nil {
			return err
		}
		if err := img.RenderVoid(w); err != nil {
			return err
		}
		if err := link.RenderClose(w); err != nil {
			return err
		}
	} else {
		if err := img.RenderVoid(w); err != nil {
			return err
		}
	}

	if err := iconTd.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr></tbody>"); err != nil {
		return err
	}
	if err := innerTable.RenderClose(w); err != nil {
		return err
	}
	if err := paddingTd.RenderClose(w); err != nil {
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

		if err := textTd.RenderOpen(w); err != nil {
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

		if err := textElement.RenderOpen(w); err != nil {
			return err
		}
		if _, err := w.WriteString(textContent); err != nil {
			return err
		}
		if err := textElement.RenderClose(w); err != nil {
			return err
		}
		if err := textTd.RenderClose(w); err != nil {
			return err
		}
	}

	if _, err := w.WriteString("</tr></tbody>"); err != nil {
		return err
	}
	if err := outerTable.RenderClose(w); err != nil {
		return err
	}

	// Close MSO conditional
	if c.wrapMSO {
		if _, err := w.WriteString("<!--[if mso | IE]></td><![endif]-->"); err != nil {
			return err
		}
	}

	return nil
}

func (c *MJSocialElementComponent) GetTagName() string {
	return "mj-social-element"
}
