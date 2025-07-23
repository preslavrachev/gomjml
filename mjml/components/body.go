package components

import (
	"strings"

	"github.com/preslavrachev/gomjml/parser"
)

// MJBodyComponent represents mj-body
type MJBodyComponent struct {
	*BaseComponent
}

// NewMJBodyComponent creates a new mj-body component
func NewMJBodyComponent(node *parser.MJMLNode) *MJBodyComponent {
	return &MJBodyComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJBodyComponent) Render() (string, error) {
	var html strings.Builder

	html.WriteString(`<div>`)

	for _, child := range c.Children {
		childHTML, err := child.Render()
		if err != nil {
			return "", err
		}
		html.WriteString(childHTML)
	}

	html.WriteString(`</div>`)
	return html.String(), nil
}

func (c *MJBodyComponent) GetTagName() string {
	return "mj-body"
}

func (c *MJBodyComponent) GetDefaultAttribute(name string) string {
	return ""
}
