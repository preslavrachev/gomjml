package components

import (
	"io"

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

func (c *MJTableComponent) RenderString() (string, error) {
	// TODO: Implement mj-table component functionality
	return "", &NotImplementedError{ComponentName: "mj-table"}
}

func (c *MJTableComponent) Render(w io.Writer) error {
	// TODO: Implement mj-table component functionality
	return &NotImplementedError{ComponentName: "mj-table"}
}

func (c *MJTableComponent) GetTagName() string {
	return "mj-table"
}

func (c *MJTableComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "align":
		return "left"
	case "border":
		return "none"
	case "cellpadding":
		return "0"
	case "cellspacing":
		return "0"
	case "color":
		return "#000000"
	case "font-family":
		return "Ubuntu, Helvetica, Arial, sans-serif"
	case "font-size":
		return "13px"
	case "line-height":
		return "22px"
	case "padding":
		return "10px 25px"
	case "table-layout":
		return "auto"
	case "width":
		return "100%"
	default:
		return ""
	}
}
