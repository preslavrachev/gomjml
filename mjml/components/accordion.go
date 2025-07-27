package components

import (
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJAccordionComponent represents the mj-accordion component
type MJAccordionComponent struct {
	*BaseComponent
}

func NewMJAccordionComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJAccordionComponent {
	return &MJAccordionComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJAccordionComponent) Render() (string, error) {
	// TODO: Implement mj-accordion component functionality
	return "", &NotImplementedError{ComponentName: "mj-accordion"}
}

func (c *MJAccordionComponent) GetTagName() string {
	return "mj-accordion"
}

// MJAccordionTextComponent represents the mj-accordion-text component
type MJAccordionTextComponent struct {
	*BaseComponent
}

func NewMJAccordionTextComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJAccordionTextComponent {
	return &MJAccordionTextComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJAccordionTextComponent) Render() (string, error) {
	// TODO: Implement mj-accordion-text component functionality
	return "", &NotImplementedError{ComponentName: "mj-accordion-text"}
}

func (c *MJAccordionTextComponent) GetTagName() string {
	return "mj-accordion-text"
}

// MJAccordionTitleComponent represents the mj-accordion-title component
type MJAccordionTitleComponent struct {
	*BaseComponent
}

func NewMJAccordionTitleComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJAccordionTitleComponent {
	return &MJAccordionTitleComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJAccordionTitleComponent) Render() (string, error) {
	// TODO: Implement mj-accordion-title component functionality
	return "", &NotImplementedError{ComponentName: "mj-accordion-title"}
}

func (c *MJAccordionTitleComponent) GetTagName() string {
	return "mj-accordion-title"
}
