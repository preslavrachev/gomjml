package components

import (
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJCarouselComponent represents the mj-carousel component
type MJCarouselComponent struct {
	*BaseComponent
}

func NewMJCarouselComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJCarouselComponent {
	return &MJCarouselComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJCarouselComponent) Render() (string, error) {
	// TODO: Implement mj-carousel component functionality
	return "", &NotImplementedError{ComponentName: "mj-carousel"}
}

func (c *MJCarouselComponent) GetTagName() string {
	return "mj-carousel"
}

// MJCarouselImageComponent represents the mj-carousel-image component
type MJCarouselImageComponent struct {
	*BaseComponent
}

func NewMJCarouselImageComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJCarouselImageComponent {
	return &MJCarouselImageComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJCarouselImageComponent) Render() (string, error) {
	// TODO: Implement mj-carousel-image component functionality
	return "", &NotImplementedError{ComponentName: "mj-carousel-image"}
}

func (c *MJCarouselImageComponent) GetTagName() string {
	return "mj-carousel-image"
}
