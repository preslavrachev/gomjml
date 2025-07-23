package mjml

import (
	"strings"

	"github.com/preslavrachev/gomjml/mjml/components"
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
	Head *components.MJHeadComponent
	Body *components.MJBodyComponent
}

// Render implements the Component interface for MJMLComponent
func (c *MJMLComponent) Render() (string, error) {
	var html strings.Builder

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

	// Responsive CSS
	html.WriteString(
		`<style type="text/css">@media only screen and (min-width:480px) { .mj-column-per-100 { width:100% !important; max-width:100%; }  }</style>`,
	)
	html.WriteString(
		`<style media="screen and (min-width:480px)">.moz-text-html .mj-column-per-100 { width:100% !important; max-width:100%; } </style>`,
	)
	html.WriteString(`<style type="text/css"></style>`)

	html.WriteString(`</head>`)

	// Body
	html.WriteString(`<body style="word-spacing:normal;">`)
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
