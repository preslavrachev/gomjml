package components

import (
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJNavbarComponent represents the mj-navbar component
type MJNavbarComponent struct {
	*BaseComponent
}

func NewMJNavbarComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJNavbarComponent {
	return &MJNavbarComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJNavbarComponent) Render() (string, error) {
	// TODO: Implement mj-navbar component functionality
	return "", &NotImplementedError{ComponentName: "mj-navbar"}
}

func (c *MJNavbarComponent) GetTagName() string {
	return "mj-navbar"
}

func (c *MJNavbarComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "align":
		return "center"
	case "ico-align":
		return "center"
	case "ico-close":
		return "&#8855;"
	case "ico-color":
		return "#000000"
	case "ico-font-family":
		return "Ubuntu, Helvetica, Arial, sans-serif"
	case "ico-font-size":
		return "30px"
	case "ico-line-height":
		return "30px"
	case "ico-open":
		return "&#9776;"
	case "ico-padding":
		return "10px"
	case "ico-text-decoration":
		return "none"
	case "ico-text-transform":
		return "uppercase"
	default:
		return ""
	}
}

// MJNavbarLinkComponent represents the mj-navbar-link component
type MJNavbarLinkComponent struct {
	*BaseComponent
}

func NewMJNavbarLinkComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJNavbarLinkComponent {
	return &MJNavbarLinkComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJNavbarLinkComponent) Render() (string, error) {
	// TODO: Implement mj-navbar-link component functionality
	return "", &NotImplementedError{ComponentName: "mj-navbar-link"}
}

func (c *MJNavbarLinkComponent) GetTagName() string {
	return "mj-navbar-link"
}

func (c *MJNavbarLinkComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "color":
		return "#000000"
	case "font-family":
		return "Ubuntu, Helvetica, Arial, sans-serif"
	case "font-size":
		return "13px"
	case "font-weight":
		return "normal"
	case "line-height":
		return "22px"
	case "padding":
		return "15px 10px"
	case "target":
		return "_blank"
	case "text-decoration":
		return "none"
	case "text-transform":
		return "uppercase"
	default:
		return ""
	}
}
