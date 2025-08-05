package constants

// CSS Properties - commonly used CSS property names
const (
	// Layout & Box Model
	CSSPadding       = "padding"
	CSSPaddingTop    = "padding-top"
	CSSPaddingRight  = "padding-right"
	CSSPaddingBottom = "padding-bottom"
	CSSPaddingLeft   = "padding-left"
	CSSMargin        = "margin"
	CSSMarginTop     = "margin-top"
	CSSMarginRight   = "margin-right"
	CSSMarginBottom  = "margin-bottom"
	CSSMarginLeft    = "margin-left"
	CSSWidth         = "width"
	CSSHeight        = "height"
	CSSMaxWidth      = "max-width"
	CSSMinWidth      = "min-width"
	CSSBorder        = "border"
	CSSBorderRadius  = "border-radius"
	CSSBorderTop     = "border-top"
	CSSBorderRight   = "border-right"
	CSSBorderBottom  = "border-bottom"
	CSSBorderLeft    = "border-left"

	// Typography
	CSSFontFamily     = "font-family"
	CSSFontSize       = "font-size"
	CSSFontWeight     = "font-weight"
	CSSFontStyle      = "font-style"
	CSSLineHeight     = "line-height"
	CSSTextAlign      = "text-align"
	CSSTextDecoration = "text-decoration"
	CSSTextTransform  = "text-transform"
	CSSLetterSpacing  = "letter-spacing"
	CSSWordSpacing    = "word-spacing"
	CSSColor          = "color"

	// Background & Visual
	CSSBackground         = "background"
	CSSBackgroundColor    = "background-color"
	CSSBackgroundImage    = "background-image"
	CSSBackgroundSize     = "background-size"
	CSSBackgroundPosition = "background-position"
	CSSBackgroundRepeat   = "background-repeat"

	// Positioning & Display
	CSSDisplay       = "display"
	CSSPosition      = "position"
	CSSTop           = "top"
	CSSRight         = "right"
	CSSBottom        = "bottom"
	CSSLeft          = "left"
	CSSVerticalAlign = "vertical-align"
	CSSFloat         = "float"
	CSSDirection     = "direction"

	// Table-specific
	CSSBorderCollapse = "border-collapse"
	CSSTableLayout    = "table-layout"

	// Flexbox & Grid (for future use)
	CSSFlexDirection  = "flex-direction"
	CSSJustifyContent = "justify-content"
	CSSAlignItems     = "align-items"
)

// HTML Attributes - commonly used HTML attribute names
const (
	// Universal attributes
	AttrClass = "class"
	AttrStyle = "style"
	AttrID    = "id"
	AttrTitle = "title"
	AttrAlt   = "alt"

	// Table attributes
	AttrBorder      = "border"
	AttrCellPadding = "cellpadding"
	AttrCellSpacing = "cellspacing"
	AttrRole        = "role"
	AttrAlign       = "align"
	AttrValign      = "valign"
	AttrWidth       = "width"
	AttrHeight      = "height"

	// Link attributes
	AttrHref   = "href"
	AttrTarget = "target"
	AttrRel    = "rel"

	// Image attributes
	AttrSrc = "src"

	// Meta attributes
	AttrCharset   = "charset"
	AttrContent   = "content"
	AttrName      = "name"
	AttrHttpEquiv = "http-equiv"

	// Form attributes
	AttrType  = "type"
	AttrValue = "value"
)

// MJML-specific attributes - commonly used MJML component attributes
const (
	// Layout attributes
	MJMLPadding       = "padding"
	MJMLPaddingTop    = "padding-top"
	MJMLPaddingRight  = "padding-right"
	MJMLPaddingBottom = "padding-bottom"
	MJMLPaddingLeft   = "padding-left"
	MJMLAlign         = "align"
	MJMLVerticalAlign = "vertical-align"
	MJMLDirection     = "direction"
	MJMLWidth         = "width"
	MJMLHeight        = "height"

	// Typography attributes
	MJMLFontFamily     = "font-family"
	MJMLFontSize       = "font-size"
	MJMLFontWeight     = "font-weight"
	MJMLFontStyle      = "font-style"
	MJMLLineHeight     = "line-height"
	MJMLTextAlign      = "text-align"
	MJMLTextDecoration = "text-decoration"
	MJMLColor          = "color"

	// Background attributes
	MJMLBackgroundColor    = "background-color"
	MJMLBackgroundImage    = "background-image"
	MJMLBackgroundSize     = "background-size"
	MJMLBackgroundPosition = "background-position"
	MJMLBackgroundRepeat   = "background-repeat"

	// Border attributes
	MJMLBorder       = "border"
	MJMLBorderRadius = "border-radius"
	MJMLBorderTop    = "border-top"
	MJMLBorderRight  = "border-right"
	MJMLBorderBottom = "border-bottom"
	MJMLBorderLeft   = "border-left"

	// Component-specific attributes
	MJMLCSSClass                 = "css-class"
	MJMLContainerBackgroundColor = "container-background-color"
	MJMLInnerPadding             = "inner-padding"
	MJMLTextPadding              = "text-padding"
	MJMLIconSize                 = "icon-size"
	MJMLIconHeight               = "icon-height"
	MJMLIconPadding              = "icon-padding"
	MJMLTableLayout              = "table-layout"
	MJMLMode                     = "mode"
	MJMLName                     = "name"
	MJMLSrc                      = "src"
	MJMLHref                     = "href"
	MJMLTarget                   = "target"
	MJMLAlt                      = "alt"
	MJMLTitle                    = "title"
	MJMLFullWidth                = "full-width"
	MJMLFluidOnMobile            = "fluid-on-mobile"
)

// Common CSS values
const (
	// Display values
	DisplayBlock       = "block"
	DisplayInline      = "inline"
	DisplayInlineBlock = "inline-block"
	DisplayInlineTable = "inline-table"
	DisplayTable       = "table"
	DisplayTableCell   = "table-cell"
	DisplayTableRow    = "table-row"
	DisplayNone        = "none"

	// Text alignment values
	AlignLeft    = "left"
	AlignCenter  = "center"
	AlignRight   = "right"
	AlignJustify = "justify"

	// Vertical alignment values
	VAlignTop      = "top"
	VAlignMiddle   = "middle"
	VAlignBottom   = "bottom"
	VAlignBaseline = "baseline"

	// Border collapse values
	BorderCollapseCollapse = "collapse"
	BorderCollapseSeparate = "separate"

	// Target values
	TargetBlank  = "_blank"
	TargetSelf   = "_self"
	TargetParent = "_parent"
	TargetTop    = "_top"

	// Direction values
	DirectionLTR = "ltr"
	DirectionRTL = "rtl"

	// Background repeat values
	BackgroundRepeatRepeat   = "repeat"
	BackgroundRepeatNoRepeat = "no-repeat"
	BackgroundRepeatRepeatX  = "repeat-x"
	BackgroundRepeatRepeatY  = "repeat-y"

	// Background size values
	BackgroundSizeAuto    = "auto"
	BackgroundSizeContain = "contain"
	BackgroundSizeCover   = "cover"

	// Font weight values
	FontWeightNormal = "normal"
	FontWeightBold   = "bold"
	FontWeight100    = "100"
	FontWeight200    = "200"
	FontWeight300    = "300"
	FontWeight400    = "400"
	FontWeight500    = "500"
	FontWeight600    = "600"
	FontWeight700    = "700"
	FontWeight800    = "800"
	FontWeight900    = "900"

	// Font style values
	FontStyleNormal  = "normal"
	FontStyleItalic  = "italic"
	FontStyleOblique = "oblique"

	// Text decoration values
	TextDecorationNone        = "none"
	TextDecorationUnderline   = "underline"
	TextDecorationOverline    = "overline"
	TextDecorationLineThrough = "line-through"
)
