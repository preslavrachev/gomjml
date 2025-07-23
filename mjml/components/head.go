package components

import "github.com/preslavrachev/gomjml/parser"

// MJHeadComponent represents mj-head
type MJHeadComponent struct {
	*BaseComponent
}

// NewMJHeadComponent creates a new mj-head component
func NewMJHeadComponent(node *parser.MJMLNode) *MJHeadComponent {
	return &MJHeadComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJHeadComponent) Render() (string, error) {
	return "", nil // Head is handled in MJML component
}

func (c *MJHeadComponent) GetTagName() string {
	return "mj-head"
}

func (c *MJHeadComponent) GetDefaultAttribute(name string) string {
	return ""
}

// MJTitleComponent represents mj-title
type MJTitleComponent struct {
	*BaseComponent
}

// NewMJTitleComponent creates a new mj-title component
func NewMJTitleComponent(node *parser.MJMLNode) *MJTitleComponent {
	return &MJTitleComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJTitleComponent) Render() (string, error) {
	return "", nil // Title is handled in MJML component head processing
}

func (c *MJTitleComponent) GetTagName() string {
	return "mj-title"
}

func (c *MJTitleComponent) GetDefaultAttribute(name string) string {
	return ""
}

// MJFontComponent represents mj-font
type MJFontComponent struct {
	*BaseComponent
}

// NewMJFontComponent creates a new mj-font component
func NewMJFontComponent(node *parser.MJMLNode) *MJFontComponent {
	return &MJFontComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJFontComponent) Render() (string, error) {
	return "", nil // Font is handled in MJML component head processing
}

func (c *MJFontComponent) GetTagName() string {
	return "mj-font"
}

func (c *MJFontComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "name":
		return ""
	case "href":
		return ""
	default:
		return ""
	}
}
