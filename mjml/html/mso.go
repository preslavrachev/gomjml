package html

import (
	"io"
	"strconv"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
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
//	RenderMSOConditional(w, "<table><tr><td>Outlook content</td></tr></table>")
//	// Output: <!--[if mso | IE]><table><tr><td>Outlook content</td></tr></table><![endif]-->
func RenderMSOConditional(w io.StringWriter, content string) error {
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}
	if _, err := w.WriteString(content); err != nil {
		return err
	}
	if _, err := w.WriteString("<![endif]-->"); err != nil {
		return err
	}
	return nil
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
		AddAttribute("role", "presentation")

	// Add attributes in MRML order: bgcolor, align, width
	if bgcolor != "" {
		tag.AddAttribute("bgcolor", bgcolor)
	}

	tag.AddAttribute("align", "center")

	if width != "" {
		tag.AddAttribute("width", width)
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
//	w: the StringWriter to write to
//	content: the HTML content to wrap
//	width: table width (e.g., "600"), can be empty
//	bgcolor: background color (e.g., "#f0f0f0"), can be empty
//
// Example:
//
//	WrapWithMSOTable(w, "<div>Content</div>", "600", "#f0f0f0")
//	// Writes: <!--[if mso | IE]><table width="600" bgcolor="#f0f0f0">...<tr><td>...</td></tr></table><![endif]-->
//	         <div>Content</div>
//	         <!--[if mso | IE]></td></tr></table><![endif]-->
func WrapWithMSOTable(w io.StringWriter, content, width, bgcolor string) error {
	msoTable := CreateMSOTable(width, bgcolor)
	msoCell := CreateMSOTableCell()

	// Opening MSO conditional with table and cell
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}
	if err := msoTable.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString(" <tr>"); err != nil {
		return err
	}
	if err := msoCell.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<![endif]-->"); err != nil {
		return err
	}

	// Content
	if _, err := w.WriteString(content); err != nil {
		return err
	}

	// Closing MSO conditional with cell and table
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}
	if err := msoCell.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr>"); err != nil {
		return err
	}
	if err := msoTable.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<![endif]-->"); err != nil {
		return err
	}

	return nil
}

// CreateMSOCompatibleWrapper creates the opening HTML for an MSO-compatible wrapper structure.
// This combines a modern div element with an MSO table fallback, providing the best of both:
// - Modern CSS div for advanced email clients
// - Table-based fallback for Outlook compatibility
//
// The function writes the opening HTML that should be followed by content and closed
// with CloseMSOCompatibleWrapper.
//
// Parameters:
//
//	w: the StringWriter to write to
//	divTag: the HTMLTag for the main div element
//	width: table width for MSO fallback (e.g., "600")
//	bgcolor: background color for both div and MSO table
//
// Example:
//
//	div := NewHTMLTag("div").AddStyle("background-color", "#f0f0f0")
//	CreateMSOCompatibleWrapper(w, div, "600", "#f0f0f0")
//	// Writes: <!--[if mso | IE]><table...><tr><td...><![endif]--><div style="background-color:#f0f0f0;">
func CreateMSOCompatibleWrapper(w io.StringWriter, divTag *HTMLTag, width, bgcolor string) error {
	// MSO table wrapper
	msoTable := CreateMSOTable(width, bgcolor)
	msoCell := CreateMSOTableCell()

	// Opening MSO conditional
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}
	if err := msoTable.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString(" <tr>"); err != nil {
		return err
	}
	if err := msoCell.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<![endif]-->"); err != nil {
		return err
	}

	// Main div (hidden from MSO)
	return divTag.RenderOpen(w)
}

// CloseMSOCompatibleWrapper closes an MSO-compatible wrapper structure created by CreateMSOCompatibleWrapper.
// This provides the closing tags for both the div element and the MSO table fallback.
//
// Parameters:
//
//	w: the StringWriter to write to
//	divTag: the same HTMLTag used in CreateMSOCompatibleWrapper (used only for tag name)
//
// Example:
//
//	div := NewHTMLTag("div")
//	CloseMSOCompatibleWrapper(w, div)
//	// Writes: </div><!--[if mso | IE]></td></tr></table><![endif]-->
func CloseMSOCompatibleWrapper(w io.StringWriter, divTag *HTMLTag) error {
	// Close main div
	if err := divTag.RenderClose(w); err != nil {
		return err
	}

	// Close MSO conditional
	return RenderMSOConditional(w, "</td></tr></table>")
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

// RenderMSOTableOpen renders an MSO table opening with tr and td tags
func RenderMSOTableOpen(w io.StringWriter, table, td *HTMLTag) error {
	if err := table.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString(" <tr>"); err != nil {
		return err
	}
	return td.RenderOpen(w)
}

// RenderMSOTableClose renders an MSO table closing with td and tr tags
func RenderMSOTableClose(w io.StringWriter, td, table *HTMLTag) error {
	if err := td.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr>"); err != nil {
		return err
	}
	return table.RenderClose(w)
}

// RenderMSOTableTrOpen renders an MSO table opening with tr tags for section background
func RenderMSOTableTrOpen(w io.StringWriter, table, tr, td *HTMLTag) error {
	if err := table.RenderOpen(w); err != nil {
		return err
	}
	if err := tr.RenderOpen(w); err != nil {
		return err
	}
	return td.RenderOpen(w)
}

// RenderMSOTableOpenConditional renders MSO table open with conditional comments directly to Writer
func RenderMSOTableOpenConditional(w io.StringWriter, table, td *HTMLTag) error {
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}
	var tableOpen strings.Builder
	if err := table.RenderOpen(&tableOpen); err != nil {
		return err
	}
	openStr := tableOpen.String()
	if strings.HasSuffix(openStr, ">") {
		if len(openStr) == 1 || openStr[len(openStr)-2] != ' ' {
			openStr = openStr[:len(openStr)-1] + " >"
		}
	}
	if _, err := w.WriteString(openStr); err != nil {
		return err
	}
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}
	if err := td.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<![endif]-->"); err != nil {
		return err
	}
	return nil
}

// RenderMSOTableCloseConditional renders MSO table close with conditional comments directly to Writer
func RenderMSOTableCloseConditional(w io.StringWriter, td, table *HTMLTag) error {
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}
	if err := td.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr>"); err != nil {
		return err
	}
	if err := table.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<![endif]-->"); err != nil {
		return err
	}
	return nil
}

// RenderMSOTableTrOpenConditional renders MSO table with tr opening with conditional comments directly to Writer
func RenderMSOTableTrOpenConditional(w io.StringWriter, table, tr, td *HTMLTag) error {
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}
	if err := table.RenderOpen(w); err != nil {
		return err
	}
	if err := tr.RenderOpen(w); err != nil {
		return err
	}
	if err := td.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<![endif]-->"); err != nil {
		return err
	}
	return nil
}

// RenderMSOWrapperTableOpen renders the opening Outlook wrapper table with all
// required attributes. The generated structure matches MRML's output exactly so
// that integration tests comparing against reference HTML don't report spurious
// differences.
//
// Example output for width=600:
//
//	<!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="" width="600px" ><table align="center" border="0" cellpadding="0" cellspacing="0" class="" role="presentation" style="width:600px;" width="600" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->
func RenderMSOWrapperTableOpen(w io.StringWriter, widthPx int, align string) error {
	return RenderMSOWrapperTableOpenWithWidths(w, widthPx, widthPx, align)
}

// RenderMSOWrapperTableOpenWithWidths renders the opening Outlook wrapper table
// while allowing different outer (MSO td) and inner table widths. This matches
// MJML's behavior for wrappers with borders where Outlook table width remains at
// the body width but the inner table shrinks by the border size.
func RenderMSOWrapperTableOpenWithWidths(w io.StringWriter, outerWidthPx int, innerWidthPx int, align string) error {
	if _, err := w.WriteString("<!--[if mso | IE]><table role=\"presentation\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\"><tr><td class=\"\""); err != nil {
		return err
	}
	if align != "" {
		if _, err := w.WriteString(" " + constants.AttrAlign + "=\""); err != nil {
			return err
		}
		if _, err := w.WriteString(align + "\""); err != nil {
			return err
		}
	}
	if _, err := w.WriteString(" " + constants.AttrWidth + "=\""); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(outerWidthPx)); err != nil {
		return err
	}
	if _, err := w.WriteString("px\" ><table align=\"center\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" class=\"\" role=\"presentation\" style=\"width:"); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(innerWidthPx)); err != nil {
		return err
	}
	if _, err := w.WriteString("px;\" " + constants.AttrWidth + "=\""); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(innerWidthPx)); err != nil {
		return err
	}
	if _, err := w.WriteString("\" ><tr><td style=\"line-height:0px;font-size:0px;mso-line-height-rule:exactly;\"><![endif]-->"); err != nil {
		return err
	}
	return nil
}

// RenderMSOWrapperTableClose renders MSO wrapper table closing directly to Writer
func RenderMSOWrapperTableClose(w io.StringWriter) error {
        return RenderMSOConditional(w, "</td></tr></table></td></tr></table>")
}

// RenderMSOSectionTransition renders MSO conditional comment that bridges between sections in a wrapper.
// This generates the pattern: <!--[if mso | IE]></td></tr><tr><td width="600px"><![endif]-->
// widthPx should typically be the body width (600 by default).
func RenderMSOSectionTransition(w io.StringWriter, widthPx int, align string) error {
	return RenderMSOSectionTransitionWithContent(w, widthPx, align, nil)
}

// RenderMSOSectionTransitionWithContent renders an MSO section transition that can inject
// additional content (e.g. mj-raw) inside the conditional comment before reopening the table row.
//
// It produces the sequence: <!--[if mso | IE]></td></tr>{content}<tr><td width="XXXpx"><![endif]-->
// where {content} is rendered via the provided callback while the conditional comment is still open.
func RenderMSOSectionTransitionWithContent(w io.StringWriter, widthPx int, align string, renderContent func(io.StringWriter) error) error {
	if _, err := w.WriteString("<!--[if mso | IE]></td></tr>"); err != nil {
		return err
	}

	if renderContent != nil {
		if err := renderContent(w); err != nil {
			return err
		}
	}

	if _, err := w.WriteString("<tr><td"); err != nil {
		return err
	}
	if align != "" {
		if _, err := w.WriteString(" " + constants.AttrAlign + "=\""); err != nil {
			return err
		}
		if _, err := w.WriteString(align + "\""); err != nil {
			return err
		}
	}
	if _, err := w.WriteString(" " + constants.AttrWidth + "=\""); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(widthPx)); err != nil {
		return err
	}
	if _, err := w.WriteString("px\"><![endif]-->"); err != nil {
		return err
	}
	return nil
}

// RenderMSOGroupTDOpen renders MSO group TD opening directly to Writer without string allocation
// RenderMSOGroupTableOpen renders the opening Outlook table wrapper for mj-group components.
//
// It produces the following structure in a single conditional comment to match the MJML reference output:
//
//	<!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="" style="width:600px;" ><![endif]-->
//
// The width is passed in pixels (without the unit) to ensure deterministic formatting and to avoid
// floating point rounding differences. The background color, when present, is applied on the inner
// Outlook table that wraps mj-column children (see RenderMSOGroupTDOpen) to mirror MJML's output.
func RenderMSOGroupTableOpen(w io.StringWriter, widthPx int, backgroundColor, outlookClass, verticalAlign string) error {
	if _, err := w.WriteString("<!--[if mso | IE]><table role=\"presentation\" border=\"0\" cellpadding=\"0\" cellspacing=\"0\""); err != nil {
		return err
	}
	if _, err := w.WriteString("><tr><td class=\""); err != nil {
		return err
	}
	if _, err := w.WriteString(outlookClass); err != nil {
		return err
	}
	if _, err := w.WriteString("\" style=\""); err != nil {
		return err
	}
	if verticalAlign != "" {
		if _, err := w.WriteString("vertical-align:"); err != nil {
			return err
		}
		if _, err := w.WriteString(verticalAlign); err != nil {
			return err
		}
		if _, err := w.WriteString(";"); err != nil {
			return err
		}
	}
	if _, err := w.WriteString("width:"); err != nil {
		return err
	}
	if _, err := w.WriteString(strconv.Itoa(widthPx)); err != nil {
		return err
	}
	if _, err := w.WriteString("px;\" ><![endif]-->"); err != nil {
		return err
	}
	return nil
}

// RenderMSOGroupTableClose renders the closing Outlook wrapper for mj-group components, matching the
// single conditional comment produced by the MJML reference implementation.
func RenderMSOGroupTableClose(w io.StringWriter) error {
	return RenderMSOConditional(w, "</td></tr></table>")
}

// RenderMSOGroupTDOpen renders the Outlook-specific table structure for each mj-column inside an mj-group.
//
// It generates the following markup to ensure both the table and td are wrapped inside the same MSO
// conditional comment, exactly like MJML's Node implementation:
//
//	<!--[if mso | IE]><table border="0" cellpadding="0" cellspacing="0" role="presentation"><tr><td style="vertical-align:top;width:600px;" ><![endif]-->
//
// The classAttr parameter allows optional attributes (for now it is typically empty, but keeping the
// parameter provides flexibility for future parity work). The width should include the unit (e.g. "600px").
//
// The backgroundColor argument mirrors MJML by applying the color to the Outlook table once for the
// first column, ensuring subsequent columns reuse the same table without duplicating attributes.
func RenderMSOGroupTDOpen(w io.StringWriter, classAttr, verticalAlign, widthPx, backgroundColor string, isFirst bool) error {
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}
	if isFirst {
		if _, err := w.WriteString("<table"); err != nil {
			return err
		}
		if backgroundColor != "" {
			if _, err := w.WriteString(" bgcolor=\""); err != nil {
				return err
			}
			if _, err := w.WriteString(backgroundColor); err != nil {
				return err
			}
			if _, err := w.WriteString("\""); err != nil {
				return err
			}
		}
		if _, err := w.WriteString(" border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\""); err != nil {
			return err
		}
		if _, err := w.WriteString(" ><tr><td"); err != nil {
			return err
		}
	} else {
		if _, err := w.WriteString("</td><td"); err != nil {
			return err
		}
	}
	if classAttr != "" {
		if _, err := w.WriteString(" "); err != nil {
			return err
		}
		if _, err := w.WriteString(classAttr); err != nil {
			return err
		}
	}
	if _, err := w.WriteString(" style=\"vertical-align:"); err != nil {
		return err
	}
	if _, err := w.WriteString(verticalAlign); err != nil {
		return err
	}
	if _, err := w.WriteString(";width:"); err != nil {
		return err
	}
	if _, err := w.WriteString(widthPx); err != nil {
		return err
	}
	if _, err := w.WriteString(";\" ><![endif]-->"); err != nil {
		return err
	}
	return nil
}

// RenderMSOGroupTDClose renders the Outlook-specific closing tags for an mj-column inside an mj-group.
func RenderMSOGroupTDClose(w io.StringWriter, isLast bool) error {
	if !isLast {
		return nil
	}
	return RenderMSOConditional(w, "</td></tr></table>")
}
