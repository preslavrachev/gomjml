package mjml

import (
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/components"
	"github.com/preslavrachev/gomjml/mjml/globals"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/mjml/styles"
	"github.com/preslavrachev/gomjml/parser"
)

// Type alias for convenience
type MJMLNode = parser.MJMLNode

// ParseMJML re-exports the parser function for convenience
var ParseMJML = parser.ParseMJML

// RenderOpts is an alias for convenience
type RenderOpts = options.RenderOpts

// RenderOption is a functional option for configuring MJML rendering
type RenderOption func(*RenderOpts)

// WithDebugTags enables or disables debug tag inclusion in the rendered output
func WithDebugTags(enabled bool) RenderOption {
	return func(opts *RenderOpts) {
		opts.DebugTags = enabled
	}
}

// RenderResult contains both the rendered HTML and the MJML AST
type RenderResult struct {
	HTML string
	AST  *MJMLNode
}

// RenderWithAST provides the internal MJML to HTML conversion function that returns both HTML and AST
func RenderWithAST(mjmlContent string, opts ...RenderOption) (*RenderResult, error) {
	// Apply render options
	renderOpts := &RenderOpts{}
	for _, opt := range opts {
		opt(renderOpts)
	}
	// Parse MJML using the parser package
	ast, err := ParseMJML(mjmlContent)
	if err != nil {
		return nil, err
	}

	// Initialize global attributes
	globalAttrs := globals.NewGlobalAttributes()

	// Process global attributes from head if it exists
	if headNode := ast.FindFirstChild("mj-head"); headNode != nil {
		globalAttrs.ProcessAttributesFromHead(headNode)
	}

	// Set the global attributes instance
	globals.SetGlobalAttributes(globalAttrs)

	// Create component tree
	component, err := CreateComponent(ast, renderOpts)
	if err != nil {
		return nil, err
	}

	// Render to HTML
	html, err := component.Render()
	if err != nil {
		return nil, err
	}

	return &RenderResult{
		HTML: html,
		AST:  ast,
	}, nil
}

// Render provides the main MJML to HTML conversion function
func Render(mjmlContent string, opts ...RenderOption) (string, error) {
	result, err := RenderWithAST(mjmlContent, opts...)
	if err != nil {
		return "", err
	}
	return result.HTML, nil
}

// RenderFromAST renders HTML from a pre-parsed AST
func RenderFromAST(ast *MJMLNode, opts ...RenderOption) (string, error) {
	// Apply render options
	renderOpts := &RenderOpts{}
	for _, opt := range opts {
		opt(renderOpts)
	}

	component, err := CreateComponent(ast, renderOpts)
	if err != nil {
		return "", err
	}

	return component.Render()
}

// NewFromAST creates a component from a pre-parsed AST (alias for CreateComponent)
func NewFromAST(ast *MJMLNode, opts ...RenderOption) (Component, error) {
	// Apply render options
	renderOpts := &RenderOpts{}
	for _, opt := range opts {
		opt(renderOpts)
	}

	return CreateComponent(ast, renderOpts)
}

// MJMLComponent represents the root MJML component
type MJMLComponent struct {
	*components.BaseComponent
	Head           *components.MJHeadComponent
	Body           *components.MJBodyComponent
	mobileCSSAdded bool                   // Track if mobile CSS has been added
	columnClasses  map[string]styles.Size // Track column classes used in the document
}

// RequestMobileCSS allows components to request mobile CSS to be added
func (c *MJMLComponent) RequestMobileCSS() {
	c.mobileCSSAdded = true
}

// hasCustomGlobalFonts checks if global attributes specify custom fonts
func (c *MJMLComponent) hasCustomGlobalFonts() bool {
	// Check if global attributes have specified font-family
	globalFontFamily := globals.GetGlobalAttribute("mj-all", "font-family")
	if globalFontFamily != "" && globalFontFamily != "Ubuntu, Helvetica, Arial, sans-serif" {
		return true
	}

	// Check if any text components have global font-family defined
	textFontFamily := globals.GetGlobalAttribute("mj-text", "font-family")
	if textFontFamily != "" && textFontFamily != "Ubuntu, Helvetica, Arial, sans-serif" {
		return true
	}

	return false
}

// prepareBodySiblings recursively sets up sibling relationships without rendering HTML
func (c *MJMLComponent) prepareBodySiblings(comp Component) {
	// Check specific component types that need to set up their children's sibling relationships
	switch v := comp.(type) {
	case *components.MJBodyComponent:
		// Body components set up their section children
		siblings := len(v.Children)
		rawSiblings := 0
		for _, child := range v.Children {
			if child.GetTagName() == "mj-raw" {
				rawSiblings++
			}
		}
		for _, child := range v.Children {
			child.SetContainerWidth(v.GetEffectiveWidth())
			child.SetSiblings(siblings)
			child.SetRawSiblings(rawSiblings)
			c.prepareBodySiblings(child)
		}
	case *components.MJSectionComponent:
		// Section components set up their column children
		siblings := len(v.Children)
		rawSiblings := 0
		for _, child := range v.Children {
			if child.GetTagName() == "mj-raw" {
				rawSiblings++
			}
		}
		for _, child := range v.Children {
			child.SetContainerWidth(v.GetEffectiveWidth())
			child.SetSiblings(siblings)
			child.SetRawSiblings(rawSiblings)
			c.prepareBodySiblings(child)
		}
	case *components.MJColumnComponent:
		// Column components set up their content children
		for _, child := range v.Children {
			child.SetContainerWidth(v.GetEffectiveWidth())
			c.prepareBodySiblings(child)
		}
	case *components.MJWrapperComponent:
		// Wrapper components set up their children
		for _, child := range v.Children {
			child.SetContainerWidth(v.GetEffectiveWidth())
			c.prepareBodySiblings(child)
		}
	case *components.MJGroupComponent:
		// Group components set up their children and distribute width equally
		columnCount := 0
		for _, child := range v.Children {
			if _, ok := child.(*components.MJColumnComponent); ok {
				columnCount++
			}
		}

		if columnCount > 0 {
			percentagePerColumn := 100.0 / float64(columnCount)

			for _, child := range v.Children {
				child.SetContainerWidth(v.GetEffectiveWidth())

				// Set width attributes on columns like the group's Render() method does
				if columnComp, ok := child.(*components.MJColumnComponent); ok {
					if columnComp.GetAttribute("width") == nil {
						percentageWidth := fmt.Sprintf("%.15f%%", percentagePerColumn)
						percentageWidth = strings.TrimRight(percentageWidth, "0")
						percentageWidth = strings.TrimRight(percentageWidth, ".")
						if !strings.HasSuffix(percentageWidth, "%") {
							percentageWidth += "%"
						}
						columnComp.Attrs["width"] = percentageWidth
					}
				}

				c.prepareBodySiblings(child)
			}
		} else {
			for _, child := range v.Children {
				child.SetContainerWidth(v.GetEffectiveWidth())
				c.prepareBodySiblings(child)
			}
		}
	}
}

// collectColumnClasses recursively collects all column classes used in the document
func (c *MJMLComponent) collectColumnClasses() {
	if c.columnClasses == nil {
		c.columnClasses = make(map[string]styles.Size)
	}
	if c.Body != nil {
		c.collectColumnClassesFromComponent(c.Body)
	}
}

// collectColumnClassesFromComponent recursively collects column classes from a component
func (c *MJMLComponent) collectColumnClassesFromComponent(comp Component) {
	// Check if this is a column component
	if columnComp, ok := comp.(*components.MJColumnComponent); ok {
		className, size := columnComp.GetColumnClass()
		c.columnClasses[className] = size
	}

	// Check specific component types that have children
	switch v := comp.(type) {
	case *components.MJBodyComponent:
		for _, child := range v.Children {
			c.collectColumnClassesFromComponent(child)
		}
	case *components.MJSectionComponent:
		for _, child := range v.Children {
			c.collectColumnClassesFromComponent(child)
		}
	case *components.MJColumnComponent:
		for _, child := range v.Children {
			c.collectColumnClassesFromComponent(child)
		}
	case *components.MJWrapperComponent:
		for _, child := range v.Children {
			c.collectColumnClassesFromComponent(child)
		}
	case *components.MJGroupComponent:
		// Register group's CSS class based on its width attribute
		groupWidth := v.GetAttribute("width")
		if groupWidth != nil && strings.HasSuffix(*groupWidth, "px") {
			// Parse pixel width and register pixel-based class
			var widthPx int
			fmt.Sscanf(*groupWidth, "%dpx", &widthPx)
			className := fmt.Sprintf("mj-column-px-%d", widthPx)
			c.columnClasses[className] = styles.NewPixelSize(float64(widthPx))
		} else {
			// Default to percentage-based class
			c.columnClasses["mj-column-per-100"] = styles.NewPercentSize(100)
		}

		// Also recurse into children to collect column classes
		for _, child := range v.Children {
			c.collectColumnClassesFromComponent(child)
		}
	}
}

// generateResponsiveCSS generates responsive CSS for collected column classes
func (c *MJMLComponent) generateResponsiveCSS() string {
	var css strings.Builder

	// Standard responsive media query
	css.WriteString(`<style type="text/css">@media only screen and (min-width:480px) { `)
	for className, size := range c.columnClasses {
		// Include both percentage and pixel-based classes
		css.WriteString(`.`)
		css.WriteString(className)
		css.WriteString(` { width:`)
		css.WriteString(size.String())
		css.WriteString(` !important; max-width:`)
		css.WriteString(size.String())
		css.WriteString(`; } `)
	}
	css.WriteString(` }</style>`)

	// Mozilla-specific responsive media query
	css.WriteString(`<style media="screen and (min-width:480px)">`)
	for className, size := range c.columnClasses {
		// Include both percentage and pixel-based classes
		css.WriteString(`.moz-text-html .`)
		css.WriteString(className)
		css.WriteString(` { width:`)
		css.WriteString(size.String())
		css.WriteString(` !important; max-width:`)
		css.WriteString(size.String())
		css.WriteString(`; } `)
	}
	css.WriteString(`</style>`)

	return css.String()
}

// generateCustomStyles generates the final mj-style content tag (MRML lines 240-244)
func (c *MJMLComponent) generateCustomStyles() string {
	var content strings.Builder

	// Collect all mj-style content (MRML mj_style_iter equivalent)
	if c.Head != nil {
		for _, child := range c.Head.Children {
			if styleComp, ok := child.(*components.MJStyleComponent); ok {
				text := strings.TrimSpace(styleComp.Node.Text)
				if text != "" {
					content.WriteString(text)
				}
			}
		}
	}

	// Always generate the style tag (MRML always includes this, even if empty)
	return fmt.Sprintf(`<style type="text/css">%s</style>`, content.String())
}

// hasMobileCSSComponents recursively checks if any component needs mobile CSS
func (c *MJMLComponent) hasMobileCSSComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.checkComponentForMobileCSS(c.Body)
}

// hasTextComponents checks if the document contains any text-based components that need fonts
func (c *MJMLComponent) hasTextComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.hasTextComponentsRecursive(c.Body)
}

// hasTextComponentsRecursive recursively checks for text components
func (c *MJMLComponent) hasTextComponentsRecursive(component Component) bool {
	// Check if this component is a text component
	switch component.(type) {
	case *components.MJTextComponent, *components.MJButtonComponent:
		return true
	}

	// Check specific component types that have children
	switch v := component.(type) {
	case *components.MJBodyComponent:
		for _, child := range v.Children {
			if c.hasTextComponentsRecursive(child) {
				return true
			}
		}
	case *components.MJSectionComponent:
		for _, child := range v.Children {
			if c.hasTextComponentsRecursive(child) {
				return true
			}
		}
	case *components.MJColumnComponent:
		for _, child := range v.Children {
			if c.hasTextComponentsRecursive(child) {
				return true
			}
		}
	case *components.MJWrapperComponent:
		for _, child := range v.Children {
			if c.hasTextComponentsRecursive(child) {
				return true
			}
		}
	case *components.MJGroupComponent:
		for _, child := range v.Children {
			if c.hasTextComponentsRecursive(child) {
				return true
			}
		}
	}
	return false
}

// checkComponentForMobileCSS recursively checks a component and its children
func (c *MJMLComponent) checkComponentForMobileCSS(comp Component) bool {
	// Check if this component needs mobile CSS (currently only mj-image)
	if comp.GetTagName() == "mj-image" {
		return true
	}

	// Check specific component types that have children
	switch v := comp.(type) {
	case *components.MJBodyComponent:
		for _, child := range v.Children {
			if c.checkComponentForMobileCSS(child) {
				return true
			}
		}
	case *components.MJSectionComponent:
		for _, child := range v.Children {
			if c.checkComponentForMobileCSS(child) {
				return true
			}
		}
	case *components.MJColumnComponent:
		for _, child := range v.Children {
			if c.checkComponentForMobileCSS(child) {
				return true
			}
		}
	case *components.MJWrapperComponent:
		for _, child := range v.Children {
			if c.checkComponentForMobileCSS(child) {
				return true
			}
		}
	case *components.MJGroupComponent:
		for _, child := range v.Children {
			if c.checkComponentForMobileCSS(child) {
				return true
			}
		}
	}

	return false
}

// Render implements the Component interface for MJMLComponent
func (c *MJMLComponent) Render() (string, error) {
	var html strings.Builder

	// First, prepare the body to establish sibling relationships without full rendering
	if c.Body != nil {
		c.prepareBodySiblings(c.Body)
	}

	// Now collect column classes after sibling relationships are established
	c.collectColumnClasses()

	// DOCTYPE and HTML opening
	html.WriteString(
		`<!doctype html><html xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">`,
	)

	// Head section - extract title from head components
	title := ""
	customFonts := make([]string, 0)

	if c.Head != nil {
		for _, child := range c.Head.Children {
			if titleComp, ok := child.(*components.MJTitleComponent); ok {
				title = titleComp.Node.Text
			}
			if fontComp, ok := child.(*components.MJFontComponent); ok {
				// Helper function to get attribute with default
				getAttr := func(name string) string {
					if attr := fontComp.GetAttribute(name); attr != nil {
						return *attr
					}
					return fontComp.GetDefaultAttribute(name)
				}

				fontName := getAttr("name")
				fontHref := getAttr("href")
				if fontName != "" && fontHref != "" {
					customFonts = append(customFonts, fontHref)
				}
			}
		}
	}

	html.WriteString(`<head><title>` + title + `</title>`)
	html.WriteString(`<!--[if !mso]><!--><meta http-equiv="X-UA-Compatible" content="IE=edge"><!--<![endif]-->`)
	html.WriteString(`<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">`)
	html.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1">`)

	// Base CSS
	html.WriteString("\n<style type=\"text/css\">\n")
	html.WriteString("#outlook a { padding: 0; }\n")
	html.WriteString("body { margin: 0; padding: 0; -webkit-text-size-adjust: 100%; -ms-text-size-adjust: 100%; }\n")
	html.WriteString("table, td { border-collapse: collapse; mso-table-lspace: 0pt; mso-table-rspace: 0pt; }\n")
	html.WriteString(
		"img { border: 0; height: auto; line-height: 100%; outline: none; text-decoration: none; -ms-interpolation-mode: bicubic; }\n",
	)
	html.WriteString("p { display: block; margin: 13px 0; }\n")
	html.WriteString("</style>\n")

	// MSO conditionals
	html.WriteString(
		"<!--[if mso]>\n<noscript>\n<xml>\n<o:OfficeDocumentSettings>\n  <o:AllowPNG/>\n  <o:PixelsPerInch>96</o:PixelsPerInch>\n</o:OfficeDocumentSettings>\n</xml>\n</noscript>\n<![endif]-->\n",
	)
	html.WriteString(
		"<!--[if lte mso 11]>\n<style type=\"text/css\">\n.mj-outlook-group-fix { width:100% !important; }\n</style>\n<![endif]-->\n",
	)

	// Font imports - check if we need Ubuntu fallback or custom fonts
	if c.hasTextComponents() {
		if len(customFonts) > 0 {
			for _, fontHref := range customFonts {
				html.WriteString(`<!--[if !mso]><!--><link href="` + fontHref + `" rel="stylesheet" type="text/css">`)
				html.WriteString(`<style type="text/css">@import url(` + fontHref + `);</style><!--<![endif]-->`)
			}
		} else if !c.hasCustomGlobalFonts() {
			// Only import Ubuntu if no global fonts are specified
			html.WriteString(`<!--[if !mso]><!--><link href="https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700" rel="stylesheet" type="text/css">`)
			html.WriteString(`<style type="text/css">@import url(https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700);</style><!--<![endif]-->`)
		}
	}

	// Dynamic responsive CSS based on collected column classes - only if we have columns
	if len(c.columnClasses) > 0 {
		html.WriteString(c.generateResponsiveCSS())
	}

	// Mobile CSS - add only if components need it (following MRML pattern)
	if c.hasMobileCSSComponents() {
		html.WriteString(`<style type="text/css">@media only screen and (max-width:479px) {
                table.mj-full-width-mobile { width: 100% !important; }
                td.mj-full-width-mobile { width: auto !important; }
            }
            </style>`)
	}

	// Custom styles from mj-style components (MRML lines 240-244)
	html.WriteString(c.generateCustomStyles())

	html.WriteString(`</head>`)

	// Body with background-color support (matching MRML's get_body_tag)
	var bodyStyles []string

	// Always add word-spacing:normal to match MRML behavior
	bodyStyles = append(bodyStyles, "word-spacing:normal")

	if c.Body != nil {
		if bgColor := c.Body.GetAttribute("background-color"); bgColor != nil && *bgColor != "" {
			bodyStyles = append(bodyStyles, "background-color:"+*bgColor)
		}
	}

	if len(bodyStyles) > 0 {
		html.WriteString(`<body style="` + strings.Join(bodyStyles, ";") + `;">`)
	} else {
		html.WriteString(`<body>`)
	}

	// Add preview text from head components right after body tag
	if c.Head != nil {
		for _, child := range c.Head.Children {
			if previewComp, ok := child.(*components.MJPreviewComponent); ok {
				previewHTML, err := previewComp.Render()
				if err != nil {
					return "", err
				}
				html.WriteString(previewHTML)
			}
		}
	}

	if c.Body != nil {
		bodyHTML, err := c.Body.Render()
		if err != nil {
			return "", err
		}
		html.WriteString(bodyHTML)
	}
	html.WriteString(`</body></html>`)

	return html.String(), nil
}

func (c *MJMLComponent) GetTagName() string {
	return "mjml"
}
