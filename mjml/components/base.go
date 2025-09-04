package components

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/debug"
	"github.com/preslavrachev/gomjml/mjml/globals"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/mjml/styles"
	"github.com/preslavrachev/gomjml/parser"
)

// Common width strings to avoid fmt.Sprintf allocations
var (
	width600px = "600px"
	width100px = "100px"
	width200px = "200px"
	width300px = "300px"
	width400px = "400px"
	width500px = "500px"
	width150px = "150px"
	width50px  = "50px"
)

// pixelWidthStringBufSize is the pre-allocated buffer size for pixel width strings
// Calculation: max 4-5 digits + "px" suffix = ~7-8 chars
const pixelWidthStringBufSize = 8

// NotImplementedError indicates a component is not yet implemented
type NotImplementedError struct {
	ComponentName string
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("component %s is not yet implemented", e.ComponentName)
}

// Component represents a renderable MJML component
type Component interface {
	Render(w io.StringWriter) error
	GetTagName() string
	GetDefaultAttribute(name string) string
	SetContainerWidth(widthPx int)
	GetContainerWidth() int
	SetSiblings(siblings int)
	SetRawSiblings(rawSiblings int)
	GetSiblings() int
	GetRawSiblings() int
	IsRawElement() bool
}

// BaseComponent provides common functionality for all components
type BaseComponent struct {
	Node           *parser.MJMLNode
	Children       []Component
	Attrs          map[string]string
	classNames     []string            // Split mj-class attribute values
	classAttrs     map[string]string   // Merged attributes from mj-class definitions
	ContainerWidth int                 // Container width in pixels (0 means use default)
	Siblings       int                 // Total siblings count
	RawSiblings    int                 // Raw siblings count (for width calculations)
	RenderOpts     *options.RenderOpts // Rendering options
}

// NewBaseComponent creates a new base component
func NewBaseComponent(node *parser.MJMLNode, opts *options.RenderOpts) *BaseComponent {
	attrs := make(map[string]string, len(node.Attrs))
	for _, attr := range node.Attrs {
		attrs[attr.Name.Local] = attr.Value
	}

	var classNames []string
	var classAttrs map[string]string
	if classAttr, ok := attrs["mj-class"]; ok && classAttr != "" {
		classNames = strings.Fields(classAttr)
		if len(classNames) > 0 {
			classAttrs = make(map[string]string)
			cssClassParts := make([]string, 0, len(classNames)) // pre-allocate with capacity
			for _, className := range classNames {
				if ca := globals.GetClassAttributes(className); ca != nil {
					for k, v := range ca {
						if k == "css-class" {
							cssClassParts = append(cssClassParts, v)
							continue
						}
						classAttrs[k] = v // last class wins
					}
				}
			}
			if len(cssClassParts) > 0 {
				classAttrs["css-class"] = strings.Join(cssClassParts, " ")
			}
		}
	}

	if opts == nil {
		opts = &options.RenderOpts{}
	}

	return &BaseComponent{
		Node:           node,
		Attrs:          attrs,
		classNames:     classNames,
		classAttrs:     classAttrs,
		Children:       make([]Component, 0, len(node.Children)),
		ContainerWidth: 0, // 0 means use default body width
		Siblings:       1,
		RawSiblings:    0,
		RenderOpts:     opts,
	}
}

// IsRawElement returns whether this component should be treated as a raw element.
// Base components are not raw by default.
func (bc *BaseComponent) IsRawElement() bool {
	return false
}

// GetAttribute gets an attribute value as a pointer, following the MRML attribute resolution order:
// 1. Element attributes
// 2. mj-class definitions (TODO: implement)
// 3. Global element defaults (via GlobalAttributes)
// 4. Component defaults (via GetDefaultAttribute)
func (bc *BaseComponent) GetAttribute(name string) *string {
	// 1. Check element attributes
	if value, exists := bc.Attrs[name]; exists && value != "" {
		return &value
	}

	// 2. Check mj-class definitions
	if classValue := bc.getClassAttribute(name); classValue != "" {
		return &classValue
	}

	// 3. Check global defaults - we can't access GetTagName from BaseComponent
	// Global attributes will be checked in GetAttributeWithDefault or by passing component

	// 4. Check component defaults
	if defaultVal := bc.GetDefaultAttribute(name); defaultVal != "" {
		return &defaultVal
	}

	return nil
}

// GetAttributeFast gets an attribute value without debug logging using full resolution order
func (bc *BaseComponent) GetAttributeFast(comp Component, name string) string {
	// 1. Element attributes
	if value, exists := bc.Attrs[name]; exists && value != "" {
		return value
	}

	// 2. mj-class definitions
	if classValue := bc.getClassAttribute(name); classValue != "" {
		return classValue
	}

	// 3. Global attributes
	if globalValue := globals.GetGlobalAttribute(comp.GetTagName(), name); globalValue != "" {
		return globalValue
	}

	// 4. Component defaults
	if defaultVal := comp.GetDefaultAttribute(name); defaultVal != "" {
		return defaultVal
	}

	return ""
}

// GetAttributeWithDefault gets an attribute with component-specific defaults
// This method properly calls the overridden GetDefaultAttribute method on the concrete component
func (bc *BaseComponent) GetAttributeWithDefault(comp Component, name string) string {
	// 1. Check element attributes first
	if value, exists := bc.Attrs[name]; exists && value != "" {
		debug.DebugLogWithData(comp.GetTagName(), "attr-element", "Using element attribute", map[string]interface{}{
			"attr_name":  name,
			"attr_value": value,
		})
		// Track font families
		if name == constants.MJMLFontFamily {
			bc.TrackFontFamily(value)
		}
		return value
	}

	// 2. Check mj-class definitions
	if classValue := bc.getClassAttribute(name); classValue != "" {
		debug.DebugLogWithData(comp.GetTagName(), "attr-class", "Using mj-class attribute", map[string]interface{}{
			"attr_name":  name,
			"attr_value": classValue,
			"classes":    bc.Attrs["mj-class"],
		})
		if name == constants.MJMLFontFamily {
			bc.TrackFontFamily(classValue)
		}
		return classValue
	}

	// 3. Check global attributes if available (we'll get this via external function)
	if globalValue := bc.getGlobalAttribute(comp.GetTagName(), name); globalValue != "" {
		debug.DebugLogWithData(comp.GetTagName(), "attr-global", "Using global attribute", map[string]interface{}{
			"attr_name":  name,
			"attr_value": globalValue,
		})
		// Track font families
		if name == constants.MJMLFontFamily {
			bc.TrackFontFamily(globalValue)
		}
		return globalValue
	}

	// 4. Check component defaults via interface method (properly calls overridden method)
	defaultValue := comp.GetDefaultAttribute(name)
	if defaultValue != "" {
		debug.DebugLogWithData(comp.GetTagName(), "attr-default", "Using default attribute", map[string]interface{}{
			"attr_name":  name,
			"attr_value": defaultValue,
		})
		// Track font families
		if name == constants.MJMLFontFamily {
			bc.TrackFontFamily(defaultValue)
		}
	}
	return defaultValue
}

// getGlobalAttribute gets a global attribute value from the global store
func (bc *BaseComponent) getGlobalAttribute(componentName, attrName string) string {
	// Access global attributes via globals package
	return globals.GetGlobalAttribute(componentName, attrName)
}

// getClassAttribute retrieves an attribute value from mj-class definitions if present
func (bc *BaseComponent) getClassAttribute(attrName string) string {
	if bc.classAttrs == nil {
		return ""
	}
	if v, ok := bc.classAttrs[attrName]; ok {
		return v
	}
	return ""
}

// GetAttributeAsPixel parses an attribute value as a CSS pixel value
func (bc *BaseComponent) GetAttributeAsPixel(name string) *styles.Pixel {
	if attr := bc.GetAttribute(name); attr != nil {
		if pixel, err := styles.ParsePixel(*attr); err == nil {
			return pixel
		}
	}
	return nil
}

// GetAttributeAsSpacing parses an attribute value as CSS spacing (padding/margin)
func (bc *BaseComponent) GetAttributeAsSpacing(name string) *styles.Spacing {
	if attr := bc.GetAttribute(name); attr != nil {
		if spacing, err := styles.ParseSpacing(*attr); err == nil {
			return spacing
		}
	}
	return nil
}

// GetAttributeAsColor parses an attribute value as a CSS color value
func (bc *BaseComponent) GetAttributeAsColor(name string) *styles.Color {
	if attr := bc.GetAttribute(name); attr != nil {
		if color, err := styles.ParseColor(*attr); err == nil {
			return color
		}
	}
	return nil
}

// GetDefaultAttribute returns the default value for an attribute.
// Override this method in specific components to provide component-specific defaults.
func (bc *BaseComponent) GetDefaultAttribute(name string) string {
	return ""
}

// SetContainerWidth sets the container width in pixels for this component
// AIDEV-NOTE: width-flow-interface; container width flows from parent to child components
func (bc *BaseComponent) SetContainerWidth(widthPx int) {
	bc.ContainerWidth = widthPx
}

// GetContainerWidth returns the container width in pixels (0 means use default body width)
// AIDEV-NOTE: width-flow-interface; used by child components to calculate their effective rendering width
func (bc *BaseComponent) GetContainerWidth() int {
	return bc.ContainerWidth
}

// SetSiblings sets the total number of siblings for this component
func (bc *BaseComponent) SetSiblings(siblings int) {
	bc.Siblings = siblings
}

// SetRawSiblings sets the number of raw siblings for this component
func (bc *BaseComponent) SetRawSiblings(rawSiblings int) {
	bc.RawSiblings = rawSiblings
}

// GetSiblings returns the total number of siblings
func (bc *BaseComponent) GetSiblings() int {
	return bc.Siblings
}

// GetRawSiblings returns the number of raw siblings
func (bc *BaseComponent) GetRawSiblings() int {
	return bc.RawSiblings
}

// GetNonRawSiblings returns the number of non-raw siblings (used for width calculations)
func (bc *BaseComponent) GetNonRawSiblings() int {
	return bc.Siblings - bc.RawSiblings
}

// GetEffectiveWidth returns the container width if set, otherwise default body width
// AIDEV-NOTE: width-flow-calculation; used to calculate actual pixel width for rendering and child width calculation
func (bc *BaseComponent) GetEffectiveWidth() int {
	if bc.ContainerWidth > 0 {
		return bc.ContainerWidth
	}
	return GetDefaultBodyWidthPixels()
}

// GetEffectiveWidthString returns the effective width as a string with px units
func (bc *BaseComponent) GetEffectiveWidthString() string {
	if bc.ContainerWidth > 0 {
		return getPixelWidthString(bc.ContainerWidth)
	}
	return GetDefaultBodyWidth()
}

// getPixelWidthString returns pixel width string, using cached values for common widths to avoid allocations
func getPixelWidthString(widthPx int) string {
	switch widthPx {
	case 600:
		return width600px
	case 500:
		return width500px
	case 400:
		return width400px
	case 300:
		return width300px
	case 200:
		return width200px
	case 150:
		return width150px
	case 100:
		return width100px
	case 50:
		return width50px
	default:
		// Fallback using strconv for uncommon widths without fmt overhead
		var b strings.Builder
		b.Grow(pixelWidthStringBufSize) // Pre-allocate reasonable size for most width values
		b.WriteString(strconv.Itoa(widthPx))
		b.WriteString("px")
		return b.String()
	}
}

// Style Mixin Methods - Common styling patterns that components can use

// ApplyBackgroundStyles applies background-related CSS styles to an HTML tag
func (bc *BaseComponent) ApplyBackgroundStyles(tag *html.HTMLTag) *html.HTMLTag {
	bgcolor := bc.GetAttribute("background-color")
	bgImage := bc.GetAttribute("background-image")
	if bgImage == nil || *bgImage == "" {
		// MJML commonly uses the "background-url" attribute. Fall back to it
		// when "background-image" is not provided to mirror MRML's behaviour.
		bgImage = bc.GetAttribute(constants.MJMLBackgroundUrl)
	}
	bgRepeat := bc.GetAttribute("background-repeat")
	bgSize := bc.GetAttribute("background-size")
	bgPosition := bc.GetAttribute("background-position")

	return styles.ApplyBackgroundStyles(tag, bgcolor, bgImage, bgRepeat, bgSize, bgPosition)
}

// ApplyBorderStyles applies border-related CSS styles to an HTML tag
func (bc *BaseComponent) ApplyBorderStyles(tag *html.HTMLTag) *html.HTMLTag {
	border := bc.GetAttribute("border")
	borderRadius := bc.GetAttribute("border-radius")
	borderTop := bc.GetAttribute("border-top")
	borderRight := bc.GetAttribute("border-right")
	borderBottom := bc.GetAttribute("border-bottom")
	borderLeft := bc.GetAttribute("border-left")

	return styles.ApplyBorderStyles(tag, border, borderRadius, borderTop, borderRight, borderBottom, borderLeft)
}

// ApplyPaddingStyles applies padding CSS styles to an HTML tag
func (bc *BaseComponent) ApplyPaddingStyles(tag *html.HTMLTag) *html.HTMLTag {
	if spacing := bc.GetAttributeAsSpacing("padding"); spacing != nil {
		tag.AddStyle("padding", spacing.String())
	}
	return tag
}

// ApplyMarginStyles applies margin CSS styles to an HTML tag
func (bc *BaseComponent) ApplyMarginStyles(tag *html.HTMLTag) *html.HTMLTag {
	if spacing := bc.GetAttributeAsSpacing("margin"); spacing != nil {
		tag.AddStyle("margin", spacing.String())
	}
	return tag
}

// TrackFontFamily tracks a font family in the render options font tracker
func (bc *BaseComponent) TrackFontFamily(fontFamily string) {
	if fontFamily != "" && bc.RenderOpts != nil && bc.RenderOpts.FontTracker != nil {
		bc.RenderOpts.FontTracker.AddFont(fontFamily)
	}
}

// ApplyFontStyles applies font-related CSS styles to an HTML tag
func (bc *BaseComponent) ApplyFontStyles(tag *html.HTMLTag) *html.HTMLTag {
	fontFamily := bc.GetAttribute("font-family")
	fontSize := bc.GetAttribute("font-size")
	fontWeight := bc.GetAttribute("font-weight")
	fontStyle := bc.GetAttribute("font-style")
	color := bc.GetAttribute("color")
	lineHeight := bc.GetAttribute("line-height")
	textAlign := bc.GetAttribute("text-align")
	textDecoration := bc.GetAttribute("text-decoration")

	// Track font family usage
	bc.TrackFontFamily(*fontFamily)

	return styles.ApplyFontStyles(
		tag,
		fontFamily,
		fontSize,
		fontWeight,
		fontStyle,
		color,
		lineHeight,
		textAlign,
		textDecoration,
	)
}

// ApplyDimensionStyles applies width/height CSS styles to an HTML tag
func (bc *BaseComponent) ApplyDimensionStyles(tag *html.HTMLTag) *html.HTMLTag {
	width := bc.GetAttribute("width")
	height := bc.GetAttribute("height")
	minWidth := bc.GetAttribute("min-width")
	maxWidth := bc.GetAttribute("max-width")
	minHeight := bc.GetAttribute("min-height")
	maxHeight := bc.GetAttribute("max-height")

	return styles.ApplyDimensionStyles(tag, width, height, minWidth, maxWidth, minHeight, maxHeight)
}

// AddDebugAttribute adds a debug attribute to an HTML tag for component traceability
// This helps identify which MJML component generated which HTML elements during testing
func (bc *BaseComponent) AddDebugAttribute(tag *html.HTMLTag, componentType string) {
	// Only add debug attributes if enabled in render options
	if bc.RenderOpts != nil && bc.RenderOpts.DebugTags {
		debugAttr := "data-mj-debug-" + componentType
		tag.AddAttribute(debugAttr, "true")
	}
}

// CSS Class Helper Methods - Generic css-class attribute handling for all components

// GetCSSClass returns the css-class attribute value
func (bc *BaseComponent) GetCSSClass() string {
	if value, exists := bc.Attrs["css-class"]; exists {
		return value
	}
	return ""
}

// BuildClassAttribute combines existing CSS classes with the css-class attribute
// Usage: component.BuildClassAttribute("mj-column-per-100", "mj-outlook-group-fix")
func (bc *BaseComponent) BuildClassAttribute(existingClasses ...string) string {
	var classes []string

	// Add existing classes
	for _, class := range existingClasses {
		if class != "" {
			classes = append(classes, class)
		}
	}

	// Determine css-class from element or mj-class definitions
	cssClass := bc.GetCSSClass()
	if cssClass == "" {
		cssClass = bc.getClassAttribute("css-class")
	}
	if cssClass != "" {
		classes = append(classes, cssClass)
	}

	if len(classes) == 0 {
		return ""
	}

	return strings.Join(classes, " ")
}

// GetMSOClassAttribute returns the MSO conditional comment class attribute with -outlook suffix
// Returns empty string if no css-class is set, or " class=\"css-class-outlook\"" if set
func (bc *BaseComponent) GetMSOClassAttribute() string {
	if cssClass := bc.GetCSSClass(); cssClass != "" {
		return " class=\"" + cssClass + "-outlook\""
	}
	return ""
}
