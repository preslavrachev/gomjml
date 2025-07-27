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
