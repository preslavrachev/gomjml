package components

import (
	"io"
	"strconv"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/mjml/styles"
	"github.com/preslavrachev/gomjml/parser"
)

// MJSectionComponent represents mj-section
type MJSectionComponent struct {
	*BaseComponent
}

// NewMJSectionComponent creates a new mj-section component
func NewMJSectionComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJSectionComponent {
	return &MJSectionComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJSectionComponent) GetTagName() string {
	return "mj-section"
}

// Render implements optimized Writer-based rendering for MJSectionComponent
func (c *MJSectionComponent) Render(w io.StringWriter) error {
	// Get section attributes using proper attribute resolution (includes mj-attributes)
	// Cache all attribute lookups at once to avoid repeated calls
	backgroundColor := c.GetAttributeWithDefault(c, "background-color")
	backgroundUrl := c.GetAttributeWithDefault(c, constants.MJMLBackgroundUrl)
	backgroundPosition := c.GetAttributeWithDefault(c, "background-position")
	backgroundPositionX := c.GetAttributeWithDefault(c, "background-position-x")
	backgroundPositionY := c.GetAttributeWithDefault(c, "background-position-y")
	backgroundRepeat := c.GetAttributeWithDefault(c, "background-repeat")
	backgroundSize := c.GetAttributeWithDefault(c, "background-size")
	padding := c.GetAttributeWithDefault(c, "padding")
	direction := c.GetAttributeWithDefault(c, "direction")
	textAlign := c.GetAttributeWithDefault(c, "text-align")
	fullWidth := c.GetAttributeWithDefault(c, "full-width")
	borderRadius := c.GetAttributeWithDefault(c, "border-radius")
	align := c.GetAttributeWithDefault(c, "align")

	// Check if we have a background image for VML generation
	hasBackgroundImage := backgroundUrl != ""

	// For full-width sections, add an outer table wrapper like MRML does.
	// This wrapper is always present for full-width sections, even when no
	// background is specified, ensuring proper structure and alignment.
	if fullWidth != "" {
		outerTable := html.NewTableTag().
			AddAttribute("align", "center").
			AddStyle("width", "100%")

		// Apply background styles properly for full-width sections
		if backgroundUrl != "" {
			// Use shorthand and explicit longhand properties for full-width background images
			posX, posY := parseBackgroundPosition(backgroundPosition)
			posX, posY = overridePosition(posX, posY, backgroundPositionX, backgroundPositionY)
			shorthandBg := buildBackgroundShorthand(backgroundColor, backgroundUrl, posX, posY, backgroundSize, backgroundRepeat)
			if shorthandBg != "" {
				outerTable.AddStyle("background", shorthandBg)
				outerTable.AddStyle("background-position", posX+" "+posY)
				outerTable.AddStyle("background-repeat", backgroundRepeat)
				outerTable.AddStyle("background-size", backgroundSize)
				// Also add the background attribute for email client compatibility (use same encoding as VML src)
				outerTable.AddAttribute("background", htmlEscape(backgroundUrl))
			}
		} else if backgroundColor != "" {
			// Apply background color only when provided
			c.ApplyBackgroundStyles(outerTable, c)
		}
		// Only border-radius applies to the outer table. Border
		// properties belong to the inner content container.
		if borderRadius != "" {
			outerTable.AddStyle("border-radius", borderRadius)
		}

		if err := outerTable.RenderOpen(w); err != nil {
			return err
		}
		if _, err := w.WriteString("<tbody><tr><td>"); err != nil {
			return err
		}

		// Write VML opening if we have background image (inside full-width outer table TD)
		if hasBackgroundImage {
			if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
				return err
			}
			// Parse background position
			posX, posY := parseBackgroundPosition(backgroundPosition)
			posX, posY = overridePosition(posX, posY, backgroundPositionX, backgroundPositionY)

			// Compute VML attributes
			vOriginX, vOriginY, vPosX, vPosY := computeVMLPosition(posX, posY, backgroundSize, backgroundRepeat)
			vSizeAttrs, vAspect := computeVMLSize(backgroundSize)
			vmlType := computeVMLType(backgroundRepeat, backgroundSize)

			// Build VML attributes
			sizeFragment := ""
			if vSizeAttrs != "" {
				sizeFragment = " " + vSizeAttrs
			}
			aspectFragment := ""
			if vAspect != "" {
				aspectFragment = ` aspect="` + vAspect + `"`
			}

			// Build VML strings
			colorFragment := ""
			if backgroundColor != "" {
				colorFragment = ` color="` + backgroundColor + `"`
			}
			// Note: VML color attribute is only included when backgroundColor is explicitly set

			vmlOpen := `<v:rect mso-width-percent="1000" xmlns:v="urn:schemas-microsoft-com:vml" fill="true" stroke="false"><v:fill origin="` + vOriginX + `, ` + vOriginY +
				`" position="` + vPosX + `, ` + vPosY + `" src="` + htmlEscape(backgroundUrl) + `"` + colorFragment +
				` type="` + vmlType + `"` + sizeFragment + aspectFragment +
				` /><v:textbox style="mso-fit-shape-to-text:true" inset="0,0,0,0"><![endif]-->`

			if _, err := w.WriteString(vmlOpen); err != nil {
				return err
			}
		}
	}

	// MSO conditional comment - table wrapper for Outlook (inside VML textbox if present)
	// Use full container width for MSO table as per MJML spec - padding only affects inner content
	msoTableWidth := c.GetEffectiveWidth()

	// Get align from attributes (including mj-class)
	alignAttr := align
	if alignAttr == "" {
		alignAttr = "center" // default align for MSO table
	}

	useMJMLSyntax := c.RenderOpts != nil && c.RenderOpts.UseMJMLSyntax
	insideWrapper := c.RenderOpts != nil && c.RenderOpts.InsideWrapper
	skipSectionMSOTable := useMJMLSyntax && insideWrapper

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

		// Custom MSO conditional
	continueMSOComment := c.RenderOpts != nil && c.RenderOpts.PendingMSOSectionClose

	if useMJMLSyntax && !skipSectionMSOTable {
		cssClassOutlook := ""
		if cssClass := c.GetCSSClass(); cssClass != "" {
			cssClassOutlook = cssClass + "-outlook"
		}

		if continueMSOComment {
			if c.RenderOpts != nil {
				c.RenderOpts.PendingMSOSectionClose = false
			}
			if _, err := w.WriteString(`<table`); err != nil {
				return err
			}
		} else {
			if _, err := w.WriteString(`<!--[if mso | IE]><table`); err != nil {
				return err
			}
		}
		if alignAttr != "" {
			if _, err := w.WriteString(` align="` + alignAttr + `"`); err != nil {
				return err
			}
		}
		if _, err := w.WriteString(` border="0" cellpadding="0" cellspacing="0"`); err != nil {
			return err
		}
		if _, err := w.WriteString(` class="` + cssClassOutlook + `"`); err != nil {
			return err
		}
		if _, err := w.WriteString(` role="presentation"`); err != nil {
			return err
		}
		if _, err := w.WriteString(` style="width:` + getPixelWidthString(msoTableWidth) + `;"`); err != nil {
			return err
		}
		if _, err := w.WriteString(` width="` + strconv.Itoa(msoTableWidth) + `"`); err != nil {
			return err
		}
		if backgroundColor != "" {
			if _, err := w.WriteString(` bgcolor="` + backgroundColor + `"`); err != nil {
				return err
			}
		}
		if _, err := w.WriteString(` >`); err != nil {
			return err
		}
		if _, err := w.WriteString(`<tr>`); err != nil {
			return err
		}
		if err := msoTd.RenderOpen(w); err != nil {
			return err
		}
	} else if !useMJMLSyntax {
		msoTable := html.NewTableTag()

		// Add bgcolor before align/width to match MRML attribute order
		if backgroundColor != "" {
			msoTable.AddAttribute("bgcolor", backgroundColor)
		}

		msoTable.AddAttribute("align", alignAttr).
			AddAttribute("width", strconv.Itoa(msoTableWidth)).
			AddStyle("width", getPixelWidthString(msoTableWidth))

		// Add css-class-outlook if present
		if cssClass := c.GetCSSClass(); cssClass != "" {
			msoTable.AddAttribute("class", cssClass+"-outlook")
		}

		if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
			return err
		}
		if err := msoTable.RenderOpen(w); err != nil {
			return err
		}
		if _, err := w.WriteString("<tr>"); err != nil {
			return err
		}
		if err := msoTd.RenderOpen(w); err != nil {
			return err
		}
	}

	if hasBackgroundImage && fullWidth == "" {
		posX, posY := parseBackgroundPosition(backgroundPosition)
		posX, posY = overridePosition(posX, posY, backgroundPositionX, backgroundPositionY)
		vOriginX, vOriginY, vPosX, vPosY := computeVMLPosition(posX, posY, backgroundSize, backgroundRepeat)
		vSizeAttrs, vAspect := computeVMLSize(backgroundSize)
		vmlType := computeVMLType(backgroundRepeat, backgroundSize)

		sizeFragment := ""
		if vSizeAttrs != "" {
			sizeFragment = " " + vSizeAttrs
		}
		aspectFragment := ""
		if vAspect != "" {
			aspectFragment = ` aspect="` + vAspect + `"`
		}
		colorFragment := ""
		if backgroundColor != "" {
			colorFragment = ` color="` + backgroundColor + `"`
		}

		vmlOpen := `<v:rect style="width:` + strconv.Itoa(msoTableWidth) + `px;" xmlns:v="urn:schemas-microsoft-com:vml" fill="true" stroke="false"><v:fill origin="` + vOriginX + `, ` + vOriginY + `" position="` + vPosX + `, ` + vPosY + `" src="` + htmlEscape(backgroundUrl) + `"` + colorFragment + ` type="` + vmlType + `"` + sizeFragment + aspectFragment + ` /><v:textbox style="mso-fit-shape-to-text:true" inset="0,0,0,0">`
		if _, err := w.WriteString(vmlOpen); err != nil {
			return err
		}
		if !skipSectionMSOTable {
			if _, err := w.WriteString("<![endif]-->"); err != nil {
				return err
			}
		}
	} else if !skipSectionMSOTable {
		if _, err := w.WriteString("<![endif]-->"); err != nil {
			return err
		}
	}

	// Main section div with styles
	sectionDiv := html.NewHTMLTag("div")
	c.AddDebugAttribute(sectionDiv, "section")

	// Add css-class if present
	c.SetClassAttribute(sectionDiv)

	// Background on main section div (MRML behavior):
	// - When not full-width and we have a background image, use shorthand background
	//   and explicitly set position/repeat/size (no extra longhands for color/image).
	// - When only background color is present (no image) and not full-width, apply color.
	if fullWidth == "" {
		if backgroundUrl != "" {
			posX, posY := parseBackgroundPosition(backgroundPosition)
			posX, posY = overridePosition(posX, posY, backgroundPositionX, backgroundPositionY)
			shorthandBg := buildBackgroundShorthand(backgroundColor, backgroundUrl, posX, posY, backgroundSize, backgroundRepeat)
			if shorthandBg != "" {
				sectionDiv.AddStyle("background", shorthandBg)
				// Add the explicit longhand properties to match MRML output
				sectionDiv.AddStyle("background-position", posX+" "+posY)
				sectionDiv.AddStyle("background-repeat", backgroundRepeat)
				sectionDiv.AddStyle("background-size", backgroundSize)
			}
		} else if backgroundColor != "" {
			// Color-only background
			c.ApplyBackgroundStyles(sectionDiv, c)
		}
	}

	// AIDEV-NOTE: width-calculation-critical; section max-width behavior
	// 1. Use containerWidth (set by parent wrapper) - already accounts for wrapper horizontal padding
	// 2. Section's own padding is internal spacing and must NOT reduce the section's max-width
	// 3. Wrapper padding="20px" → child section gets containerWidth=560px → max-width:560px
	// 4. Section padding="15px" → section keeps full containerWidth for max-width, padding affects inner content only
	sectionDiv.AddStyle("margin", "0px auto").
		AddStyle("max-width", strconv.Itoa(c.GetContainerWidth())+"px")

	// Add border-radius if specified
	if borderRadius != "" {
		sectionDiv.AddStyle("border-radius", borderRadius)
	}

	if err := sectionDiv.RenderOpen(w); err != nil {
		return err
	}

	// Add intermediate div wrapper when we have background image (matches MRML structure)
	var intermediateDiv *html.HTMLTag
	if backgroundUrl != "" {
		intermediateDiv = html.NewHTMLTag("div").
			AddStyle("line-height", "0").
			// Match MRML: font-size should be 0 (unitless), not 0px
			AddStyle("font-size", "0")
		if err := intermediateDiv.RenderOpen(w); err != nil {
			return err
		}
	}

	// Inner table with styles
	innerTable := html.NewTableTag().
		AddAttribute("align", "center")

	// Apply background styles to inner table (only for non-full-width sections)
	if fullWidth == "" && backgroundUrl != "" {
		// Use shorthand and explicit longhand properties (avoid extra background-color/image longhands)
		posX, posY := parseBackgroundPosition(backgroundPosition)
		posX, posY = overridePosition(posX, posY, backgroundPositionX, backgroundPositionY)
		shorthandBg := buildBackgroundShorthand(backgroundColor, backgroundUrl, posX, posY, backgroundSize, backgroundRepeat)
		if shorthandBg != "" {
			innerTable.AddStyle("background", shorthandBg)
			innerTable.AddStyle("background-position", posX+" "+posY)
			innerTable.AddStyle("background-repeat", backgroundRepeat)
			innerTable.AddStyle("background-size", backgroundSize)
			// Also add the background attribute for email client compatibility (use same encoding as VML src)
			innerTable.AddAttribute("background", htmlEscape(backgroundUrl))
		}
	} else if fullWidth == "" {
		// No background image: apply defaults (color-only etc.)
		if backgroundColor == "" || (backgroundColor != "" && fullWidth == "") {
			c.ApplyBackgroundStyles(innerTable, c)
		}
	}

	// Then add width and border-radius
	innerTable.AddStyle("width", "100%")

	// Add border-radius if specified
	if borderRadius != "" {
		innerTable.AddStyle("border-radius", borderRadius)
	}

	if err := innerTable.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody><tr>"); err != nil {
		return err
	}

	// TD with padding and text alignment
	tdTag := html.NewHTMLTag("td").
		AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding)

	// Add specific padding overrides in MRML order: left, right, bottom, top
	if paddingLeftAttr := c.GetAttribute(constants.MJMLPaddingLeft); paddingLeftAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingLeft, *paddingLeftAttr)
	}
	if paddingRightAttr := c.GetAttribute(constants.MJMLPaddingRight); paddingRightAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingRight, *paddingRightAttr)
	}
	if paddingBottomAttr := c.GetAttribute(constants.MJMLPaddingBottom); paddingBottomAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingBottom, *paddingBottomAttr)
	}
	if paddingTopAttr := c.GetAttribute(constants.MJMLPaddingTop); paddingTopAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingTop, *paddingTopAttr)
	}

	// Apply border styles to the content container. Global attributes for
	// mj-section define border properties that should apply to the inner
	// TD rather than the wrapping tables.
	toPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	styles.ApplyBorderStyles(tdTag,
		toPtr(c.GetAttributeFast(c, constants.MJMLBorder)),
		nil,
		toPtr(c.GetAttributeFast(c, "border-top")),
		toPtr(c.GetAttributeFast(c, "border-right")),
		toPtr(c.GetAttributeFast(c, "border-bottom")),
		toPtr(c.GetAttributeFast(c, "border-left")),
	)

	tdTag.AddStyle("text-align", textAlign)

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// Always render inner MSO table wrapper for section content (MRML behavior)
	// This provides the MSO table structure that content (including comments) sits within
	// Only add inner MSO table wrapper for sections that have ONLY text content (no child components)
	// Sections with columns/components already get MSO tables from their children
	textContent := c.Node.Text
	trimmedText := strings.TrimSpace(textContent)
	hasTextContent := trimmedText != "" || (textContent != "" && !strings.Contains(textContent, "\n"))
	hasChildContent := len(c.Children) > 0
	needsContentMSOTable := hasTextContent && !hasChildContent

	if needsContentMSOTable {
		// Section has only text/comments - needs MSO table wrapper
		if c.RenderOpts.InsideWrapper {
			// Inside wrapper: use split conditional pattern for text content
			if useMJMLSyntax {
				if _, err := w.WriteString(`<!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><![endif]-->`); err != nil {
					return err
				}
			} else {
				if _, err := w.WriteString("<!--[if mso | IE]><table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr><![endif]-->"); err != nil {
					return err
				}
			}
			// Render text content
			if _, err := w.WriteString(c.Node.Text); err != nil {
				return err
			}
			if _, err := w.WriteString("<!--[if mso | IE]></tr></table><![endif]-->"); err != nil {
				return err
			}
		} else {
			// Standalone: use simple pattern for text content
			innerMsoTable := html.NewHTMLTag("table").
				AddAttribute(constants.AttrRole, "presentation").
				AddAttribute("border", "0").
				AddAttribute("cellpadding", "0").
				AddAttribute("cellspacing", "0")
			innerMsoTr := html.NewHTMLTag("tr")

			if useMJMLSyntax {
				if _, err := w.WriteString(`<!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><![endif]-->`); err != nil {
					return err
				}
			} else {
				if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
					return err
				}
				if err := innerMsoTable.RenderOpen(w); err != nil {
					return err
				}
				if err := innerMsoTr.RenderOpen(w); err != nil {
					return err
				}
				if _, err := w.WriteString("<![endif]-->"); err != nil {
					return err
				}
			}

			// Render text content (including comments) - goes directly in TR, no TD
			if _, err := w.WriteString(c.Node.Text); err != nil {
				return err
			}
			if useMJMLSyntax {
				if _, err := w.WriteString("<!--[if mso | IE]></tr></table><![endif]-->"); err != nil {
					return err
				}
			} else {
				if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
					return err
				}
				if err := innerMsoTr.RenderClose(w); err != nil {
					return err
				}
				if err := innerMsoTable.RenderClose(w); err != nil {
					return err
				}
				if _, err := w.WriteString("<![endif]-->"); err != nil {
					return err
				}
			}
		}
	} else if !hasChildContent && !hasTextContent {
		// Empty section: MJML emits a single MSO conditional wrapper containing an empty table.
		// Match that exact output to avoid duplicated conditional comment pairs.
		innerMsoTable := html.NewHTMLTag("table").
			AddAttribute(constants.AttrRole, "presentation").
			AddAttribute("border", "0").
			AddAttribute("cellpadding", "0").
			AddAttribute("cellspacing", "0")

		var innerContent strings.Builder
		if err := innerMsoTable.RenderOpen(&innerContent); err != nil {
			return err
		}
		if _, err := innerContent.WriteString("<tr></tr>"); err != nil {
			return err
		}
		if err := innerMsoTable.RenderClose(&innerContent); err != nil {
			return err
		}

		if err := html.RenderMSOConditional(w, innerContent.String()); err != nil {
			return err
		}
	}

	// Calculate sibling counts for width calculations (following MRML logic)
	siblings := len(c.Children)
	rawSiblings := 0
	columnCount := 0
	for _, child := range c.Children {
		if child.IsRawElement() {
			rawSiblings++
		}
		if _, ok := child.(*MJColumnComponent); ok {
			columnCount++
		}
	}

	// Outlook expects a shared table wrapper even when a section only contains mj-raw blocks.
	// Match MRML by opening the wrapper when we have multiple column children or any raw children.
	needsSharedMSOTable := columnCount > 1 || rawSiblings > 0

	// Shared MSO table management mirrors MJML's Outlook markup, where the
	// opening <table><tr> sequence lives in the first conditional block.
	// Raw-only sections still use the original wrapper pattern.
	var sharedMsoTable *html.HTMLTag
	var sharedMsoTr *html.HTMLTag
	sharedTableOpenedForColumns := false
	sharedTableOpenedForRaw := false

	if needsSharedMSOTable {
		sharedMsoTable = html.NewHTMLTag("table").
			AddAttribute(constants.AttrRole, "presentation").
			AddAttribute("border", "0").
			AddAttribute("cellpadding", "0").
			AddAttribute("cellspacing", "0")
		sharedMsoTr = html.NewHTMLTag("tr")

		if columnCount == 0 {
			if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
				return err
			}
			if err := sharedMsoTable.RenderOpen(w); err != nil {
				return err
			}
			if err := sharedMsoTr.RenderOpen(w); err != nil {
				return err
			}
			if _, err := w.WriteString("<![endif]-->"); err != nil {
				return err
			}
			sharedTableOpenedForRaw = true
		}
	}

	// Compute the effective content width after accounting for horizontal padding.
	innerContentWidth := c.getInnerContentWidth()

	// Render child columns and groups (section provides shared MSO TR, columns provide MSO TDs)
	// AIDEV-NOTE: width-flow-start; section initiates width flow by passing effective width to columns
	for _, child := range c.Children {
		if child.IsRawElement() {
			if err := child.Render(w); err != nil {
				return err
			}
			continue
		}

		// Pass the effective width and sibling counts to the child
		child.SetContainerWidth(innerContentWidth)
		child.SetSiblings(siblings)
		child.SetRawSiblings(rawSiblings)

		// Generate MSO TD for each column (within shared MSO table)
		if columnComp, ok := child.(*MJColumnComponent); ok {
			getAttr := func(name string) string {
				if attr := columnComp.GetAttribute(name); attr != nil {
					return *attr
				}
				return columnComp.GetDefaultAttribute(name)
			}

			if needsSharedMSOTable && columnCount > 0 {
				cssClass := ""
				if css := columnComp.GetAttribute(constants.MJMLCSSClass); css != nil {
					cssClass = *css
				}
				outlookClass := cssClass + "-outlook"
				if cssClass == "" {
					outlookClass = ""
				}

				msoTd := html.NewHTMLTag("td").
					AddAttribute("class", outlookClass).
					AddStyle("vertical-align", getAttr("vertical-align")).
					AddStyle("width", columnComp.GetWidthAsPixel())

				if !sharedTableOpenedForColumns {
					if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
						return err
					}
					if err := sharedMsoTable.RenderOpen(w); err != nil {
						return err
					}
					if err := sharedMsoTr.RenderOpen(w); err != nil {
						return err
					}
					sharedTableOpenedForColumns = true
				} else {
					if _, err := w.WriteString("<!--[if mso | IE]></td>"); err != nil {
						return err
					}
				}

				var tdOpen strings.Builder
				if err := msoTd.RenderOpen(&tdOpen); err != nil {
					return err
				}
				tdString := tdOpen.String()
				if strings.HasSuffix(tdString, ">") {
					tdString = tdString[:len(tdString)-1] + " >"
				}
				if _, err := w.WriteString(tdString); err != nil {
					return err
				}
				if _, err := w.WriteString("<![endif]-->"); err != nil {
					return err
				}

				if err := columnComp.Render(w); err != nil {
					return err
				}

				continue
			}

			if useMJMLSyntax && columnCount == 1 {
				cssClass := ""
				if css := columnComp.GetAttribute(constants.MJMLCSSClass); css != nil {
					cssClass = *css
				}
				outlookClass := cssClass + "-outlook"
				if cssClass == "" {
					outlookClass = ""
				}

				if _, err := w.WriteString(`<!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="`); err != nil {
					return err
				}
				if _, err := w.WriteString(outlookClass); err != nil {
					return err
				}
				if _, err := w.WriteString(`" style="vertical-align:` + getAttr("vertical-align") + `;width:` + columnComp.GetWidthAsPixel() + `;" ><![endif]-->`); err != nil {
					return err
				}

				if err := columnComp.Render(w); err != nil {
					return err
				}

				if _, err := w.WriteString(`<!--[if mso | IE]></td></tr></table><![endif]-->`); err != nil {
					return err
				}

				continue
			}

			msoTd := html.NewHTMLTag("td")
			// Add class attribute if css-class is set (with -outlook suffix)
			if css := columnComp.GetAttribute(constants.MJMLCSSClass); css != nil && *css != "" {
				msoTd.AddAttribute("class", *css+"-outlook")
			}

			// Add styles in MRML insertion order: vertical-align first, then width
			msoTd.AddStyle("vertical-align", getAttr("vertical-align"))
			msoTd.AddStyle("width", columnComp.GetWidthAsPixel())

			if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
				return err
			}
			var tdOpen strings.Builder
			if err := msoTd.RenderOpen(&tdOpen); err != nil {
				return err
			}
			tdString := tdOpen.String()
			if strings.HasSuffix(tdString, ">") {
				tdString = tdString[:len(tdString)-1] + " >"
			}
			if _, err := w.WriteString(tdString); err != nil {
				return err
			}
			if _, err := w.WriteString("<![endif]-->"); err != nil {
				return err
			}

			if err := columnComp.Render(w); err != nil {
				return err
			}

			if _, err := w.WriteString("<!--[if mso | IE]></td><![endif]-->"); err != nil {
				return err
			}

			continue
		}

		if groupComp, ok := child.(*MJGroupComponent); ok {
			// Ensure the group receives the effective section width for its internal calculations
			groupComp.SetContainerWidth(innerContentWidth)
		}

		// Use optimized rendering with fallback to string-based
		if err := child.Render(w); err != nil {
			return err
		}
	}

	// Close shared MSO table structure for columns
	if needsSharedMSOTable {
		if sharedTableOpenedForColumns {
			if _, err := w.WriteString("<!--[if mso | IE]></td></tr></table><![endif]-->"); err != nil {
				return err
			}
		} else if sharedTableOpenedForRaw {
			if _, err := w.WriteString("<!--[if mso | IE]></tr></table><![endif]-->"); err != nil {
				return err
			}
		}
	}

	if err := tdTag.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr></tbody>"); err != nil {
		return err
	}
	if err := innerTable.RenderClose(w); err != nil {
		return err
	}

	// Close intermediate div if we added one
	if intermediateDiv != nil {
		if err := intermediateDiv.RenderClose(w); err != nil {
			return err
		}
	}

	if err := sectionDiv.RenderClose(w); err != nil {
		return err
	}

	// Close MSO table structure
	if hasBackgroundImage && fullWidth == "" {
		if skipSectionMSOTable {
			if _, err := w.WriteString("<!--[if mso | IE]></v:textbox></v:rect><![endif]-->"); err != nil {
				return err
			}
		} else {
			if _, err := w.WriteString("<!--[if mso | IE]></v:textbox></v:rect></td></tr></table><![endif]-->"); err != nil {
				return err
			}
		}
	} else if !skipSectionMSOTable {
		if c.RenderOpts != nil && c.RenderOpts.RemainingBodySections > 0 {
			if _, err := w.WriteString("<!--[if mso | IE]></td></tr></table>"); err != nil {
				return err
			}
			c.RenderOpts.PendingMSOSectionClose = true
		} else {
			if _, err := w.WriteString("<!--[if mso | IE]></td></tr></table><![endif]-->"); err != nil {
				return err
			}
			if c.RenderOpts != nil {
				c.RenderOpts.PendingMSOSectionClose = false
			}
		}
	}

	// Close outer table if we added one for full-width sections
	if fullWidth != "" {
		// Close VML first if present, then outer table
		if hasBackgroundImage {
			if _, err := w.WriteString("<!--[if mso | IE]></v:textbox></v:rect><![endif]-->"); err != nil {
				return err
			}
		}
		if _, err := w.WriteString("</td></tr></tbody></table>"); err != nil {
			return err
		}
	}

	return nil
}

func (c *MJSectionComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "background-color":
		return ""
	case constants.MJMLBackgroundUrl:
		return ""
	case "background-position":
		return "top center"
	case "background-position-x":
		return ""
	case "background-position-y":
		return ""
	case "background-repeat":
		return "repeat"
	case "background-size":
		return "auto"
	case "direction":
		return "ltr"
	case "full-width":
		return ""
	case "padding":
		return "20px 0"
	case "text-align":
		return "center"
	case "text-padding":
		return "4px 4px 4px 0"
	default:
		return ""
	}
}

// getInnerContentWidth calculates the inner content width for the section after accounting for
// horizontal padding overrides. The value is used for width propagation to child columns/groups
// so MSO fallback tables match MJML's Outlook output.
func (c *MJSectionComponent) getInnerContentWidth() int {
	effectiveWidth := c.GetEffectiveWidth()
	paddingValue := c.GetAttributeWithDefault(c, "padding")

	var spacing *styles.Spacing
	if paddingValue != "" {
		if parsed, err := styles.ParseSpacing(paddingValue); err == nil && parsed != nil {
			spacing = parsed
			effectiveWidth -= int(parsed.Left + parsed.Right)
		}
	}

	if paddingLeftAttr := c.GetAttribute(constants.MJMLPaddingLeft); paddingLeftAttr != nil && *paddingLeftAttr != "" {
		if px, err := styles.ParsePixel(*paddingLeftAttr); err == nil && px != nil {
			if spacing != nil {
				effectiveWidth += int(spacing.Left)
			}
			effectiveWidth -= int(px.Value)
		}
	}

	if paddingRightAttr := c.GetAttribute(constants.MJMLPaddingRight); paddingRightAttr != nil && *paddingRightAttr != "" {
		if px, err := styles.ParsePixel(*paddingRightAttr); err == nil && px != nil {
			if spacing != nil {
				effectiveWidth += int(spacing.Right)
			}
			effectiveWidth -= int(px.Value)
		}
	}

	if effectiveWidth <= 0 {
		return c.GetEffectiveWidth()
	}
	return effectiveWidth
}
