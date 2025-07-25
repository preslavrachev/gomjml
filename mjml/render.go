package mjml

import (
	"strings"

	"github.com/preslavrachev/gomjml/mjml/components"
	"github.com/preslavrachev/gomjml/mjml/styles"
	"github.com/preslavrachev/gomjml/parser"
)

// Type alias for convenience
type MJMLNode = parser.MJMLNode

// ParseMJML re-exports the parser function for convenience
var ParseMJML = parser.ParseMJML

// Render provides the main MJML to HTML conversion function
func Render(mjmlContent string) (string, error) {
	// Parse MJML using the parser package
	ast, err := ParseMJML(mjmlContent)
	if err != nil {
		return "", err
	}

	// Create component tree
	component, err := CreateComponent(ast)
	if err != nil {
		return "", err
	}

	// Render to HTML
	return component.Render()
}

// RenderFromAST renders HTML from a pre-parsed AST
func RenderFromAST(ast *MJMLNode) (string, error) {
	component, err := CreateComponent(ast)
	if err != nil {
		return "", err
	}

	return component.Render()
}

// NewFromAST creates a component from a pre-parsed AST (alias for CreateComponent)
func NewFromAST(ast *MJMLNode) (Component, error) {
	return CreateComponent(ast)
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
		if size.IsPercent() {
			css.WriteString(`.`)
			css.WriteString(className)
			css.WriteString(` { width:`)
			css.WriteString(size.String())
			css.WriteString(` !important; max-width:`)
			css.WriteString(size.String())
			css.WriteString(`; } `)
		}
	}
	css.WriteString(` }</style>`)

	// Mozilla-specific responsive media query
	css.WriteString(`<style media="screen and (min-width:480px)">.moz-text-html `)
	for className, size := range c.columnClasses {
		if size.IsPercent() {
			css.WriteString(`.`)
			css.WriteString(className)
			css.WriteString(` { width:`)
			css.WriteString(size.String())
			css.WriteString(` !important; max-width:`)
			css.WriteString(size.String())
			css.WriteString(`; } `)
		}
	}
	css.WriteString(`</style>`)

	return css.String()
}

// generateCustomStyles collects and renders mj-style content (MRML mj_style_iter)
func (c *MJMLComponent) generateCustomStyles() string {
	var css strings.Builder
	var hasContent bool
	
	// Collect styles from mj-style components in head
	if c.Head != nil {
		for _, child := range c.Head.Children {
			if styleComp, ok := child.(*components.MJStyleComponent); ok {
				// Get the text content from the style component
				content := strings.TrimSpace(styleComp.Node.Text)
				if content != "" {
					if !hasContent {
						css.WriteString(`<style type="text/css">`)
						hasContent = true
					}
					css.WriteString(content)
				}
			}
		}
		if hasContent {
			css.WriteString(`</style>`)
		}
	}
	
	// Always add empty style tag (MRML always includes this)
	css.WriteString(`<style type="text/css"></style>`)
	
	return css.String()
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
	switch comp := component.(type) {
	case *components.MJTextComponent, *components.MJButtonComponent:
		return true
	case interface{ GetChildren() []Component }:
		for _, child := range comp.GetChildren() {
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
	}

	return false
}

// Render implements the Component interface for MJMLComponent
func (c *MJMLComponent) Render() (string, error) {
	var html strings.Builder

	// Collect column classes before rendering to generate dynamic responsive CSS
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

	// Font imports - use custom fonts if provided, otherwise default
	if len(customFonts) > 0 {
		for _, fontHref := range customFonts {
			html.WriteString(`<!--[if !mso]><!--><link href="` + fontHref + `" rel="stylesheet" type="text/css">`)
			html.WriteString(`<style type="text/css">@import url(` + fontHref + `);</style><!--<![endif]-->`)
		}
	} else {
		html.WriteString(`<!--[if !mso]><!--><link href="https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700" rel="stylesheet" type="text/css">`)
		html.WriteString(`<style type="text/css">@import url(https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700);</style><!--<![endif]-->`)
	}

	// Dynamic responsive CSS based on collected column classes
	html.WriteString(c.generateResponsiveCSS())

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
	bodyStyle := "word-spacing:normal;"
	if c.Body != nil {
		if bgColor := c.Body.GetAttribute("background-color"); bgColor != nil && *bgColor != "" {
			bodyStyle += "background-color:" + *bgColor + ";"
		}
	}
	html.WriteString(`<body style="` + bodyStyle + `">`)
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
