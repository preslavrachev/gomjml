// Package styles provides utilities for parsing and applying CSS values to HTML elements.
// It includes type-safe CSS value parsing and common style application patterns
// used throughout the MJML to HTML conversion process.
package styles

import (
	"fmt"

	"github.com/preslavrachev/gomjml/mjml/html"
)

// ApplyBackgroundStyles applies CSS background-related styles to an HTML tag.
// It handles background color, background images, and related properties commonly
// used in email templates.
//
// Parameters:
//
//	tag: the HTMLTag to modify
//	bgcolor: background color (e.g., "#f0f0f0")
//	bgImage: background image URL
//	bgRepeat: background repeat value (e.g., "no-repeat")
//	bgSize: background size value (e.g., "cover")
//	bgPosition: background position value (e.g., "center center")
//
// Example:
//
//	tag := html.NewHTMLTag("div")
//	ApplyBackgroundStyles(tag, ptrString("#f0f0f0"), ptrString("bg.jpg"), ptrString("no-repeat"), ptrString("cover"), ptrString("center"))
func ApplyBackgroundStyles(tag *html.HTMLTag, bgcolor, bgImage, bgRepeat, bgSize, bgPosition *string) *html.HTMLTag {
	// Apply both "background" and "background-color" for compatibility
	// with email clients and to match MRML's output.
	tag.MaybeAddStyle("background", bgcolor)
	tag.MaybeAddStyle("background-color", bgcolor)

	if bgImage != nil && *bgImage != "" {
		tag.AddStyle("background-image", fmt.Sprintf("url('%s')", *bgImage))
		tag.MaybeAddStyle("background-repeat", bgRepeat)
		tag.MaybeAddStyle("background-size", bgSize)
		tag.MaybeAddStyle("background-position", bgPosition)
	}

	return tag
}

// ApplyBorderStyles applies CSS border-related styles to an HTML tag.
// This includes border properties and border-radius for rounded corners.
//
// Parameters:
//
//	tag: the HTMLTag to modify
//	border: border specification (e.g., "1px solid #ccc")
//	borderRadius: border radius value (e.g., "4px")
//	borderTop, borderRight, borderBottom, borderLeft: individual border specifications
//
// Example:
//
//	tag := html.NewHTMLTag("div")
//	ApplyBorderStyles(tag, ptrString("1px solid #ccc"), ptrString("4px"), nil, nil, nil, nil)
func ApplyBorderStyles(tag *html.HTMLTag, border, borderRadius, borderTop, borderRight, borderBottom, borderLeft *string) *html.HTMLTag {
	tag.MaybeAddStyle("border", border)
	tag.MaybeAddStyle("border-radius", borderRadius)
	tag.MaybeAddStyle("border-top", borderTop)
	tag.MaybeAddStyle("border-right", borderRight)
	tag.MaybeAddStyle("border-bottom", borderBottom)
	tag.MaybeAddStyle("border-left", borderLeft)
	return tag
}

// ApplyPaddingStyles applies CSS padding styles to an HTML tag using type-safe spacing parsing.
// The padding value is parsed into a Spacing struct to ensure proper CSS formatting.
//
// Parameters:
//
//	tag: the HTMLTag to modify
//	padding: padding specification (e.g., "10px", "10px 20px", "10px 20px 30px 40px")
//
// Example:
//
//	tag := html.NewHTMLTag("div")
//	ApplyPaddingStyles(tag, "10px 20px") // top/bottom: 10px, left/right: 20px
func ApplyPaddingStyles(tag *html.HTMLTag, padding string) *html.HTMLTag {
	if spacing, err := ParseSpacing(padding); err == nil && spacing != nil {
		tag.AddStyle("padding", spacing.String())
	}
	return tag
}

// ApplyMarginStyles applies CSS margin styles to an HTML tag using type-safe spacing parsing.
// The margin value is parsed into a Spacing struct to ensure proper CSS formatting.
//
// Parameters:
//
//	tag: the HTMLTag to modify
//	margin: margin specification (e.g., "10px", "10px 20px", "10px 20px 30px 40px")
//
// Example:
//
//	tag := html.NewHTMLTag("div")
//	ApplyMarginStyles(tag, "0px auto") // centered with no top/bottom margin
func ApplyMarginStyles(tag *html.HTMLTag, margin string) *html.HTMLTag {
	if spacing, err := ParseSpacing(margin); err == nil && spacing != nil {
		tag.AddStyle("margin", spacing.String())
	}
	return tag
}

// ApplyFontStyles applies CSS font-related styles to an HTML tag.
// This includes font family, size, weight, and color properties commonly used in text elements.
//
// Parameters:
//
//	tag: the HTMLTag to modify
//	fontFamily: font family specification (e.g., "Arial, sans-serif")
//	fontSize: font size value (e.g., "14px", "1.2em")
//	fontWeight: font weight value (e.g., "bold", "400")
//	fontStyle: font style value (e.g., "italic", "normal")
//	color: text color value (e.g., "#333", "rgba(0,0,0,0.8)")
//	lineHeight: line height value (e.g., "1.4", "20px")
//	textAlign: text alignment value (e.g., "left", "center", "right")
//	textDecoration: text decoration value (e.g., "underline", "none")
//
// Example:
//
//	tag := html.NewHTMLTag("span")
//	ApplyFontStyles(tag, ptrString("Arial, sans-serif"), ptrString("14px"), ptrString("bold"), nil, ptrString("#333"), nil, nil, nil)
func ApplyFontStyles(tag *html.HTMLTag, fontFamily, fontSize, fontWeight, fontStyle, color, lineHeight, textAlign, textDecoration *string) *html.HTMLTag {
	tag.MaybeAddStyle("font-family", fontFamily)
	tag.MaybeAddStyle("font-size", fontSize)
	tag.MaybeAddStyle("font-weight", fontWeight)
	tag.MaybeAddStyle("font-style", fontStyle)
	tag.MaybeAddStyle("color", color)
	tag.MaybeAddStyle("line-height", lineHeight)
	tag.MaybeAddStyle("text-align", textAlign)
	tag.MaybeAddStyle("text-decoration", textDecoration)
	return tag
}

// ApplyTextAlignStyles applies CSS text alignment styles to an HTML tag.
//
// Parameters:
//
//	tag: the HTMLTag to modify
//	align: text alignment value (e.g., "left", "center", "right", "justify")
//
// Example:
//
//	tag := html.NewHTMLTag("p")
//	ApplyTextAlignStyles(tag, ptrString("center"))
func ApplyTextAlignStyles(tag *html.HTMLTag, align *string) *html.HTMLTag {
	tag.MaybeAddStyle("text-align", align)
	return tag
}

// ApplyDimensionStyles applies CSS width and height styles to an HTML tag.
//
// Parameters:
//
//	tag: the HTMLTag to modify
//	width: width value (e.g., "100px", "50%", "auto")
//	height: height value (e.g., "200px", "auto")
//	minWidth, maxWidth, minHeight, maxHeight: min/max dimension constraints
//
// Example:
//
//	tag := html.NewHTMLTag("img")
//	ApplyDimensionStyles(tag, ptrString("100%"), ptrString("auto"), nil, nil, nil, nil)
func ApplyDimensionStyles(tag *html.HTMLTag, width, height, minWidth, maxWidth, minHeight, maxHeight *string) *html.HTMLTag {
	tag.MaybeAddStyle("width", width)
	tag.MaybeAddStyle("height", height)
	tag.MaybeAddStyle("min-width", minWidth)
	tag.MaybeAddStyle("max-width", maxWidth)
	tag.MaybeAddStyle("min-height", minHeight)
	tag.MaybeAddStyle("max-height", maxHeight)
	return tag
}

// ApplyMSOBackgroundStyles applies both standard CSS background styles and MSO-compatible
// background attributes to an HTML tag. This ensures proper background rendering in both
// modern email clients and Microsoft Outlook.
//
// Outlook has limited CSS background support and often requires HTML bgcolor attributes
// for reliable background color rendering.
//
// Parameters:
//
//	tag: the HTMLTag to modify
//	bgcolor: background color (e.g., "#f0f0f0")
//	bgImage: background image URL
//	bgRepeat: background repeat value (e.g., "no-repeat")
//	bgSize: background size value (e.g., "cover")
//	bgPosition: background position value (e.g., "center center")
//
// Example:
//
//	tag := html.NewHTMLTag("table")
//	ApplyMSOBackgroundStyles(tag, ptrString("#f0f0f0"), nil, nil, nil, nil) // adds both CSS and bgcolor attribute
func ApplyMSOBackgroundStyles(tag *html.HTMLTag, bgcolor, bgImage, bgRepeat, bgSize, bgPosition *string) *html.HTMLTag {
	// Apply standard background styles
	ApplyBackgroundStyles(tag, bgcolor, bgImage, bgRepeat, bgSize, bgPosition)

	// Add MSO-specific bgcolor attribute if background color is set
	if bgcolor != nil && *bgcolor != "" {
		tag.MaybeAddAttribute("bgcolor", bgcolor)
	}

	return tag
}
