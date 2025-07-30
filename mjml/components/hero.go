package components

import (
	"io"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJHeroComponent represents the mj-hero component
type MJHeroComponent struct {
	*BaseComponent
}

func NewMJHeroComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJHeroComponent {
	return &MJHeroComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJHeroComponent) Render(w io.Writer) error {
	// TODO: Implement mj-hero component functionality
	return &NotImplementedError{ComponentName: "mj-hero"}
}

func (c *MJHeroComponent) GetTagName() string {
	return "mj-hero"
}

func (c *MJHeroComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "background-color":
		return "#ffffff"
	case "background-position":
		return "center center"
	case "height":
		return "0px"
	case "mode":
		return "fixed-height"
	case "padding":
		return "0px"
	case "vertical-align":
		return "top"
	default:
		return ""
	}
}
