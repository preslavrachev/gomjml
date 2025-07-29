package components

import (
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJSpacerComponent represents the mj-spacer component
type MJSpacerComponent struct {
	*BaseComponent
}

func NewMJSpacerComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJSpacerComponent {
	return &MJSpacerComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJSpacerComponent) Render() (string, error) {
	// TODO: Implement mj-spacer component functionality
	return "", &NotImplementedError{ComponentName: "mj-spacer"}
}

func (c *MJSpacerComponent) GetTagName() string {
	return "mj-spacer"
}

func (c *MJSpacerComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "height":
		return "20px"
	default:
		return ""
	}
}
