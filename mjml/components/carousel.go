package components

import (
	"io"

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

func (c *MJCarouselComponent) Render(w io.Writer) error {
	// TODO: Implement mj-carousel component functionality
	return &NotImplementedError{ComponentName: "mj-carousel"}
}

func (c *MJCarouselComponent) GetTagName() string {
	return "mj-carousel"
}

func (c *MJCarouselComponent) GetDefaultAttribute(name string) string {
	// TODO: Consider more performant approaches to attribute matching than switch statements,
	// such as static map[string]string lookups or compile-time generated code for components
	// with many default attributes (10+ attributes). Switch statements may have O(n) lookup
	// time while map lookups are O(1) average case.
	switch name {
	case "align":
		return "center"
	case "border-radius":
		return "6px"
	case "icon-width":
		return "44px"
	case "left-icon":
		return "https://i.imgur.com/xTh3hln.png"
	case "right-icon":
		return "https://i.imgur.com/os7o9kz.png"
	case "tb-border":
		return "2px solid transparent"
	case "tb-border-radius":
		return "6px"
	case "tb-hover-border-color":
		return "#fead0d"
	case "tb-selected-border-color":
		return "#cccccc"
	case "thumbnails":
		return "visible"
	default:
		return ""
	}
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

func (c *MJCarouselImageComponent) Render(w io.Writer) error {
	// TODO: Implement mj-carousel-image component functionality
	return &NotImplementedError{ComponentName: "mj-carousel-image"}
}

func (c *MJCarouselImageComponent) GetTagName() string {
	return "mj-carousel-image"
}

func (c *MJCarouselImageComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "target":
		return "_blank"
	default:
		return ""
	}
}
