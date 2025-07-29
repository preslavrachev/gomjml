// Package html provides utilities for generating email-compatible HTML from MJML components.
// It includes HTML tag building with inline styles, email client compatibility features,
// and MSO/Outlook-specific rendering support.
package html

import (
	"strings"
)

// HTMLTag represents an HTML element with its name, attributes, CSS classes, and inline styles.
// It provides a fluent API for building HTML tags with proper email client compatibility.
//
// HTMLTag maintains the order of both attributes and style properties as they are added,
// which is important for consistency with MRML reference implementation.
type HTMLTag struct {
	name       string
	attributes []AttributeProperty
	classes    []string
	styles     []StyleProperty
}

// AttributeProperty represents a single HTML attribute with its name and value.
// Attributes are stored in order to maintain consistency with MRML output.
type AttributeProperty struct {
	Name  string // HTML attribute name (e.g., "border")
	Value string // HTML attribute value (e.g., "0")
}

// StyleProperty represents a single CSS property with its name and value.
// Properties are stored in order to maintain CSS specificity and ensure
// consistent rendering across email clients.
type StyleProperty struct {
	Name  string // CSS property name (e.g., "background-color")
	Value string // CSS property value (e.g., "#f0f0f0")
}

// NewHTMLTag creates a new HTMLTag with the specified element name.
// The tag is initialized with empty attributes, classes, and styles.
//
// Example:
//
//	tag := html.NewHTMLTag("div")
//	tag.AddStyle("margin", "0px auto").AddAttribute("class", "wrapper")
func NewHTMLTag(name string) *HTMLTag {
	return &HTMLTag{
		name:       name,
		attributes: make([]AttributeProperty, 0),
		classes:    make([]string, 0),
		styles:     make([]StyleProperty, 0),
	}
}

// AddStyle adds a CSS style property to the HTML tag.
// Styles are added in order and will appear in the rendered style attribute
// in the same sequence they were added.
//
// Returns the HTMLTag to enable method chaining.
//
// Example:
//
//	tag.AddStyle("background-color", "#f0f0f0").AddStyle("margin", "0px auto")
func (t *HTMLTag) AddStyle(name, value string) *HTMLTag {
	t.styles = append(t.styles, StyleProperty{name, value})
	return t
}

// MaybeAddStyle conditionally adds a CSS style property to the HTML tag.
// The style is only added if the value pointer is not nil and the value is not empty.
// This is useful for applying styles based on optional MJML attributes.
//
// Returns the HTMLTag to enable method chaining.
//
// Example:
//
//	var bgcolor *string = getBackgroundColor() // might be nil
//	tag.MaybeAddStyle("background-color", bgcolor)
func (t *HTMLTag) MaybeAddStyle(name string, value *string) *HTMLTag {
	if value != nil && *value != "" {
		t.AddStyle(name, *value)
	}
	return t
}

// AddAttribute adds an HTML attribute to the tag.
// If an attribute with the same name already exists, it will be overwritten.
//
// Returns the HTMLTag to enable method chaining.
//
// Example:
//
//	tag.AddAttribute("id", "main-content").AddAttribute("role", "presentation")
func (t *HTMLTag) AddAttribute(name, value string) *HTMLTag {
	// Check if attribute already exists and update it
	for i, attr := range t.attributes {
		if attr.Name == name {
			t.attributes[i].Value = value
			return t
		}
	}
	// Add new attribute
	t.attributes = append(t.attributes, AttributeProperty{name, value})
	return t
}

// MaybeAddAttribute conditionally adds an HTML attribute to the tag.
// The attribute is only added if the value pointer is not nil and the value is not empty.
// This is commonly used for email client compatibility attributes like bgcolor.
//
// Returns the HTMLTag to enable method chaining.
//
// Example:
//
//	var bgcolor *string = getBackgroundColor() // might be nil
//	tag.MaybeAddAttribute("bgcolor", bgcolor) // for Outlook compatibility
func (t *HTMLTag) MaybeAddAttribute(name string, value *string) *HTMLTag {
	if value != nil && *value != "" {
		t.AddAttribute(name, *value)
	}
	return t
}

// NewTableTag creates a new table tag with MRML-compatible attribute ordering
// Attributes are added in the order: border, cellpadding, cellspacing, role, align, width
func NewTableTag() *HTMLTag {
	return NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation")
}

// AddClass adds a CSS class name to the tag.
// Multiple classes can be added and will be joined with spaces in the class attribute.
//
// Returns the HTMLTag to enable method chaining.
//
// Example:
//
//	tag.AddClass("wrapper").AddClass("section")
func (t *HTMLTag) AddClass(class string) *HTMLTag {
	t.classes = append(t.classes, class)
	return t
}

// RenderOpen renders the opening HTML tag with all attributes, classes, and styles.
// Styles are rendered as an inline style attribute with properties in the order they were added.
//
// Example output:
//
//	<div class="wrapper" style="background-color:#f0f0f0;margin:0px auto;" bgcolor="#f0f0f0">
func (t *HTMLTag) RenderOpen() string {
	var b strings.Builder
	b.WriteString("<")
	b.WriteString(t.name)

	// Add HTML attributes in order
	for _, attr := range t.attributes {
		b.WriteByte(' ')
		b.WriteString(attr.Name)
		b.WriteString(`="`)
		b.WriteString(attr.Value)
		b.WriteByte('"')
	}

	// Add CSS classes
	if len(t.classes) > 0 {
		b.WriteString(` class="`)
		b.WriteString(strings.Join(t.classes, " "))
		b.WriteByte('"')
	}

	// Add inline styles
	if len(t.styles) > 0 {
		b.WriteString(` style="`)
		for _, style := range t.styles {
			b.WriteString(style.Name)
			b.WriteByte(':')
			b.WriteString(style.Value)
			b.WriteByte(';')
		}
		b.WriteByte('"')
	}

	b.WriteByte('>')
	return b.String()
}

// RenderClose renders the closing HTML tag.
//
// Example output:
//
//	</div>
func (t *HTMLTag) RenderClose() string {
	var b strings.Builder
	b.Grow(len(t.name) + 3) // Pre-allocate for "</name>"
	b.WriteString("</")
	b.WriteString(t.name)
	b.WriteByte('>')
	return b.String()
}

// RenderSelfClosing renders a self-closing HTML tag with all attributes, classes, and styles.
// This is used for void elements like <img>, <br>, <hr>, etc.
//
// Example output:
//
//	<img src="image.jpg" style="width:100%;" />
func (t *HTMLTag) RenderSelfClosing() string {
	var b strings.Builder
	b.WriteString("<")
	b.WriteString(t.name)

	// Add HTML attributes in order
	for _, attr := range t.attributes {
		b.WriteByte(' ')
		b.WriteString(attr.Name)
		b.WriteString(`="`)
		b.WriteString(attr.Value)
		b.WriteByte('"')
	}

	// Add CSS classes
	if len(t.classes) > 0 {
		b.WriteString(` class="`)
		b.WriteString(strings.Join(t.classes, " "))
		b.WriteByte('"')
	}

	// Add inline styles
	if len(t.styles) > 0 {
		b.WriteString(` style="`)
		for _, style := range t.styles {
			b.WriteString(style.Name)
			b.WriteByte(':')
			b.WriteString(style.Value)
			b.WriteByte(';')
		}
		b.WriteByte('"')
	}

	b.WriteString(" />")
	return b.String()
}
