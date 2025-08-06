package mjml

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/preslavrachev/gomjml/mjml/components"
	"github.com/preslavrachev/gomjml/mjml/debug"
	"github.com/preslavrachev/gomjml/mjml/fonts"
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

// calculateOptimalBufferSize determines the optimal buffer size based on template complexity
func calculateOptimalBufferSize(mjmlContent string) int {
	mjmlSize := len(mjmlContent)
	componentCount := strings.Count(mjmlContent, "<mj-")

	// Prevent division by zero for empty MJML content
	if mjmlSize == 0 {
		// Return a reasonable default buffer size for empty input
		return 1024
	}

	// Calculate component density (components per 1000 characters)
	complexity := float64(componentCount) / float64(mjmlSize) * 1000

	if complexity > 10 {
		// Very dense template - needs more buffer per component
		return mjmlSize*5 + componentCount*180
	} else if complexity > 5 {
		// Medium density - balanced approach
		return mjmlSize*4 + componentCount*140
	} else {
		// Light template - more conservative
		return mjmlSize*3 + componentCount*100
	}
}

// WithDebugTags enables or disables debug tag inclusion in the rendered output
func WithDebugTags(enabled bool) RenderOption {
	return func(opts *RenderOpts) {
		opts.DebugTags = enabled
	}
}

// WithOutputFormat sets the output format for rendering
func WithOutputFormat(format options.OutputFormat) RenderOption {
	return func(opts *RenderOpts) {
		opts.OutputFormat = format
	}
}

// RenderResult contains both the rendered HTML and the MJML AST
type RenderResult struct {
	Result string
	AST    *MJMLNode
}

// RenderWithAST provides the internal MJML to HTML conversion function that returns both HTML and AST
func RenderWithAST(mjmlContent string, opts ...RenderOption) (*RenderResult, error) {
	startTime := time.Now()
	debug.DebugLogWithData("mjml", "render-start", "Starting MJML rendering", map[string]interface{}{
		"content_length": len(mjmlContent),
		"has_debug":      len(opts) > 0,
	})

	// Apply render options
	renderOpts := &RenderOpts{
		FontTracker: options.NewFontTracker(),
	}
	for _, opt := range opts {
		opt(renderOpts)
	}

	// Parse MJML using the parser package
	debug.DebugLog("mjml", "parse-start", "Starting MJML parsing")
	ast, err := ParseMJML(mjmlContent)
	if err != nil {
		debug.DebugLogError("mjml", "parse-error", "Failed to parse MJML", err)
		return nil, err
	}
	debug.DebugLog("mjml", "parse-complete", "MJML parsing completed successfully")

	// Initialize global attributes
	globalAttrs := globals.NewGlobalAttributes()

	// Process global attributes from head if it exists
	if headNode := ast.FindFirstChild("mj-head"); headNode != nil {
		globalAttrs.ProcessAttributesFromHead(headNode)
	}

	// Set the global attributes instance
	globals.SetGlobalAttributes(globalAttrs)

	// Create component tree
	debug.DebugLog("mjml", "component-tree-start", "Creating component tree from AST")
	component, err := CreateComponent(ast, renderOpts)
	if err != nil {
		debug.DebugLogError("mjml", "component-tree-error", "Failed to create component tree", err)
		return nil, err
	}
	debug.DebugLog("mjml", "component-tree-complete", "Component tree created successfully")

	// Render to HTML with optimized pre-allocation based on template complexity
	bufferSize := calculateOptimalBufferSize(mjmlContent)
	debug.DebugLogWithData("mjml", "render-html-start", "Starting HTML rendering", map[string]interface{}{
		"buffer_size": bufferSize,
	})
	var html strings.Builder
	html.Grow(bufferSize) // Pre-allocate with complexity-aware sizing

	renderStart := time.Now()
	err = component.RenderHTML(&html)
	if err != nil {
		debug.DebugLogError("mjml", "render-html-error", "Failed to render HTML", err)
		return nil, err
	}
	renderDuration := time.Since(renderStart).Milliseconds()

	htmlOutput := html.String()
	totalDuration := time.Since(startTime).Milliseconds()

	debug.DebugLogWithData("mjml", "render-complete", "MJML rendering completed", map[string]interface{}{
		"output_length":    len(htmlOutput),
		"render_time_ms":   renderDuration,
		"total_time_ms":    totalDuration,
		"expansion_factor": float64(len(htmlOutput)) / float64(len(mjmlContent)),
	})

	return &RenderResult{
		Result: htmlOutput,
		AST:    ast,
	}, nil
}

// RenderHTML provides the main MJML to HTML conversion function
func RenderHTML(mjmlContent string, opts ...RenderOption) (string, error) {
	result, err := RenderWithAST(mjmlContent, opts...)
	if err != nil {
		return "", err
	}
	return result.Result, nil
}

// RenderFromAST renders from a pre-parsed AST to the specified output format
func RenderFromAST(ast *MJMLNode, opts ...RenderOption) (string, error) {
	// Apply render options with FontTracker initialized
	renderOpts := &RenderOpts{
		FontTracker: options.NewFontTracker(),
	}
	for _, opt := range opts {
		opt(renderOpts)
	}

	component, err := CreateComponent(ast, renderOpts)
	if err != nil {
		return "", err
	}

	// Switch based on output format
	switch renderOpts.OutputFormat {
	case options.OutputHTML:
		return RenderComponentString(component)
	case options.OutputMJML:
		return RenderComponentMJMLString(component)
	default:
		return RenderComponentString(component) // Default to HTML
	}
}

// NewFromAST creates a component from a pre-parsed AST (alias for CreateComponent)
func NewFromAST(ast *MJMLNode, opts ...RenderOption) (Component, error) {
	// Apply render options
	renderOpts := &RenderOpts{
		FontTracker: options.NewFontTracker(),
	}
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
	if globalFontFamily != "" && globalFontFamily != fonts.DefaultFontStack {
		return true
	}

	// Check if any text components have global font-family defined
	textFontFamily := globals.GetGlobalAttribute("mj-text", "font-family")
	if textFontFamily != "" && textFontFamily != fonts.DefaultFontStack {
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

				// Set width attributes on columns like the group's RenderHTML() method does
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

// generateAccordionCSS generates the CSS styles needed for accordion functionality
func (c *MJMLComponent) generateAccordionCSS() string {
	return `<style type="text/css">noinput.mj-accordion-checkbox { display: block! important; }
@media yahoo, only screen and (min-width:0) {
  .mj-accordion-element { display:block; }
  input.mj-accordion-checkbox, .mj-accordion-less { display: none !important; }
  input.mj-accordion-checkbox+* .mj-accordion-title { cursor: pointer; touch-action: manipulation; -webkit-user-select: none; -moz-user-select: none; user-select: none; }
  input.mj-accordion-checkbox+* .mj-accordion-content { overflow: hidden; display: none; }
  input.mj-accordion-checkbox+* .mj-accordion-more { display: block !important; }
  input.mj-accordion-checkbox:checked+* .mj-accordion-content { display: block; }
  input.mj-accordion-checkbox:checked+* .mj-accordion-more { display: none !important; }
  input.mj-accordion-checkbox:checked+* .mj-accordion-less { display: block !important; }
}
.moz-text-html input.mj-accordion-checkbox+* .mj-accordion-title { cursor: auto; touch-action: auto; -webkit-user-select: auto; -moz-user-select: auto; user-select: auto; }
.moz-text-html input.mj-accordion-checkbox+* .mj-accordion-content { overflow: hidden; display: block; }
.moz-text-html input.mj-accordion-checkbox+* .mj-accordion-ico { display: none; }
@goodbye { @gmail }
</style>`
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

// hasSocialComponents checks if the MJML contains any social components
func (c *MJMLComponent) hasSocialComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.checkChildrenForCondition(c.Body, func(comp Component) bool {
		switch comp.GetTagName() {
		case "mj-social", "mj-social-element":
			return true
		}
		return false
	})
}

// hasButtonComponents checks if the MJML contains any button components
func (c *MJMLComponent) hasButtonComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.checkChildrenForCondition(c.Body, func(comp Component) bool {
		return comp.GetTagName() == "mj-button"
	})
}

// hasAccordionComponents checks if the MJML contains any accordion components
func (c *MJMLComponent) hasAccordionComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.checkChildrenForCondition(c.Body, func(comp Component) bool {
		switch comp.GetTagName() {
		case "mj-accordion", "mj-accordion-element", "mj-accordion-title", "mj-accordion-text":
			return true
		}
		return false
	})
}

// hasTextComponentsRecursive recursively checks for text components
func (c *MJMLComponent) hasTextComponentsRecursive(component Component) bool {
	// Check if this component is a text component
	switch component.(type) {
	case *components.MJTextComponent, *components.MJButtonComponent:
		return true
	}

	// Check specific component types that have children
	return c.checkChildrenForCondition(component, c.hasTextComponentsRecursive)
}

// checkComponentForMobileCSS recursively checks a component and its children
func (c *MJMLComponent) checkComponentForMobileCSS(comp Component) bool {
	// Check if this component needs mobile CSS (currently only mj-image)
	if comp.GetTagName() == "mj-image" {
		return true
	}

	// Check specific component types that have children
	return c.checkChildrenForCondition(comp, c.checkComponentForMobileCSS)
}

// checkChildrenForCondition is a helper function that checks if any children of a component meet a condition
func (c *MJMLComponent) checkChildrenForCondition(component Component, condition func(Component) bool) bool {
	// Check all children recursively
	switch v := component.(type) {
	case *components.MJBodyComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJSectionComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJColumnComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJWrapperComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJGroupComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJSocialComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJAccordionComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	}
	return false
}

func (c *MJMLComponent) GetTagName() string {
	return "mjml"
}

// RenderHTML implements optimized Writer-based rendering for MJMLComponent
func (c *MJMLComponent) RenderHTML(w io.StringWriter) error {
	debug.DebugLog("mjml-root", "render-start", "Starting root MJML component rendering")

	// First, prepare the body to establish sibling relationships without full rendering
	debug.DebugLog("mjml-root", "prepare-siblings", "Preparing body sibling relationships")
	if c.Body != nil {
		c.prepareBodySiblings(c.Body)
	}

	// Now collect column classes after sibling relationships are established
	debug.DebugLog("mjml-root", "collect-column-classes", "Collecting column classes for responsive CSS")
	c.collectColumnClasses()
	debug.DebugLogWithData("mjml-root", "column-classes-collected", "Column classes collected", map[string]interface{}{
		"class_count": len(c.columnClasses),
	})

	// Generate body content once for both font detection and final output
	debug.DebugLog("mjml-root", "render-body", "Rendering body content for font analysis and output")
	var bodyBuffer strings.Builder
	if c.Body != nil {
		if err := c.Body.RenderHTML(&bodyBuffer); err != nil {
			debug.DebugLogError("mjml-root", "render-body-error", "Failed to render body", err)
			return err
		}
	}
	bodyContent := bodyBuffer.String()
	debug.DebugLogWithData("mjml-root", "render-complete", "Body rendering completed", map[string]interface{}{
		"body_length": len(bodyContent),
	})

	// DOCTYPE and HTML opening
	if _, err := w.WriteString(`<!doctype html><html xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">`); err != nil {
		return err
	}

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

	if _, err := w.WriteString(`<head><title>` + title + `</title>`); err != nil {
		return err
	}
	if _, err := w.WriteString(`<!--[if !mso]><!--><meta http-equiv="X-UA-Compatible" content="IE=edge"><!--<![endif]-->`); err != nil {
		return err
	}
	if _, err := w.WriteString(`<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">`); err != nil {
		return err
	}
	if _, err := w.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1">`); err != nil {
		return err
	}

	// Base CSS
	baseCSSText := "\n<style type=\"text/css\">\n" +
		"#outlook a { padding: 0; }\n" +
		"body { margin: 0; padding: 0; -webkit-text-size-adjust: 100%; -ms-text-size-adjust: 100%; }\n" +
		"table, td { border-collapse: collapse; mso-table-lspace: 0pt; mso-table-rspace: 0pt; }\n" +
		"img { border: 0; height: auto; line-height: 100%; outline: none; text-decoration: none; -ms-interpolation-mode: bicubic; }\n" +
		"p { display: block; margin: 13px 0; }\n" +
		"</style>\n"
	if _, err := w.WriteString(baseCSSText); err != nil {
		return err
	}

	// MSO conditionals
	msoText := "<!--[if mso]>\n<noscript>\n<xml>\n<o:OfficeDocumentSettings>\n  <o:AllowPNG/>\n  <o:PixelsPerInch>96</o:PixelsPerInch>\n</o:OfficeDocumentSettings>\n</xml>\n</noscript>\n<![endif]-->\n" +
		"<!--[if lte mso 11]>\n<style type=\"text/css\">\n.mj-outlook-group-fix { width:100% !important; }\n</style>\n<![endif]-->\n"
	if _, err := w.WriteString(msoText); err != nil {
		return err
	}

	// Font imports - auto-detect fonts from content and add custom fonts from mj-font
	var allFontsToImport []string

	// Add explicit custom fonts from mj-font components
	allFontsToImport = append(allFontsToImport, customFonts...)

	// Get fonts tracked during component rendering
	trackedFonts := c.RenderOpts.FontTracker.GetFonts()
	detectedFonts := fonts.ConvertFontFamiliesToURLs(trackedFonts)
	debug.DebugLogWithData(
		"font-detection",
		"component-tracking",
		"Fonts tracked from components",
		map[string]interface{}{
			"tracked_count": len(trackedFonts),
			"url_count":     len(detectedFonts),
			"fonts":         strings.Join(trackedFonts, ","),
		},
	)

	for _, detectedFont := range detectedFonts {
		// Only add if not already in custom fonts from mj-font
		alreadyExists := false
		for _, customFont := range customFonts {
			if customFont == detectedFont {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			allFontsToImport = append(allFontsToImport, detectedFont)
		}
	}

	// Also check for default fonts based on component presence (like MRML does)
	// Note: MRML only imports fonts when specific conditions are met, not just any text presence
	hasSocial := c.hasSocialComponents()
	hasButtons := c.hasButtonComponents()
	hasText := c.hasTextComponents()

	// Only auto-import default fonts if no fonts were already detected from content
	// This matches MRML's behavior: explicit fonts override default font imports
	if len(detectedFonts) == 0 && hasSocial {
		debug.DebugLogWithData(
			"font-detection",
			"check-defaults",
			"No content fonts detected, checking defaults",
			map[string]interface{}{
				"has_social": hasSocial,
			},
		)
		defaultFonts := fonts.DetectDefaultFonts(hasText, hasSocial, hasButtons)
		debug.DebugLogWithData("font-detection", "default-fonts", "Default fonts to import", map[string]interface{}{
			"count": len(defaultFonts),
			"fonts": strings.Join(defaultFonts, ","),
		})
		for _, defaultFont := range defaultFonts {
			// Only add if not already in existing fonts
			alreadyExists := false
			for _, existingFont := range allFontsToImport {
				if existingFont == defaultFont {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				allFontsToImport = append(allFontsToImport, defaultFont)
			}
		}
	} else {
		debug.DebugLogWithData("font-detection", "skip-defaults", "Skipping default fonts", map[string]interface{}{
			"detected_count": len(detectedFonts),
			"has_social":     hasSocial,
		})
	}

	// Generate font import HTML
	debug.DebugLogWithData("font-detection", "final-list", "Final fonts to import", map[string]interface{}{
		"total_count": len(allFontsToImport),
		"fonts":       strings.Join(allFontsToImport, ","),
	})
	if len(allFontsToImport) > 0 {
		fontImportsHTML := fonts.BuildFontsTags(allFontsToImport)
		if _, err := w.WriteString(fontImportsHTML); err != nil {
			return err
		}
	}

	// Dynamic responsive CSS based on collected column classes - only if we have columns
	if len(c.columnClasses) > 0 {
		responsiveCSS := c.generateResponsiveCSS()
		if _, err := w.WriteString(responsiveCSS); err != nil {
			return err
		}
	}

	// Mobile CSS - add only if components need it (following MRML pattern)
	if c.hasMobileCSSComponents() {
		mobileCSSText := `<style type="text/css">@media only screen and (max-width:479px) {
                table.mj-full-width-mobile { width: 100% !important; }
                td.mj-full-width-mobile { width: auto !important; }
            }
            </style>`
		if _, err := w.WriteString(mobileCSSText); err != nil {
			return err
		}
	}

	// Accordion CSS - add only if components need it (following MRML pattern)
	if c.hasAccordionComponents() {
		accordionCSSText := c.generateAccordionCSS()
		if _, err := w.WriteString(accordionCSSText); err != nil {
			return err
		}
	}

	// Custom styles from mj-style components (MRML lines 240-244)
	customStyles := c.generateCustomStyles()
	if _, err := w.WriteString(customStyles); err != nil {
		return err
	}

	if _, err := w.WriteString(`</head>`); err != nil {
		return err
	}

	// Body with background-color support (matching MRML's get_body_tag)
	var bodyStyles []string

	// Always add word-spacing:normal to match MRML behavior
	bodyStyles = append(bodyStyles, "word-spacing:normal")

	if c.Body != nil {
		if bgColor := c.Body.GetAttribute("background-color"); bgColor != nil && *bgColor != "" {
			bodyStyles = append(bodyStyles, "background-color:"+*bgColor)
		}
	}

	bodyTag := `<body>`
	if len(bodyStyles) > 0 {
		bodyTag = `<body style="` + strings.Join(bodyStyles, ";") + `;">`
	}
	if _, err := w.WriteString(bodyTag); err != nil {
		return err
	}

	// Add preview text from head components right after body tag
	if c.Head != nil {
		for _, child := range c.Head.Children {
			if previewComp, ok := child.(*components.MJPreviewComponent); ok {
				if err := previewComp.RenderHTML(w); err != nil {
					return err
				}
			}
		}
	}

	// Write the body content (already rendered once above)
	if _, err := w.WriteString(bodyContent); err != nil {
		return err
	}
	if _, err := w.WriteString(`</body></html>`); err != nil {
		return err
	}

	return nil
}

func (c *MJMLComponent) RenderMJML(w io.StringWriter) error {
	if _, err := w.WriteString("<mjml>"); err != nil {
		return err
	}

	// Render head if present
	if c.Head != nil {
		if err := c.Head.RenderMJML(w); err != nil {
			return err
		}
	}

	// Render body if present
	if c.Body != nil {
		if err := c.Body.RenderMJML(w); err != nil {
			return err
		}
	}

	if _, err := w.WriteString("\n</mjml>"); err != nil {
		return err
	}

	return nil
}
