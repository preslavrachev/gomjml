package components

import (
	"fmt"
	"io"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJHeadComponent represents mj-head
type MJHeadComponent struct {
	*BaseComponent
}

// NewMJHeadComponent creates a new mj-head component
func NewMJHeadComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJHeadComponent {
	return &MJHeadComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJHeadComponent) Render(w io.StringWriter) error {
	return nil // Head is handled in MJML component
}

func (c *MJHeadComponent) GetTagName() string {
	return "mj-head"
}

func (c *MJHeadComponent) GetDefaultAttribute(name string) string {
	return ""
}

// MJTitleComponent represents mj-title
type MJTitleComponent struct {
	*BaseComponent
}

// NewMJTitleComponent creates a new mj-title component
func NewMJTitleComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJTitleComponent {
	return &MJTitleComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJTitleComponent) Render(w io.StringWriter) error {
	return nil // Title is handled in MJML component head processing
}

func (c *MJTitleComponent) GetTagName() string {
	return "mj-title"
}

func (c *MJTitleComponent) GetDefaultAttribute(name string) string {
	return ""
}

// MJFontComponent represents mj-font
type MJFontComponent struct {
	*BaseComponent
}

// NewMJFontComponent creates a new mj-font component
func NewMJFontComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJFontComponent {
	return &MJFontComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJFontComponent) Render(w io.StringWriter) error {
	return nil // Font is handled in MJML component head processing
}

func (c *MJFontComponent) GetTagName() string {
	return "mj-font"
}

func (c *MJFontComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "name":
		return ""
	case "href":
		return ""
	default:
		return ""
	}
}

// MJPreviewComponent represents mj-preview
type MJPreviewComponent struct {
	*BaseComponent
}

// NewMJPreviewComponent creates a new mj-preview component
func NewMJPreviewComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJPreviewComponent {
	return &MJPreviewComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJPreviewComponent) Render(w io.StringWriter) error {
	// Preview text is rendered as hidden div in body
	if c.Node.Text != "" {
		previewHTML := fmt.Sprintf(
			`<div style="display:none;font-size:1px;color:#ffffff;line-height:1px;max-height:0px;max-width:0px;opacity:0;overflow:hidden;">%s</div>`,
			c.Node.Text,
		)
		_, err := w.WriteString(previewHTML)
		return err
	}
	return nil
}

func (c *MJPreviewComponent) GetTagName() string {
	return "mj-preview"
}

func (c *MJPreviewComponent) GetDefaultAttribute(name string) string {
	return ""
}

// MJStyleComponent represents mj-style
type MJStyleComponent struct {
	*BaseComponent
}

// NewMJStyleComponent creates a new mj-style component
func NewMJStyleComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJStyleComponent {
	return &MJStyleComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJStyleComponent) Render(w io.StringWriter) error {
	// Custom CSS styles - render as style tag
	if c.Node.Text != "" {
		styleHTML := fmt.Sprintf(`<style type="text/css">%s</style>`, c.Node.Text)
		_, err := w.WriteString(styleHTML)
		return err
	}
	return nil
}

func (c *MJStyleComponent) GetTagName() string {
	return "mj-style"
}

func (c *MJStyleComponent) GetDefaultAttribute(name string) string {
	return ""
}

// MJAttributesComponent represents mj-attributes
type MJAttributesComponent struct {
	*BaseComponent
}

// NewMJAttributesComponent creates a new mj-attributes component
func NewMJAttributesComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJAttributesComponent {
	return &MJAttributesComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJAttributesComponent) Render(w io.StringWriter) error {
	return nil // Attributes are processed during parsing, no HTML output
}

func (c *MJAttributesComponent) GetTagName() string {
	return "mj-attributes"
}

func (c *MJAttributesComponent) GetDefaultAttribute(name string) string {
	return ""
}

// MJAllComponent represents mj-all (global attributes)
type MJAllComponent struct {
	*BaseComponent
}

// NewMJAllComponent creates a new mj-all component
func NewMJAllComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJAllComponent {
	return &MJAllComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJAllComponent) Render(w io.StringWriter) error {
	return nil // Global attributes are processed during parsing, no HTML output
}

func (c *MJAllComponent) GetTagName() string {
	return "mj-all"
}

func (c *MJAllComponent) GetDefaultAttribute(name string) string {
	return ""
}
