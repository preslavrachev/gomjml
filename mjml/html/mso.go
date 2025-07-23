package html

import (
	"fmt"
	"strings"
)

// RenderMSOConditional wraps content in MSO/Outlook conditional comments.
// This ensures the content is only rendered in Microsoft Outlook and Internet Explorer,
// while being hidden from other email clients.
//
// The conditional comment format <!--[if mso | IE]>...<![endif]--> is specifically
// recognized by Outlook and IE to provide fallback HTML structures.
//
// Example:
//
//	html := RenderMSOConditional("<table><tr><td>Outlook content</td></tr></table>")
//	// Output: <!--[if mso | IE]><table><tr><td>Outlook content</td></tr></table><![endif]-->
func RenderMSOConditional(content string) string {
	return fmt.Sprintf("<!--[if mso | IE]>%s<![endif]-->", content)
}

// CreateMSOTable creates an HTML table specifically designed for MSO/Outlook compatibility.
// The table includes all necessary attributes for proper rendering in Outlook:
// - border="0", cellpadding="0", cellspacing="0" for consistent spacing
// - role="presentation" for accessibility
// - align="center" for horizontal centering
// - bgcolor attribute for background color support (Outlook doesn't fully support CSS backgrounds)
//
// Parameters:
//
//	width: table width (e.g., "600" for 600px width), can be empty
//	bgcolor: background color (e.g., "#f0f0f0"), can be empty
//
// Example:
//
//	table := CreateMSOTable("600", "#f0f0f0")
//	html := table.RenderOpen() // <table border="0" cellpadding="0" cellspacing="0" role="presentation" align="center" width="600" bgcolor="#f0f0f0">
func CreateMSOTable(width, bgcolor string) *HTMLTag {
	tag := NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center")

	if width != "" {
		tag.AddAttribute("width", width)
	}

	if bgcolor != "" {
		tag.AddAttribute("bgcolor", bgcolor)
	}

	return tag
}

// CreateMSOTableCell creates a table cell (td) with MSO-specific styles for proper line height handling.
// Outlook has issues with line-height and font-size, so this cell includes:
// - line-height: 0px to prevent spacing issues
// - font-size: 0px to eliminate text node spacing
// - mso-line-height-rule: exactly for precise Outlook line height control
//
// This is typically used as a container cell within MSO conditional tables.
//
// Example:
//
//	cell := CreateMSOTableCell()
//	html := cell.RenderOpen() // <td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;">
func CreateMSOTableCell() *HTMLTag {
	return NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")
}

// ApplyMSOStyles applies Microsoft Outlook-specific CSS styles to an HTML tag.
// These styles help ensure consistent rendering in Outlook by:
// - Setting exact line height rules (mso-line-height-rule: exactly)
// - Removing default padding alternatives (mso-padding-alt: 0px)
//
// This function should be used on elements that need precise Outlook compatibility.
//
// Example:
//
//	tag := NewHTMLTag("div")
//	ApplyMSOStyles(tag) // Adds mso-line-height-rule and mso-padding-alt styles
func ApplyMSOStyles(tag *HTMLTag) *HTMLTag {
	return tag.
		AddStyle("mso-line-height-rule", "exactly").
		AddStyle("mso-padding-alt", "0px")
}

// ApplyMSOTableStyles applies MSO-compatible attributes to a table element.
// This includes standard email table attributes and optional background color for Outlook.
//
// The bgcolor attribute is crucial for Outlook background color support, as Outlook
// has limited CSS background support and relies on HTML attributes instead.
//
// Parameters:
//
//	tag: the HTMLTag to modify (should be a table element)
//	bgcolor: background color for Outlook (e.g., "#f0f0f0"), can be empty
//
// Example:
//
//	table := NewHTMLTag("table")
//	ApplyMSOTableStyles(table, "#f0f0f0")
func ApplyMSOTableStyles(tag *HTMLTag, bgcolor string) *HTMLTag {
	tag.AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation")

	if bgcolor != "" {
		tag.AddAttribute("bgcolor", bgcolor)
	}

	return tag
}

// WrapWithMSOTable wraps content with a complete MSO-compatible table structure.
// This creates a full MSO conditional comment containing a table and cell wrapper
// around the provided content.
//
// This is useful for ensuring content renders properly in Outlook by providing
// a table-based layout fallback while maintaining modern CSS for other clients.
//
// Parameters:
//
//	content: the HTML content to wrap
//	width: table width (e.g., "600"), can be empty
//	bgcolor: background color (e.g., "#f0f0f0"), can be empty
//
// Example:
//
//	wrapped := WrapWithMSOTable("<div>Content</div>", "600", "#f0f0f0")
//	// Output: <!--[if mso | IE]><table width="600" bgcolor="#f0f0f0">...<tr><td>...</td></tr></table><![endif]-->
//	         <div>Content</div>
//	         <!--[if mso | IE]></td></tr></table><![endif]-->
func WrapWithMSOTable(content, width, bgcolor string) string {
	var html strings.Builder

	msoTable := CreateMSOTable(width, bgcolor)
	msoCell := CreateMSOTableCell()

	html.WriteString(RenderMSOConditional(
		msoTable.RenderOpen() +
			"<tr>" +
			msoCell.RenderOpen(),
	))

	html.WriteString(content)

	html.WriteString(RenderMSOConditional(
		msoCell.RenderClose() +
			"</tr>" +
			msoTable.RenderClose(),
	))

	return html.String()
}

// CreateMSOCompatibleWrapper creates the opening HTML for an MSO-compatible wrapper structure.
// This combines a modern div element with an MSO table fallback, providing the best of both:
// - Modern CSS div for advanced email clients
// - Table-based fallback for Outlook compatibility
//
// The function returns the opening HTML that should be followed by content and closed
// with CloseMSOCompatibleWrapper.
//
// Parameters:
//
//	divTag: the HTMLTag for the main div element
//	width: table width for MSO fallback (e.g., "600")
//	bgcolor: background color for both div and MSO table
//
// Example:
//
//	div := NewHTMLTag("div").AddStyle("background-color", "#f0f0f0")
//	opening := CreateMSOCompatibleWrapper(div, "600", "#f0f0f0")
//	// Returns: <!--[if mso | IE]><table...><tr><td...><![endif]--><div style="background-color:#f0f0f0;">
func CreateMSOCompatibleWrapper(divTag *HTMLTag, width, bgcolor string) string {
	var html strings.Builder

	// MSO table wrapper
	msoTable := CreateMSOTable(width, bgcolor)
	msoCell := CreateMSOTableCell()

	html.WriteString(RenderMSOConditional(
		msoTable.RenderOpen() +
			"<tr>" +
			msoCell.RenderOpen(),
	))

	// Main div (hidden from MSO)
	html.WriteString(divTag.RenderOpen())

	return html.String()
}

// CloseMSOCompatibleWrapper closes an MSO-compatible wrapper structure created by CreateMSOCompatibleWrapper.
// This provides the closing tags for both the div element and the MSO table fallback.
//
// Parameters:
//
//	divTag: the same HTMLTag used in CreateMSOCompatibleWrapper (used only for tag name)
//
// Example:
//
//	div := NewHTMLTag("div")
//	closing := CloseMSOCompatibleWrapper(div)
//	// Returns: </div><!--[if mso | IE]></td></tr></table><![endif]-->
func CloseMSOCompatibleWrapper(divTag *HTMLTag) string {
	var html strings.Builder

	html.WriteString(divTag.RenderClose())

	html.WriteString(RenderMSOConditional(
		"</td></tr></table>",
	))

	return html.String()
}

// ApplyMSOFontFallback applies MSO-specific font fallback styles to an HTML tag.
// Outlook has limited font support and requires special handling for web fonts.
//
// This function adds the mso-font-alt style property which provides a fallback
// font name for Outlook when web fonts are not available.
//
// Parameters:
//
//	tag: the HTMLTag to modify
//	fontFamily: the font family name (e.g., "Arial, sans-serif")
//
// Example:
//
//	tag := NewHTMLTag("div")
//	ApplyMSOFontFallback(tag, "Helvetica, Arial, sans-serif")
//	// Adds: mso-font-alt: "Helvetica, Arial, sans-serif"
func ApplyMSOFontFallback(tag *HTMLTag, fontFamily string) *HTMLTag {
	if fontFamily != "" {
		// Remove quotes from font family for MSO compatibility
		msoFontFamily := strings.ReplaceAll(fontFamily, "'", "")
		tag.AddStyle("mso-font-alt", msoFontFamily)
	}
	return tag
}

// ApplyMSOLineHeight applies MSO-specific line height styles to an HTML tag.
// Outlook requires special handling for line height to ensure consistent text spacing.
//
// This function adds both the standard line-height CSS property and the MSO-specific
// mso-line-height-rule property set to "exactly" for precise Outlook control.
//
// Parameters:
//
//	tag: the HTMLTag to modify
//	lineHeight: the line height value (e.g., "1.4", "20px")
//
// Example:
//
//	tag := NewHTMLTag("p")
//	ApplyMSOLineHeight(tag, "1.4")
//	// Adds: line-height: 1.4; mso-line-height-rule: exactly;
func ApplyMSOLineHeight(tag *HTMLTag, lineHeight string) *HTMLTag {
	if lineHeight != "" {
		tag.AddStyle("line-height", lineHeight).
			AddStyle("mso-line-height-rule", "exactly")
	}
	return tag
}
