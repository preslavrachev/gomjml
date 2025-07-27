package components

import (
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

func (c *MJHeroComponent) Render() (string, error) {
	// TODO: Implement mj-hero component functionality
	return "", &NotImplementedError{ComponentName: "mj-hero"}
}

func (c *MJHeroComponent) GetTagName() string {
	return "mj-hero"
}
