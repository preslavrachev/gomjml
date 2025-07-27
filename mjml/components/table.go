package components

import (
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJTableComponent represents the mj-table component
type MJTableComponent struct {
	*BaseComponent
}

func NewMJTableComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJTableComponent {
	return &MJTableComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJTableComponent) Render() (string, error) {
	// TODO: Implement mj-table component functionality
	return "", &NotImplementedError{ComponentName: "mj-table"}
}

func (c *MJTableComponent) GetTagName() string {
	return "mj-table"
}
