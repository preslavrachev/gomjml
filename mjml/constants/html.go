package constants

// Additional HTML Attributes - attributes not already defined in css.go
const (
	// Email-specific table attributes
	AttrBackground    = "background"
	AttrBgcolor       = "bgcolor"
	AttrVerticalAlign = "vertical-align"

	// Image attributes
	AttrUsemap = "usemap"
	AttrIsmap  = "ismap"

	// Form attributes
	AttrFor      = "for"
	AttrChecked  = "checked"
	AttrDisabled = "disabled"

	// Accessibility attributes
	AttrAriaLabel  = "aria-label"
	AttrAriaHidden = "aria-hidden"
	AttrTabindex   = "tabindex"

	// XML/Namespace attributes (for Outlook VML)
	AttrXmlns  = "xmlns"
	AttrXmlnsV = "xmlns:v"
	AttrXmlnsO = "xmlns:o"

	// Language and Direction constants
	// LangUndetermined represents "undetermined" language code per RFC 3066/BCP 47
	// Per emailmarkup.org: "It's not nearly as good as setting a language but
	// it's much better than setting nothing"
	LangUndetermined = "und"
	DirAuto          = "auto"
)
