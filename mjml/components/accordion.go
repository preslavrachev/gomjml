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

func (c *MJAccordionComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "border":
		return "2px solid black"
	case "font-family":
		return "Ubuntu, Helvetica, Arial, sans-serif"
	case "icon-align":
		return "middle"
	case "icon-height":
		return "32px"
	case "icon-position":
		return "right"
	case "icon-unwrapped-alt":
		return "-"
	case "icon-unwrapped-url":
		return "https://i.imgur.com/w4uTygT.png"
	case "icon-width":
		return "32px"
	case "icon-wrapped-alt":
		return "+"
	case "icon-wrapped-url":
		return "https://i.imgur.com/bIXv1bk.png"
	case "padding":
		return "10px 25px"
	default:
		return ""
	}
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

func (c *MJAccordionTextComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "font-size":
		return "13px"
	case "line-height":
		return "1"
	case "padding":
		return "16px"
	default:
		return ""
	}
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

func (c *MJAccordionTitleComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "font-size":
		return "13px"
	case "padding":
		return "16px"
	default:
		return ""
	}
}
